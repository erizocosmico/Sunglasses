package tests

import (
	. "github.com/mvader/sunglasses/models"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestUser(t *testing.T) {
	// TODO revisit
	conn := getConnection()
	defer conn.Session.Close()

	Convey("Subject: Creating an user and removing it", t, func() {
		Convey("Given a database connection", func() {

			Convey("and an user instance", func() {
				user := new(User)
				user.Username = "John Doe"
				err := user.SetPassword("testing")
				if err != nil {
					panic(err)
				}
				err = user.SetEmail("test@test.com")
				if err != nil {
					panic(err)
				}
				user.Role = RoleUser
				user.Active = true

				Convey("User should be saved correctly", func() {
					err := user.Save(conn)
					So(err, ShouldEqual, nil)
				})

				Convey("And the password must match 'testing'", func() {
					valid := user.CheckPassword("testing")
					So(valid, ShouldEqual, true)
				})

				Convey("The email hash must match 'test@test.com'", func() {
					valid := user.CheckEmail("test@test.com")
					So(valid, ShouldEqual, true)
				})

				Convey("But with the same username the user should not be inserted", func() {
					uid := user.ID
					user.ID = ""

					err := user.Save(conn)
					So(err, ShouldNotEqual, nil)

					user.ID = uid
				})

				Convey("Deleting the user should not produce any errors", func() {
					err := user.Remove(conn)
					So(err, ShouldEqual, nil)
				})
			})
		})
	})
}
