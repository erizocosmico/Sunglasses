package handlers

import (
	"errors"
	. "github.com/mvader/mask/error"
	"github.com/mvader/mask/middleware"
	"github.com/mvader/mask/models"
	"github.com/mvader/mask/modules/timeline"
	"github.com/mvader/mask/modules/upload"
	"github.com/mvader/mask/modules/video"
	"github.com/mvader/mask/util"
	"labix.org/v2/mgo/bson"
	"os"
	"strconv"
	"strings"
)

// CreatePost creates a new post
func CreatePost(c middleware.Context) {
	postType := c.Form("post_type")

	if c.User == nil {
		c.Error(400, CodeInvalidData, MsgInvalidData)
		return
	}

	switch postType {
	case "photo":
		postPhoto(c)
		break
	case "video":
		postVideo(c)
		break
	case "link":
		postLink(c)
		break
	default:
		// Default post type is status
		postStatus(c)
	}
}

// ShowPost returns all data about a post including comments and likes
func ShowPost(c middleware.Context) {
	var post models.Post

	// TODO post is liked?

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

	var like models.PostLike
	err := c.Find("likes", bson.M{"post_id": post.ID, "user_id": c.User.ID}).One(&like)
	post.Liked = err == nil

	if !post.CanBeAccessedBy(c.User, c.Conn) {
		c.Error(403, CodeUnauthorized, MsgUnauthorized)
		return
	}

	uids := make([]bson.ObjectId, 0, 11)

	count, _ := c.Count("comments", bson.M{"post_id": post.ID})
	if count > 0 {
		var comments []models.Comment
		iter := c.Find("comments", bson.M{"post_id": post.ID}).Limit(10).Sort("-created").Iter()
		err := iter.All(&comments)
		if err == nil {
			post.Comments = comments
		}

		for _, v := range comments {
			uids = append(uids, v.UserID)
		}
	}

	uids = append(uids, post.UserID)

	data := models.GetUsersData(uids, c.User, c.Conn)
	if len(data) == 0 {
		c.Error(500, CodeUnexpected, MsgUnexpected)
		return
	}

	for i, _ := range post.Comments {
		post.Comments[i].User = data[post.Comments[i].UserID]
	}

	post.User = data[post.UserID]

	c.Success(200, map[string]interface{}{
		"post": post,
	})
}

// DeletePost deletes a post owned by the user making the request
func DeletePost(c middleware.Context) {
	var post models.Post

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

	if c.User.ID.Hex() != post.UserID.Hex() {
		c.Error(403, CodeUnauthorized, MsgUnauthorized)
		return
	}

	if err := c.Query("posts").RemoveId(post.ID); err != nil {
		c.Error(500, CodeUnexpected, MsgUnexpected)
		return
	}

	if post.Type == models.PostPhoto {
		os.Remove(upload.ToLocalImagePath(post.PhotoURL, c.Config))
		os.Remove(upload.ToLocalThumbnailPath(post.Thumbnail, c.Config))
	}

	c.RemoveAll("comments", bson.M{"post_id": post.ID})
	c.RemoveAll("likes", bson.M{"post_id": post.ID})
	c.RemoveAll("notifications", bson.M{"post_id": post.ID})

	go timeline.PropagatePostsOnDeletion(c, post.ID)

	c.Success(200, map[string]interface{}{
		"deleted": true,
		"message": "Post deleted successfully",
	})
}

// LikePost likes a post (or unlikes it if the post has already been liked)
func LikePost(c middleware.Context) {
	var post models.Post

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

	if !post.CanBeAccessedBy(c.User, c.Conn) {
		c.Error(403, CodeUnauthorized, MsgUnauthorized)
		return
	}

	count, _ := c.Find("likes", bson.M{"post_id": post.ID, "user_id": c.User.ID}).Count()

	// Post was already liked by the user, unlike it
	if count > 0 {
		post.Likes--
		if err := (&post).Save(c.Conn); err != nil {
			c.Error(500, CodeUnexpected, MsgUnexpected)
			return
		}

		if _, err := c.RemoveAll("likes", bson.M{"post_id": post.ID, "user_id": c.User.ID}); err != nil {
			post.Likes++
			(&post).Save(c.Conn)

			c.Error(500, CodeUnexpected, MsgUnexpected)
			return
		}

		go timeline.PropagatePostOnLike(c, post.ID, false)

		c.Success(200, map[string]interface{}{
			"liked":   false,
			"message": "Post unliked successfully",
		})
		return
	}

	// Like post
	post.Likes++
	if err := (&post).Save(c.Conn); err != nil {
		c.Error(500, CodeUnexpected, MsgUnexpected)
		return
	}

	like := models.PostLike{bson.NewObjectId(), c.User.ID, post.ID}
	if err := c.Query("likes").Insert(like); err != nil {
		post.Likes--
		(&post).Save(c.Conn)

		c.Error(500, CodeUnexpected, MsgUnexpected)
		return
	}

	go timeline.PropagatePostOnLike(c, post.ID, true)

	c.Success(200, map[string]interface{}{
		"liked":   true,
		"message": "Post liked successfully",
	})
}

