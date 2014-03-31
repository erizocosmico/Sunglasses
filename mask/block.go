package mask

import (
	"labix.org/v2/mgo/bson"
	"time"
)

// Block model (same as follow)
type Block Follow

// BlockUser blocks an user ("from" blocks "to")
func BlockUser(from, to bson.ObjectId, conn *Connection) error {
	f := Block{}
	f.ID = bson.NewObjectId()
	f.To = to
	f.From = from
	f.Time = float64(time.Now().Unix())

	if err := conn.Save("blocks", f.ID, f); err != nil {
		// Remove user blocked from follows
		return err
	}

	return nil
}

// UnblockUser unblocks an user ("from" unblocks "to")
func UnblockUser(from, to bson.ObjectId, conn *Connection) error {
	if err := conn.Db.C("blocks").Remove(bson.M{"user_from": from, "user_to": to}); err != nil {
		return err
	}

	return nil
}

// UserIsBlocked returns if the user is blocked
func UserIsBlocked(from, to bson.ObjectId, conn *Connection) bool {
	var (
		err   error
		count int
	)

	if count, err = conn.Db.C("blocks").Find(bson.M{"user_from": from, "user_to": to}).Count(); err != nil {
		return false
	}

	return count > 0
}

// Block blocks an user
func BlockHandler(c Context) {
	if c.User != nil {
		userTo := c.Form("user_to")

		if userTo != "" && bson.IsObjectIdHex(userTo) {
			userToID := bson.ObjectIdHex(userTo)

			if toUser := UserExists(c.Conn, userToID); toUser != nil {
				if UserIsBlocked(c.User.ID, userToID, c.Conn) {
					c.Success(200, map[string]interface{}{
						"error":   false,
						"message": "User was already blocked",
					})
					return
				}

				if err := BlockUser(c.User.ID, userToID, c.Conn); err != nil {
					c.Error(500, CodeUnexpected, MsgUnexpected)
					return
				}

				c.Success(200, map[string]interface{}{
					"error":   false,
					"message": "User blocked successfully",
				})
				return
			} else {
				c.Error(404, CodeUserDoesNotExist, MsgUserDoesNotExist)
				return
			}
		}
	}

	c.Error(400, CodeInvalidData, MsgInvalidData)
}

// Unblock unblocks an user
func Unblock(c Context) {
	if c.User != nil {
		userTo := c.Form("user_to")

		if userTo != "" && bson.IsObjectIdHex(userTo) {
			userToID := bson.ObjectIdHex(userTo)

			if toUser := UserExists(c.Conn, userToID); toUser != nil {
				if !UserIsBlocked(c.User.ID, userToID, c.Conn) {
					c.Success(200, map[string]interface{}{
						"error":   false,
						"message": "User was not blocked",
					})
					return
				}

				if err := UnblockUser(c.User.ID, userToID, c.Conn); err != nil {
					c.Error(500, CodeUnexpected, MsgUnexpected)
					return
				}

				c.Success(200, map[string]interface{}{
					"message": "User unblocked successfully",
				})
				return
			}

			c.Error(404, CodeUserDoesNotExist, MsgUserDoesNotExist)
			return
		}
	}

	c.Error(400, CodeInvalidData, MsgInvalidData)
}

// ListBlocks retrieves a list with the users the user has blocked
func ListBlocks(c Context) {
	count, offset := c.ListCountParams()
	var result Block
	blocks := make([]Block, 0, count)

	if c.User != nil {
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

		usersData := GetUsersData(users, c.User, c.Conn)
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
		return
	}

	c.Error(403, CodeUnauthorized, MsgUnauthorized)
}
