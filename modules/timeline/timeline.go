package timeline

import (
	"github.com/mvader/mask/middleware"
	"github.com/mvader/mask/models"
	"github.com/mvader/mask/services"
	"labix.org/v2/mgo/bson"
)

// PropagatePostOnCreation propagates the post to all timelines when a new post is created.
func PropagatePostOnCreation(c middleware.Context, post *models.Post) {
	if !c.Config.Debug {
		t := models.TimelineEntry{
			Post:     post.ID,
			PostUser: post.UserID,
			Liked:    false,
			Time:     post.Created,
		}

		ID := c.Tasks.PushTask("create_post", post)

		c.AsyncQuery(func(conn *services.Connection) {
			var f models.Follow
			allCompleted := true
			iter := conn.Db.C("follows").Find(bson.M{"user_to": post.UserID}).Iter()
			for iter.Next(&f) {
				var u models.User
				if err := conn.Db.C("users").FindId(f.From).One(&u); err == nil {
					if post.CanBeAccessedBy(&u, conn) {
						t.ID = bson.NewObjectId()
						t.User = u.ID

						if _, err := conn.Db.C("timelines").UpsertId(t.ID, t); err != nil {
							allCompleted = false
							c.Tasks.PushFail("create_post", ID, u.ID)
						}
					}
				} else {
					allCompleted = false
					break
				}
			}
            
            // Propagate the post on the users timeline
			t.ID = bson.NewObjectId()
			t.User = c.User.ID

			if _, err := conn.Db.C("timelines").UpsertId(t.ID, t); err != nil {
				allCompleted = false
				c.Tasks.PushFail("create_post", ID, c.User.ID)
			}

			if allCompleted {
				c.Tasks.TaskDone("create_post", ID)
			}

			iter.Close()
		})
	}
}

// PropagateSinglePostOnCreation propagates a post to a certain timeline because it failed before
func PropagateSinglePostOnCreation(conn *services.Connection, ts *services.TaskService, p *models.Post, u, task, op bson.ObjectId) error {
	t := models.TimelineEntry{
		ID:       bson.NewObjectId(),
		Post:     p.ID,
		PostUser: p.UserID,
		Liked:    false,
		Time:     p.Created,
		User:     u,
	}

	if _, err := conn.Db.C("timelines").UpsertId(t.ID, t); err != nil {
		return err
	}

	if err := ts.FailedOpSolved("create_post", task, op); err != nil {
		return err
	}

	return nil
}

// PropagatePostOnPrivacyChange propagates the privacy changes across all timelines, removing and adding the post
// according to the new privacy settings.
func PropagatePostOnPrivacyChange(c middleware.Context, post *models.Post) {
	if !c.Config.Debug {
		PropagatePostsOnDeletion(c, post.ID)
		PropagatePostOnCreation(c, post)
	}
}

// PropagatePostsOnUserFollow propagates the posts to the timeline when a new user is followed
func PropagatePostsOnUserFollow(c middleware.Context, userID bson.ObjectId) {
	if !c.Config.Debug {

		ID := c.Tasks.PushTask("follow_user", c.User.ID)

		c.AsyncQuery(func(conn *services.Connection) {
			var p models.Post

			allCompleted := true
			iter := conn.Db.C("posts").Find(bson.M{"user_id": userID}).Iter()
			for iter.Next(&p) {
				t := models.TimelineEntry{
					User:     c.User.ID,
					Post:     p.ID,
					PostUser: p.UserID,
					Liked:    false,
					Time:     p.Created,
				}

				if (&p).CanBeAccessedBy(c.User, conn) {
					t.ID = bson.NewObjectId()
					t.User = c.User.ID

					if _, err := conn.Db.C("timelines").UpsertId(t.ID, t); err != nil {
						allCompleted = false
						c.Tasks.PushFail("follow_user", ID, p.ID.Hex())
					}
				}
			}

			if allCompleted {
				c.Tasks.TaskDone("follow_user", ID)
			}

			iter.Close()
		})
	}
}

// PropagateSinglePostOnUserFollow propagates a single post to the timeline when a new user is followed
func PropagateSinglePostOnUserFollow(conn *services.Connection, ts *services.TaskService, user, post, task, op bson.ObjectId) error {
	var (
		p models.Post
		u models.User
	)

	err := conn.Db.C("posts").FindId(post).One(&p)
	if err != nil {
		return err
	}

	err = conn.Db.C("users").FindId(user).One(&u)
	if err != nil {
		return err
	}

	t := models.TimelineEntry{
		User:     user,
		Post:     p.ID,
		PostUser: p.UserID,
		Liked:    false,
		Time:     p.Created,
	}

	if (&p).CanBeAccessedBy(&u, conn) {
		t.ID = bson.NewObjectId()
		t.User = user

		if _, err := conn.Db.C("timelines").UpsertId(t.ID, t); err != nil {
			return err
		}

		if err := ts.FailedOpSolved("follow_user", task, op); err != nil {
			return err
		}
	}

	return nil
}

// PropagatePostsOnUserFollow propagates the posts to the timeline when a new user is followed
func PropagatePostsOnUserUnfollow(c middleware.Context, userID bson.ObjectId) {
	if !c.Config.Debug {
		ID := c.Tasks.PushTask("user_unfollow", userID.Hex(), c.User.ID.Hex())

		c.AsyncQuery(func(conn *services.Connection) {
			_, err := conn.Db.C("timelines").RemoveAll(bson.M{"user_id": c.User.ID, "post_user_id": userID})
			if err == nil {
				c.Tasks.TaskDone("user_unfollow", ID)
			}
		})
	}
}

