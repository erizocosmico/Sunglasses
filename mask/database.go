package mask

import (
	"github.com/garyburd/redigo/redis"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"os"
)

// Connection represents the database session
type Connection struct {
	Session *mgo.Session
	Db      *mgo.Database
	Redis   redis.Conn
}

// NewDatabaseConn initializes the database connection
func NewDatabaseConn(config *Config) (*Connection, error) {
	conn := new(Connection)
	var err error

	if os.Getenv("WERCKER_MONGODB_HOST") != "" {
		config.DatabaseUrl = os.Getenv("WERCKER_MONGODB_HOST")
	}

	if conn.Session, err = mgo.Dial(config.DatabaseUrl); err != nil {
		return nil, err
	}
	conn.Session.SetMode(mgo.Strong, true)
	conn.Db = conn.Session.DB(config.DatabaseName)

	if config.Debug {
		if err := createIndexes(conn); err != nil {
			return nil, err
		}
	}

	if os.Getenv("WERCKER_REDIS_HOST") != "" {
		config.RedisAddress = os.Getenv("WERCKER_REDIS_HOST")
	}

	if conn.Redis, err = redis.Dial("tcp", config.RedisAddress); err != nil {
		conn.Session.Close()
		return nil, err
	}

	return conn, nil
}

// Save inserts an item or updates it if it has already been created
func (c *Connection) Save(collection string, ID bson.ObjectId, item interface{}) error {
	if _, err := c.Db.C(collection).UpsertId(ID, item); err != nil {
		return err
	}

	return nil
}

// Remove removes an item with the specified id on the given collection
func (c *Connection) Remove(collection string, ID bson.ObjectId) error {
	if err := c.Db.C(collection).RemoveId(ID); err != nil {
		return err
	}

	return nil
}

// Close closes both redis and mongodb open connections
func (c *Connection) Close() {
	c.Session.Close()
	c.Redis.Close()
}

func createIndexes(conn *Connection) error {
	indexes := map[string][]string{
		"posts":         []string{"user_id"},
		"albums":        []string{"user_id"},
		"notifications": []string{"user_id"},
		"tokens":        []string{"user_id", "hash"},
		"requests":      []string{"user_to", "user_from"},
		"follows":       []string{"user_to", "user_from"},
		"reports":       []string{"user_id", "post_id"},
		"blocks":        []string{"user_to", "user_from"},
		"likes":         []string{"user_id", "post_id"},
		"comments":      []string{"user_id", "post_id"},
	}

	for col, colIndexes := range indexes {
		for _, index := range colIndexes {
			if err := conn.Db.C(col).EnsureIndexKey(index); err != nil {
				return err
			}
		}
	}

	return nil
}

// PushTask pushes a task to Redis
func (c *Connection) PushTask(task string, args ...interface{}) bson.ObjectId {
	taskID := bson.NewObjectId()
	var empty bson.ObjectId

	c.Redis.Do("SADD", "tasks", taskID.Hex())

	taskName := task + ":" + taskID.Hex()

	switch task {
	case "create_post":
		if len(args) < 1 {
			return empty
		}

		p := args[0].(*Post)

		c.Redis.Do("HMSET", taskName, "post_id", p.ID.Hex(), "post_user", p.UserID.Hex(), "created", p.Created, "has_children", true)
		break
	case "follow_user":
		if len(args) < 2 {
			return empty
		}

		c.Redis.Do("HMSET", taskName, "user_followed", args[0], "user", args[1], "has_children", true)
		break
	case "unfollow_user":
		if len(args) < 2 {
			return empty
		}

		c.Redis.Do("HMSET", taskName, "user_unfollowed", args[0], "user", args[0], "has_children", false)
		break
	case "post_delete":
		if len(args) < 1 {
			return empty
		}

		c.Redis.Do("HMSET", taskName, "post_id", args[0], "has_children", false)
		break
	case "post_like":
		if len(args) < 3 {
			return empty
		}

		c.Redis.Do("HMSET", taskName, "user_id", args[0], "post_id", args[1], "liked", args[2], "has_children", false)
		break
	case "create_comment":
		if len(args) < 2 {
			return empty
		}

		c.Redis.Do("HMSET", taskName, "post_id", args[0], "comment_id", args[1], "has_children", true)
		break
	case "delete_comment":
		if len(args) < 2 {
			return empty
		}

		c.Redis.Do("HMSET", taskName, "post_id", args[0], "comment_id", args[1], "has_children", true)
		break
	case "delete_user":
		if len(args) < 1 {
			return empty
		}

		c.Redis.Do("HMSET", taskName, "user_id", "has_children", false)
		break
	default:
		return empty
	}

	return taskID
}

// PushFail pushes a failed operation to Redis
func (c *Connection) PushFail(task string, taskID bson.ObjectId, args ...interface{}) {
	var name string

	if taskID.Hex() == "" {
		return
	}

	ID := bson.NewObjectId()
	c.Redis.Do("SADD", task+":"+taskID.Hex()+":fail", ID.Hex())
	name = "task_op_fail:" + ID.Hex()

	if len(args) < 1 {
		return
	}

	switch task {
	case "create_post":
		c.Redis.Do("HMSET", name, "user", args[0])
		break
	case "follow_user":
		c.Redis.Do("HMSET", name, "post", args[0])
		break
	case "create_comment":
		c.Redis.Do("HMSET", name, "timeline", args[0])
		break
	case "delete_comment":
		c.Redis.Do("HMSET", name, "timeline", args[0])
		break
	}
}

// TaskDone clears all the data for the given task
func (c *Connection) TaskDone(task string, taskID bson.ObjectId) {
	taskName := task + ":" + taskID.Hex()

	if task == "create_post" || task == "follow_user" || task == "create_comment" || task == "delete_comment" {
		v, err := c.Redis.Do("SMEMBERS", taskName+":fail")
		keys, err := redis.Strings(v, err)
		if err != nil {
			return
		}

		for _, k := range keys {
			c.Redis.Do("DEL", "task_op_fail:"+k)
		}

		c.Redis.Do("DEL", taskName+":fail")
	}

	c.Redis.Do("DEL", taskName)
	c.Redis.Do("SREM", "tasks", taskID.Hex())
}

// TaskResolver is a process that takes care of the failed operations pushed to redis.
// The resolver tries to run again the task. If the task can't be completed by the resolver it will get discarded.
func (c *Connection) TaskResolver() {
	// TODO implement
}
