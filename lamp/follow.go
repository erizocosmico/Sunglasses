package lamp

import (
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	"labix.org/v2/mgo/bson"
	"net/http"
	"time"
)

// FollowRequest model
type FollowRequest struct {
	ID   bson.ObjectId `json:"id" bson:"_id"`
	From bson.ObjectId `json:"id" bson:"from_id"`
	To   bson.ObjectId `json:"id" bson:"to_id"`
	Msg  string        `json:"msg,omitempty" bson:"msg,omitempty"`
	Time float64       `json:"time" bson:"time"`
}

// Follow model
type Follow struct {
	ID   bson.ObjectId `json:"id" bson:"_id"`
	From bson.ObjectId `json:"from_id" bson:"from_id"`
	To   bson.ObjectId `json:"to_id" bson:"to_id"`
	Time float64       `json:"time" bson:"time"`
}

// FollowUser follows an user ("from" follows "to")
func FollowUser(from, to bson.ObjectId, conn *Connection) error {
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

// Save inserts the FollowRequest instance if it hasn't been created yet or updates it if it has
func (fr *FollowRequest) Save(conn *Connection) error {
	if fr.ID.Hex() == "" {
		fr.ID = bson.NewObjectId()
	}

	if err := conn.Save("requests", fr.ID, fr); err != nil {
		return err
	}

	return nil
}

// Remove deletes the follow request instance
func (fr *FollowRequest) Remove(conn *Connection) error {
	return conn.Remove("requests", fr.ID)
}

// SendFollowRequests sends a follow request from an user to another user
func SendFollowRequest(r *http.Request, conn *Connection, res render.Render, s sessions.Session) {
	// TODO check if the user can be requested (privacy)
	userFrom := GetRequestUser(r, conn, s)
	if userFrom != nil {
		userTo := r.PostFormValue("user_to")
		if userTo != "" && bson.IsObjectIdHex(userTo) {
			userToID := bson.ObjectIdHex(userTo)
			if UserExists(conn, userToID) {
				fr := new(FollowRequest)
				fr.From = userFrom.ID
				fr.To = userToID
				fr.Msg = r.PostFormValue("request_message")
				fr.Time = float64(time.Now().Unix())

				if err := fr.Save(conn); err == nil {
					SendNotification(NotificationFollowRequest, nil, userToID, userFrom.ID, conn)

					res.JSON(200, map[string]interface{}{
						"error":   false,
						"message": "Follow request sent successfully",
					})
				}
			}
		}
	}

	RenderError(res, 5, 400, "Error sending the follow request")
}

// ReplyFollowRequests replies a follow request
func ReplyFollowRequest(r *http.Request, conn *Connection, res render.Render, s sessions.Session) {
	user := GetRequestUser(r, conn, s)
	if user != nil {
		reqIDStr := r.PostFormValue("request_id")
		if reqIDStr != "" && bson.IsObjectIdHex(reqIDStr) {
			reqID := bson.ObjectIdHex(reqIDStr)
			var fr FollowRequest
			if err := conn.Db.C("requests").FindId(reqID).One(&fr); err == nil {
				if err := (&fr).Remove(conn); err == nil {
					if r.PostFormValue("accept") == "yes" {
						if err := FollowUser(fr.From, user.ID, conn); err == nil {
							SendNotification(NotificationFollowRequestAccepted, nil, fr.From, fr.To, conn)
						}
					}

					res.JSON(200, map[string]interface{}{
						"error": false,
					})
				}
			}
		}
	}

	RenderError(res, 6, 400, "Error replying to follow request")
}