func postPhoto(c middleware.Context) {
	file, err := upload.RetrieveUploadedImage(c.Request, "post_picture")
	if err != nil {
		code, msg := upload.CodeAndMessageForUploadError(err)
		c.Error(400, code, msg)
		return
	}

	imagePath, thumbnailPath, err := upload.StoreImage(file, upload.DefaultUploadOptions(c.Config))
	if err != nil {
		code, msg := upload.CodeAndMessageForUploadError(err)
		c.Error(400, code, msg)
		return
	}

	p := models.NewPost(models.PostPhoto, c.User)
	p.PhotoURL = imagePath
	p.Thumbnail = thumbnailPath
	p.Caption = strings.TrimSpace(c.Form("caption"))
	p.Text = strings.TrimSpace(c.Form("post_text"))
	privacy, err := getPostPrivacy(models.PostPhoto, c)
	if err != nil {
		c.Error(400, CodeInvalidUserList, MsgInvalidUserList)
		return
	}

	p.Privacy = privacy

	if util.Strlen(p.Text) > 1500 {
		c.Error(400, CodeInvalidStatusText, MsgInvalidStatusText)
		return
	}

	if util.Strlen(p.Caption) > 255 {
		c.Error(400, CodeInvalidCaption, MsgInvalidCaption)
		return
	}

	if err := p.Save(c.Conn); err != nil {
		c.Error(500, CodeUnexpected, MsgUnexpected)

		os.Remove(p.PhotoURL)
		os.Remove(p.Thumbnail)
		return
	}

	go timeline.PropagatePostOnCreation(c, p)

	c.Success(201, map[string]interface{}{
		"message": "Photo posted successfully",
	})
}

func postVideo(c middleware.Context) {
	statusText := strings.TrimSpace(c.Form("post_text"))

	if util.Strlen(statusText) > 1500 {
		c.Error(400, CodeInvalidStatusText, MsgInvalidStatusText)
		return
	}

	post := models.NewPost(models.PostVideo, c.User)
	post.Text = statusText
	privacy, err := getPostPrivacy(models.PostVideo, c)
	if err != nil {
		c.Error(400, CodeInvalidUserList, MsgInvalidUserList)
		return
	}

	post.Privacy = privacy

	valid, videoID, service, title := video.IsValidVideo(strings.TrimSpace(c.Form("video_url")))

	if !valid {
		c.Error(400, CodeInvalidVideoURL, MsgInvalidVideoURL)
		return
	}

	post.VideoID = videoID
	post.Service = service
	post.Title = title

	if err := post.Save(c.Conn); err != nil {
		c.Error(500, CodeUnexpected, MsgUnexpected)
		return
	}

	go timeline.PropagatePostOnCreation(c, post)

	c.Success(201, map[string]interface{}{
		"message": "Video posted successfully",
	})
}

func postLink(c middleware.Context) {
	statusText := strings.TrimSpace(c.Form("post_text"))

	if util.Strlen(statusText) > 1500 {
		c.Error(400, CodeInvalidStatusText, MsgInvalidStatusText)
		return
	}

	post := models.NewPost(models.PostVideo, c.User)
	post.Text = statusText
	privacy, err := getPostPrivacy(models.PostLink, c)
	if err != nil {
		c.Error(400, CodeInvalidUserList, MsgInvalidUserList)
		return
	}

	post.Privacy = privacy

	valid, link, title := util.IsValidLink(strings.TrimSpace(c.Form("link_url")))

	if !valid {
		c.Error(400, CodeInvalidLinkURL, MsgInvalidLinkURL)
		return
	}

	post.URL = link
	post.Title = title

	if err := post.Save(c.Conn); err != nil {
		c.Error(500, CodeUnexpected, MsgUnexpected)
		return
	}

	go timeline.PropagatePostOnCreation(c, post)

	c.Success(201, map[string]interface{}{
		"message": "Link posted successfully",
	})
}

func postStatus(c middleware.Context) {
	statusText := strings.TrimSpace(c.Form("post_text"))

	if util.Strlen(statusText) < 1 || util.Strlen(statusText) > 1500 {
		c.Error(400, CodeInvalidStatusText, MsgInvalidStatusText)
		return
	}

	post := models.NewPost(models.PostStatus, c.User)
	post.Text = statusText
	privacy, err := getPostPrivacy(models.PostStatus, c)
	if err != nil {
		c.Error(400, CodeInvalidUserList, MsgInvalidUserList)
		return
	}

	post.Privacy = privacy
	if err := post.Save(c.Conn); err != nil {
		c.Error(500, CodeUnexpected, MsgUnexpected)
		return
	}

	go timeline.PropagatePostOnCreation(c, post)

	c.Success(201, map[string]interface{}{
		"message": "Status posted successfully",
	})
}

func getPostPrivacy(postType models.ObjectType, c middleware.Context) (models.PrivacySettings, error) {
	p := models.PrivacySettings{}
	var pType int64
	var err error

	if pType, err = strconv.ParseInt(c.Form("privacy_type"), 10, 8); err != nil {
		pType = 0
	}

	privacyType := models.PrivacyType(pType)
	defaultSettings := c.User.Settings.GetPrivacySettings(postType)
	if privacyType == 0 {
		p.Type = defaultSettings.Type
	} else {
		if models.IsValidPrivacyType(privacyType) {
			p.Type = privacyType
		} else {
			p.Type = defaultSettings.Type
		}
	}

	if p.Type == models.PrivacyAllBut || p.Type == models.PrivacyNoneBut {
		if privacyType == 0 {
			p.Users = defaultSettings.Users
		} else {
			us, ok := c.Request.PostForm["privacy_users"]
			if ok && len(us) > 0 {
				p.Users = make([]bson.ObjectId, 0, len(us))
				for _, u := range us {
					if bson.IsObjectIdHex(u) {
						p.Users = append(p.Users, bson.ObjectIdHex(u))
					}
				}

				count, err := c.Count("follows", bson.M{"user_from": c.User.ID, "user_to": bson.M{"$in": p.Users}})
				if err != nil || count != len(p.Users) {
					count2, err := c.Count("follows", bson.M{"user_to": c.User.ID, "user_from": bson.M{"$in": p.Users}})
					if err != nil || count+count2 != len(p.Users) {
						return p, errors.New("invalid user list provided")
					}
				}
			}
		}
	}

	return p, nil
}
