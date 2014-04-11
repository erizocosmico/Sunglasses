package mask

import (
	. "github.com/smartystreets/goconvey/convey"
	"labix.org/v2/mgo/bson"
	"testing"
)

func TestPushTask(t *testing.T) {
	config, err := NewConfig("../config.sample.json")
	if err != nil {
		panic(err)
	}

	ts, err := NewTaskService(config)
	if err != nil {
		panic(err)
	}

	defer func() {
		ts.Do("DEL", "tasks")
		ts.Close()
	}()

	Convey("Pushing tasks", t, func() {
		Convey("create_post task", func() {
			p := &Post{ID: bson.NewObjectId(), UserID: bson.NewObjectId(), Created: 0}
			ID := ts.PushTask("create_post")

			So(ID.Hex(), ShouldEqual, "")

			ID = ts.PushTask("create_post", p)

			So(ID.Hex(), ShouldNotEqual, "")

			ts.Do("DEL", "create_post:"+ID.Hex())
		})

		Convey("follow_user task", func() {
			ID := ts.PushTask("follow_user")

			So(ID.Hex(), ShouldEqual, "")

			ID = ts.PushTask("follow_user", bson.NewObjectId(), bson.NewObjectId())

			So(ID.Hex(), ShouldNotEqual, "")

			ts.Do("DEL", "follow_user:"+ID.Hex())
		})

		Convey("unfollow_user task", func() {
			ID := ts.PushTask("unfollow_user")

			So(ID.Hex(), ShouldEqual, "")

			ID = ts.PushTask("unfollow_user", bson.NewObjectId(), bson.NewObjectId())

			So(ID.Hex(), ShouldNotEqual, "")

			ts.Do("DEL", "unfollow_user:"+ID.Hex())
		})

		Convey("post_delete task", func() {
			ID := ts.PushTask("post_delete")

			So(ID.Hex(), ShouldEqual, "")

			ID = ts.PushTask("post_delete", bson.NewObjectId())

			So(ID.Hex(), ShouldNotEqual, "")

			ts.Do("DEL", "post_delete:"+ID.Hex())
		})

		Convey("post_like task", func() {
			ID := ts.PushTask("post_like")

			So(ID.Hex(), ShouldEqual, "")

			ID = ts.PushTask("post_like", bson.NewObjectId(), bson.NewObjectId(), true)

			So(ID.Hex(), ShouldNotEqual, "")

			ts.Do("DEL", "post_like:"+ID.Hex())
		})

		Convey("create_comment task", func() {
			ID := ts.PushTask("create_comment")

			So(ID.Hex(), ShouldEqual, "")

			ID = ts.PushTask("create_comment", bson.NewObjectId(), bson.NewObjectId())

			So(ID.Hex(), ShouldNotEqual, "")

			ts.Do("DEL", "create_comment:"+ID.Hex())
		})

		Convey("delete_comment task", func() {
			ID := ts.PushTask("delete_comment")

			So(ID.Hex(), ShouldEqual, "")

			ID = ts.PushTask("delete_comment", bson.NewObjectId(), bson.NewObjectId())

			So(ID.Hex(), ShouldNotEqual, "")

			ts.Do("DEL", "delete_comment:"+ID.Hex())
		})

		Convey("delete_user task", func() {
			ID := ts.PushTask("delete_user")

			So(ID.Hex(), ShouldEqual, "")

			ID = ts.PushTask("delete_user", bson.NewObjectId())

			So(ID.Hex(), ShouldNotEqual, "")

			ts.Do("DEL", "delete_user:"+ID.Hex())
		})
	})
}
