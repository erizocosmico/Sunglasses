package mask

import (
	"github.com/codegangsta/martini"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
)

// NewApp instantiates the application
func NewApp(configPath string) (*martini.ClassicMartini, string, error) {
	// Create app
	m := martini.Classic()

	// Create config service
	config, err := NewConfig(configPath)
	if err != nil {
		return nil, "", err
	}

	// Create database service
	conn, err := NewDatabaseConn(config)
	if err != nil {
		return nil, "", err
	}

	// Map services
	m.Map(config)
	m.Map(conn)
	m.Use(render.Renderer())
	store := sessions.NewCookieStore([]byte(config.SecretKey), []byte(config.EncriptionKey))
	store.Options(sessions.Options{
		MaxAge:   UserTokenExpirationDays * 86400,
		Secure:   config.SecureCookies,
		HttpOnly: true,
	})
	m.Use(sessions.Sessions(config.SessionName, store))

	// Add routes

	return m, config.Port, nil
}
