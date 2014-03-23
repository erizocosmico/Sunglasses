package mask

import (
	"errors"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	"labix.org/v2/mgo/bson"
	"net/http"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
)

// CreateAccount creates a new user account
func CreateAccount(r *http.Request, conn *Connection, res render.Render, s sessions.Session) {
	var (
		username                = r.PostFormValue("username")
		password                = r.PostFormValue("password")
		passwordRepeat          = r.PostFormValue("password_repeat")
		question, answer, email string
		errorList               = make([]string, 0)
		codeList                = make([]int, 0)
		responseStatus          = 400
	)

	recoveryMethod, err := strconv.ParseInt(r.PostFormValue("recovery_method"), 10, 0)
	if err != nil {
		recoveryMethod = RecoveryNone
	}

	switch RecoveryMethod(recoveryMethod) {
	case RecoveryNone:
		break
	case RecoverByEMail:
		email = r.PostFormValue("email")

		if _, err := mail.ParseAddress(email); err != nil {
			errorList = append(errorList, MsgInvalidEmail)
			codeList = append(codeList, CodeInvalidEmail)
		}
		break
	case RecoverByQuestion:
		question = r.PostFormValue("recovery_question")
		answer = r.PostFormValue("recovery_answer")

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

		if err = user.Save(conn); err == nil {
			r.PostForm.Add("token_type", "session")
			r.PostForm.Add("username", user.Username)
			r.PostForm.Add("password", password)
			GetUserToken(r, conn, res, s)
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

	RenderErrors(res, responseStatus, codeList, errorList)
}

// GetAccountInfo retrieves the info of the user
func GetAccountInfo(r *http.Request, conn *Connection, res render.Render, s sessions.Session) {
	user := GetRequestUser(r, conn, s)

	if user != nil {
		res.JSON(200, map[string]interface{}{
			"error":        false,
			"account_info": user.Info,
		})
		return
	}

	RenderError(res, CodeInvalidData, 400, MsgInvalidData)
}

// UpdateAccountInfo updates the user's information
func UpdateAccountInfo(r *http.Request, conn *Connection, res render.Render, s sessions.Session) {
	user := GetRequestUser(r, conn, s)

	if user != nil {
		info := UserInfo{}

		for _, v := range r.PostForm {
			for _, f := range v {
				if strlen(f) > 500 {
					RenderError(res, CodeInvalidInfoLength, 400, MsgInvalidInfoLength)
					return
				}
			}
		}

		if v, ok := r.PostForm["websites"]; ok {
			sites := make([]string, 0, len(v))

			for _, site := range v {
				if !strings.HasPrefix(site, "http://") && !strings.HasPrefix(site, "https://") {
					site = "http://" + site
				}

				if !isValidURL(site) {
					RenderError(res, CodeInvalidWebsites, 400, MsgInvalidWebsites)
					return
				}

				sites = append(sites, site)
			}
			info.Websites = sites
		}

		if gender := r.PostFormValue("gender"); gender != "" {
			gender, err := strconv.ParseInt(gender, 10, 8)
			if err == nil {
				if gender == Male || gender == Female || gender == Other {
					info.Gender = Gender(gender)
				} else {
					err = errors.New("invalid gender")
				}
			}

			if err != nil {
				RenderError(res, CodeInvalidGender, 400, MsgInvalidGender)
				return
			}
		}

		if status := r.PostFormValue("status"); status != "" {
			status, err := strconv.ParseInt(status, 10, 8)
			if err == nil {
				if status >= 0 && status <= 4 {
					info.Status = UserStatus(status)
				} else {
					err = errors.New("invalid statusr")
				}
			}

			if err != nil {
				RenderError(res, CodeInvalidStatus, 400, MsgInvalidStatus)
				return
			}
		}

		info.Work = strings.TrimSpace(r.PostFormValue("work"))
		info.Education = strings.TrimSpace(r.PostFormValue("education"))
		info.Hobbies = strings.TrimSpace(r.PostFormValue("hobbies"))
		info.Books = strings.TrimSpace(r.PostFormValue("books"))
		info.Movies = strings.TrimSpace(r.PostFormValue("movies"))
		info.TV = strings.TrimSpace(r.PostFormValue("tv"))
		info.About = strings.TrimSpace(r.PostFormValue("about"))

		user.Info = info
		if err := user.Save(conn); err != nil {
			RenderError(res, CodeUnexpected, 500, MsgUnexpected)
			return
		}

		res.JSON(200, map[string]interface{}{
			"error":   false,
			"message": "User info updated successfully",
		})
		return
	}

	RenderError(res, CodeInvalidData, 400, MsgInvalidData)
}

// GetAccountSettings retrieves the settings of the user
func GetAccountSettings(r *http.Request, conn *Connection, res render.Render, s sessions.Session) {
	user := GetRequestUser(r, conn, s)

	if user != nil {
		res.JSON(200, map[string]interface{}{
			"error":            false,
			"account_settings": user.Settings,
		})
		return
	}

	RenderError(res, CodeInvalidData, 400, MsgInvalidData)
}

// UpdateAccountSettings updates the user's settings
func UpdateAccountSettings(r *http.Request, conn *Connection, res render.Render, s sessions.Session) {
	user := GetRequestUser(r, conn, s)

	if user != nil {
		s := UserSettings{}
		getPrivacy := func(r *http.Request, kind string) (PrivacySettings, error) {
			p := PrivacySettings{}

			if privacyType := r.PostFormValue("privacy_" + kind + "_type"); privacyType != "" {
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

				// TODO check if users are followed by the user

				p.Users = uids
			} else if p.Type > PrivacyNone {
				return p, errors.New("users param required for this privacy type")
			}

			return p, nil
		}

		s.OverrideDefaultPrivacy = getBoolean(r, "override_default_privacy")
		if s.OverrideDefaultPrivacy {
			p, err := getPrivacy(r, "status")
			if err != nil {
				RenderError(res, CodeInvalidPrivacySettings, 400, MsgInvalidPrivacySettings)
				return
			}

			s.DefaultStatusPrivacy = p
		} else {
			for _, k := range []string{"status", "video", "photo", "link", "album"} {
				p, err := getPrivacy(r, k)
				if err != nil {
					RenderError(res, CodeInvalidPrivacySettings, 400, MsgInvalidPrivacySettings)
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

		s.Invisible = getBoolean(r, "invisible")
		s.CanReceiveRequests = getBoolean(r, "can_receive_requests")
		s.FollowApprovalRequired = getBoolean(r, "follow_approval_required")
		s.DisplayAvatarBeforeApproval = getBoolean(r, "display_avatar_before_approval")
		s.NotifyNewComment = getBoolean(r, "notify_new_comment")
		s.NotifyNewCommentOthers = getBoolean(r, "notify_new_comment_others")
		s.NotifyPostsInMyProfile = getBoolean(r, "notify_new_posts_in_my_profile")
		s.NotifyLikes = getBoolean(r, "notify_likes")
		s.AllowPostsInMyProfile = getBoolean(r, "allow_posts_in_my_profile")
		s.AllowCommentsInPosts = getBoolean(r, "allow_comments_in_posts")
		s.DisplayEmail = getBoolean(r, "display_email")
		s.DisplayInfoFollowersOnly = getBoolean(r, "display_info_followers_only")

		if recoveryMethod := r.PostFormValue("recovery_method"); recoveryMethod != "" {
			method, err := strconv.ParseInt(recoveryMethod, 10, 8)
			if err != nil || (method != RecoveryNone && method != RecoverByQuestion && method != RecoverByEMail) {
				s.PasswordRecoveryMethod = RecoveryNone
			}

			s.PasswordRecoveryMethod = RecoveryMethod(method)
			if s.PasswordRecoveryMethod == RecoverByQuestion {
				s.RecoveryQuestion = r.PostFormValue("recovery_question")
				s.RecoveryAnswer = r.PostFormValue("recovery_answer")

				if s.RecoveryAnswer == "" || s.RecoveryQuestion == "" {
					RenderError(res, CodeInvalidRecoveryQuestion, 400, MsgInvalidRecoveryQuestion)
					return
				}
			}
		} else {
			s.PasswordRecoveryMethod = RecoveryNone
		}

		user.Settings = s
		if err := user.Save(conn); err != nil {
			RenderError(res, CodeUnexpected, 500, MsgUnexpected)
			return
		}

		res.JSON(200, map[string]interface{}{
			"error":   false,
			"message": "User settings updated successfully",
		})
		return
	}

	RenderError(res, CodeInvalidData, 400, MsgInvalidData)
}
