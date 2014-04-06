package mask

import (
	"errors"
	"labix.org/v2/mgo/bson"
	"os"
	"strconv"
	"strings"
	"time"
)

type ObjectType int

const (
	// Post types
	PostStatus = 1
	PostPhoto  = 2
	PostVideo  = 3
	PostLink   = 4
	Album      = 5
)

// Post model
type Post struct {
	ID          bson.ObjectId          `json:"id" bson:"_id"`
	User        map[string]interface{} `json:"user" bson:"-"`
	UserID      bson.ObjectId          `json:"-" bson:"user_id"`
	Created     float64                `json:"created" bson:"created"`
	Type        ObjectType             `json:"post_type" bson:"post_type"`
	Likes       float64                `json:"likes" bson:"likes"`
	CommentsNum float64                `json:"comments_num" bson:"comments_num"`
	Comments    []Comment              `json:"comments" bson:"-"`
	Reported    float64                `json:"reported" bson:"reported"`
	Privacy     PrivacySettings        `json:"privacy" bson:"privacy"`
	Text        string                 `json:"text,omitempty" bson:"text,omitempty"`
	Liked       bool                   `json:"liked,omitempty" bson:"-"`

	// Video specific fields
	Service VideoService `json:"video_service,omitempty" bson:"video_service,omitempty"`
	VideoID string       `json:"video_id,omitempty" bson:"video_id,omitempty"`
	// Also used in link
	Title string `json:"title,omitempty" bson:"title,omitempty"`

	// Photo specific fields
	PhotoURL  string        `json:"photo_url,omitempty" bson:"photo_url,omitempty"`
	Caption   string        `json:"caption,omitempty" bson:"caption,omitempty"`
	AlbumID   bson.ObjectId `json:"album_id,omitempty" bson:"album_id,omitempty"`
	Thumbnail string        `json:"thumbnail,omitempty" bson:"thumbnail,omitempty"`

	// Link specific fields
	URL string `json:"link_url,omitempty" bson:"link_url,omitempty"`
}

// PostLike model
type PostLike struct {
	ID     bson.ObjectId `json:"id" bson:"_id"`
	UserID bson.ObjectId `json:"user_id" bson:"user_id"`
	PostID bson.ObjectId `json:"post_id" bson:"post_id"`
}

// NewPost returns a new post instance
func NewPost(t ObjectType, user *User) *Post {
	p := new(Post)
	p.Type = t
	p.Created = float64(time.Now().Unix())
	p.UserID = user.ID

	return p
}

// Save inserts the Post instance if it hasn't been created yet or updates it if it has
func (p *Post) Save(conn *Connection) error {
	if p.ID.Hex() == "" {
		p.ID = bson.NewObjectId()
	}

	if err := conn.Save("posts", p.ID, p); err != nil {
		return err
	}

	return nil
}

