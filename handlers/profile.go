package handlers

import (
	"github.com/go-martini/martini"
	. "github.com/mvader/sunglasses/error"
	"github.com/mvader/sunglasses/middleware"
	"github.com/mvader/sunglasses/models"
	"labix.org/v2/mgo/bson"
	"strconv"
	"strings"
)

// ShowUserProfile retrieves a list of the first 25 posts and the data of the requested user
func ShowUserProfile(c middleware.Context, params martini.Params) {
	var u models.User

	user := params["username"]
	if err := c.Find("users", bson.M{"username_lower": strings.ToLower(user)}).One(&u); err != nil {
		c.Error(404, CodeNotFound, MsgNotFound)
		return
	}

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

	users := models.GetUsersData([]bson.ObjectId{u.ID}, c.User, c.Conn)
	if len(users) != 1 {
		c.Error(404, CodeNotFound, MsgNotFound)
		return
	}

	if !users[u.ID]["protected"].(bool) {
		followers, err := c.Count("follows", bson.M{"user_to": u.ID})
		if err != nil {
			followers = 0
		}

		following, err := c.Count("follows", bson.M{"user_from": u.ID})
		if err != nil {
			following = 0
		}

		users[u.ID]["followers"] = followers
		users[u.ID]["following"] = following
	}

	numPosts, err := c.Count("posts", bson.M{"user_id": u.ID})
	if err != nil {
		numPosts = 0
	}

	users[u.ID]["num_posts"] = numPosts

	posts := getPostsFromUser(c, u.ID, timeConstraint)
	c.Success(200, map[string]interface{}{
		"user":        users[u.ID],
		"posts":       posts,
		"posts_count": len(posts),
	})
}

// GetUserPosts retrieves a list of posts from an user
func GetUserPosts(c middleware.Context) {
	userID := c.Form("user_id")
	if userID == "" || !bson.IsObjectIdHex(userID) {
		c.Error(400, CodeInvalidData, MsgInvalidData)
		return
	}

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

	posts := getPostsFromUser(c, bson.ObjectIdHex(userID), timeConstraint)
	if posts == nil {
		c.Error(404, CodeNotFound, MsgNotFound)
		return
	}

	c.Success(200, map[string]interface{}{
		"posts": posts,
		"count": len(posts),
	})
}

// The main reason to not retrieve the posts from the user's generated timeline is that
// the timeline for the user may have not been processed yet when the user browses the profile
func getPostsFromUser(c middleware.Context, user bson.ObjectId, constraint bson.M) []models.Post {
	var (
		posts = make([]models.Post, 0, 25)
		ids   = make([]bson.ObjectId, 0, 25)
		p     models.Post
	)

	iter := c.Find("posts", bson.M{"user_id": user, "created": constraint}).Sort("-created").Iter()
	for len(posts) < 25 && iter.Next(&p) {
		if (&p).CanBeAccessedBy(c.User, c.Conn) {
			comments := models.GetCommentsForPost(p.ID, c.User, 5, c.Conn)
			if comments != nil {
				p.Comments = comments
			}

			posts = append(posts, p)
			ids = append(ids, p.ID)
		}
	}

	iter.Close()

	udata := models.GetUsersData([]bson.ObjectId{user}, c.User, c.Conn)

	if len(udata) == 0 {
		return nil
	}

	likes := models.GetLikesForPosts(ids, c.User.ID, c.Conn)

	for i, v := range posts {
		posts[i].User = udata[v.UserID]
		if likes != nil {
			if l, ok := likes[v.ID]; ok {
				posts[i].Liked = l
			}
		}
	}

	return posts
}
