package models

import (
	"code.google.com/p/go.crypto/bcrypt"
	"errors"
	"github.com/mvader/mask/services/interfaces"
	"github.com/mvader/mask/util"
	"labix.org/v2/mgo/bson"
	"strings"
)

// UserRole represents an user role
type UserRole int

// Gender represents an user gender
type Gender int

// UserStatus represents the civil status of the user
type UserStatus int

// RecoveryMethod represents the type of recovery for an account
type RecoveryMethod int

const (
	// Roles
	RoleUser  = 0
	RoleAdmin = 1

	// Genders
	Male   = 0
	Female = 1
	Other  = 2

	// User statuses
	Single          = 0
	Married         = 1
	InARelationship = 2
	ItsComplicated  = 3
	OtherStatus     = 4

	// Recovery methods
	RecoveryNone      = 0
	RecoverByEMail    = 1
	RecoverByQuestion = 2
)

// User represents an application user
type User struct {
	ID                    bson.ObjectId `json:"id,omitempty" bson:"_id"`
	Username              string        `json:"username" bson:"username"`
	UsernameLower         string        `json:"username_lower,omitempty" bson:"username_lower"`
	Password              string        `json:"-" bson:"password"`
	EMail                 string        `json:"-" bson:"email,omitempty"`
	PublicName            string        `json:"public_name" bson:"public_name,omitempty"`
	PrivateName           string        `json:"private_name" bson:"private_name,omitempty"`
	Role                  UserRole      `json:"role,omitempty" bson:"role"`
	PreferredLanguage     string        `json:"preferred_lang,omitempty" bson:"preferred_lang,omitempty"`
	Timezone              int           `json:"timezone,omitempty" bson:"timezone,omitempty"`
	Avatar                string        `json:"avatar" bson:"avatar,omitempty"`
	AvatarThumbnail       string        `json:"avatar_thumbnail" bson:"avatar_thumbnail,omitempty"`
	PublicAvatar          string        `json:"public_avatar" bson:"public_avatar,omitempty"`
	PublicAvatarThumbnail string        `json:"public_avatar_thumbnail" bson:"public_avatar_thumbnail,omitempty"`
	Active                bool          `json:"active,omitempty" bson:"active"`
	Info                  UserInfo      `json:"info,omitempty" bson:"info"`
	Settings              UserSettings  `json:"settings,omitempty" bson:"settings"`
}

// UserInfo stores all personal information about the user
type UserInfo struct {
	Work      string     `json:"work,omitempty" bson:"work,omitempty"`
	Education string     `json:"education,omitempty" bson:"education,omitempty"`
	Hobbies   string     `json:"hobbies,omitempty" bson:"hobbies,omitempty"`
	Books     string     `json:"books,omitempty" bson:"books,omitempty"`
	Movies    string     `json:"movies,omitempty" bson:"movies,omitempty"`
	TV        string     `json:"tv,omitempty" bson:"tv,omitempty"`
	Gender    Gender     `json:"gender,omitempty" bson:"gender,omitempty"`
	Websites  []string   `json:"websites,omitempty" bson:"websites,omitempty"`
	Status    UserStatus `json:"status,omitempty" bson:"status,omitempty"`
	About     string     `json:"about,omitempty" bson:"about,omitempty"`
}