// CreatePost creates a new post
func CreatePost(c Context) {
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
func ShowPost(c Context) {
	var post Post

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

	var like PostLike
	err := c.Find("likes", bson.M{"post_id": post.ID, "user_id": c.User.ID}).One(&like)
	post.Liked = err == nil

	if !post.CanBeAccessedBy(c.User, c.Conn) {
		c.Error(403, CodeUnauthorized, MsgUnauthorized)
		return
	}

	uids := make([]bson.ObjectId, 0, 11)

	count, _ := c.Count("comments", bson.M{"post_id": post.ID})
	if count > 0 {
		var comments []Comment
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

	data := GetUsersData(uids, c.User, c.Conn)
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
func DeletePost(c Context) {
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

	if c.User.ID.Hex() != post.UserID.Hex() {
		c.Error(403, CodeUnauthorized, MsgUnauthorized)
		return
	}

	if err := c.Query("posts").RemoveId(post.ID); err != nil {
		c.Error(500, CodeUnexpected, MsgUnexpected)
		return
	}

	c.RemoveAll("comments", bson.M{"post_id": post.ID})
	c.RemoveAll("likes", bson.M{"post_id": post.ID})
	c.RemoveAll("notifications", bson.M{"post_id": post.ID})

	go PropagatePostsOnDeletion(c, post.ID)

	c.Success(200, map[string]interface{}{
		"deleted": true,
		"message": "Post deleted successfully",
	})
}

// LikePost likes a post (or unlikes it if the post has already been liked)
func LikePost(c Context) {
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

		go PropagatePostOnLike(c, post.ID, false)

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

	like := PostLike{bson.NewObjectId(), c.User.ID, post.ID}
	if err := c.Query("likes").Insert(like); err != nil {
		post.Likes--
		(&post).Save(c.Conn)

		c.Error(500, CodeUnexpected, MsgUnexpected)
		return
	}

	go PropagatePostOnLike(c, post.ID, true)

	c.Success(200, map[string]interface{}{
		"liked":   true,
		"message": "Post liked successfully",
	})
}

func postPhoto(c Context) {
	file, err := RetrieveUploadedImage(c.Request, "post_picture")
	if err != nil {
		code, msg := CodeAndMessageForUploadError(err)
		c.Error(400, code, msg)
		return
	}

	imagePath, thumbnailPath, err := StoreImage(file, DefaultUploadOptions(c.Config))
	if err != nil {
		code, msg := CodeAndMessageForUploadError(err)
		c.Error(400, code, msg)
		return
	}

	p := NewPost(PostPhoto, c.User)
	p.PhotoURL = imagePath
	p.Thumbnail = thumbnailPath
	p.Caption = strings.TrimSpace(c.Form("caption"))
	p.Text = strings.TrimSpace(c.Form("post_text"))
	privacy, err := getPostPrivacy(PostPhoto, c)
	if err != nil {
		c.Error(400, CodeInvalidUserList, MsgInvalidUserList)
		return
	}

	p.Privacy = privacy

	if strlen(p.Text) > 1500 {
		c.Error(400, CodeInvalidStatusText, MsgInvalidStatusText)
		return
	}

	if strlen(p.Caption) > 255 {
		c.Error(400, CodeInvalidCaption, MsgInvalidCaption)
		return
	}

	if err := p.Save(c.Conn); err != nil {
		c.Error(500, CodeUnexpected, MsgUnexpected)

		os.Remove(p.PhotoURL)
		os.Remove(p.Thumbnail)
		return
	}

	go PropagatePostOnCreation(c, p)

	c.Success(201, map[string]interface{}{
		"message": "Photo posted successfully",
	})
}

func postVideo(c Context) {
	statusText := strings.TrimSpace(c.Form("post_text"))

	if strlen(statusText) > 1500 {
		c.Error(400, CodeInvalidStatusText, MsgInvalidStatusText)
		return
	}

	post := NewPost(PostVideo, c.User)
	post.Text = statusText
	privacy, err := getPostPrivacy(PostVideo, c)
	if err != nil {
		c.Error(400, CodeInvalidUserList, MsgInvalidUserList)
		return
	}

	post.Privacy = privacy

	valid, videoID, service, title := isValidVideo(strings.TrimSpace(c.Form("video_url")))

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

	go PropagatePostOnCreation(c, post)

	c.Success(201, map[string]interface{}{
		"message": "Video posted successfully",
	})
}

func postLink(c Context) {
	statusText := strings.TrimSpace(c.Form("post_text"))

	if strlen(statusText) > 1500 {
		c.Error(400, CodeInvalidStatusText, MsgInvalidStatusText)
		return
	}

	post := NewPost(PostVideo, c.User)
	post.Text = statusText
	privacy, err := getPostPrivacy(PostLink, c)
	if err != nil {
		c.Error(400, CodeInvalidUserList, MsgInvalidUserList)
		return
	}

	post.Privacy = privacy

	valid, link, title := isValidLink(strings.TrimSpace(c.Form("link_url")))

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

	go PropagatePostOnCreation(c, post)

	c.Success(201, map[string]interface{}{
		"message": "Link posted successfully",
	})
}

func postStatus(c Context) {
	statusText := strings.TrimSpace(c.Form("post_text"))

	if strlen(statusText) < 1 || strlen(statusText) > 1500 {
		c.Error(400, CodeInvalidStatusText, MsgInvalidStatusText)
		return
	}

	post := NewPost(PostStatus, c.User)
	post.Text = statusText
	privacy, err := getPostPrivacy(PostStatus, c)
	if err != nil {
		c.Error(400, CodeInvalidUserList, MsgInvalidUserList)
		return
	}

	post.Privacy = privacy
	if err := post.Save(c.Conn); err != nil {
		c.Error(500, CodeUnexpected, MsgUnexpected)
		return
	}

	go PropagatePostOnCreation(c, post)

	c.Success(201, map[string]interface{}{
		"message": "Status posted successfully",
	})
}

// CanBeAccessedBy determines if the current post can be accessed by the given user
func (p *Post) CanBeAccessedBy(user *User, conn *Connection) bool {
	if p.UserID.Hex() == user.ID.Hex() {
		return true
	}

	inUsersArray := func() bool {
		for _, i := range p.Privacy.Users {
			if i.Hex() == user.ID.Hex() {
				return true
			}
		}

		return false
	}()

	switch p.Privacy.Type {
	case PrivacyPublic:
		return true
	case PrivacyNone:
		return false
	case PrivacyFollowersOnly:
		return user.Follows(p.UserID, conn)
	case PrivacyFollowingOnly:
		return user.FollowedBy(p.UserID, conn)
	case PrivacyAllBut:
		return !inUsersArray
	case PrivacyNoneBut:
		return inUsersArray
	case PrivacyFollowersBut:
		return user.Follows(p.UserID, conn) && !inUsersArray
	case PrivacyFollowingBut:
		return user.FollowedBy(p.UserID, conn) && !inUsersArray
	}

	return false
}

func getPostPrivacy(postType ObjectType, c Context) (PrivacySettings, error) {
	p := PrivacySettings{}
	var pType int64
	var err error

	if pType, err = strconv.ParseInt(c.Form("privacy_type"), 10, 8); err != nil {
		pType = 0
	}

	privacyType := PrivacyType(pType)
	defaultSettings := c.User.Settings.GetPrivacySettings(postType)
	if privacyType == 0 {
		p.Type = defaultSettings.Type
	} else {
		if isValidPrivacyType(privacyType) {
			p.Type = privacyType
		} else {
			p.Type = defaultSettings.Type
		}
	}

	if p.Type == PrivacyAllBut || p.Type == PrivacyNoneBut {
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
