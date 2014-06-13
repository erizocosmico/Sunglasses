package handlers

import (
	"errors"
	. "github.com/mvader/sunglasses/error"
	"github.com/mvader/sunglasses/middleware"
	"github.com/mvader/sunglasses/models"
	"github.com/mvader/sunglasses/modules/timeline"
	"github.com/mvader/sunglasses/modules/upload"
	"github.com/mvader/sunglasses/util"
	"labix.org/v2/mgo/bson"
	"net/http"
	"net/mail"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// CreateAccount is a handler in charge of creating an user account. After a successful registration the user will be automatically logged in.
//
// This handler needs the following parameters:
// - username: The username provided by the user
// - password: The password chosen by the user
// - password_repeat: The password, again
// - recovery_method (optional): The recovery method chosen by the user (look at the values for the type RecoveryMethod on the models package)
// - recovery_answer and recovery_question (optional): A recovery answer and question if the provided recovery method is question.
func CreateAccount(c middleware.Context) {
	// TODO: Max retries
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
		recoveryMethod = models.RecoveryNone
	}

	switch models.RecoveryMethod(recoveryMethod) {
	case models.RecoveryNone:
		break
	case models.RecoverByEMail:
		email = c.Form("email")

		if _, err := mail.ParseAddress(email); err != nil {
			errorList = append(errorList, MsgInvalidEmail)
			codeList = append(codeList, CodeInvalidEmail)
		}
		break
	case models.RecoverByQuestion:
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

	if util.Strlen(password) < 6 {
		errorList = append(errorList, MsgPasswordLength)
		codeList = append(codeList, CodePasswordLength)
	}

	if password != passwordRepeat {
		errorList = append(errorList, MsgPasswordMatch)
		codeList = append(codeList, CodePasswordMatch)
	}

	if len(errorList) == 0 {
		user := models.NewUser()
		user.Username = username
		user.SetPassword(password)
		user.Settings.PasswordRecoveryMethod = models.RecoveryMethod(recoveryMethod)

		switch recoveryMethod {
		case models.RecoverByEMail:
			if err := user.SetEmail(email); err != nil {
				user.Settings.PasswordRecoveryMethod = models.RecoveryNone
			}
			break
		case models.RecoverByQuestion:
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
				codeList = append(codeList, CodeUnexpected)
			}
		}
	}

	c.Errors(responseStatus, codeList, errorList)
}

// IsUsernameTaken checks if the given username is already taken
func IsUsernameTaken(c middleware.Context) {
	count, _ := c.Count("users", bson.M{"username_lower": strings.ToLower(c.Form("username"))})

	c.Success(200, map[string]interface{}{"taken": count > 0})
}

// GetAccountInfo retrieves the info of the user
func GetAccountInfo(c middleware.Context) {
	c.Success(200, map[string]interface{}{
		"account_info": c.User.Info,
	})
}

// UpdateAccountInfo updates the user's information
func UpdateAccountInfo(c middleware.Context) {
	info := c.User.Info

	// If any field is more than 500 characters long directly return an error
	for _, v := range c.Request.PostForm {
		for _, f := range v {
			if util.Strlen(f) > 500 {
				c.Error(400, CodeInvalidInfoLength, MsgInvalidInfoLength)
				return
			}
		}
	}

	// Check user websites
	if v, ok := c.Request.PostForm["websites"]; ok {
		sites := make([]string, 0, len(v))

		for _, site := range v {
			if !strings.HasPrefix(site, "http://") && !strings.HasPrefix(site, "https://") {
				site = "http://" + site
			}

			if !util.IsValidURL(site) {
				c.Error(400, CodeInvalidWebsites, MsgInvalidWebsites)
				return
			}

			sites = append(sites, site)
		}
		info.Websites = sites
	}

	// Check gender
	if gender := c.Form("gender"); gender != "" {
		gender, err := strconv.ParseInt(gender, 10, 8)
		if err == nil {
			if gender == models.Male || gender == models.Female || gender == models.Other {
				info.Gender = models.Gender(gender)
			} else {
				err = errors.New("invalid gender")
			}
		}

		if err != nil {
			c.Error(400, CodeInvalidGender, MsgInvalidGender)
			return
		}
	}

	// Check status
	if status := c.Form("status"); status != "" {
		status, err := strconv.ParseInt(status, 10, 8)
		if err == nil {
			if status >= 0 && status <= 4 {
				info.Status = models.UserStatus(status)
			} else {
				err = errors.New("invalid status")
			}
		}

		if err != nil {
			c.Error(400, CodeInvalidStatus, MsgInvalidStatus)
			return
		}
	}

	// The rest of the fields need no validation since their length was already checked before
	info.Work = strings.TrimSpace(c.Form("work"))
	info.Education = strings.TrimSpace(c.Form("education"))
	info.Hobbies = strings.TrimSpace(c.Form("hobbies"))
	info.Books = strings.TrimSpace(c.Form("books"))
	info.Movies = strings.TrimSpace(c.Form("movies"))
	info.TV = strings.TrimSpace(c.Form("tv"))
	info.About = strings.TrimSpace(c.Form("about"))

	// Update the user info in the database
	c.User.Info = info
	if err := c.User.Save(c.Conn); err != nil {
		c.Error(500, CodeUnexpected, MsgUnexpected)
		return
	}

	c.Success(200, map[string]interface{}{
		"message": "User info updated successfully",
	})
}

