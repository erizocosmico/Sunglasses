package handlers

import (
	"fmt"
	. "github.com/mvader/mask/error"
	"github.com/mvader/mask/middleware"
	"github.com/mvader/mask/models"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

// Search is a handler for searching users
func Search(c middleware.Context) {
	var user models.User
	count, offset := c.ListCountParams()

	search := c.Form("q")
	if c.User == nil || search == "" {
		c.Error(400, CodeInvalidData, MsgInvalidData)
		return
	}

	justFollowings := c.GetBoolean("just_followings")

	getUserIter := func(count, offset int) *mgo.Iter {
		regex := bson.M{"$regex": fmt.Sprintf(`(?i)%s`, search)}

		iter := c.Find("users", bson.M{"$and": []bson.M{
			bson.M{"$or": []bson.M{
				bson.M{"username": regex},
				bson.M{"public_name": regex},
				bson.M{"private_name": regex},
			}},
			bson.M{"active": true},
			bson.M{"settings.invisible": false},
		}}).Limit(count).Skip(offset).Iter()

		return iter
	}

	users := make([]map[string]interface{}, 0, count)
	for len(users) < count {
		ids := make([]bson.ObjectId, 0, count)
		iter := getUserIter(count, offset)
		for iter.Next(&user) {
			ids = append(ids, user.ID)
		}

		if justFollowings {
			var follows []models.Follow
			followsIter := c.Find("follows", bson.M{"user_from": user.ID, "user_to": bson.M{"$in": ids}}).Iter()

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

		if len(ids) == 0 {
			break
		}
	}

	c.Success(200, map[string]interface{}{
		"count": len(users),
		"users": users,
	})
}