// PropagatePostsOnDeletion erases a deleted post from all timelines
func PropagatePostsOnDeletion(c middleware.Context, postID bson.ObjectId) {
	if !c.Config.Debug {
		ID := c.Tasks.PushTask("post_delete", postID.Hex())

		c.AsyncQuery(func(conn *services.Connection) {
			if _, err := conn.Db.C("timelines").RemoveAll(bson.M{"post_id": postID}); err == nil {
				c.Tasks.TaskDone("post_delete", ID)
			}
		})
	}
}

// PropagatePostOnLike sets the new like value for the user's timeline
func PropagatePostOnLike(c middleware.Context, postID bson.ObjectId, liked bool) {
	if !c.Config.Debug {
		ID := c.Tasks.PushTask("post_like", c.User.ID.Hex(), postID.Hex(), liked)
		c.AsyncQuery(func(conn *services.Connection) {
			var t models.TimelineEntry

			err := conn.Db.C("timelines").Find(bson.M{"post_id": postID, "user_id": c.User.ID}).One(&t)
			if err == nil {
				t.Liked = liked
				if _, err := conn.Db.C("timelines").UpsertId(t.ID, t); err == nil {
					c.Tasks.TaskDone("post_like", ID)
				}
			}
		})
	}
}

// PropagatePostOnNewComment adds a reference to the new comment on all user timelines
func PropagatePostOnNewComment(c middleware.Context, postID, commentID bson.ObjectId) {
	if !c.Config.Debug {
		ID := c.Tasks.PushTask("create_comment", commentID.Hex())

		c.AsyncQuery(func(conn *services.Connection) {
			var t models.TimelineEntry

			allCompleted := true
			iter := conn.Db.C("timelines").Find(bson.M{"post_id": postID}).Iter()
			for iter.Next(&t) {
				t.Comments = append(t.Comments, commentID)

				if _, err := conn.Db.C("timelines").UpsertId(t.ID, t); err != nil {
					allCompleted = false
					c.Tasks.PushFail("create_comment", ID, t.ID.Hex())
				}
			}

			if allCompleted {
				c.Tasks.TaskDone("create_comment", ID)
			}

			iter.Close()
		})
	}
}

// PropagateSinglePostOnNewComment adds a reference to the new comment on a timeline
func PropagateSinglePostOnNewComment(conn *services.Connection, ts *services.TaskService, c, tID, task, op bson.ObjectId) error {
	var t models.TimelineEntry
	if err := conn.Db.C("timelines").FindId(tID).One(&t); err != nil {
		return err
	}

	t.Comments = append(t.Comments, c)

	if _, err := conn.Db.C("timelines").UpsertId(t.ID, t); err != nil {
		return err
	}

	if err := ts.FailedOpSolved("create_comment", task, op); err != nil {
		return err
	}

	return nil
}

// PropagatePostOnNewComment deletes a reference to the new comment on all user timelines
func PropagatePostOnCommentDeleted(c middleware.Context, postID, commentID bson.ObjectId) {
	if !c.Config.Debug {
		ID := c.Tasks.PushTask("delete_comment", commentID.Hex())

		c.AsyncQuery(func(conn *services.Connection) {
			var t models.TimelineEntry

			allCompleted := true
			iter := conn.Db.C("timelines").Find(bson.M{"post_id": postID}).Iter()
			for iter.Next(&t) {
				cmts := make([]bson.ObjectId, 0, len(t.Comments)-1)
				for _, v := range t.Comments {
					if v != commentID {
						cmts = append(cmts, v)
					}
				}
				t.Comments = cmts

				if _, err := conn.Db.C("timelines").UpsertId(t.ID, t); err != nil {
					allCompleted = false
					c.Tasks.PushFail("delete_comment", ID, t.ID.Hex())
				}
			}

			if allCompleted {
				c.Tasks.TaskDone("delete_comment", ID)
			}

			iter.Close()
		})
	}
}

// PropagateSinglePostOnCommentDeleted removes a reference of the comment on a timeline
func PropagateSinglePostOnCommentDeleted(conn *services.Connection, ts *services.TaskService, c, tID, task, op bson.ObjectId) error {
	var t models.TimelineEntry

	if err := conn.Db.C("timelines").FindId(tID).One(&t); err != nil {
		return err
	}

	cmts := make([]bson.ObjectId, 0, len(t.Comments)-1)
	for _, v := range t.Comments {
		if v != c {
			cmts = append(cmts, v)
		}
	}
	t.Comments = cmts

	if _, err := conn.Db.C("timelines").UpsertId(t.ID, t); err != nil {
		return err
	}

	if err := ts.FailedOpSolved("delete_comment", task, op); err != nil {
		return err
	}

	return nil
}

// PropagatePostOnUserDeleted erases all posts owned by the deleted user from all timelines
func PropagatePostOnUserDeleted(c middleware.Context, userID bson.ObjectId) {
	if !c.Config.Debug {
		ID := c.Tasks.PushTask("delete_user", userID.Hex())

		c.AsyncQuery(func(conn *services.Connection) {
			if _, err := conn.Db.C("timelines").RemoveAll(bson.M{"post_user_id": userID}); err == nil {
				c.Tasks.TaskDone("delete_user", ID)
			}
		})
	}
}