// GetAccountSettings retrieves the settings of the user
func GetAccountSettings(c middleware.Context) {
	c.Success(200, map[string]interface{}{
		"account_info": c.User.Info,
	})
}

// UpdateAccountData updates some data about the user like the names and the preferred language
func UpdateAccountData(c middleware.Context) {
	var (
		publicName    = c.Form("public_name")
		privateName   = c.Form("private_name")
		preferredLang = c.Form("preferred_lang")
	)

	if util.Strlen(publicName) > 35 || util.Strlen(privateName) > 35 {
		c.Error(400, CodeInvalidData, MsgInvalidData)
		return
	}

	// TODO: Check lang is available

	c.User.PublicName = publicName
	c.User.PrivateName = privateName
	c.User.PreferredLanguage = preferredLang

	if err := c.User.Save(c.Conn); err != nil {
		c.Error(500, CodeUnexpected, MsgUnexpected)
		return
	}

	c.Success(200, map[string]interface{}{
		"message": "Data updated successfully",
	})
}

// UpdateAccountPassword updates the user password
func UpdateAccountPassword(c middleware.Context) {
	var (
		password        = c.Form("password")
		passwordRepeat  = c.Form("password_repeat")
		currentPassword = c.Form("current_password")
	)

	if !c.User.CheckPassword(currentPassword) {
		c.Error(400, CodePasswordCurrentError, MsgPasswordCurrentError)
		return
	}

	if util.Strlen(password) < 6 {
		c.Error(400, CodePasswordLength, MsgPasswordLength)
		return
	}

	if password != passwordRepeat {
		c.Error(400, CodePasswordMatch, MsgPasswordMatch)
		return
	}

	if err := c.User.SetPassword(password); err != nil {
		c.Error(500, CodeUnexpected, MsgUnexpected)
		return
	}

	if err := c.User.Save(c.Conn); err != nil {
		c.Error(500, CodeUnexpected, MsgUnexpected)
		return
	}

	c.Success(200, map[string]interface{}{
		"message": "Password updated successfully",
	})
}

