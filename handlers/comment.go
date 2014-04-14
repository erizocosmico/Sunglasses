package handlers

import (
	"github.com/go-martini/martini"
	. "github.com/mvader/mask/error"
	"github.com/mvader/mask/middleware"
	"github.com/mvader/mask/models"
	"github.com/mvader/mask/modules/timeline"
	"github.com/mvader/mask/util"
	"labix.org/v2/mgo/bson"
)

// CreateComment adds a comment to a post
func CreateComment(c middleware.Context) {
	var post models.Post

	postID := c.Form("post_id")
	if !bson.IsObjectIdHex(postID) {
		c.Error(400, CodeInvalidData, MsgInvalidData)
		return
	}

	if err := c.FindId("posts", bson.ObjectIdHex(postID)).One(&post); err != nil {
		c.Error(404, CodeNotFound, MsgNotFound)
		return
	}

	if !(&post).CanBeAccessedBy(c.User, c.Conn) {
		c.Error(403, CodeUnauthorized, MsgUnauthorized)
		return
	}

	message := c.Form("comment_text")
	if util.Strlen(message) < 1 || util.Strlen(message) > 500 {
		c.Error(400, CodeInvalidCommentText, MsgInvalidCommentText)
		return
	}

	comment := models.NewComment(c.User.ID, post.ID)
	comment.Message = message

	if err := comment.Save(c.Conn); err != nil {
		c.Error(500, CodeUnexpected, MsgUnexpected)
		return
	}

	post.CommentsNum++
	(&post).Save(c.Conn)

	go timeline.PropagatePostOnNewComment(c, post.ID, comment.ID)

	c.Success(201, map[string]interface{}{
		"created": true,
		"message": "Comment posted successfully",
	})
}

// DeleteComment removes a comment from a post
func RemoveComment(c middleware.Context, params martini.Params) {
	var (
		post    models.Post
		comment models.Comment
	)

	commentID := params["comment_id"]
	if !bson.IsObjectIdHex(commentID) {
		c.Error(400, CodeInvalidData, MsgInvalidData)
		return
	}

	if err := c.FindId("comments", bson.ObjectIdHex(commentID)).One(&comment); err != nil {
		c.Error(404, CodeNotFound, MsgNotFound)
		return
	}

	if c.User.ID.Hex() != comment.UserID.Hex() {
		c.Error(403, CodeUnauthorized, MsgUnauthorized)
		return
	}

	if err := c.FindId("posts", comment.PostID).One(&post); err != nil {
		c.Error(404, CodeNotFound, MsgNotFound)
		return
	}

	if err := c.Query("comments").RemoveId(comment.ID); err != nil {
		c.Error(500, CodeUnexpected, MsgUnexpected)
		return
	}

	if c.GetBoolean("confirmed") {
		post.CommentsNum--
		(&post).Save(c.Conn)

		go timeline.PropagatePostOnCommentDeleted(c, post.ID, comment.ID)

		c.Success(200, map[string]interface{}{
			"deleted": true,
			"message": "Comment deleted successfully",
		})
	} else {
		c.Success(200, map[string]interface{}{
			"deleted": false,
			"message": "Comment was not deleted",
		})
	}
}

// CommentsForPost returns a list with the comments for a post
func CommentsForPost(c middleware.Context, params martini.Params) {
	var (
		post   models.Post
		result models.Comment
	)
	count, offset := c.ListCountParams()

	postID := params["post_id"]
	if !bson.IsObjectIdHex(postID) {
		c.Error(400, CodeInvalidData, MsgInvalidData)
		return
	}

	if err := c.FindId("posts", bson.ObjectIdHex(postID)).One(&post); err != nil {
		c.Error(404, CodeNotFound, MsgNotFound)
		return
	}

	if !(&post).CanBeAccessedBy(c.User, c.Conn) {
		c.Error(403, CodeUnauthorized, MsgUnauthorized)
		return
	}

	comments := make([]models.Comment, 0, count)
	cursor := c.Find("comments", bson.M{"post_id": post.ID}).Sort("-created").Limit(count).Skip(offset).Iter()
	for cursor.Next(&result) {
		comments = append(comments, result)
	}

	if err := cursor.Close(); err != nil {
		c.Error(500, CodeUnexpected, MsgUnexpected)
		return
	}

	users := make([]bson.ObjectId, 0, len(comments))
	for _, c := range comments {
		if c.UserID.Hex() != "" {
			users = append(users, c.UserID)
		}
	}

	usersData := models.GetUsersData(users, c.User, c.Conn)
	if usersData == nil {
		c.Error(500, CodeUnexpected, MsgUnexpected)
		return
	}

	commentsResult := make([]models.Comment, 0, len(comments))
	for i, _ := range comments {
		if v, ok := usersData[comments[i].UserID]; ok {
			comments[i].User = v
			commentsResult = append(commentsResult, comments[i])
		}
	}

	c.Success(200, map[string]interface{}{
		"comments": commentsResult,
		"count":    len(commentsResult),
	})
}
