package app

import (
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	"github.com/mvader/mask/middleware"
	"github.com/mvader/mask/models"
	"github.com/mvader/mask/services"
)

// NewApp instantiates the application
func NewApp(configPath string) (*martini.ClassicMartini, string, error) {
	// Create app
	m := martini.Classic()

	// Create config service
	config, err := services.NewConfig(configPath)
	if err != nil {
		return nil, "", err
	}

	// Create database service
	conn, err := services.NewDatabaseConn(config)
	if err != nil {
		return nil, "", err
	}

	ts, err := services.NewTaskService(config)
	if err != nil {
		return nil, "", err
	}

	// Map services
	m.Map(config)
	m.Map(conn)
	m.Map(ts)
	m.Use(render.Renderer())
	m.Use(martini.Static(config.StorePath))
	m.Use(martini.Static(config.ThumbnailStorePath))
	store := sessions.NewCookieStore([]byte(config.SecretKey), []byte(config.EncriptionKey))
	store.Options(sessions.Options{
		MaxAge:   models.UserTokenExpirationDays * 86400,
		Secure:   config.SecureCookies,
		HttpOnly: true,
	})
	m.Use(sessions.Sessions(config.SessionName, store))
	m.Use(middleware.CreateContext)

	// Add routes

	return m, config.Port, nil
}
