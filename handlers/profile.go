package handlers

import (
	"github.com/go-martini/martini"
	. "github.com/mvader/mask/error"
	"github.com/mvader/mask/middleware"
	"github.com/mvader/mask/models"
	"labix.org/v2/mgo/bson"
	"strings"
)

// ShowUserProfile retrieves a list of the first 25 posts and the data of the requested user
func ShowUserProfile(c middleware.Context, params martini.Params) {
	// TODO untested
	var u models.User

	user := params["username"]
	if err := c.Find("users", bson.M{"username_lower": strings.ToLower(user)}).One(&u); err != nil {
		c.Error(404, CodeNotFound, MsgNotFound)
		return
	}

	hasAccess := false
	if c.User.ID.Hex() == u.ID.Hex() {
		hasAccess = true
	} else if models.Follows(c.User.ID, u.ID, c.Conn) {
		hasAccess = true
	} else if models.FollowedBy(c.User.ID, u.ID, c.Conn) {
		hasAccess = true
	}

	posts := getPostsFromUser(c, u.ID, 25, 0)
	c.Success(200, map[string]interface{}{
		"user":        models.UserForDisplay(u, hasAccess, true),
		"posts":       posts,
		"posts_count": len(posts),
	})
}

// GetUserPosts retrieves a list of posts from an user
func GetUserPosts(c middleware.Context) {
	// TODO untested
	userID := c.Form("user_id")
	if userID == "" || !bson.IsObjectIdHex(userID) {
		c.Error(400, CodeInvalidData, MsgInvalidData)
		return
	}

	count, offset := c.ListCountParams()
	posts := getPostsFromUser(c, bson.ObjectIdHex(userID), count, offset)
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
func getPostsFromUser(c middleware.Context, user bson.ObjectId, limit, offset int) []models.Post {
	var (
		posts = make([]models.Post, 0, limit)
		ids   = make([]bson.ObjectId, 0, limit)
		p     models.Post
	)

	iter := c.Find("posts", bson.M{"user_id": user}).Sort("-created").Skip(offset).Iter()
	for len(posts) < offset && iter.Next(&p) {
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
	if likes != nil {
		for i, v := range posts {
			if l, ok := likes[v.ID]; ok {
				posts[i].Liked = l
				posts[i].User = udata[v.ID]
			}
		}
	}

	return posts
}
