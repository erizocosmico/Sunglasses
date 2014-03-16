package lamp

import (
	"labix.org/v2/mgo/bson"
	"time"
	"net/http"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
)

// Block model (same as follow)
type Block Follow

// BlockUser blocks an user ("from" blocks "to")
func BlockUser(from, to bson.ObjectId, conn *Connection) error {
	f := Block{}
	f.ID = bson.NewObjectId()
	f.To = to
	f.From = from
	f.Time = float64(time.Now().Unix())

	if err := conn.Save("blocks", f.ID, f); err != nil {
		return err
	}

	return nil
}

// UnblockUser unblocks an user ("from" unblocks "to")
func UnblockUser(from, to bson.ObjectId, conn *Connection) error {
	if err := conn.Db.C("blocks").Remove(bson.M{"from": from, "to": to}); err != nil {
		return err
	}

	return nil
}

// UserIsBlocked returns if the user is blocked
func UserIsBlocked(from, to bson.ObjectId, conn *Connection) bool {
	var (
		err   error
		count int
	)

	if count, err = conn.Db.C("blocks").Find(bson.M{"from": from, "to": to}).Count(); err != nil {
		return false
	}

	return count > 0
}

// Block blocks an user
func Block(r *http.Request, conn *Connection, res render.Render, s sessions.Session) {
	user := GetRequestUser(r, conn, s)

	if user != nil {
		userTo := r.PostFormValue("user_to")

		if userTo != "" && bson.IsObjectIdHex(userTo) {
			userToID := bson.ObjectIdHex(userTo)

			if toUser := UserExists(conn, userToID); toUser != nil {
					if err := BlockUser(user.ID, userToID, conn); err != nil {
						RenderError(res, CodeUnexpected, 500, MsgUnexpected)
						return
					}

					res.JSON(200, map[string]interface{}{
						"error": false,
						"message": "User blocked successfully",
					})
					return
			} else {
				RenderError(res, CodeUserDoesNotExist, 404, MsgUserDoesNotExist)
				return
			}
		}
	}

	RenderError(res, CodeInvalidData, 400, MsgInvalidData)
}

// Unblock unblocks an user
func Unblock(r *http.Request, conn *Connection, res render.Render, s sessions.Session) {
	user := GetRequestUser(r, conn, s)

	if user != nil {
		userTo := r.PostFormValue("user_to")

		if userTo != "" && bson.IsObjectIdHex(userTo) {
			userToID := bson.ObjectIdHex(userTo)

			if toUser := UserExists(conn, userToID); toUser != nil {
				if err := UnblockUser(user.ID, userToID, conn); err != nil {
					RenderError(res, CodeUnexpected, 500, MsgUnexpected)
					return
				}

				res.JSON(200, map[string]interface{}{
					"error": false,
					"message": "User unblocked successfully",
				})
				return
			} else {
				RenderError(res, CodeUserDoesNotExist, 404, MsgUserDoesNotExist)
				return
			}
		}
	}

	RenderError(res, CodeInvalidData, 400, MsgInvalidData)
}
