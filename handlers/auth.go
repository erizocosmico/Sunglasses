package handlers

import (
	. "github.com/mvader/mask/error"
	"github.com/mvader/mask/middleware"
	"github.com/mvader/mask/models"
	"github.com/mvader/mask/modules/auth"
	"github.com/mvader/mask/util"
	"labix.org/v2/mgo/bson"
	"strings"
	"time"
)

// GetAccessToken is a handler to retrieve an access token
func GetAccessToken(c middleware.Context) {
	token := new(models.Token)
	token.Expires = float64(time.Now().Add(models.AccessTokenExpirationHours * time.Hour).Unix())
	token.Hash = util.NewRandomHash()
	token.Type = models.AccessToken

	if err := token.Save(c.Conn); err == nil {
		c.Success(200, map[string]interface{}{
			"access_token": token.Hash,
			"expires":      token.Expires,
		})
	} else {
		c.Error(500, CodeUnexpected, MsgUnexpected)
	}
}

// Login is a handler to log the user in
func Login(c middleware.Context) {
	c.Request.ParseForm()
	c.Request.Form.Add("token_type", "session")
	GetUserToken(c)
}

// GetUserToken is a handler to retrieve an user token
func GetUserToken(c middleware.Context) {
    // TODO: Max retries
	username := strings.ToLower(c.Form("username"))
	password := c.Form("password")

	user := new(models.User)

	if err := c.Find("users", (bson.M{"username_lower": username})).One(user); err != nil {
		c.Error(400, CodeInvalidUsernameOrPassword, MsgInvalidUsernameOrPassword)
		return
	}

	if user.CheckPassword(password) && user.Active {
		token := new(models.Token)
		token.Hash = util.NewRandomHash()
		token.Expires = float64(time.Now().AddDate(0, 0, models.UserTokenExpirationDays).Unix())
		token.UserID = user.ID
		if c.Form("token_type") == "session" {
			token.Type = models.SessionToken
		} else {
			token.Type = models.UserToken
		}

		if err := token.Save(c.Conn); err != nil {
			c.Error(500, CodeUnexpected, MsgUnexpected)
		} else {
			if c.Form("token_type") == "session" {
				c.Session.Set("user_token", token.Hash)

				c.Success(200, map[string]interface{}{
					"expires": token.Expires,
				})
			} else {
				c.Success(200, map[string]interface{}{
					"user_token": token.Hash,
					"expires":    token.Expires,
				})
			}
		}
	} else {
		c.Error(400, CodeInvalidUsernameOrPassword, MsgInvalidUsernameOrPassword)
	}
}

// DestroyUserToken destroys the current user token
func DestroyUserToken(c middleware.Context) {
	tokenID, tokenType := auth.GetRequestToken(c.Request, false, c.Session)
	if valid, _ := auth.IsTokenValid(tokenID, tokenType, c.Conn); valid {
		if err := c.Remove("tokens", bson.M{"hash": tokenID}); err != nil {
			c.Error(404, CodeTokenNotFound, MsgTokenNotFound)
			return
		} else {
			if tokenType == models.SessionToken {
				c.Session.Delete("user_token")
				c.Session.Delete("csrf_key")
			}

			c.Success(200, map[string]interface{}{
				"deleted": true,
				"message": "Token destroyed successfully",
			})
		}
	} else {
		c.Error(403, CodeInvalidUserToken, MsgInvalidUserToken)
	}
}
