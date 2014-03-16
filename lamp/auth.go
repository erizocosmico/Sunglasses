package lamp

import (
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	"labix.org/v2/mgo/bson"
	"net/http"
	"strings"
	"time"
)

// TokenType represents the type of the token
type TokenType int

const (
	// Token types
	AccessToken  = 0
	UserToken    = 1
	SessionToken = 2

	// Expiration times
	AccessTokenExpirationHours = 1
	UserTokenExpirationDays    = 15
)

// Token represents a token which can be an access token, an user token or a session token
type Token struct {
	ID      bson.ObjectId `json:"id,omitempty" bson:"_id"`
	Type    TokenType     `json:"type" bson:"type"`
	Expires float64       `json:"expires" bson:"expires"`
	UserID  bson.ObjectId `json:"user_id,omitempty" bson:"user_id,omitempty"`
}

// ValidateAccessToken validates the supplied access token
func ValidateAccessToken(req *http.Request, conn *Connection, resp render.Render, s sessions.Session) {
	tokenID, _ := GetRequestToken(req, true, s)

	if !bson.IsObjectIdHex(tokenID) {
		RenderError(resp, CodeInvalidAccessToken, 403, MsgInvalidAccessToken)
		return
	}

	var token Token
	if err := conn.Db.C("tokens").FindId(bson.ObjectIdHex(tokenID)).One(&token); err != nil {
		RenderError(resp, CodeInvalidAccessToken, 403, MsgInvalidAccessToken)
	} else {
		if token.ID.Hex() == "" || token.Type != AccessToken || token.Expires < float64(time.Now().Unix()) {
			RenderError(resp, CodeInvalidAccessToken, 403, MsgInvalidAccessToken)
		}
	}

	_ = eraseExpiredTokens(conn)
}

// ValidateUserToken validates the supplied user token
func ValidateUserToken(req *http.Request, conn *Connection, resp render.Render, s sessions.Session) {
	tokenID, _ := GetRequestToken(req, false, s)

	if !bson.IsObjectIdHex(tokenID) {
		RenderError(resp, CodeInvalidUserToken, 403, MsgInvalidUserToken)
		return
	}

	var token Token
	if err := conn.Db.C("tokens").FindId(bson.ObjectIdHex(tokenID)).One(&token); err != nil {
		RenderError(resp, CodeInvalidAccessToken, 403, MsgInvalidAccessToken)
	} else {
		if token.ID.Hex() == "" || token.Type != UserToken || token.Expires < float64(time.Now().Unix()) {
			RenderError(resp, CodeInvalidAccessToken, 403, MsgInvalidAccessToken)
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
		RenderError(resp, CodeUnexpected, 500, MsgUnexpected)
	}
}

// Login is a handler to log the user in
func Login(req *http.Request, conn *Connection, resp render.Render, s sessions.Session) {
	req.PostForm.Add("token_type", "session")
	GetUserToken(req, conn, resp, s)
}

// GetUserToken is a handler to retrieve an user token
func GetUserToken(req *http.Request, conn *Connection, resp render.Render, s sessions.Session) {
	username := strings.ToLower(req.PostFormValue("username"))
	password := req.PostFormValue("password")

	user := new(User)

	if err := conn.Db.C("users").Find((bson.M{"username_lower": username})).One(user); err != nil {
		RenderError(resp, CodeInvalidUsernameOrPassword, 400, MsgInvalidUsernameOrPassword)
		return
	}

	if user.CheckPassword(password) && user.Active {
		token := new(Token)
		token.Expires = float64(time.Now().AddDate(0, 0, UserTokenExpirationDays).Unix())
		token.UserID = user.ID
		if req.PostFormValue("token_type") == "session" {
			token.Type = SessionToken
		} else {
			token.Type = UserToken
		}

		if err := token.Save(conn); err != nil {
			RenderError(resp, CodeUnexpected, 500, MsgUnexpected)
		} else {
			if req.PostFormValue("token_type") == "session" {
				s.Set("user_token", token.ID.Hex())

				resp.JSON(200, map[string]interface{}{
					"error": false,
				})
			} else {
				resp.JSON(200, map[string]interface{}{
					"error":      false,
					"user_token": token.ID,
					"expires":    token.Expires,
				})
			}
		}
	} else {
		RenderError(resp, CodeInvalidUsernameOrPassword, 400, MsgInvalidUsernameOrPassword)
	}
}

// DestroyUserToken destroys the current user token
func DestroyUserToken(req *http.Request, conn *Connection, resp render.Render, s sessions.Session) {
	tokenID, tokenType := GetRequestToken(req, false, s)
	if valid, _ := IsTokenValid(tokenID, tokenType, conn); valid {
		if err := conn.Db.C("tokens").RemoveId(bson.ObjectIdHex(tokenID)); err != nil {
			RenderError(resp, CodeTokenNotFound, 404, MsgTokenNotFound)
			return
		} else {
			resp.JSON(200, map[string]interface{}{
				"error":   false,
				"deleted": true,
				"message": "Token destroyed successfully",
			})
		}
	} else {
		RenderError(resp, CodeInvalidUserToken, 403, MsgInvalidUserToken)
	}
}

// GetRequestToken returns the token associated with the request
func GetRequestToken(r *http.Request, isAccessToken bool, s sessions.Session) (string, TokenType) {
	var (
		token     string
		tokenType TokenType
	)
	if isAccessToken {
		return r.Header.Get("X-Access-Token"), AccessToken
	}

	token = r.Header.Get("X-User-Token")
	tokenType = UserToken

	if token == "" {
		// We're accessing via web
		tokenType = SessionToken
		v := s.Get("user_token")
		if v != nil {
			token = s.Get("user_token").(string)
		}
	}

	return token, tokenType
}

// IsTokenValid returns if the provided token is a valid token
func IsTokenValid(tokenID string, tokenType TokenType, conn *Connection) (bool, bson.ObjectId) {
	var userID bson.ObjectId
	if !bson.IsObjectIdHex(tokenID) {
		return false, userID
	}

	var token Token
	if err := conn.Db.C("tokens").FindId(bson.ObjectIdHex(tokenID)).One(&token); err == nil {
		if token.Expires > float64(time.Now().Unix()) && token.Type == tokenType {
			return true, token.ID
		}
	}

	return false, userID
}

// GetRequestUser returns the user associated with the request
func GetRequestUser(r *http.Request, conn *Connection, s sessions.Session) *User {
	var (
		userID bson.ObjectId
		valid  bool
		user   User
	)
	token, tokenType := GetRequestToken(r, false, s)

	if valid, userID = IsTokenValid(token, tokenType, conn); valid {
		if err := conn.Db.C("users").FindId(userID).One(&user); err != nil {
			return &user
		}
	}

	return nil
}

// Save inserts the Token instance if it hasn't been created yet or updates it if it has
func (t *Token) Save(conn *Connection) error {
	if t.ID.Hex() == "" {
		t.ID = bson.NewObjectId()
	}

	if err := conn.Save("tokens", t.ID, t); err != nil {
		return err
	}

	return nil
}