// UserSettings stores the user preferences
type UserSettings struct {
	Invisible                   bool           `json:"invisible,omitempty" bson:"invisible"`
	CanReceiveRequests          bool           `json:"can_receive_requests,omitempty" bson:"can_receive_requests"`
	FollowApprovalRequired      bool           `json:"follow_approval_required,omitempty" bson:"follow_approval_required"`
	DisplayAvatarBeforeApproval bool           `json:"display_avatar_before_approval,omitempty" bson:"display_avatar_before_approval"`
	NotifyNewComment            bool           `json:"notify_new_comment,omitempty" bson:"notify_new_comment"`
	NotifyNewCommentOthers      bool           `json:"notify_new_comment_others,omitempty" bson:"notify_new_comment_others"`
	NotifyPostsInMyProfile      bool           `json:"notify_posts_in_my_profile,omitempty" bson:"notify_posts_in_my_profile"`
	NotifyLikes                 bool           `json:"notify_likes,omitempty" bson:"notify_likes"`
	AllowPostsInMyProfile       bool           `json:"allow_posts_in_my_profile,omitempty" bson:"allow_posts_in_my_profile"`
	AllowCommentsInPosts        bool           `json:"allow_comments_in_posts,omitempty" bson:"allow_comments_in_posts"`
	DisplayInfoFollowersOnly    bool           `json:"display_info_followers_only,omitempty" bson:"display_info_followers_only"`
	PasswordRecoveryMethod      RecoveryMethod `json:"recovery_method,omitempty" bson:"recovery_method"`
	RecoveryQuestion            string         `json:"recovery_question,omitempty" bson:"recovery_question,omitempty"`
	RecoveryAnswer              string         `json:"recovery_answer,omitempty" bson:"recovery_answer,omitempty"`
	// If this is true DefaultStatusPrivacy will override all the other settings
	OverrideDefaultPrivacy bool            `json:"override_default_privacy,omitempty" bson:"override_default_privacy,omitempty"`
	DefaultStatusPrivacy   PrivacySettings `json:"default_status_privacy,omitempty" bson:"default_status_privacy,omitempty"`
	DefaultVideoPrivacy    PrivacySettings `json:"default_video_privacy,omitempty" bson:"default_video_privacy,omitempty"`
	DefaultPhotoPrivacy    PrivacySettings `json:"default_photo_privacy,omitempty" bson:"default_photo_privacy,omitempty"`
	DefaultLinkPrivacy     PrivacySettings `json:"default_link_privacy,omitempty" bson:"default_link_privacy,omitempty"`
	DefaultAlbumPrivacy    PrivacySettings `json:"default_album_privacy,omitempty" bson:"default_album_privacy,omitempty"`
}

// NewUser returns a new User instance
func NewUser() *User {
	user := new(User)
	user.Settings = UserSettings{}
	user.Info = UserInfo{}

	return user
}

// Save inserts the User instance if it hasn't been created yet or updates it if it has
func (u *User) Save(conn interfaces.Conn) error {
	var count int
	var err error

	u.UsernameLower = strings.ToLower(u.Username)

	// Check if the username is already in use
	if count, err = conn.C("users").Find(bson.M{"username_lower": u.UsernameLower}).Count(); err != nil {
		return err
	}

	if u.ID.Hex() != "" && count > 1 {
		return errors.New("username already in use")
	} else if u.ID.Hex() == "" && count > 0 {
		return errors.New("username already in use")
	}

	// That means we're creating an user
	if u.ID.Hex() == "" {
		pvSet := NewPrivacySettings()
		u.ID = bson.NewObjectId()
		u.Settings.Invisible = true
		u.Settings.CanReceiveRequests = false
		u.Settings.FollowApprovalRequired = true
		u.Settings.NotifyPostsInMyProfile = false
		u.Settings.NotifyLikes = false
		u.Settings.DisplayAvatarBeforeApproval = false
		u.Settings.NotifyNewComment = false
		u.Settings.NotifyNewCommentOthers = false
		u.Settings.AllowPostsInMyProfile = false
		u.Settings.AllowCommentsInPosts = false
		u.Settings.DisplayInfoFollowersOnly = true
		u.Settings.PasswordRecoveryMethod = RecoveryNone
		u.Settings.DefaultStatusPrivacy = pvSet
		u.Settings.DefaultPhotoPrivacy = pvSet
		u.Settings.DefaultAlbumPrivacy = pvSet
		u.Settings.DefaultLinkPrivacy = pvSet
		u.Settings.DefaultVideoPrivacy = pvSet
		u.Active = true
	}

	if err = conn.Save("users", u.ID, u); err != nil {
		return err
	}

	return nil
}

// Remove deletes the user instance
func (u *User) Remove(conn interfaces.Remover) error {
	return conn.Remove("users", u.ID)
}

