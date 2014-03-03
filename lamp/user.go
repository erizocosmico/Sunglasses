package lamp

import (
	"code.google.com/p/go.crypto/bcrypt"
	r "github.com/dancannon/gorethink"
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
	ID                string       `json:"id,omitempty" gorethink:"id,omitempty"`
	Username          string       `json:"username" gorethink:"username"`
	Password          string       `json:"password" gorethink:"password"`
	EMail             string       `json:"email,omitempty" gorethink:"email,omitempty"`
	PublicName        string       `json:"public_name,omitempty" gorethink:"public_name,omitempty"`
	PrivateName       string       `json:"private_name,omitempty" gorethink:"private_name,omitempty"`
	Role              UserRole     `json:"role" gorethink:"role"`
	PreferredLanguage string       `json:"preferred_lang,omitempty" gorethink:"preferred_lang,omitempty"`
	Timezone          int          `json:"timezone,omitempty" gorethink:"timezone,omitempty"`
	Avatar            string       `json:"avatar,omitempty" gorethink:"avatar,omitempty"`
	PublicAvatar      string       `json:"public_avatar,omitempty" gorethink:"public_avatar,omitempty"`
	Active            bool         `json:"active" gorethink:"active"`
	Info              UserInfo     `json:"info" gorethink:"info"`
	Settings          UserSettings `json:"settings" gorethink:"settings"`
}

// UserInfo stores all personal information about the user
type UserInfo struct {
	Work      string     `json:"work,omitempty" gorethink:"work,omitempty"`
	Education string     `json:"education,omitempty" gorethink:"education,omitempty"`
	Hobbies   string     `json:"hobbies,omitempty" gorethink:"hobbies,omitempty"`
	Books     string     `json:"books,omitempty" gorethink:"books,omitempty"`
	Movies    string     `json:"movies,omitempty" gorethink:"movies,omitempty"`
	TV        string     `json:"tv,omitempty" gorethink:"tv,omitempty"`
	Gender    Gender     `json:"gender,omitempty" gorethink:"gender,omitempty"`
	Websites  []string   `json:"websites,omitempty" gorethink:"websites,omitempty"`
	Status    UserStatus `json:"status,omitempty" gorethink:"status,omitempty"`
	About     string     `json:"about,omitempty" gorethink:"about,omitempty"`
}

// UserSettings stores the user preferences
type UserSettings struct {
	Invisible                   bool           `json:"invisible" gorethink:"invisible"`
	CanReceiveRequests          bool           `json:"can_receive_requests" gorethink:"can_receive_requests"`
	DisplayAvatarBeforeApproval bool           `json:"display_avatar_before_approval" gorethink:"display_avatar_before_approval"`
	NotifyNewComment            bool           `json:"notify_new_comment" gorethink:"notify_new_comment"`
	NotifyNewCommentOthers      bool           `json:"notify_new_comment_others" gorethink:"notify_new_comment_others"`
	AllowPostsInMyProfile       bool           `json:"allow_posts_in_my_profile" gorethink:"allow_posts_in_my_profile"`
	AllowCommentsInPosts        bool           `json:"allow_comments_in_posts" gorethink:"allow_comments_in_posts"`
	DisplayEmail                bool           `json:"display_email" gorethink:"display_email"`
	PasswordRecoveryMethod      RecoveryMethod `json:"recovery_method" gorethink:"recovery_method"`
	RecoveryQuestion            string         `json:"recovery_question,omitempty" gorethink:"recovery_question,omitempty"`
	RecoveryAnswer              string         `json:"recovery_answer,omitempty" gorethink:"recovery_answer,omitempty"`
}

// Save inserts the User instance if it hasn't been reated yet ot updates it if it has
func (u *User) Save(conn *Connection) (bool, error) {
	var count int64
	var err error
	var res *r.ResultRow

	// Check if username is already in use
	if u.ID != "" {
		res, err = conn.Db.Table("user").
			Filter(r.Row.Field("username").
			Eq(u.Username).
			And(r.Row.Field("id").
			Ne(u.ID))).Count().RunRow(conn.Session)
	} else {
		res, err = conn.Db.Table("user").Filter(r.Row.Field("username").Eq(u.Username)).Count().RunRow(conn.Session)
	}

	if err != nil {
		return false, err
	}
	res.Scan(&count)

	if count > 0 {
		return false, nil
	}

	// That means we're creating an user
	if u.ID != "" {
		info := UserInfo{}

		settings := UserSettings{}
		settings.Invisible = true
		settings.CanReceiveRequests = false
		settings.DisplayAvatarBeforeApproval = false
		settings.NotifyNewComment = false
		settings.NotifyNewCommentOthers = false
		settings.AllowPostsInMyProfile = false
		settings.AllowCommentsInPosts = false
		settings.DisplayEmail = false
		settings.PasswordRecoveryMethod = RecoveryNone

		u.Info = info
		u.Settings = settings
	}

	success, err, ID := conn.Save("user", u.ID, u)
	if err != nil {
		return false, err
	}

	if !success {
		return false, nil
	}

	if u.ID == "" {
		u.ID = ID
	}

	return true, nil
}

// Remove deletes the user instance
func (u *User) Remove(conn *Connection) (bool, error) {
	return conn.Remove("user", u.ID)
}

// SetPassword sets a new encrypted password for the user
func (u *User) SetPassword(password string) error {
	pwBytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return err
	}

	u.Password = string(pwBytes[:])
	return nil
}

// CheckPassword checks if the given password matches the current password hash
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

func (u *User) table(conn *Connection) r.RqlTerm {
	return conn.Db.Table("user")
}
