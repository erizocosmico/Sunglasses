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

func crypt(str string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(str), 10)
	if err != nil {
		return "", err
	}

	return string(bytes[:]), nil
}