// SetEmail sets the email of the user
func (u *User) SetEmail(email string) error {
	emailHash, err := util.Crypt(email)
	if err != nil {
		return err
	}

	u.EMail = emailHash
	return nil
}

// CheckEmail checks if the given email matches the current user email
func (u *User) CheckEmail(email string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.EMail), []byte(email))
	return err == nil
}

// SetPassword sets a new encrypted password for the user
func (u *User) SetPassword(password string) error {
	pwHash, err := util.Crypt(password)
	if err != nil {
		return err
	}

	u.Password = pwHash
	return nil
}

// CheckPassword checks if the given password matches the current password hash
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// GetPrivacySettings returns the privacy settings of the user for the given object type
func (us UserSettings) GetPrivacySettings(objectType ObjectType) PrivacySettings {
	if us.OverrideDefaultPrivacy {
		return us.DefaultStatusPrivacy
	} else {
		switch objectType {
		case PostLink:
			return us.DefaultLinkPrivacy
		case PostPhoto:
			return us.DefaultPhotoPrivacy
		case PostVideo:
			return us.DefaultVideoPrivacy
		case Album:
			return us.DefaultAlbumPrivacy
		default:
			return us.DefaultStatusPrivacy
		}
	}
}

// UserExists returns the user if exists or nil
func UserExists(conn interfaces.Conn, ID bson.ObjectId) *User {
	var (
		err  error
		user = new(User)
	)

	if err = conn.C("users").FindId(ID).One(user); err != nil {
		return nil
	}

	return user
}

// GetUserData retrieves basic data from users for responses
func GetUsersData(ids []bson.ObjectId, user *User, conn interfaces.Conn) map[bson.ObjectId]map[string]interface{} {
	var u User
	var follows []Follow
	users := make(map[bson.ObjectId]map[string]interface{})
	cursor := conn.C("users").Find(bson.M{"_id": bson.M{"$in": ids}}).Iter()

	followsIter := conn.C("follows").Find(bson.M{"$or": []bson.M{
		bson.M{"user_from": user.ID, "user_to": bson.M{"$in": ids}},
		bson.M{"user_to": user.ID, "user_from": bson.M{"$in": ids}},
	}}).Iter()

	if err := followsIter.All(&follows); err != nil {
		return nil
	}

	if err := followsIter.Close(); err != nil {
		return nil
	}

	for cursor.Next(&u) {
		hasAccess := false
		for _, v := range follows {
			if v.From.Hex() == u.ID.Hex() || v.To.Hex() == u.ID.Hex() {
				hasAccess = true
				break
			}
		}

		if u.ID.Hex() == user.ID.Hex() {
			hasAccess = true
		}

		if _, ok := users[u.ID]; !ok {
			users[u.ID] = UserForDisplay(u, hasAccess, false)
		}
	}

	if err := cursor.Close(); err != nil {
		return nil
	}

	return users
}

// UserForDisplay returns a displayable version of the user model based on the users permissions
func UserForDisplay(u User, hasAccess, includeInfo bool) map[string]interface{} {
	var user map[string]interface{}

	if !(u.Settings.Invisible && !hasAccess) {
		user = map[string]interface{}{
			"id":                      u.ID,
			"username":                u.Username,
			"public_avatar":           u.PublicAvatar,
			"public_avatar_thumbnail": u.PublicAvatarThumbnail,
			"public_name":             u.PublicName,
			"avatar":                  "",
			"avatar_thumbnail":        "",
			"private_name":            "",
		}

		if u.Settings.DisplayAvatarBeforeApproval || hasAccess {
			user["avatar"] = u.Avatar
			user["avatar_thumbnail"] = u.AvatarThumbnail
			user["private_name"] = u.PrivateName
		}

		if hasAccess && includeInfo {
			user["info"] = u.Info
		}
	} else {
		user = map[string]interface{}{
			"username":                "Protected",
			"public_avatar":           "",
			"public_avatar_thumbnail": "",
			"public_name":             "",
			"avatar":                  "",
			"avatar_thumbnail":        "",
			"private_name":            "",
		}
	}

	return user
}
