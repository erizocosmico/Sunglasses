package mask

import (
	"github.com/garyburd/redigo/redis"
	"labix.org/v2/mgo/bson"
	"os"
)

var (
	empty bson.ObjectId
)

type TaskService struct {
	redis.Conn
}

// NewTaskSercice initializes the task service
func NewTaskService(config *Config) (*TaskService, error) {
	var (
		conn redis.Conn
		err  error
	)

	if os.Getenv("WERCKER_REDIS_HOST") != "" {
		config.RedisAddress = os.Getenv("WERCKER_REDIS_HOST")
	}

	if conn, err = redis.Dial("tcp", config.RedisAddress); err != nil {
		return nil, err
	}

	return &TaskService{conn}, nil
}

// Do performs a Redis commant
func (ts *TaskService) Do(commandName string, args ...interface{}) (reply interface{}, err error) {
	return ts.Conn.Do(commandName, args...)
}

// Close closes mongodb open connection
func (ts *TaskService) Close() {
	ts.Conn.Close()
}

// PushTask pushes a task to Redis
func (ts *TaskService) PushTask(task string, args ...interface{}) bson.ObjectId {
	var (
		err    error
		taskID = bson.NewObjectId()
	)

	_, err = ts.Do("SADD", "tasks", taskID.Hex())
	if err != nil {
		return empty
	}

	taskName := task + ":" + taskID.Hex()

	switch task {
	case "create_post":
		if len(args) < 1 {
			return empty
		}

		p := args[0].(*Post)

		_, err = ts.Do("HMSET", taskName, "post_id", p.ID.Hex(), "post_user", p.UserID.Hex(), "created", p.Created, "has_children", true)
		break
	case "follow_user":
		if len(args) < 1 {
			return empty
		}

		_, err = ts.Do("HMSET", taskName, "user", args[0], "has_children", true)
		break
	case "unfollow_user":
		if len(args) < 2 {
			return empty
		}

		_, err = ts.Do("HMSET", taskName, "user_unfollowed", args[0], "user", args[0], "has_children", false)
		break
	case "post_delete":
		if len(args) < 1 {
			return empty
		}

		_, err = ts.Do("HMSET", taskName, "post_id", args[0], "has_children", false)
		break
	case "post_like":
		if len(args) < 3 {
			return empty
		}

		_, err = ts.Do("HMSET", taskName, "user_id", args[0], "post_id", args[1], "liked", args[2], "has_children", false)
		break
	case "create_comment":
		if len(args) < 1 {
			return empty
		}

		_, err = ts.Do("HMSET", taskName, "comment_id", args[0], "has_children", true)
		break
	case "delete_comment":
		if len(args) < 1 {
			return empty
		}

		_, err = ts.Do("HMSET", taskName, "comment_id", args[0], "has_children", true)
		break
	case "delete_user":
		if len(args) < 1 {
			return empty
		}

		_, err = ts.Do("HMSET", taskName, "user_id", args[0], "has_children", false)
		break
	default:
		return empty
	}

	if err != nil {
		return empty
	}

	return taskID
}

// PushFail pushes a failed operation to Redis
func (ts *TaskService) PushFail(task string, taskID bson.ObjectId, args ...interface{}) bson.ObjectId {
	var (
		name string
		err  error
	)

	if taskID.Hex() == "" {
		return empty
	}

	ID := bson.NewObjectId()
	_, err = ts.Do("SADD", task+":"+taskID.Hex()+":fail", ID.Hex())
	if err != nil {
		return empty
	}

	name = "task_op_fail:" + ID.Hex()

	if len(args) < 1 {
		return empty
	}

	switch task {
	case "create_post":
		_, err = ts.Do("HMSET", name, "user", args[0])
		break
	case "follow_user":
		_, err = ts.Do("HMSET", name, "post", args[0])
		break
	case "create_comment":
		_, err = ts.Do("HMSET", name, "timeline", args[0])
		break
	case "delete_comment":
		_, err = ts.Do("HMSET", name, "timeline", args[0])
		break
	}

	if err != nil {
		return empty
	}

	return ID
}

// TaskDone clears all the data for the given task
func (ts *TaskService) TaskDone(task string, taskID bson.ObjectId) error {
	var err error
	taskName := task + ":" + taskID.Hex()

	if task == "create_post" || task == "follow_user" || task == "create_comment" || task == "delete_comment" {
		v, err := ts.Do("SMEMBERS", taskName+":fail")
		keys, err := redis.Strings(v, err)
		if err != nil {
			return err
		}

		for _, k := range keys {
			_, err = ts.Do("DEL", "task_op_fail:"+k)
			if err != nil {
				return err
			}
		}

		_, err = ts.Do("DEL", taskName+":fail")
		if err != nil {
			return err
		}
	}

	_, err = ts.Do("DEL", taskName)
	if err != nil {
		return err
	}

	_, err = ts.Do("SREM", "tasks", taskID.Hex())
	if err != nil {
		return err
	}

	return nil
}

// FailedOpSolved marks a previously failed operation as solved and erases its data
func (ts *TaskService) FailedOpSolved(task string, taskID, opID bson.ObjectId) error {
	if _, err := ts.Do("DEL", "task_op_fail:"+opID.Hex()); err != nil {
		return err
	}

	if _, err := ts.Do("SREM", task+":"+taskID.Hex()+":fail"); err != nil {
		return err
	}

	v, err := ts.Do("SMEMBERS", task+":"+taskID.Hex()+":fail")
	count, err := redis.Int(v, err)
	if err != nil {
		return err
	}

	if count == 0 {
		if err := ts.TaskDone(task, taskID); err != nil {
			return err
		}
	}

	return nil
}

// TaskResolver is a process that takes care of the failed operations pushed to redis.
// The resolver tries to run again the task. If the task can't be completed by the resolver it will get discarded.
func (ts *TaskService) TaskResolver(conn *Connection) {
	// TODO implement
}
