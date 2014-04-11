package models

import (
	"github.com/mvader/mask/services/interfaces"
	"labix.org/v2/mgo/bson"
	"time"
)

// NotificationType is the type of the notification
type NotificationType int

// Notification model
type Notification struct {
	ID           bson.ObjectId          `json:"id" bson:"_id"`
	Type         NotificationType       `json:"notification_type" bson:"notification_type"`
	PostID       bson.ObjectId          `json:"post_id" bson:"post_id,omitempty"`
	User         bson.ObjectId          `json:"user_id" bson:"user_id"`
	UserActionID bson.ObjectId          `json:"-" bson:"user_action_id,omitempty"`
	UserAction   map[string]interface{} `json:"user_action" bson:"-"`
	Time         float64                `json:"time" bson:"time"`
	Read         bool                   `json:"read" bson:"read"`
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
func SendNotification(notificationType NotificationType, user *User, postID, userActionID bson.ObjectId, conn interfaces.Saver) error {
	switch int(notificationType) {
	case NotificationPostLiked:
		if !user.Settings.NotifyLikes {
			return nil
		}
		break

	case NotificationPostCommented:
		if !user.Settings.NotifyNewComment {
			return nil
		}
		break

	case NotificationPostOnMyWall:
		if !user.Settings.NotifyPostsInMyProfile {
			return nil
		}
		break
	}

	n := Notification{}
	n.ID = bson.NewObjectId()
	n.Type = notificationType
	n.User = user.ID
	n.Read = false
	n.Time = float64(time.Now().Unix())

	if postID.Hex() != "" {
		n.PostID = postID
	}

	if userActionID.Hex() != "" {
		n.UserActionID = userActionID
	}

	if err := conn.Save("notifications", n.ID, n); err != nil {
		return err
	}

	return nil
}
