package mask

import (
	"labix.org/v2/mgo/bson"
	"time"
)

// NotificationType is the type of the notification
type NotificationType int

// Notification model
type Notification struct {
	ID           bson.ObjectId    `json:"id" bson:"_id"`
	Type         NotificationType `json:"notification_type" bson:"notification_type"`
	PostID       bson.ObjectId    `json:"post_id" bson:"post_id,omitempty"`
	User         bson.ObjectId    `json:"user_id" bson:"user_id"`
	UserActionID bson.ObjectId    `json:"-" bson:"user_action_id,omitempty"`
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
func SendNotification(notificationType NotificationType, user *User, postID, userActionID bson.ObjectId, conn *Connection) error {
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

// MarkNotificationRead marks a notification as read
func MarkNotificationRead(c Context) {
	var n Notification

	if c.User != nil {
		nid := c.Form("notification_id")

		if nid != "" && bson.IsObjectIdHex(nid) {
			notificationID := bson.ObjectIdHex(nid)

			if err := c.FindId("notifications", notificationID).One(&n); err != nil {
				c.Error(404, CodeNotFound, MsgNotFound)
				return
			}

			if n.User != c.User.ID {
				c.Error(403, CodeUnauthorized, MsgUnauthorized)
				return
			}

			if !n.Read {
				n.Read = true

				if _, err := c.Query("notifications").UpsertId(n.ID, n); err != nil {
					c.Error(500, CodeUnexpected, MsgUnexpected)
					return
				}
			}

			c.Success(200, map[string]interface{}{
				"message": "Notification marked successfully as read",
			})
			return
		}
	}

	c.Error(400, CodeInvalidData, MsgInvalidData)
}

// ListNotifications list all the user's notifications
func ListNotifications(c Context) {
	count, offset := c.ListCountParams()
	var result Notification
	notifications := make([]Notification, 0, count)

	if c.User != nil {
		cursor := c.Find("notifications", bson.M{"user_id": c.User.ID}).Limit(count).Skip(offset).Iter()
		for cursor.Next(&result) {
			notifications = append(notifications, result)
		}

		if err := cursor.Close(); err != nil {
			c.Error(500, CodeUnexpected, MsgUnexpected)
			return
		}

		users := make([]bson.ObjectId, 0, len(notifications))
		for _, n := range notifications {
			if n.UserActionID.Hex() != "" {
				users = append(users, n.UserActionID)
			}
		}

		usersData := GetUsersData(users, c.User, c.Conn)
		if usersData == nil {
			c.Error(500, CodeUnexpected, MsgUnexpected)
			return
		}

		for i, n := range notifications {
			if u, ok := usersData[n.UserActionID]; ok {
				notifications[i].UserAction = u
			}
		}

		c.Success(200, map[string]interface{}{
			"notifications": notifications,
			"count":         len(notifications),
		})
		return
	}

	c.Error(403, CodeUnauthorized, MsgUnauthorized)
}
