package mask

import (
	"labix.org/v2/mgo/bson"
	"time"
)

type Comment struct {
	ID      bson.ObjectId          `json:"id" bson:"_id"`
	UserID  bson.ObjectId          `json:"-" bson:"user_id"`
	User    map[string]interface{} `json:"user" bson:"-"`
	PostID  bson.ObjectId          `json:"post_id" bson:"post_id"`
	Created float64                `json:"created" bson:"created"`
	Message string                 `json:"message" bson:"message"`
}

// NewComment returns a new instance of Comment
func NewComment(user, post bson.ObjectId) *Comment {
	c := new(Comment)
	c.UserID = user
	c.PostID = post
	c.Created = float64(time.Now().Unix())

	return c
}

// Save inserts the Comment instance if it hasn't been created yet or updates it if it has
func (c *Comment) Save(conn *Connection) error {
	if c.ID.Hex() == "" {
		c.ID = bson.NewObjectId()
	}

	if err := conn.Save("comments", c.ID, c); err != nil {
		return err
	}

	return nil
}

// CreateComment adds a comment to a post
func CreateComment(c Context) {
	var post Post

	if c.User == nil {
		c.Error(400, CodeInvalidData, MsgInvalidData)
		return
	}

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
	if strlen(message) < 1 || strlen(message) > 500 {
		c.Error(400, CodeInvalidCommentText, MsgInvalidCommentText)
		return
	}

	comment := NewComment(c.User.ID, post.ID)
	comment.Message = message

	if err := comment.Save(c.Conn); err != nil {
		c.Error(500, CodeUnexpected, MsgUnexpected)
		return
	}

	post.CommentsNum++
	(&post).Save(c.Conn)

	c.Success(201, map[string]interface{}{
		"created": true,
		"message": "Comment posted successfully",
	})
}

// DeleteComment removes a comment from a post
func RemoveComment(c Context) {
	var (
		post    Post
		comment Comment
	)

	if c.User == nil {
		c.Error(400, CodeInvalidData, MsgInvalidData)
		return
	}

	commentID := c.Form("comment_id")
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

	post.CommentsNum--
	(&post).Save(c.Conn)

	c.Success(200, map[string]interface{}{
		"deleted": true,
		"message": "Comment deleted successfully",
	})
}

// CommentsForPost returns a list with the comments for a post
func CommentsForPost(c Context) {
	var (
		post   Post
		result Comment
	)
	count, offset := c.ListCountParams()

	if c.User == nil {
		c.Error(400, CodeInvalidData, MsgInvalidData)
		return
	}

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

	comments := make([]Comment, 0, count)
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

	usersData := GetUsersData(users, c.User, c.Conn)
	if usersData == nil {
		c.Error(500, CodeUnexpected, MsgUnexpected)
		return
	}

	commentsResult := make([]Comment, 0, len(comments))
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
