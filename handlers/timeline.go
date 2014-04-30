package handlers

import (
	. "github.com/mvader/mask/error"
	"github.com/mvader/mask/middleware"
	"github.com/mvader/mask/models"
	"labix.org/v2/mgo/bson"
)

// GetUserTimeline gets all posts, comments and likes needed to render the timeline for the user
func GetUserTimeline(c middleware.Context) {
	count, offset := c.ListCountParams()

	var (
		t           models.TimelineEntry
		comments    = make(map[bson.ObjectId][]models.Comment)
		posts       = make([]bson.ObjectId, 0, count)
		users       = make([]bson.ObjectId, 0, count)
		postsResult = make([]models.Post, 0, count)
		p           models.Post
	)

	iter := c.Find("timelines", bson.M{"user_id": c.User.ID}).Sort("-time").Limit(count).Skip(offset).Iter()
	for iter.Next(&t) {
		cmts := models.GetCommentsForPost(t.Post, c.User, 5, c.Conn)
		if cmts != nil {
			comments[t.Post] = cmts
		}

		posts = append(posts, t.Post)
		users = append(users, t.PostUser)
	}

	iter.Close()

	udata := models.GetUsersData(users, c.User, c.Conn)

	if len(udata) == 0 {
		c.Error(500, CodeUnexpected, MsgUnexpected)
		return
	}

	iter = c.Find("posts", bson.M{"_id": bson.M{"$in": posts}}).Sort("-created").Iter()

	likes := models.GetLikesForPosts(posts, c.User.ID, c.Conn)

	for iter.Next(&p) {
		if c, ok := comments[p.ID]; ok {
			p.Comments = c
		}

		if l, ok := likes[p.ID]; ok {
			p.Liked = l
		}

		if u, ok := udata[p.ID]; ok {
			p.User = u
		}

		postsResult = append(postsResult, p)
	}

	c.Success(200, map[string]interface{}{
		"posts": postsResult,
		"count": len(postsResult),
	})
}
