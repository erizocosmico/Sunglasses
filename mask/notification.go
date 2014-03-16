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
	// TODO check if user can be notified
	n := Notification{}
	n.ID = bson.NewObjectId()
	n.Type = notificationType
	n.User = userID
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