// UpdateAccountSettings updates the user's settings
func UpdateAccountSettings(c middleware.Context) {
	s := c.User.Settings

	getPrivacy := func(r *http.Request, kind string) (models.PrivacySettings, error) {
		p := models.PrivacySettings{}

		// Get the privacy type for the current kind of privacy
		if privacyType := c.Form("privacy_" + kind + "_type"); privacyType != "" {
			pType, err := strconv.ParseInt(privacyType, 10, 8)
			if !models.IsValidPrivacyType(models.PrivacyType(pType)) || err != nil {
				return p, errors.New("invalid data provided")
			}

			p.Type = models.PrivacyType(pType)
		} else {
			return p, errors.New("privacy type is required")
		}

		// Get users for the current kind of privacy
		if users, ok := r.Form["privacy_"+kind+"_users"]; ok {
			uids := make([]bson.ObjectId, 0, len(users))
			for _, u := range users {
				if !bson.IsObjectIdHex(u) {
					return p, errors.New("invalid data provided")
				}
				uids = append(uids, bson.ObjectIdHex(u))
			}

			// Check that you only selected users you follow or that follow you
			count, err := c.Count("follows", bson.M{"user_from": c.User.ID, "user_to": bson.M{"$in": uids}})
			if err != nil || count != len(p.Users) {
				count2, err := c.Count("follows", bson.M{"user_to": c.User.ID, "user_from": bson.M{"$in": uids}})
				if err != nil || count+count2 != len(uids) {
					return p, errors.New("invalid user list provided")
				}
			}

			p.Users = uids
		} else if p.Type > models.PrivacyNone {
			return p, errors.New("users param required for this privacy type")
		}

		return p, nil
	}

	// If override_default_privacy is true the privacy settings used will be just status for
	// all privacy kinds
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

	// Get other boolean settings
	s.Invisible = c.GetBoolean("invisible")
	s.CanReceiveRequests = c.GetBoolean("can_receive_requests")
	s.FollowApprovalRequired = c.GetBoolean("follow_approval_required")
	s.DisplayAvatarBeforeApproval = c.GetBoolean("display_avatar_before_approval")
	s.NotifyNewComment = c.GetBoolean("notify_new_comment")
	s.NotifyNewCommentOthers = c.GetBoolean("notify_new_comment_others")
	s.NotifyPostsInMyProfile = c.GetBoolean("notify_posts_in_my_profile")
	s.NotifyLikes = c.GetBoolean("notify_likes")
	s.AllowPostsInMyProfile = c.GetBoolean("allow_posts_in_my_profile")
	s.AllowCommentsInPosts = c.GetBoolean("allow_comments_in_posts")
	s.DisplayInfoFollowersOnly = c.GetBoolean("display_info_followers_only")

	// Recovery method settings. If no recovery_method value is sent or an invalid recovery_method
	// is entered RecoveryNone will be set!
	if recoveryMethod := c.Form("recovery_method"); recoveryMethod != "" {
		method, err := strconv.ParseInt(recoveryMethod, 10, 8)
		if err != nil || (method != models.RecoveryNone && method != models.RecoverByQuestion && method != models.RecoverByEMail) {
			s.PasswordRecoveryMethod = models.RecoveryNone
		}

		s.PasswordRecoveryMethod = models.RecoveryMethod(method)
		if s.PasswordRecoveryMethod == models.RecoverByQuestion {
			s.RecoveryQuestion = c.Form("recovery_question")
			s.RecoveryAnswer = c.Form("recovery_answer")

			if s.RecoveryAnswer == "" || s.RecoveryQuestion == "" {
				c.Error(400, CodeInvalidRecoveryQuestion, MsgInvalidRecoveryQuestion)
				return
			}
		}
		// TODO: Recover by email
	} else {
		s.PasswordRecoveryMethod = models.RecoveryNone
	}

	// Update the user settings in the database
	c.User.Settings = s
	if err := c.User.Save(c.Conn); err != nil {
		c.Error(500, CodeUnexpected, MsgUnexpected)
		return
	}

	c.Success(200, map[string]interface{}{
		"message": "User settings updated successfully",
	})
}

// UpldateProfilePicture updates the current profile picture of the user
func UpdateProfilePicture(c middleware.Context) {
	// Try to retrieve the uploaded image
	file, err := upload.RetrieveUploadedImage(c.Request, "account_picture")
	if err != nil {
		code, msg := upload.CodeAndMessageForUploadError(err)
		c.Error(400, code, msg)
		return
	}

	// Try to upload the image
	imagePath, thumbnailPath, err := upload.StoreImage(file, upload.ProfileUploadOptions(c.Config))
	if err != nil {
		code, msg := upload.CodeAndMessageForUploadError(err)
		c.Error(400, code, msg)
		return
	}

	var prevAvatar, prevThumbnail string
	// If there was any previously set avatar we store its value to destroy them later
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

	// Update the data
	if c.User.Save(c.Conn); err != nil {
		c.Error(500, CodeUnexpected, MsgUnexpected)

		// If it's been an error updating the changes in the database the files
		// need to be erased
		os.Remove(upload.ToLocalImagePath(imagePath, c.Config))
		os.Remove(upload.ToLocalThumbnailPath(thumbnailPath, c.Config))
		return
	} else {
		// Remove previous avatar
		if prevAvatar != "" {
			os.Remove(upload.ToLocalImagePath(prevAvatar, c.Config))
		}

		if prevThumbnail != "" {
			os.Remove(upload.ToLocalThumbnailPath(prevThumbnail, c.Config))
		}
	}

	c.Success(200, map[string]interface{}{
		"message": "User settings updated successfully",
	})
}

