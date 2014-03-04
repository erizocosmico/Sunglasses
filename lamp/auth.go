package lamp

import (
	r "github.com/dancannon/gorethink"
	"github.com/martini-contrib/render"
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
	ID      string    `json:"id,omitempty" gorethink:"id,omitempty"`
	Type    TokenType `json:"type" gorethink:"type"`
	Expires float64   `json:"expires" gorethink:"expires"`
	UserID  string    `json:"user_id,omitempty" gorethink:"user_id,omitempty"`
}

// ValidateAccessToken validates the supplied access token
func ValidateAccessToken(req *http.Request, conn *Connection, resp render.Render) {
	token := req.Header.Get("X-Access-Token")

	result, err := conn.Db.Table("token").
		Filter(r.Row.Field("type").
		Eq(AccessToken).
		And(r.Row.Field("id").Eq(token)).
		And(r.Row.Field("expires").Gt(time.Now().Unix()))).
		Count().
		RunRow(conn.Session)
	if err != nil {
		RenderError(resp, 2, 500, "Unknown error occurred")
	} else {
		var count int64
		result.Scan(&count)

		if count < 1 {
			RenderError(resp, 3, 403, "Invalid access token provided")
		}
	}

	eraseExpiredTokens(conn)
}

// ValidateUserToken validates the supplied user token
func ValidateUserToken(req *http.Request, conn *Connection, resp render.Render) {
	token := req.Header.Get("X-User-Token")

	result, err := conn.Db.Table("token").
		Filter(r.Row.Field("type").
		Eq(UserToken).
		And(r.Row.Field("id").Eq(token)).
		And(r.Row.Field("expires").Gt(time.Now().Unix()))).
		Count().
		RunRow(conn.Session)
	if err != nil {
		RenderError(resp, 2, 500, "Unknown error occurred")
	} else {
		var count int64
		result.Scan(&count)

		if count < 1 {
			RenderError(resp, 4, 403, "Invalid user token provided")
		}
	}
}

func eraseExpiredTokens(conn *Connection) {
	_, err := conn.Db.Table("token").
		Filter(r.Row.Field("expires").Lt(time.Now().Unix())).
		Delete().
		RunWrite(conn.Session)
	if err != nil {
		// TODO log error
	}
}

// GetAccessToken is a handler to retrieve an access token
func GetAccessToken(conn *Connection, resp render.Render) {
	token := new(Token)
	token.Expires = float64(time.Now().Add(AccessTokenExpirationHours * time.Hour).Unix())
	token.Type = AccessToken

	success, err := token.Save(conn)
	if err == nil && success {
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

	res, err := conn.Db.Table("user").
		Filter(r.Row.Field("username_lower").Eq(username).And(r.Row.Field("active").Eq(true))).
		RunRow(conn.Session)
	if err != nil {
		RenderError(resp, 2, 500, "Unknown error occurred")
		return
	}
	res.Scan(user)

	if user.CheckPassword(password) {
		token := new(Token)
		token.Expires = float64(time.Now().AddDate(0, 0, UserTokenExpirationDays).Unix())
		token.UserID = user.ID
		token.Type = UserToken

		success, err := token.Save(conn)
		if err == nil && !success {
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

	res, err := conn.Db.Table("token").Get(tokenID).Delete().RunWrite(conn.Session)
	if err != nil {
		RenderError(resp, 2, 500, "Unknown error occurred")
		return
	} else if res.Deleted < 1 {
		resp.JSON(200, map[string]interface{}{
			"error":   false,
			"deleted": false,
			"message": "Token destroyed successfully",
		})
	} else {
		resp.JSON(200, map[string]interface{}{
			"error":   false,
			"deleted": true,
			"message": "Token destroyed successfully",
		})
	}
}

// Save inserts the Token instance if it hasn't been reated yet ot updates it if it has
func (t *Token) Save(conn *Connection) (bool, error) {
	success, err, ID := conn.Save("token", t.ID, t)
	if err != nil {
		return false, err
	}

	if !success {
		return false, nil
	}

	if t.ID == "" {
		t.ID = ID
	}

	return true, nil
}
