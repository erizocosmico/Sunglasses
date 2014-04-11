package tests

import (
	"github.com/garyburd/redigo/redis"
	. "github.com/smartystreets/goconvey/convey"
	"labix.org/v2/mgo/bson"
	"testing"
	. "github.com/mvader/mask/models"
	. "github.com/mvader/mask/services"
)

func newTaskService() *TaskService {
	config, err := NewConfig("../config.sample.json")
	if err != nil {
		panic(err)
	}

	ts, err := NewTaskService(config)
	if err != nil {
		panic(err)
	}

	return ts
}

func TestPushTask(t *testing.T) {
	ts := newTaskService()

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

			ID = ts.PushTask("follow_user", bson.NewObjectId())

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

func TestPushFail(t *testing.T) {
	ts := newTaskService()

	defer func() {
		ts.Do("DEL", "tasks")
		ts.Close()
	}()

	Convey("Pushing fails", t, func() {
		Convey("create_post fail", func() {
			p := &Post{ID: bson.NewObjectId(), UserID: bson.NewObjectId(), Created: 0}

			taskID := ts.PushTask("create_post", p)

			So(taskID.Hex(), ShouldNotEqual, "")

			ID := ts.PushFail("create_post", taskID)

			So(ID.Hex(), ShouldEqual, "")

			ID = ts.PushFail("create_post", taskID, bson.NewObjectId())

			So(ID.Hex(), ShouldNotEqual, "")

			ts.Do("DEL", "create_post:"+taskID.Hex())
			ts.Do("DEL", "create_post:"+taskID.Hex()+":fail")
			ts.Do("DEL", "task_op_fail:"+ID.Hex())
		})

		Convey("follow_user fail", func() {
			taskID := ts.PushTask("follow_user", bson.NewObjectId())

			So(taskID.Hex(), ShouldNotEqual, "")

			ID := ts.PushFail("follow_user", taskID)

			So(ID.Hex(), ShouldEqual, "")

			ID = ts.PushFail("follow_user", taskID, bson.NewObjectId())

			So(ID.Hex(), ShouldNotEqual, "")

			ts.Do("DEL", "follow_user:"+taskID.Hex())
			ts.Do("DEL", "follow_user:"+taskID.Hex()+":fail")
			ts.Do("DEL", "task_op_fail:"+ID.Hex())
		})

		Convey("create_comment fail", func() {
			taskID := ts.PushTask("create_comment", bson.NewObjectId(), bson.NewObjectId())

			So(taskID.Hex(), ShouldNotEqual, "")

			ID := ts.PushFail("create_comment", taskID)

			So(ID.Hex(), ShouldEqual, "")

			ID = ts.PushFail("create_comment", taskID, bson.NewObjectId())

			So(ID.Hex(), ShouldNotEqual, "")

			ts.Do("DEL", "create_comment:"+taskID.Hex())
			ts.Do("DEL", "create_comment:"+taskID.Hex()+":fail")
			ts.Do("DEL", "task_op_fail:"+ID.Hex())
		})

		Convey("delete_comment fail", func() {
			taskID := ts.PushTask("delete_comment", bson.NewObjectId(), bson.NewObjectId())

			So(taskID.Hex(), ShouldNotEqual, "")

			ID := ts.PushFail("delete_comment", taskID)

			So(ID.Hex(), ShouldEqual, "")

			ID = ts.PushFail("delete_comment", taskID, bson.NewObjectId())

			So(ID.Hex(), ShouldNotEqual, "")

			ts.Do("DEL", "delete_comment:"+taskID.Hex())
			ts.Do("DEL", "delete_comment:"+taskID.Hex()+":fail")
			ts.Do("DEL", "task_op_fail:"+ID.Hex())
		})
	})
}

func TestTaskDone(t *testing.T) {
	ts := newTaskService()

	defer func() {
		ts.Do("DEL", "tasks")
		ts.Close()
	}()

	Convey("Marking tasks as done", t, func() {
		ops := make([]bson.ObjectId, 0, 5)
		ID := ts.PushTask("follow_user", bson.NewObjectId())
		So(ID.Hex(), ShouldNotEqual, "")

		for i := 0; i < 5; i++ {
			op := ts.PushFail("follow_user", ID, bson.NewObjectId())
			So(op.Hex(), ShouldNotEqual, "")

			if op.Hex() != "" {
				ops = append(ops, op)
			}
		}

		testExistance := func(n, m int) {
			v, err := ts.Do("SCARD", "follow_user:"+ID.Hex()+":fail")
			count, err := redis.Int(v, err)
			if err != nil {
				panic(err)
			}

			So(count, ShouldEqual, n)

			for _, op := range ops {
				v, err := ts.Do("HLEN", "task_op_fail:"+op.Hex())
				count, err := redis.Int(v, err)
				if err != nil {
					panic(err)
				}

				So(count, ShouldEqual, m)
			}
		}

		testExistance(5, 1)

		err := ts.TaskDone("follow_user", ID)
		So(err, ShouldEqual, nil)

		testExistance(0, 0)

		ts.Do("DEL", "follow_user:"+ID.Hex())
	})
}
