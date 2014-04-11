package auth

import (
	"github.com/martini-contrib/sessions"
	"github.com/mvader/mask/models"
	"github.com/mvader/mask/services/interfaces"
	"github.com/mvader/mask/util"
	"labix.org/v2/mgo/bson"
	"net/http"
	"time"
)

// GetRequestToken returns the token associated with the request
func GetRequestToken(r *http.Request, isAccessToken bool, s sessions.Session) (string, models.TokenType) {
	var (
		token     string
		tokenType models.TokenType
	)

	if isAccessToken {
		return util.Hash(r.Header.Get("X-Access-Token")), models.AccessToken
	}

	token = r.Header.Get("X-User-Token")
	tokenType = models.UserToken

	if token == "" {
		// We're accessing via web
		tokenType = models.SessionToken
		v := s.Get("user_token")
		if v != nil {
			token = s.Get("user_token").(string)
		}
	}

	return util.Hash(token), tokenType
}

// IsTokenValid returns if the provided token is a valid token
func IsTokenValid(tokenID string, tokenType models.TokenType, conn interfaces.Conn) (bool, bson.ObjectId) {
	var userID bson.ObjectId

	var token models.Token
	if err := conn.C("tokens").Find(bson.M{"hash": tokenID}).One(&token); err == nil {
		if token.Expires > float64(time.Now().Unix()) && token.Type == tokenType {
			return true, token.UserID
		}
	}

	return false, userID
}

// GetRequestUser returns the user associated with the request
func GetRequestUser(r *http.Request, conn interfaces.Conn, s sessions.Session) *models.User {
	var (
		userID bson.ObjectId
		valid  bool
		user   models.User
	)
	token, tokenType := GetRequestToken(r, false, s)

	if valid, userID = IsTokenValid(token, tokenType, conn); valid {
		if err := conn.C("users").FindId(userID).One(&user); err == nil {
			return &user
		}
	}

	return nil
}

// EraseExpiredTokens removes all expired tokens from the database
func EraseExpiredTokens(conn interfaces.Conn) error {
	_, err := conn.C("tokens").RemoveAll(bson.M{"expires": bson.M{"$lt": float64(time.Now().Unix())}})
	if err != nil {
		// TODO log error
	}

	return err
}
