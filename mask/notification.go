package mask

import (
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	"labix.org/v2/mgo/bson"
	"net/http"
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
func MarkNotificationRead(r *http.Request, conn *Connection, res render.Render, s sessions.Session) {
	user := GetRequestUser(r, conn, s)
	var n Notification

	if user != nil {
		nid := r.PostFormValue("notification_id")

		if nid != "" && bson.IsObjectIdHex(nid) {
			notificationID := bson.ObjectIdHex(nid)

			if err := conn.Db.C("notifications").FindId(notificationID).One(&n); err != nil {
				RenderError(res, CodeNotFound, 404, MsgNotFound)
				return
			}

			if n.User != user.ID {
				RenderError(res, CodeUnauthorized, 403, MsgUnauthorized)
				return
			}

			if !n.Read {
				n.Read = true

				if _, err := conn.Db.C("notifications").UpsertId(n.ID, n); err != nil {
					RenderError(res, CodeUnexpected, 500, MsgUnexpected)
					return
				}
			}

			res.JSON(200, map[string]interface{}{
				"error":   false,
				"message": "Notification marked successfully as read",
			})
			return
		}
	}

	RenderError(res, CodeInvalidData, 400, MsgInvalidData)
}

// ListNotifications list all the user's notifications
func ListNotifications(r *http.Request, conn *Connection, res render.Render, s sessions.Session) {
	user := GetRequestUser(r, conn, s)
	count, offset := ListCountParams(r)
	var result Notification
	notifications := make([]Notification, 0, count)

	if user != nil {
		cursor := conn.Db.C("notifications").Find(bson.M{"user_id": user.ID}).Limit(count).Skip(offset).Iter()
		for cursor.Next(&result) {
			notifications = append(notifications, result)
		}

		if err := cursor.Close(); err != nil {
			RenderError(res, CodeUnexpected, 500, MsgUnexpected)
			return
		}

		users := make([]bson.ObjectId, 0, len(notifications))
		for _, n := range notifications {
			if n.UserActionID.Hex() != "" {
				users = append(users, n.UserActionID)
			}
		}

		usersData := GetUsersData(users, false, conn)
		if usersData == nil {
			RenderError(res, CodeUnexpected, 500, MsgUnexpected)
			return
		}

		for i, n := range notifications {
			if u, ok := usersData[n.UserActionID]; ok {
				notifications[i].UserAction = u
			}
		}

		res.JSON(200, map[string]interface{}{
			"error":         false,
			"notifications": notifications,
			"count":         len(notifications),
		})
		return
	}

	RenderError(res, CodeUnauthorized, 403, MsgUnauthorized)
}
