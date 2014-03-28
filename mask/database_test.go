package mask

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestNewDatabaseConn(t *testing.T) {
	Convey("Subject: Establishing a database connection", t, func() {
		Convey("Given a valid config", func() {
			config, err := NewConfig("../config.sample.json")
			if err != nil {
				panic(err)
			}

			Convey("And a valid mongodb address", func() {
				conn, err := NewDatabaseConn(config)
				defer conn.Session.Close()
				So(conn, ShouldNotEqual, nil)
				So(err, ShouldEqual, nil)
			})
		})
	})
}
