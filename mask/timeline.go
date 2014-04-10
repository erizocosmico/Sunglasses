package mask

import "labix.org/v2/mgo/bson"

type TimelineEntry struct {
	ID       bson.ObjectId   `bson:"_id"`
	User     bson.ObjectId   `bson:"user_id"`
	Post     bson.ObjectId   `bson:"post_id"`
	PostUser bson.ObjectId   `bson:"post_user_id"`
	Liked    bool            `bson:"liked"`
	Comments []bson.ObjectId `bson:"comments"`
	Time     float64         `bson:"time"`
}

// PropagatePostOnCreation propagates the post to all timelines when a new post is created.
func PropagatePostOnCreation(c Context, post *Post) {
	if !c.Config.Debug {
		t := TimelineEntry{
			Post:     post.ID,
			PostUser: post.UserID,
			Liked:    false,
			Time:     post.Created,
		}

		ID := c.Conn.PushTask("create_post", post)

		c.AsyncQuery(func(conn *Connection) {
			var f Follow
			allCompleted := true
			iter := conn.Db.C("follows").Find(bson.M{"user_to": c.User.ID}).Iter()
			for iter.Next(&f) {
				var u User
				if err := conn.Db.C("users").FindId(f.From).One(&u); err == nil {
					if post.CanBeAccessedBy(&u, conn) {
						t.ID = bson.NewObjectId()
						t.User = u.ID

						if _, err := conn.Db.C("timelines").UpsertId(t.ID, t); err != nil {
							allCompleted = false
							c.Conn.PushFail("create_post", ID, u.ID)
						}
					}
				} else {
					allCompleted = false
					break
				}
			}

			if allCompleted {
				c.Conn.TaskDone("create_post", ID)
			}

			iter.Close()
		})
	}
}

// PropagatePostOnPrivacyChange propagates the privacy changes across all timelines, removing and adding the post
// according to the new privacy settings.
func PropagatePostOnPrivacyChange(c Context, post *Post) {
	// TODO: Implement, requires handler for changing posts privacy
}

// PropagatePostsOnUserFollow propagates the posts to the timeline when a new user is followed
func PropagatePostsOnUserFollow(c Context, userID bson.ObjectId) {
	if !c.Config.Debug {

		ID := c.Conn.PushTask("follow_user", userID, c.User.ID)

		c.AsyncQuery(func(conn *Connection) {
			var p Post

			allCompleted := true
			iter := conn.Db.C("posts").Find(bson.M{"user_id": userID}).Iter()
			for iter.Next(&p) {
				t := TimelineEntry{
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
						c.Conn.PushFail("follow_user", ID, p.ID.Hex())
					}
				}
			}

			if allCompleted {
				c.Conn.TaskDone("follow_user", ID)
			}

			iter.Close()
		})
	}
}

// PropagatePostsOnUserFollow propagates the posts to the timeline when a new user is followed
func PropagatePostsOnUserUnfollow(c Context, userID bson.ObjectId) {
	if !c.Config.Debug {
		ID := c.Conn.PushTask("user_unfollow", userID.Hex(), c.User.ID.Hex())

		c.AsyncQuery(func(conn *Connection) {
			_, err := conn.Db.C("timelines").RemoveAll(bson.M{"user_id": c.User.ID, "post_user_id": userID})
			if err == nil {
				c.Conn.TaskDone("user_unfollow", ID)
			}
		})
	}
}

// PropagatePostsOnDeletion erases a deleted post from all timelines
func PropagatePostsOnDeletion(c Context, postID bson.ObjectId) {
	if !c.Config.Debug {
		ID := c.Conn.PushTask("post_delete", postID.Hex())

		c.AsyncQuery(func(conn *Connection) {
			if _, err := conn.Db.C("timelines").RemoveAll(bson.M{"post_id": postID}); err == nil {
				c.Conn.TaskDone("post_delete", ID)
			}
		})
	}
}

// PropagatePostOnLike sets the new like value for the user's timeline
func PropagatePostOnLike(c Context, postID bson.ObjectId, liked bool) {
	if !c.Config.Debug {

		ID := c.Conn.PushTask("post_like", c.User.ID.Hex(), postID.Hex(), liked)
		c.AsyncQuery(func(conn *Connection) {
			var t TimelineEntry

			err := conn.Db.C("timelines").Find(bson.M{"post_id": postID, "user_id": c.User.ID}).One(&t)
			if err == nil {
				t.Liked = liked
				if _, err := conn.Db.C("timelines").UpsertId(t.ID, t); err == nil {
					c.Conn.TaskDone("post_like", ID)
				}
			}
		})
	}
}

// PropagatePostOnNewComment adds a reference to the new comment on all user timelines
func PropagatePostOnNewComment(c Context, postID, commentID bson.ObjectId) {
	if !c.Config.Debug {
		ID := c.Conn.PushTask("create_comment", postID.Hex(), commentID.Hex())

		c.AsyncQuery(func(conn *Connection) {
			var t TimelineEntry

			allCompleted := true
			iter := conn.Db.C("timelines").Find(bson.M{"post_id": postID}).Iter()
			for iter.Next(&t) {
				t.Comments = append(t.Comments, commentID)

				if _, err := conn.Db.C("timelines").UpsertId(t.ID, t); err != nil {
					allCompleted = false
					c.Conn.PushFail("create_comment", ID, t.ID.Hex())
				}
			}

			if allCompleted {
				c.Conn.TaskDone("create_comment", ID)
			}

			iter.Close()
		})
	}
}

// PropagatePostOnNewComment deletes a reference to the new comment on all user timelines
func PropagatePostOnCommentDeleted(c Context, postID, commentID bson.ObjectId) {
	if !c.Config.Debug {
		ID := c.Conn.PushTask("delete_comment", postID.Hex(), commentID.Hex())

		c.AsyncQuery(func(conn *Connection) {
			var t TimelineEntry

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
					c.Conn.PushFail("delete_comment", ID, t.ID.Hex())
				}
			}

			if allCompleted {
				c.Conn.TaskDone("delete_comment", ID)
			}

			iter.Close()
		})
	}
}

// PropagatePostOnUserDeleted erases all posts owned by the deleted user from all timelines
func PropagatePostOnUserDeleted(c Context, userID bson.ObjectId) {
	if !c.Config.Debug {
		ID := c.Conn.PushTask("delete_user", userID.Hex())

		c.AsyncQuery(func(conn *Connection) {
			if _, err := conn.Db.C("timelines").RemoveAll(bson.M{"post_user_id": userID}); err == nil {
				c.Conn.TaskDone("delete_user", ID)
			}
		})
	}
}
