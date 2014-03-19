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
	From bson.ObjectId `json:"user_from" bson:"user_from"`
	To   bson.ObjectId `json:"user_to" bson:"user_to"`
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
							SendNotification(NotificationFollowed, toUser, blankID, userFrom.ID, conn)
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
							SendNotification(NotificationFollowRequest, toUser, blankID, userFrom.ID, conn)

							res.JSON(200, map[string]interface{}{
								"error":   false,
								"message": "Follow request sent successfully",
							})
							return
						}

						RenderError(res, CodeUnexpected, 500, MsgUnexpected)
						return
					}
				}

				res.JSON(200, map[string]interface{}{
					"error":   false,
					"message": "You already follow that user",
				})
				return
			}

			RenderError(res, CodeUserDoesNotExist, 404, MsgUserDoesNotExist)
			return
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
				if fr.To.Hex() == user.ID.Hex() {
					if err := (&fr).Remove(conn); err == nil {
						if r.PostFormValue("accept") == "yes" {
							if err := FollowUser(fr.From, user.ID, conn); err == nil {
								SendNotification(NotificationFollowRequestAccepted, user, blankID, fr.To, conn)
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

					RenderError(res, CodeUnexpected, 500, MsgUnexpected)
					return
				}

				RenderError(res, CodeUnauthorized, 404, MsgUnauthorized)
				return
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
				}

				res.JSON(200, map[string]interface{}{
					"error":   false,
					"message": "You can't unfollow that user",
				})
				return
			}

			RenderError(res, CodeUserDoesNotExist, 404, MsgUserDoesNotExist)
			return
		}
	}

	RenderError(res, CodeInvalidData, 400, MsgInvalidData)
}

// ListFollowers retrieves a list with the user's followers
func ListFollowers(r *http.Request, conn *Connection, res render.Render, s sessions.Session) {
	listFollows(r, conn, res, s, true)
}

// ListFollowing retrieves a list with the users followed by the user
func ListFollowing(r *http.Request, conn *Connection, res render.Render, s sessions.Session) {
	listFollows(r, conn, res, s, false)
}

func listFollows(r *http.Request, conn *Connection, res render.Render, s sessions.Session, listFollowers bool) {
	user := GetRequestUser(r, conn, s)
	count, offset := ListCountParams(r)
	var result Follow
	follows := make([]Follow, 0)

	if user != nil {
		var key, outputKey string
		if listFollowers {
			key = "user_to"
			outputKey = "followers"
		} else {
			key = "user_from"
			outputKey = "followings"
		}

		cursor := conn.Db.C("follows").Find(bson.M{key: user.ID}).Limit(count).Skip(offset).Iter()
		for cursor.Next(&result) {
			follows = append(follows, result)
		}

		if err := cursor.Close(); err != nil {
			RenderError(res, CodeUnexpected, 500, MsgUnexpected)
			return
		}

		users := make([]bson.ObjectId, 0, len(follows))
		for _, f := range follows {
			if listFollowers {
				if f.From.Hex() != "" {
					users = append(users, f.From)
				}
			} else {
				if f.To.Hex() != "" {
					users = append(users, f.To)
				}
			}

		}

		usersData := GetUsersData(users, false, conn)
		if usersData == nil {
			RenderError(res, CodeUnexpected, 500, MsgUnexpected)
			return
		}

		followsResponse := make([]map[string]interface{}, 0, len(usersData))
		for _, f := range follows {
			var u bson.ObjectId
			if listFollowers {
				u = f.From
			} else {
				u = f.To
			}

			if v, ok := usersData[u]; ok {
				followsResponse = append(followsResponse, map[string]interface{}{
					"time": f.Time,
					"user": v,
				})
			}
		}

		res.JSON(200, map[string]interface{}{
			"error":   false,
			outputKey: followsResponse,
			"count":   len(followsResponse),
		})
		return
	}

	RenderError(res, CodeUnauthorized, 403, MsgUnauthorized)
}

// ListFollowRequests returns all the user's follow requests
func ListFollowRequests(r *http.Request, conn *Connection, res render.Render, s sessions.Session) {
	user := GetRequestUser(r, conn, s)
	count, offset := ListCountParams(r)
	var result FollowRequest
	requests := make([]FollowRequest, 0)

	if user != nil {
		cursor := conn.Db.C("requests").Find(bson.M{"user_to": user.ID}).Limit(count).Skip(offset).Iter()
		for cursor.Next(&result) {
			requests = append(requests, result)
		}

		if err := cursor.Close(); err != nil {
			RenderError(res, CodeUnexpected, 500, MsgUnexpected)
			return
		}

		users := make([]bson.ObjectId, 0, len(requests))
		for _, f := range requests {
			if f.From.Hex() != "" {
				users = append(users, f.From)
			}
		}

		usersData := GetUsersData(users, false, conn)
		if usersData == nil {
			RenderError(res, CodeUnexpected, 500, MsgUnexpected)
			return
		}

		requestsResponse := make([]map[string]interface{}, 0, len(usersData))
		for _, f := range requests {
			if v, ok := usersData[f.From]; ok {
				requestsResponse = append(requestsResponse, map[string]interface{}{
					"time":    f.Time,
					"user":    v,
					"message": f.Msg,
				})
			}
		}

		res.JSON(200, map[string]interface{}{
			"error":           false,
			"follow_requests": requestsResponse,
			"count":           len(requestsResponse),
		})
		return
	}

	RenderError(res, CodeUnauthorized, 403, MsgUnauthorized)
}
