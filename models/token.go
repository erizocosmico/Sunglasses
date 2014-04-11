package models

import (
	"github.com/mvader/mask/services/interfaces"
	"github.com/mvader/mask/util"
	"labix.org/v2/mgo/bson"
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
	UserTokenExpirationDays    = 30
)

// Token represents a token which can be an access token, an user token or a session token
type Token struct {
	ID      bson.ObjectId `json:"id,omitempty" bson:"_id"`
	Type    TokenType     `json:"type" bson:"type"`
	Hash    string        `json:"hash" bson:"hash"`
	Expires float64       `json:"expires" bson:"expires"`
	UserID  bson.ObjectId `json:"user_id,omitempty" bson:"user_id,omitempty"`
}

// Save inserts the Token instance if it hasn't been created yet or updates it if it has
func (t *Token) Save(conn interfaces.Saver) error {
	var hashTmp string
	if t.ID.Hex() == "" {
		t.ID = bson.NewObjectId()
	}

	if t.Hash == "" {
		t.Hash = util.NewRandomHash()
	}

	hashTmp = t.Hash
	t.Hash = util.Hash(t.Hash)

	if err := conn.Save("tokens", t.ID, t); err != nil {
		return err
	}

	t.Hash = hashTmp

	return nil
}

// Remove removes the Token instance
func (t *Token) Remove(conn interfaces.Remover) error {
	if err := conn.Remove("tokens", t.ID); err != nil {
		return err
	}

	return nil
}
