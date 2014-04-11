package models

import (
	"labix.org/v2/mgo/bson"
	"github.com/mvader/mask/services/interfaces"
	"time"
)

// Follow model
type Follow struct {
	ID   bson.ObjectId `json:"id" bson:"_id"`
	From bson.ObjectId `json:"user_from" bson:"user_from"`
	To   bson.ObjectId `json:"user_to" bson:"user_to"`
	Time float64       `json:"time" bson:"time"`
}

// FollowUser follows an user ("from" follows "to")
func FollowUser(from, to bson.ObjectId, conn interfaces.Saver) error {
	f := Follow{}
f.ID = bson.NewObjectId()
f.To = to
f.From = from
f.Time = float64(time.Now().Unix())

if err := conn.Save("follows", f.ID, f); err != nil {
return err
}

return nil
}

// UnfollowUser unfollows an user ("from" unfollows "to")
func UnfollowUser(from, to bson.ObjectId, conn interfaces.Conn) error {
	if err := conn.C("follows").Remove(bson.M{"user_from": from, "user_to": to}); err != nil {
		return err
	}

	return nil
}

// Follows checks if the user follows another user
func Follows(from, to bson.ObjectId, conn interfaces.Conn) bool {
	var (
		err   error
		count int
	)

	if count, err = conn.C("follows").Find(bson.M{"user_from": from, "user_to": to}).Count(); err != nil {
		return false
	}

	return count > 0
}

// FollowedBy checks if the user is followed by another user
func FollowedBy(to, from bson.ObjectId, conn interfaces.Conn) bool {
	var (
		err   error
		count int
	)

	if count, err = conn.C("follows").Find(bson.M{"user_to": to, "user_from": from}).Count(); err != nil {
		return false
	}

	return count > 0
}
