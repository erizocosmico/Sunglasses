package lamp

import (
	"github.com/martini-contrib/render"
	"labix.org/v2/mgo/bson"
	"net/http"
	"strings"
	"time"
)

// TokenType represents the type of the token
type TokenType int

const (
	// Token types
	AccessToken = 0
	UserToken   = 1

	// Expiration times
	AccessTokenExpirationHours = 1
	UserTokenExpirationDays    = 15
)

// Token represents a token which can be an access token or an user token
type Token struct {
	ID      bson.ObjectId `json:"id,omitempty" bson:"_id"`
	Type    TokenType     `json:"type" bson:"type"`
	Expires float64       `json:"expires" bson:"expires"`
	UserID  bson.ObjectId `json:"user_id,omitempty" bson:"user_id,omitempty"`
}

// ValidateAccessToken validates the supplied access token
func ValidateAccessToken(req *http.Request, conn *Connection, resp render.Render) {
	tokenID := req.Header.Get("X-Access-Token")

	if !bson.IsObjectIdHex(tokenID) {
		RenderError(resp, 3, 403, "Invalid access token provided")
		return
	}

	var token Token
	if err := conn.Db.C("tokens").FindId(bson.ObjectIdHex(tokenID)).One(&token); err != nil {
		RenderError(resp, 3, 403, "Invalid access token provided")
	} else {
		if token.ID.Hex() == "" || token.Type != AccessToken || token.Expires < float64(time.Now().Unix()) {
			RenderError(resp, 3, 403, "Invalid access token provided")
		}
	}

	_ := eraseExpiredTokens(conn)
}

// ValidateUserToken validates the supplied user token
func ValidateUserToken(req *http.Request, conn *Connection, resp render.Render) {
	tokenID := req.Header.Get("X-User-Token")

	if !bson.IsObjectIdHex(tokenID) {
		RenderError(resp, 3, 403, "Invalid user token provided")
		return
	}

	var token Token
	if err := conn.Db.C("tokens").FindId(bson.ObjectIdHex(tokenID)).One(&token); err != nil {
		RenderError(resp, 3, 403, "Invalid access token provided")
	} else {
		if token.ID.Hex() == "" || token.Type != UserToken || token.Expires < float64(time.Now().Unix()) {
			RenderError(resp, 3, 403, "Invalid access token provided")
		}
	}
}

func eraseExpiredTokens(conn *Connection) error {
	_, err := conn.Db.C("tokens").RemoveAll(bson.M{"expires": bson.M{"$lt": float64(time.Now().Unix())}})
	if err != nil {
		// TODO log error
	}

	return err
}

// GetAccessToken is a handler to retrieve an access token
func GetAccessToken(conn *Connection, resp render.Render) {
	token := new(Token)
	token.Expires = float64(time.Now().Add(AccessTokenExpirationHours * time.Hour).Unix())
	token.Type = AccessToken

	if err := token.Save(conn); err == nil {
		resp.JSON(200, map[string]interface{}{
			"error":        false,
			"access_token": token.ID,
			"expires":      token.Expires,
		})
	} else {
		RenderError(resp, 2, 500, "Unknown error occurred")
	}
}

// GetUserToken is a handler to retrieve an user token
func GetUserToken(req *http.Request, conn *Connection, resp render.Render) {
	username := strings.ToLower(req.PostFormValue("username"))
	password := req.PostFormValue("password")

	user := new(User)

	if err := conn.Db.C("users").Find((bson.M{"username_lower": username})).One(user); err != nil {
		RenderError(resp, 1, 400, "Invalid username or password")
		return
	}

	if user.CheckPassword(password) && user.Active {
		token := new(Token)
		token.Expires = float64(time.Now().AddDate(0, 0, UserTokenExpirationDays).Unix())
		token.UserID = user.ID
		token.Type = UserToken

		if err := token.Save(conn); err != nil {
			RenderError(resp, 2, 500, "Unexpected error occurred")
		} else {
			resp.JSON(200, map[string]interface{}{
				"error":      false,
				"user_token": token.ID,
				"expires":    token.Expires,
			})
		}
	} else {
		RenderError(resp, 1, 400, "Invalid username or password")
	}
}

// DestroyUserToken destroys the current user token
func DestroyUserToken(req *http.Request, conn *Connection, resp render.Render) {
	tokenID := req.Header.Get("X-User-Token")

	if !bson.IsObjectIdHex(tokenID) {
		RenderError(resp, 3, 403, "Invalid user token provided")
		return
	}

	if err := conn.Db.C("tokens").RemoveId(bson.ObjectIdHex(tokenID)); err != nil {
		RenderError(resp, 2, 404, "Token not found")
		return
	} else {
		resp.JSON(200, map[string]interface{}{
			"error":   false,
			"deleted": true,
			"message": "Token destroyed successfully",
		})
	}
}

// Save inserts the Token instance if it hasn't been reated yet ot updates it if it has
func (t *Token) Save(conn *Connection) error {
	if t.ID.Hex() == "" {
		t.ID = bson.NewObjectId()
	}

	if err := conn.Save("tokens", t.ID, t); err != nil {
		return err
	}

	return nil
}
