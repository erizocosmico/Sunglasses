package mask

import (
	"errors"
	"labix.org/v2/mgo/bson"
	"net/http"
	"net/mail"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// CreateAccount creates a new user account
func CreateAccount(c Context) {
	var (
		username                = c.Form("username")
		password                = c.Form("password")
		passwordRepeat          = c.Form("password_repeat")
		question, answer, email string
		errorList               = make([]string, 0)
		codeList                = make([]int, 0)
		responseStatus          = 400
	)

	recoveryMethod, err := strconv.ParseInt(c.Form("recovery_method"), 10, 0)
	if err != nil {
		recoveryMethod = RecoveryNone
	}

	switch RecoveryMethod(recoveryMethod) {
	case RecoveryNone:
		break
	case RecoverByEMail:
		email = c.Form("email")

		if _, err := mail.ParseAddress(email); err != nil {
			errorList = append(errorList, MsgInvalidEmail)
			codeList = append(codeList, CodeInvalidEmail)
		}
		break
	case RecoverByQuestion:
		question = c.Form("recovery_question")
		answer = c.Form("recovery_answer")

		if question == "" || answer == "" {
			errorList = append(errorList, MsgInvalidRecoveryQuestion)
			codeList = append(codeList, CodeInvalidRecoveryQuestion)
		}
		break
	default:
		errorList = append(errorList, MsgInvalidRecoveryMethod)
		codeList = append(codeList, CodeInvalidRecoveryMethod)
	}

	reg, err := regexp.Compile("^[a-zA-Z_0-9]{2,30}$")
	if err != nil {
		errorList = append(errorList, MsgInvalidUsername)
		codeList = append(codeList, CodeInvalidUsername)
	}

	if !reg.MatchString(username) {
		errorList = append(errorList, MsgInvalidUsername)
		codeList = append(codeList, CodeInvalidUsername)
	}

	if strlen(password) < 6 {
		errorList = append(errorList, MsgPasswordLength)
		codeList = append(codeList, CodePasswordLength)
	}

	if password != passwordRepeat {
		errorList = append(errorList, MsgPasswordMatch)
		codeList = append(codeList, CodePasswordMatch)
	}

	if len(errorList) == 0 {
		user := NewUser()
		user.Username = username
		user.SetPassword(password)
		user.Settings.PasswordRecoveryMethod = RecoveryMethod(recoveryMethod)

		switch recoveryMethod {
		case RecoverByEMail:
			if err := user.SetEmail(email); err != nil {
				user.Settings.PasswordRecoveryMethod = RecoveryNone
			}
			break
		case RecoverByQuestion:
			user.Settings.RecoveryQuestion = question
			user.Settings.RecoveryAnswer = answer
			break
		}

		if err = user.Save(c.Conn); err == nil {
			c.Request.PostForm.Add("token_type", "session")
			c.Request.PostForm.Add("username", user.Username)
			c.Request.PostForm.Add("password", password)
			GetUserToken(c)
			return
		} else {
			if err.Error() == "username already in use" {
				errorList = append(errorList, MsgUsernameTaken)
				codeList = append(codeList, CodeInvalidRecoveryQuestion)
			} else {
				responseStatus = 500
				errorList = append(errorList, MsgUnexpected)
				codeList = append(codeList, CodeInvalidRecoveryQuestion)
			}
		}
	}

	c.Errors(responseStatus, codeList, errorList)
}

// GetAccountInfo retrieves the info of the user
func GetAccountInfo(c Context) {
	if c.User != nil {
		c.Success(200, map[string]interface{}{
			"error":        false,
			"account_info": c.User.Info,
		})
		return
	}

	c.Error(400, CodeInvalidData, MsgInvalidData)
}

// UpdateAccountInfo updates the user's information
func UpdateAccountInfo(c Context) {
	if c.User != nil {
		info := UserInfo{}

		for _, v := range c.Request.PostForm {
			for _, f := range v {
				if strlen(f) > 500 {
					c.Error(400, CodeInvalidInfoLength, MsgInvalidInfoLength)
					return
				}
			}
		}

		if v, ok := c.Request.PostForm["websites"]; ok {
			sites := make([]string, 0, len(v))

			for _, site := range v {
				if !strings.HasPrefix(site, "http://") && !strings.HasPrefix(site, "https://") {
					site = "http://" + site
				}

				if !isValidURL(site) {
					c.Error(400, CodeInvalidWebsites, MsgInvalidWebsites)
					return
				}

				sites = append(sites, site)
			}
			info.Websites = sites
		}

		if gender := c.Form("gender"); gender != "" {
			gender, err := strconv.ParseInt(gender, 10, 8)
			if err == nil {
				if gender == Male || gender == Female || gender == Other {
					info.Gender = Gender(gender)
				} else {
					err = errors.New("invalid gender")
				}
			}

			if err != nil {
				c.Error(400, CodeInvalidGender, MsgInvalidGender)
				return
			}
		}

		if status := c.Form("status"); status != "" {
			status, err := strconv.ParseInt(status, 10, 8)
			if err == nil {
				if status >= 0 && status <= 4 {
					info.Status = UserStatus(status)
				} else {
					err = errors.New("invalid statusr")
				}
			}

			if err != nil {
				c.Error(400, CodeInvalidStatus, MsgInvalidStatus)
				return
			}
		}

		info.Work = strings.TrimSpace(c.Form("work"))
		info.Education = strings.TrimSpace(c.Form("education"))
		info.Hobbies = strings.TrimSpace(c.Form("hobbies"))
		info.Books = strings.TrimSpace(c.Form("books"))
		info.Movies = strings.TrimSpace(c.Form("movies"))
		info.TV = strings.TrimSpace(c.Form("tv"))
		info.About = strings.TrimSpace(c.Form("about"))

		c.User.Info = info
		if err := c.User.Save(c.Conn); err != nil {
			c.Error(500, CodeUnexpected, MsgUnexpected)
			return
		}

		c.Success(200, map[string]interface{}{
			"message": "User info updated successfully",
		})
		return
	}

	c.Error(400, CodeInvalidData, MsgInvalidData)
}

// GetAccountSettings retrieves the settings of the user
func GetAccountSettings(c Context) {
	if c.User != nil {
		c.Success(200, map[string]interface{}{
			"account_settings": c.User.Settings,
		})
		return
	}

	c.Error(400, CodeInvalidData, MsgInvalidData)
}

// UpdateAccountSettings updates the user's settings
func UpdateAccountSettings(c Context) {
	if c.User != nil {
		s := UserSettings{}
		getPrivacy := func(r *http.Request, kind string) (PrivacySettings, error) {
			p := PrivacySettings{}

			if privacyType := c.Form("privacy_" + kind + "_type"); privacyType != "" {
				pType, err := strconv.ParseInt(privacyType, 10, 8)
				if !isValidPrivacyType(PrivacyType(pType)) || err != nil {
					return p, errors.New("invalid data provided")
				}

				p.Type = PrivacyType(pType)
			} else {
				return p, errors.New("privacy type is required")
			}

			if users, ok := r.Form["privacy_"+kind+"_users"]; ok {
				uids := make([]bson.ObjectId, 0, len(users))
				for _, u := range users {
					if !bson.IsObjectIdHex(u) {
						return p, errors.New("invalid data provided")
					}
					uids = append(uids, bson.ObjectIdHex(u))
				}

				count, err := c.Count("follows", bson.M{"user_from": c.User.ID, "user_to": bson.M{"$in": uids}})
				if err != nil || count != len(p.Users) {
					count2, err := c.Count("follows", bson.M{"user_to": c.User.ID, "user_from": bson.M{"$in": uids}})
					if err != nil || count+count2 != len(uids) {
						return p, errors.New("invalid user list provided")
					}
				}

				p.Users = uids
			} else if p.Type > PrivacyNone {
				return p, errors.New("users param required for this privacy type")
			}

			return p, nil
		}

		s.OverrideDefaultPrivacy = c.GetBoolean("override_default_privacy")
		if s.OverrideDefaultPrivacy {
			p, err := getPrivacy(c.Request, "status")
			if err != nil {
				c.Error(400, CodeInvalidPrivacySettings, MsgInvalidPrivacySettings)
				return
			}

			s.DefaultStatusPrivacy = p
		} else {
			for _, k := range []string{"status", "video", "photo", "link", "album"} {
				p, err := getPrivacy(c.Request, k)
				if err != nil {
					c.Error(400, CodeInvalidPrivacySettings, MsgInvalidPrivacySettings)
					return
				}

				switch k {
				case "status":
					s.DefaultStatusPrivacy = p
					break
				case "video":
					s.DefaultVideoPrivacy = p
					break
				case "photo":
					s.DefaultPhotoPrivacy = p
					break
				case "link":
					s.DefaultLinkPrivacy = p
					break
				case "album":
					s.DefaultAlbumPrivacy = p
					break
				}
			}
		}

		s.Invisible = c.GetBoolean("invisible")
		s.CanReceiveRequests = c.GetBoolean("can_receive_requests")
		s.FollowApprovalRequired = c.GetBoolean("follow_approval_required")
		s.DisplayAvatarBeforeApproval = c.GetBoolean("display_avatar_before_approval")
		s.NotifyNewComment = c.GetBoolean("notify_new_comment")
		s.NotifyNewCommentOthers = c.GetBoolean("notify_new_comment_others")
		s.NotifyPostsInMyProfile = c.GetBoolean("notify_new_posts_in_my_profile")
		s.NotifyLikes = c.GetBoolean("notify_likes")
		s.AllowPostsInMyProfile = c.GetBoolean("allow_posts_in_my_profile")
		s.AllowCommentsInPosts = c.GetBoolean("allow_comments_in_posts")
		s.DisplayEmail = c.GetBoolean("display_email")
		s.DisplayInfoFollowersOnly = c.GetBoolean("display_info_followers_only")

		if recoveryMethod := c.Form("recovery_method"); recoveryMethod != "" {
			method, err := strconv.ParseInt(recoveryMethod, 10, 8)
			if err != nil || (method != RecoveryNone && method != RecoverByQuestion && method != RecoverByEMail) {
				s.PasswordRecoveryMethod = RecoveryNone
			}

			s.PasswordRecoveryMethod = RecoveryMethod(method)
			if s.PasswordRecoveryMethod == RecoverByQuestion {
				s.RecoveryQuestion = c.Form("recovery_question")
				s.RecoveryAnswer = c.Form("recovery_answer")

				if s.RecoveryAnswer == "" || s.RecoveryQuestion == "" {
					c.Error(400, CodeInvalidRecoveryQuestion, MsgInvalidRecoveryQuestion)
					return
				}
			}
		} else {
			s.PasswordRecoveryMethod = RecoveryNone
		}

		c.User.Settings = s
		if err := c.User.Save(c.Conn); err != nil {
			c.Error(500, CodeUnexpected, MsgUnexpected)
			return
		}

		c.Success(200, map[string]interface{}{
			"message": "User settings updated successfully",
		})
		return
	}

	c.Error(400, CodeInvalidData, MsgInvalidData)
}

// UpldateProfilePicture updates the current profile picture of the user
func UpdateProfilePicture(c Context, config *Config) {
	if c.User != nil {
		file, err := RetrieveUploadedImage(c.Request, "account_picture")
		if err != nil {
			code, msg := CodeAndMessageForUploadError(err)
			c.Error(400, code, msg)
			return
		}

		imagePath, thumbnailPath, err := StoreImage(file, ProfileUploadOptions(config))
		if err != nil {
			code, msg := CodeAndMessageForUploadError(err)
			c.Error(400, code, msg)
			return
		}

		var prevAvatar, prevThumbnail string

		if c.Form("picture_type") == "public" {
			prevAvatar = c.User.PublicAvatar
			c.User.PublicAvatar = imagePath
			prevThumbnail = c.User.PublicAvatarThumbnail
			c.User.PublicAvatarThumbnail = thumbnailPath
		} else {
			prevAvatar = c.User.Avatar
			c.User.Avatar = imagePath
			prevThumbnail = c.User.AvatarThumbnail
			c.User.AvatarThumbnail = thumbnailPath
		}

		if c.User.Save(c.Conn); err != nil {
			c.Error(500, CodeUnexpected, MsgUnexpected)

			os.Remove(toLocalImagePath(imagePath, config))
			os.Remove(toLocalThumbnailPath(thumbnailPath, config))
			return
		} else {
			if prevAvatar != "" {
				os.Remove(toLocalImagePath(prevAvatar, config))
			}

			if prevThumbnail != "" {
				os.Remove(toLocalThumbnailPath(prevThumbnail, config))
			}
		}

		c.Success(200, map[string]interface{}{
			"message": "User settings updated successfully",
		})
		return
	}

	c.Error(400, CodeInvalidData, MsgInvalidData)
}
