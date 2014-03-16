package mask

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
	From bson.ObjectId `json:"user_from" bson:"user_from"`
	To   bson.ObjectId `json:"user_to" bson:"user_to"`
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

// UnfollowUser unfollows an user ("from" unfollows "to")
func UnfollowUser(from, to bson.ObjectId, conn *Connection) error {
	if err := conn.Db.C("follows").Remove(bson.M{"user_from": from, "user_to": to}); err != nil {
		return err
	}

	return nil
}

// Follows checks if the user follows another user
func (u *User) Follows(to bson.ObjectId, conn *Connection) bool {
	var (
		err   error
		count int
	)

	if count, err = conn.Db.C("follows").Find(bson.M{"user_from": u.ID, "user_to": to}).Count(); err != nil {
		return false
	}

	return count > 0
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
	var toUser *User
	var blankID bson.ObjectId

	userFrom := GetRequestUser(r, conn, s)

	if userFrom != nil {
		userTo := r.PostFormValue("user_to")

		if userTo != "" && bson.IsObjectIdHex(userTo) {
			userToID := bson.ObjectIdHex(userTo)

			if toUser = UserExists(conn, userToID); toUser != nil {

				if !userFrom.Follows(userToID, conn) {
					// If the user we want to follow already follows us, skip privacy settings
					if toUser.Follows(userFrom.ID, conn) || !toUser.Settings.FollowApprovalRequired {
						if err := FollowUser(userFrom.ID, userToID, conn); err == nil {
							SendNotification(NotificationFollowed, blankID, userFrom.ID, userToID, conn)
						} else {
							RenderError(res, CodeUnexpected, 500, MsgUnexpected)
							return
						}

						res.JSON(200, map[string]interface{}{
							"error":   false,
							"message": "User followed successfully",
						})
						return
					} else {
						if !toUser.Settings.CanReceiveRequests || UserIsBlocked(userFrom.ID, userToID, conn) {
							RenderError(res, CodeUserCantBeRequested, 403, MsgUserCantBeRequested)
							return
						}

						fr := new(FollowRequest)
						fr.From = userFrom.ID
						fr.To = userToID
						fr.Msg = r.PostFormValue("request_message")
						fr.Time = float64(time.Now().Unix())

						if err := fr.Save(conn); err == nil {
							SendNotification(NotificationFollowRequest, blankID, userToID, userFrom.ID, conn)

							res.JSON(200, map[string]interface{}{
								"error":   false,
								"message": "Follow request sent successfully",
							})
							return
						} else {
							RenderError(res, CodeUnexpected, 500, MsgUnexpected)
							return
						}
					}
				} else {
					res.JSON(200, map[string]interface{}{
						"error":   false,
						"message": "You already follow that user",
					})
					return
				}
			} else {
				RenderError(res, CodeUserDoesNotExist, 404, MsgUserDoesNotExist)
				return
			}
		}
	}

	RenderError(res, CodeInvalidData, 400, MsgInvalidData)
}

// ReplyFollowRequests replies a follow request
func ReplyFollowRequest(r *http.Request, conn *Connection, res render.Render, s sessions.Session) {
	var blankID bson.ObjectId
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
							SendNotification(NotificationFollowRequestAccepted, blankID, fr.From, fr.To, conn)
						} else {
							RenderError(res, CodeUnexpected, 500, MsgUnexpected)
							return
						}
					}

					res.JSON(200, map[string]interface{}{
						"error":   false,
						"message": "Successfully replied to follow request",
					})
					return
				}
			}

			RenderError(res, CodeFollowRequestDoesNotExist, 404, MsgFollowRequestDoesNotExist)
			return
		}
	}

	RenderError(res, CodeInvalidData, 400, MsgInvalidData)
}

// Unfollow unfollows an user
func Unfollow(r *http.Request, conn *Connection, res render.Render, s sessions.Session) {
	user := GetRequestUser(r, conn, s)

	if user != nil {
		userTo := r.PostFormValue("user_to")

		if userTo != "" && bson.IsObjectIdHex(userTo) {
			userToID := bson.ObjectIdHex(userTo)

			if toUser := UserExists(conn, userToID); toUser != nil {
				if user.Follows(userToID, conn) {
					if err := UnfollowUser(user.ID, userToID, conn); err != nil {
						RenderError(res, CodeUnexpected, 500, MsgUnexpected)
						return
					}

					res.JSON(200, map[string]interface{}{
						"error":   false,
						"message": "User unfollowed successfully",
					})
					return
				} else {
					res.JSON(200, map[string]interface{}{
						"error":   false,
						"message": "You can't unfollow that user",
					})
					return
				}
			} else {
				RenderError(res, CodeUserDoesNotExist, 404, MsgUserDoesNotExist)
				return
			}
		}
	}

	RenderError(res, CodeInvalidData, 400, MsgInvalidData)
}
