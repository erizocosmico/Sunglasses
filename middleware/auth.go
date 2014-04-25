package middleware

import . "github.com/mvader/mask/error"

// LoginRequired returns an error if the user is not logged in
func LoginRequired(c Context) {
	if c.User == nil {
		c.Error(403, CodeNotLoggedIn, MsgNotLoggedIn)
	}
}

// LoginForbidden returns an error if the user is logged in
func LoginForbidden(c Context) {
	if c.User != nil {
		c.Error(403, CodeLoggedIn, MsgLoggedIn)
	}
}

// WebOnly returns an error if the user token is not a session token
func WebOnly(c Context) {
	if !c.IsWebToken {
		c.Error(403, CodeUnauthorized, MsgUnauthorized)
	}
}

// RequiresValidSignature returns an error if the request signature is not valid
func RequiresValidSignature(c Context) {
	if !c.RequestIsValid(c.Request.Header.Get("X-Access-Token") != "") {
		c.Error(400, CodeInvalidSignature, MsgInvalidSignature)
	}
}
