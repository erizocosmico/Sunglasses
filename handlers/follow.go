package handlers

import (
	. "github.com/mvader/mask/error"
	"github.com/mvader/mask/middleware"
	"github.com/mvader/mask/models"
	"github.com/mvader/mask/modules/timeline"
	"labix.org/v2/mgo/bson"
	"time"
)

// SendFollowRequests sends a follow request from an user to another user
func SendFollowRequest(c middleware.Context) {
	var toUser *models.User
	var blankID bson.ObjectId

	userFrom := c.User

	if userFrom != nil {
		userTo := c.Form("user_to")

		if userTo != "" && bson.IsObjectIdHex(userTo) {
			userToID := bson.ObjectIdHex(userTo)

			if toUser = models.UserExists(c.Conn, userToID); toUser != nil {

				if !models.Follows(userFrom.ID, userToID, c.Conn) {
					// If the user we want to follow already follows us, skip privacy settings
					if models.Follows(toUser.ID, userFrom.ID, c.Conn) || !toUser.Settings.FollowApprovalRequired {
						if err := models.FollowUser(userFrom.ID, userToID, c.Conn); err == nil {
							go timeline.PropagatePostsOnUserFollow(c, userToID)
							models.SendNotification(models.NotificationFollowed, toUser, blankID, userFrom.ID, c.Conn)
						} else {
							c.Error(500, CodeUnexpected, MsgUnexpected)
							return
						}

						c.Success(200, map[string]interface{}{
							"message": "User followed successfully",
						})
						return
					} else {
						if !toUser.Settings.CanReceiveRequests || models.UserIsBlocked(userFrom.ID, userToID, c.Conn) {
							c.Error(403, CodeUserCantBeRequested, MsgUserCantBeRequested)
							return
						}

						fr := new(models.FollowRequest)
						fr.From = userFrom.ID
						fr.To = userToID
						fr.Msg = c.Form("request_message")
						fr.Time = float64(time.Now().Unix())

						if err := fr.Save(c.Conn); err == nil {
							models.SendNotification(models.NotificationFollowRequest, toUser, blankID, userFrom.ID, c.Conn)

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
func ReplyFollowRequest(c middleware.Context) {
	var blankID bson.ObjectId

	if c.User != nil {
		reqIDStr := c.Form("request_id")

		if reqIDStr != "" && bson.IsObjectIdHex(reqIDStr) {
			reqID := bson.ObjectIdHex(reqIDStr)
			var fr models.FollowRequest

			if err := c.Query("requests").FindId(reqID).One(&fr); err == nil {
				if fr.To.Hex() == c.User.ID.Hex() {
					if err := (&fr).Remove(c.Conn); err == nil {
						if c.Form("accept") == "yes" {
							if err := models.FollowUser(fr.From, c.User.ID, c.Conn); err == nil {
								models.SendNotification(models.NotificationFollowRequestAccepted, c.User, blankID, fr.To, c.Conn)
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
func Unfollow(c middleware.Context) {
	if c.User != nil {
		userTo := c.Form("user_to")

		if userTo != "" && bson.IsObjectIdHex(userTo) {
			userToID := bson.ObjectIdHex(userTo)

			if toUser := models.UserExists(c.Conn, userToID); toUser != nil {
				if models.Follows(c.User.ID, userToID, c.Conn) {
					if err := models.UnfollowUser(c.User.ID, userToID, c.Conn); err != nil {
						c.Error(500, CodeUnexpected, MsgUnexpected)
						return
					}

					go timeline.PropagatePostsOnUserUnfollow(c, userToID)

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
func ListFollowers(c middleware.Context) {
	listFollows(c, true)
}

// ListFollowing retrieves a list with the users followed by the user
func ListFollowing(c middleware.Context) {
	listFollows(c, false)
}

func listFollows(c middleware.Context, listFollowers bool) {
	count, offset := c.ListCountParams()
	var result models.Follow
	follows := make([]models.Follow, 0, count)

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

		usersData := models.GetUsersData(users, c.User, c.Conn)
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
func ListFollowRequests(c middleware.Context) {
	count, offset := c.ListCountParams()
	var result models.FollowRequest
	requests := make([]models.FollowRequest, 0)

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

		usersData := models.GetUsersData(users, c.User, c.Conn)
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