// DestroyAccount destroys the user account and all its related content such as comments, posts, images, etc.
func DestroyAccount(c middleware.Context) {
	// TODO: Log errors removing images and push image remove to tasks
	confirmed := c.GetBoolean("confirmed")

	if confirmed {
		var (
			p   models.Post
			cmt models.Comment
		)

		// Destroy all user tokens
		c.RemoveAll("tokens", bson.M{"user_id": c.User.ID})

		// Destroy all user material (posts + comments + images)
		iter := c.Find("posts", bson.M{"user_id": c.User.ID}).Iter()
		for iter.Next(&p) {
			if p.Type == models.PostPhoto {
				os.Remove(upload.ToLocalImagePath(p.PhotoURL, c.Config))
				os.Remove(upload.ToLocalThumbnailPath(p.Thumbnail, c.Config))
			}

			go timeline.PropagatePostsOnDeletion(c, p.ID)
		}

		iter.Close()

		// Destroy all user comments
		iter = c.Find("comments", bson.M{"user_id": c.User.ID}).Iter()
		for iter.Next(&cmt) {
			go timeline.PropagatePostOnCommentDeleted(c, cmt.PostID, cmt.ID)
		}

		iter.Close()

		// Remove other stuff related to the user
		c.RemoveAll("posts", bson.M{"user_id": c.User.ID})
		c.RemoveAll("comments", bson.M{"user_id": c.User.ID})
		c.RemoveAll("follows", bson.M{"user_to": c.User.ID})
		c.RemoveAll("follows", bson.M{"user_from": c.User.ID})
		c.RemoveAll("blocks", bson.M{"user_to": c.User.ID})
		c.RemoveAll("blocks", bson.M{"user_from": c.User.ID})
		c.RemoveAll("timelines", bson.M{"user_id": c.User.ID})
		c.RemoveAll("likes", bson.M{"user_id": c.User.ID})
		go timeline.PropagatePostOnUserDeleted(c, c.User.ID)

		// Logout user
		c.Session.Values["user_token"] = nil
		c.Session.Values["csrf_key"] = nil
		c.Session.Save(c.Request, c.ResponseWriter)

		// Destroy user
		c.Remove("users", bson.M{"_id": c.User.ID})

		// Destroy all user notifications
		c.RemoveAll("notifications", bson.M{"user_id": c.User.ID})

		// Remove user avatars
		if c.User.Avatar != "" {
			os.Remove(upload.ToLocalImagePath(c.User.Avatar, c.Config))
		}

		if c.User.AvatarThumbnail != "" {
			os.Remove(upload.ToLocalThumbnailPath(c.User.AvatarThumbnail, c.Config))
		}

		if c.User.PublicAvatar != "" {
			os.Remove(upload.ToLocalImagePath(c.User.PublicAvatar, c.Config))
		}

		if c.User.PublicAvatarThumbnail != "" {
			os.Remove(upload.ToLocalThumbnailPath(c.User.PublicAvatarThumbnail, c.Config))
		}

		c.Success(200, map[string]interface{}{"message": "User account has been successfully destroyed"})
	} else {
		c.Success(200, map[string]interface{}{"message": "User account has not been destroyed"})
	}
}

// ChangeLanguage changes the default language for the user
func ChangeLanguage(c middleware.Context) {
	preferredLang := ""
	lang := c.Form("lang")

	for _, l := range models.AvailableLanguages {
		if lang == l {
			preferredLang = l
		}
	}

	if preferredLang == "" {
		c.Error(400, CodeInvalidData, MsgInvalidData)
		return
	}

	c.User.PreferredLanguage = preferredLang
	if err := c.User.Save(c.Conn); err != nil {
		c.Error(500, CodeUnexpected, MsgUnexpected)
		return
	}

	c.Success(200, map[string]interface{}{"message": "Language changed successfully"})
}
