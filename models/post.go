package models

import (
	"github.com/mvader/sunglasses/services/interfaces"
	"labix.org/v2/mgo/bson"
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
func (p *Post) Save(conn interfaces.Saver) error {
	if p.ID.Hex() == "" {
		p.ID = bson.NewObjectId()
	}

	if err := conn.Save("posts", p.ID, p); err != nil {
		return err
	}

	return nil
}

// CanBeAccessedBy determines if the current post can be accessed by the given user
func (p *Post) CanBeAccessedBy(u *User, conn interfaces.Conn) bool {
	if p.UserID.Hex() == u.ID.Hex() {
		return true
	}

	inUsersArray := false
	for _, i := range p.Privacy.Users {
		if i.Hex() == u.ID.Hex() {
			inUsersArray = true
			break
		}
	}

	switch int(p.Privacy.Type) {
	case PrivacyPublic:
		return true
	case PrivacyNone:
		return false
	case PrivacyFollowersOnly:
		return Follows(u.ID, p.UserID, conn)
	case PrivacyFollowingOnly:
		return FollowedBy(u.ID, p.UserID, conn)
	case PrivacyAllBut:
		return !inUsersArray
	case PrivacyNoneBut:
		return inUsersArray
	case PrivacyFollowersBut:
		return Follows(u.ID, p.UserID, conn) && !inUsersArray
	case PrivacyFollowingBut:
		return FollowedBy(u.ID, p.UserID, conn) && !inUsersArray
	}

	return false
}

// GetLikesForPosts gets all user likes for a list of posts
func GetLikesForPosts(posts []bson.ObjectId, user bson.ObjectId, conn interfaces.Conn) map[bson.ObjectId]bool {
	var like PostLike
	result := make(map[bson.ObjectId]bool)

	for _, v := range posts {
		result[v] = false
	}

	iter := conn.C("likes").Find(bson.M{"post_id": bson.M{"$in": posts}, "user_id": user}).Iter()
	for iter.Next(&like) {
		result[like.PostID] = true
	}

	return result
}

// GetCommentsForPost returns up to N comments for the given post
func GetCommentsForPost(post bson.ObjectId, user *User, n int, conn interfaces.Conn) []Comment {
	uids := make([]bson.ObjectId, 0, n)
	var comments []Comment

	iter := conn.C("comments").Find(bson.M{"post_id": post}).Limit(n).Sort("created").Iter()
	err := iter.All(&comments)
	if err != nil {
		return nil
	}

	for _, v := range comments {
		uids = append(uids, v.UserID)
	}

	data := GetUsersData(uids, user, conn)
	if len(data) == 0 {
		return nil
	}

	for i, _ := range comments {
		comments[i].User = data[comments[i].UserID]
	}

	return comments
}
