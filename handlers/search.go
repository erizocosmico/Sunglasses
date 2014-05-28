package handlers

import (
	"fmt"
	. "github.com/mvader/sunglasses/error"
	"github.com/mvader/sunglasses/middleware"
	"github.com/mvader/sunglasses/models"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"strings"
)

// Search is a handler for searching users
func Search(c middleware.Context) {
	var user models.User
	count, offset := c.ListCountParams()

	search := c.Form("q")
	justFollowings := c.GetBoolean("just_followings")

	getUserIter := func(count, offset int) *mgo.Iter {
		if strings.TrimSpace(strings.ToLower(search)) == "" {
			return nil
		}

		regex := bson.M{"$regex": fmt.Sprintf(`(?i)^%s(.*)`, search)}

		iter := c.Find("users", bson.M{"$and": []bson.M{
			bson.M{"$or": []bson.M{
				bson.M{"username": regex},
				bson.M{"public_name": regex},
				bson.M{"private_name": regex},
			}},
			bson.M{"active": true},
		}}).Sort("username_lower").Limit(count).Skip(offset).Iter()

		return iter
	}

	users := make([]map[string]interface{}, 0, count)
	for len(users) < count {
		ids := make([]bson.ObjectId, 0, count)
		allIds := make([]bson.ObjectId, 0, count)

		iter := getUserIter(count, offset)
		if iter == nil {
			break
		}

		for iter.Next(&user) {
			if user.ID.Hex() == c.User.ID.Hex() {
				ids = append(ids, user.ID)
			} else if justFollowings || !user.Settings.Invisible {
				ids = append(ids, user.ID)
			} else if user.Settings.Invisible && !justFollowings {
				allIds = append(allIds, user.ID)
			}
		}

		if justFollowings {
			var follows []models.Follow
			followsIter := c.Find("follows", bson.M{"user_from": c.User.ID, "user_to": bson.M{"$in": ids}}).Iter()

			if err := followsIter.All(&follows); err != nil {
				c.Error(500, CodeUnexpected, MsgUnexpected)
				return
			}

			if len(follows) != len(ids) {
				ids = make([]bson.ObjectId, 0, len(follows))
				for _, f := range follows {
					ids = append(ids, f.To)
				}
			}

			if err := followsIter.Close(); err != nil {
				c.Error(500, CodeUnexpected, MsgUnexpected)
				return
			}
		} else {
			var follows []models.Follow
			followsIter := c.Find("follows", bson.M{"$or": []bson.M{
				bson.M{"user_from": c.User.ID, "user_to": bson.M{"$in": allIds}},
				bson.M{"user_to": c.User.ID, "user_from": bson.M{"$in": allIds}},
			}}).Iter()

			if err := followsIter.All(&follows); err != nil {
				c.Error(500, CodeUnexpected, MsgUnexpected)
				return
			}

			for _, f := range follows {
				if f.To.Hex() == c.User.ID.Hex() {
					ids = append(ids, f.From)
				} else {
					ids = append(ids, f.To)
				}
			}

			if err := followsIter.Close(); err != nil {
				c.Error(500, CodeUnexpected, MsgUnexpected)
				return
			}
		}

		userMap := models.GetUsersData(ids, c.User, c.Conn)
		if userMap != nil {
			for _, v := range userMap {
				users = append(users, v)
			}
		}

		if err := iter.Close(); err != nil {
			c.Error(500, CodeUnexpected, MsgUnexpected)
			return
		}

		// In case another iteration is needed update the offset
		offset += count

		if len(ids) == 0 {
			break
		}
	}

	c.Success(200, map[string]interface{}{
		"count": len(users),
		"users": users,
	})
}
