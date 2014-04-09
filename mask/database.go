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
func (c *Connection) PushTask(task string, args ...interface{}) {
	// TODO implement
}

// PushFail pushes a failed operation to Redis
func (c *Connection) PushFail(task string, args ...interface{}) {
	// TODO implement
}

// PushSuccess pushes a successful operation to redis
func (c *Connection) PushSuccess(task string, args ...interface{}) {
	// TODO implement
}

// TaskDone clears all data related to a certain task, that means the task has been completed successfully
func (c *Connection) TaskDone(task string, args ...interface{}) {
	// TODO implement
}

// TaskDone clears all the successful operations and leaves the failed ones to be taken care of later
func (c *Connection) TaskFailed(task string, args ...interface{}) {
	// TODO implement
}

// TaskResolver is a process that takes care of the failed operations pushed to redis.
// The resolver tries to run again the task. If the task can't be completed by the resolver it will get discarded.
func (c *Connection) TaskResolver() {
	// TODO implement
}
