package lamp

import (
	"labix.org/v2/mgo/bson"
	"time"
)

type NotificationType int

type Notification struct {
	ID           bson.ObjectId    `json:"id" bson:"_id"`
	Type         NotificationType `json:"notification_type" bson:"notification_type"`
	PostID       bson.ObjectId    `json:"-" bson:"post_id,omitempty"`
	User         bson.ObjectId    `json:"user_id" bson:"user_id"`
	UserActionID bson.ObjectId    `json:"-" bson:"user_action_id,omitempty"`
	Post         Post             `json:"post" bson:"-"`
	UserAction   User             `json:"user_action" bson:"-"`
	Time         float64          `json:"time" bson:"time"`
	Read         bool             `json:"read" bson:"read"`
}

const (
	NotificationFollowRequest         = 1
	NotificationFollowRequestAccepted = 2
	NotificationFollowed              = 3
	NotificationPostLiked             = 4
	NotificationPostCommented         = 5
	NotificationPostOnMyWall          = 6
)

// SendNotification sends a new notification to the user
func SendNotification(notificationType NotificationType, postID, userID, userActionID bson.ObjectId, conn *Connection) error {
	n := Notification{}
	n.ID = bson.NewObjectId()
	n.Type = notificationType
	n.User = userID
	n.Read = false
	n.Time = float64(time.Now().Unix())

	if postID != nil {
		n.PostID = postID
	}

	if userActionID != nil {
		n.UserActionID = userActionID
	}

	if err := conn.Save("notifications", n.ID, n); err != nil {
		return err
	}

	return nil
}
