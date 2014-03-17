package mask

import (
	"github.com/martini-contrib/render"
	"net/http"
	"net/mail"
	"regexp"
	"strconv"
)

// CreateAccount creates a new user account
func CreateAccount(r *http.Request, conn *Connection, res render.Render) {
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

	if len(password) < 6 {
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
			res.JSON(200, map[string]interface{}{
				"error":   false,
				"message": "User has been successfully created",
			})
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
