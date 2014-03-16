package mask

import (
	"github.com/martini-contrib/render"
	"net/http"
	"net/mail"
	"regexp"
	"strconv"
)

// CreateUser creates a new user account
func CreateUser(r *http.Request, conn *Connection, res render.Render) {
	var (
		username                = r.PostFormValue("username")
		password                = r.PostFormValue("password")
		passwordRepeat          = r.PostFormValue("password_repeat")
		question, answer, email string
		errorList               = make([]string, 0)
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
			errorList = append(errorList, "Invalid email address")
		}
		break
	case RecoverByQuestion:
		question = r.PostFormValue("recovery_question")
		answer = r.PostFormValue("recovery_answer")

		if question == "" || answer == "" {
			errorList = append(errorList, "Recovery question and answer can not be blank")
		}
		break
	default:
		errorList = append(errorList, "Invalid recovery method")
	}

	reg, err := regexp.Compile("^[a-zA-Z_0-9]{2,30}$")
	if err != nil {
		errorList = append(errorList, "Invalid username")
	}

	if !reg.MatchString(username) {
		errorList = append(errorList, "Invalid username")
	}

	if password == "" {
		errorList = append(errorList, "Password can not be blank")
	}

	if password != passwordRepeat {
		errorList = append(errorList, "Passwords don't match")
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
			res.JSON(200, map[string]interface{}{
				"error":   false,
				"message": "User has been successfully created",
			})
			return
		} else {
			if err.Error() == "username already in use" {
				errorList = append(errorList, "The username is taken")
			} else {
				responseStatus = 500
				errorList = append(errorList, "Unexpected error occurred")
			}
		}
	}

	RenderErrors(res, responseStatus, errorList)
}
