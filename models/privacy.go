package models

import "labix.org/v2/mgo/bson"

type PrivacyType int

// PrivacySettings model
type PrivacySettings struct {
	Type  PrivacyType     `json:"privacy_type,omitempty" bson:"privacy_type"`
	Users []bson.ObjectId `json:"users,omitempty" bson:"users,omitempty"`
}

const (
	PrivacyPublic        = 1
	PrivacyFollowersOnly = 2
	PrivacyFollowingOnly = 3
	PrivacyNone          = 4
	PrivacyAllBut        = 5
	PrivacyFollowersBut  = 6
	PrivacyFollowingBut  = 7
	PrivacyNoneBut       = 8
)

// NewPrivacySettings returns a new empty instance of PrivacySettings
func NewPrivacySettings() PrivacySettings {
	p := PrivacySettings{}
	p.Type = PrivacyNone

	return p
}

// IsValidPrivacyType determines if the given PrivacyType is valid or not
func IsValidPrivacyType(t PrivacyType) bool {
	return t > 0 && t <= PrivacyNoneBut
}
