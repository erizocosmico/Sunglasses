package handlers

import (
	. "github.com/mvader/sunglasses/error"
	"github.com/mvader/sunglasses/middleware"
	"github.com/mvader/sunglasses/models"
	"labix.org/v2/mgo/bson"
	"strconv"
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
		newerThan   float64
	)

	n, err := strconv.ParseInt(c.Form("newer_than"), 10, 64)
	if err != nil {
		newerThan = 0
	} else {
		newerThan = float64(n)
	}

	iter := c.Find("timelines", bson.M{
		"user_id": c.User.ID,
		"time":    bson.M{"$gt": newerThan},
	}).
		Sort("-time").
		Limit(count).
		Skip(offset).
		Iter()

	for iter.Next(&t) {
		cmts := models.GetCommentsForPost(t.Post, c.User, 5, c.Conn)
		if cmts != nil {
			comments[t.Post] = cmts
		}

		posts = append(posts, t.Post)
		users = append(users, t.PostUser)
	}

	iter.Close()

	// Return empty response if there are no posts
	if len(posts) == 0 {
		c.Success(200, map[string]interface{}{
			"posts": []string{},
			"count": 0,
		})
		return
	}

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

		if u, ok := udata[p.UserID]; ok {
			p.User = u
		}

		postsResult = append(postsResult, p)
	}

	c.Success(200, map[string]interface{}{
		"posts": postsResult,
		"count": len(postsResult),
	})
}
