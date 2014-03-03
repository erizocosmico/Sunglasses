package lamp

import (
	r "github.com/dancannon/gorethink"
	"os"
	"time"
)

// Connection represents the database session
type Connection struct {
	Session *r.Session
	Db      r.RqlTerm
}

// NewDatabaseConn initializes the database connection
func NewDatabaseConn(config *Config) (*Connection, error) {
	// Needed for wercker
	url := os.Getenv("WERCKER_RETHINKDB_URL")
	if url == "" {
		url = config.DatabaseUrl
	}

	args := make(map[string]interface{})
	args["address"] = url
	if config.DatabaseAuthKey != "" {
		args["authkey"] = config.DatabaseAuthKey
	}
	if config.DatabaseName != "" {
		args["database"] = config.DatabaseName
	}
	args["maxIdle"] = config.DatabaseMaxIdleConnections
	args["idleTimeout"] = time.Second * time.Duration(config.DatabaseIdleTimeout)
	args["maxActive"] = config.DatabaseMaxActiveConnections

	session, err := r.Connect(args)
	if err != nil {
		return nil, err
	}

	var conn *Connection = new(Connection)
	conn.Session = session

	// Run this only on debug. We don't want to waste time with this on production
	if config.Debug {
		err = createDatabase(conn, config)
		if err != nil {
			return nil, err
		}
	}

	conn.Db = r.Db(config.DatabaseName)

	return conn, nil
}

// Save inserts an item or updates it if it has already been created
func (c *Connection) Save(table, ID string, item interface{}) (bool, error, string) {
	var err error
	var res r.WriteResponse

	if ID == "" {
		res, err = c.Db.Table(table).Insert(item).RunWrite(c.Session)
	} else {
		res, err = c.Db.Table(table).Replace(item).RunWrite(c.Session)
	}

	if err != nil {
		return false, err, ""
	}

	if ID == "" {
		if res.Inserted != 1 {
			return false, nil, ""
		} else {
			if len(res.GeneratedKeys) < 1 {
				return false, nil, ""
			} else {
				return true, nil, res.GeneratedKeys[0]
			}
		}
	} else {
		if res.Updated != 1 && res.Replaced != 1 {
			return false, nil, ""
		}
	}

	return true, nil, ""
}

func (c *Connection) Remove(table, ID string) (bool, error) {
	res, err := c.Db.Table(table).Get(ID).Delete().RunWrite(c.Session)
	if err != nil || res.Deleted < 1 {
		return false, err
	}

	return true, nil
}

func createDatabase(conn *Connection, config *Config) error {
	tables := []string{
		"user",
		"post",
		"album",
		"notification",
		"request",
		"block",
		"follow",
		"report",
		"token",
	}
	indexes := map[string][]string{
		"post":         []string{"user_id"},
		"album":        []string{"user_id"},
		"notification": []string{"user_id"},
		"token":        []string{"user_id", "token"},
		"request":      []string{"user_to", "user_from"},
		"follow":       []string{"user_to", "user_from"},
		"report":       []string{"user_id", "post_id"},
		"block":        []string{"user_to", "user_from"},
	}

	_, err := r.DbCreate(config.DatabaseName).RunRow(conn.Session)
	if err != nil {
		return nil
	}

	for _, table := range tables {
		_, err := r.Db(config.DatabaseName).TableCreate(table).RunWrite(conn.Session)
		if err != nil {
			return err
		}
	}

	for table, tableIndexes := range indexes {
		for _, index := range tableIndexes {
			_, err := r.Db(config.DatabaseName).Table(table).IndexCreate(index).RunWrite(conn.Session)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
