package models

import (
	"github.com/mvader/sunglasses/services/interfaces"
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
func (c *Comment) Save(conn interfaces.Saver) error {
	if c.ID.Hex() == "" {
		c.ID = bson.NewObjectId()
	}

	if err := conn.Save("comments", c.ID, c); err != nil {
		return err
	}

	return nil
}
