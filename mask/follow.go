package mask

import (
	"labix.org/v2/mgo/bson"
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

// FollowedBy checks if the user is followed by another user
func (u *User) FollowedBy(from bson.ObjectId, conn *Connection) bool {
	var (
		err   error
		count int
	)

	if count, err = conn.Db.C("follows").Find(bson.M{"user_to": u.ID, "user_from": from}).Count(); err != nil {
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
func SendFollowRequest(c Context) {
	var toUser *User
	var blankID bson.ObjectId

	userFrom := c.User

	if userFrom != nil {
		userTo := c.Form("user_to")

		if userTo != "" && bson.IsObjectIdHex(userTo) {
			userToID := bson.ObjectIdHex(userTo)

			if toUser = UserExists(c.Conn, userToID); toUser != nil {

				if !userFrom.Follows(userToID, c.Conn) {
					// If the user we want to follow already follows us, skip privacy settings
					if toUser.Follows(userFrom.ID, c.Conn) || !toUser.Settings.FollowApprovalRequired {
						if err := FollowUser(userFrom.ID, userToID, c.Conn); err == nil {
							go PropagatePostsOnUserFollow(c, userToID)
							SendNotification(NotificationFollowed, toUser, blankID, userFrom.ID, c.Conn)
						} else {
							c.Error(500, CodeUnexpected, MsgUnexpected)
							return
						}

						c.Success(200, map[string]interface{}{
							"message": "User followed successfully",
						})
						return
					} else {
						if !toUser.Settings.CanReceiveRequests || UserIsBlocked(userFrom.ID, userToID, c.Conn) {
							c.Error(403, CodeUserCantBeRequested, MsgUserCantBeRequested)
							return
						}

						fr := new(FollowRequest)
						fr.From = userFrom.ID
						fr.To = userToID
						fr.Msg = c.Form("request_message")
						fr.Time = float64(time.Now().Unix())

						if err := fr.Save(c.Conn); err == nil {
							SendNotification(NotificationFollowRequest, toUser, blankID, userFrom.ID, c.Conn)

							c.Success(200, map[string]interface{}{
								"message": "Follow request sent successfully",
							})
							return
						}

						c.Error(500, CodeUnexpected, MsgUnexpected)
						return
					}
				}

				c.Success(200, map[string]interface{}{
					"message": "You already follow that user",
				})
				return
			}

			c.Error(404, CodeUserDoesNotExist, MsgUserDoesNotExist)
			return
		}
	}

	c.Error(400, CodeInvalidData, MsgInvalidData)
}

// ReplyFollowRequests replies a follow request
func ReplyFollowRequest(c Context) {
	var blankID bson.ObjectId

	if c.User != nil {
		reqIDStr := c.Form("request_id")

		if reqIDStr != "" && bson.IsObjectIdHex(reqIDStr) {
			reqID := bson.ObjectIdHex(reqIDStr)
			var fr FollowRequest

			if err := c.Query("requests").FindId(reqID).One(&fr); err == nil {
				if fr.To.Hex() == c.User.ID.Hex() {
					if err := (&fr).Remove(c.Conn); err == nil {
						if c.Form("accept") == "yes" {
							if err := FollowUser(fr.From, c.User.ID, c.Conn); err == nil {
								SendNotification(NotificationFollowRequestAccepted, c.User, blankID, fr.To, c.Conn)
							} else {
								c.Error(500, CodeUnexpected, MsgUnexpected)
								return
							}
						}

						c.Success(200, map[string]interface{}{
							"message": "Successfully replied to follow request",
						})
						return
					}

					c.Error(500, CodeUnexpected, MsgUnexpected)
					return
				}

				c.Error(404, CodeUnauthorized, MsgUnauthorized)
				return
			}

			c.Error(404, CodeFollowRequestDoesNotExist, MsgFollowRequestDoesNotExist)
			return
		}
	}

	c.Error(400, CodeInvalidData, MsgInvalidData)
}

// Unfollow unfollows an user
func Unfollow(c Context) {
	if c.User != nil {
		userTo := c.Form("user_to")

		if userTo != "" && bson.IsObjectIdHex(userTo) {
			userToID := bson.ObjectIdHex(userTo)

			if toUser := UserExists(c.Conn, userToID); toUser != nil {
				if c.User.Follows(userToID, c.Conn) {
					if err := UnfollowUser(c.User.ID, userToID, c.Conn); err != nil {
						c.Error(500, CodeUnexpected, MsgUnexpected)
						return
					}

					go PropagatePostsOnUserUnfollow(c, userToID)

					c.Success(200, map[string]interface{}{
						"message": "User unfollowed successfully",
					})
					return
				}

				c.Success(200, map[string]interface{}{
					"message": "You can't unfollow that user",
				})
				return
			}

			c.Error(404, CodeUserDoesNotExist, MsgUserDoesNotExist)
			return
		}
	}

	c.Error(400, CodeInvalidData, MsgInvalidData)
}

// ListFollowers retrieves a list with the user's followers
func ListFollowers(c Context) {
	listFollows(c, true)
}

// ListFollowing retrieves a list with the users followed by the user
func ListFollowing(c Context) {
	listFollows(c, false)
}

func listFollows(c Context, listFollowers bool) {
	count, offset := c.ListCountParams()
	var result Follow
	follows := make([]Follow, 0, count)

	if c.User != nil {
		var key, outputKey string
		if listFollowers {
			key = "user_to"
			outputKey = "followers"
		} else {
			key = "user_from"
			outputKey = "followings"
		}

		cursor := c.Find("follows", bson.M{key: c.User.ID}).Limit(count).Skip(offset).Iter()
		for cursor.Next(&result) {
			follows = append(follows, result)
		}

		if err := cursor.Close(); err != nil {
			c.Error(500, CodeUnexpected, MsgUnexpected)
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

		usersData := GetUsersData(users, c.User, c.Conn)
		if usersData == nil {
			c.Error(500, CodeUnexpected, MsgUnexpected)
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

		c.Success(200, map[string]interface{}{
			outputKey: followsResponse,
			"count":   len(followsResponse),
		})
		return
	}

	c.Error(403, CodeUnauthorized, MsgUnauthorized)
}

// ListFollowRequests returns all the user's follow requests
func ListFollowRequests(c Context) {
	count, offset := c.ListCountParams()
	var result FollowRequest
	requests := make([]FollowRequest, 0)

	if c.User != nil {
		cursor := c.Find("requests", bson.M{"user_to": c.User.ID}).Limit(count).Skip(offset).Iter()
		for cursor.Next(&result) {
			requests = append(requests, result)
		}

		if err := cursor.Close(); err != nil {
			c.Error(500, CodeUnexpected, MsgUnexpected)
			return
		}

		users := make([]bson.ObjectId, 0, len(requests))
		for _, f := range requests {
			if f.From.Hex() != "" {
				users = append(users, f.From)
			}
		}

		usersData := GetUsersData(users, c.User, c.Conn)
		if usersData == nil {
			c.Error(500, CodeUnexpected, MsgUnexpected)
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

		c.Success(200, map[string]interface{}{
			"follow_requests": requestsResponse,
			"count":           len(requestsResponse),
		})
		return
	}

	c.Error(403, CodeUnauthorized, MsgUnauthorized)
}
