package models

import (
	"github.com/mvader/sunglasses/services/interfaces"
	"labix.org/v2/mgo/bson"
)

// FollowRequest model
type FollowRequest struct {
	ID   bson.ObjectId `json:"id" bson:"_id"`
	From bson.ObjectId `json:"user_from" bson:"user_from"`
	To   bson.ObjectId `json:"user_to" bson:"user_to"`
	Msg  string        `json:"msg,omitempty" bson:"msg,omitempty"`
	Time float64       `json:"time" bson:"time"`
}

// Save inserts the FollowRequest instance if it hasn't been created yet or updates it if it has
func (fr *FollowRequest) Save(conn interfaces.Saver) error {
	if fr.ID.Hex() == "" {
		fr.ID = bson.NewObjectId()
	}

	if err := conn.Save("requests", fr.ID, fr); err != nil {
		return err
	}

	return nil
}

// Remove deletes the follow request instance
func (fr *FollowRequest) Remove(conn interfaces.Remover) error {
	return conn.Remove("requests", fr.ID)
}
