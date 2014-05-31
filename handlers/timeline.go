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
	var (
		t           models.TimelineEntry
		comments    = make(map[bson.ObjectId][]models.Comment)
		posts       = make([]bson.ObjectId, 0, 25)
		users       = make([]bson.ObjectId, 0, 25)
		postsResult = make([]models.Post, 0, 25)
		p           models.Post
		newerThan   int64
	)

	newerThan, err := strconv.ParseInt(c.Form("newer_than"), 10, 64)
	if err != nil {
		newerThan = 0
	}

	olderThan, err := strconv.ParseInt(c.Form("older_than"), 10, 64)
	if err != nil {
		olderThan = 0
	}

	var timeConstraint bson.M
	if olderThan > 0 {
		timeConstraint = bson.M{"$lt": olderThan}
	} else {
		timeConstraint = bson.M{"$gt": newerThan}
	}

	iter := c.Find("timelines", bson.M{
		"user_id": c.User.ID,
		"time":    timeConstraint,
	}).
		Sort("-time").
		Limit(25).
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
