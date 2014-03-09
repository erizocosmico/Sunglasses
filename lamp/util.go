package lamp

import (
	"code.google.com/p/go.crypto/bcrypt"
	"github.com/martini-contrib/render"
)

// RenderError renders an error message
func RenderError(resp render.Render, code, status int, message string) {
	resp.JSON(status, map[string]interface{}{
		"error":   true,
		"code":    code,
		"message": message,
	})
}

// RenderErrors renders a json response with an array of errors
func RenderErrors(resp render.Render, status int, messages []string) {
	resp.JSON(status, map[string]interface{}{
		"error":    true,
		"messages": messages,
	})
}

func crypt(str string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(str), 10)
	if err != nil {
		return "", err
	}

	return string(bytes[:]), nil
}
