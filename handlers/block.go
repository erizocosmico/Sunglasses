package handlers

import (
	. "github.com/mvader/mask/error"
	"github.com/mvader/mask/middleware"
	"github.com/mvader/mask/models"
	"github.com/mvader/mask/modules/timeline"
	"labix.org/v2/mgo/bson"
)

// BlockHandler blocks an user
func BlockHandler(c middleware.Context) {
	userTo := c.Form("user_to")

	if userTo == "" || !bson.IsObjectIdHex(userTo) {
		c.Error(400, CodeInvalidData, MsgInvalidData)
		return
	}

	userToID := bson.ObjectIdHex(userTo)

	if toUser := models.UserExists(c.Conn, userToID); toUser == nil {
		c.Error(404, CodeUserDoesNotExist, MsgUserDoesNotExist)
		return
	}

	if models.UserIsBlocked(c.User.ID, userToID, c.Conn) {
		c.Success(200, map[string]interface{}{
			"error":   false,
			"message": "User was already blocked",
		})
		return
	}

	if err := models.BlockUser(c.User.ID, userToID, c.Conn); err != nil {
		c.Error(500, CodeUnexpected, MsgUnexpected)
		return
	}

	models.UnfollowUser(c.User.ID, userToID, c.Conn)
	go timeline.PropagatePostsOnUserUnfollow(c, userToID)

	c.Success(200, map[string]interface{}{
		"error":   false,
		"message": "User blocked successfully",
	})
}

// Unblock unblocks an user
func Unblock(c middleware.Context) {
	userTo := c.Form("user_to")

	if userTo == "" || !bson.IsObjectIdHex(userTo) {
		c.Error(400, CodeInvalidData, MsgInvalidData)
		return
	}

	userToID := bson.ObjectIdHex(userTo)

	if toUser := models.UserExists(c.Conn, userToID); toUser == nil {
		c.Error(404, CodeUserDoesNotExist, MsgUserDoesNotExist)
		return
	}

	if !models.UserIsBlocked(c.User.ID, userToID, c.Conn) {
		c.Success(200, map[string]interface{}{
			"error":   false,
			"message": "User was not blocked",
		})
		return
	}

	if err := models.UnblockUser(c.User.ID, userToID, c.Conn); err != nil {
		c.Error(500, CodeUnexpected, MsgUnexpected)
		return
	}

	c.Success(200, map[string]interface{}{
		"message": "User unblocked successfully",
	})
}

// ListBlocks retrieves a list with the users the user has blocked
func ListBlocks(c middleware.Context) {
	count, offset := c.ListCountParams()
	var result models.Block
	blocks := make([]models.Block, 0, count)

	cursor := c.Find("blocks", bson.M{"user_from": c.User.ID}).Limit(count).Skip(offset).Iter()
	for cursor.Next(&result) {
		blocks = append(blocks, result)
	}

	if err := cursor.Close(); err != nil {
		c.Error(500, CodeUnexpected, MsgUnexpected)
		return
	}

	users := make([]bson.ObjectId, 0, len(blocks))
	for _, b := range blocks {
		if b.To.Hex() != "" {
			users = append(users, b.To)
		}
	}

	usersData := models.GetUsersData(users, c.User, c.Conn)
	if usersData == nil {
		c.Error(500, CodeUnexpected, MsgUnexpected)
		return
	}

	blocksResponse := make([]map[string]interface{}, 0, len(usersData))
	for _, b := range blocks {
		if v, ok := usersData[b.To]; ok {
			blocksResponse = append(blocksResponse, map[string]interface{}{
				"time":         b.Time,
				"user_blocked": v,
			})
		}
	}

	c.Success(200, map[string]interface{}{
		"blocks": blocksResponse,
		"count":  len(blocksResponse),
	})
}
