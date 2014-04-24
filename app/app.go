package app

import (
	"errors"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/cors"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	"github.com/martini-contrib/strict"
	"github.com/mvader/mask/handlers"
	"github.com/mvader/mask/middleware"
	"github.com/mvader/mask/models"
	"github.com/mvader/mask/services"
	"log"
	"os"
	"strings"
)

type App struct {
	Martini    *martini.ClassicMartini
	Config     *services.Config
	Connection *services.Connection
	LogFile    *os.File
}

// NewApp creates a new application
func NewApp(configPath string) (*App, error) {
	var err error

	// Create app
	m := martini.Classic()

	// Create config service
	config, err := services.NewConfig(configPath)
	if err != nil {
		return nil, err
	}

	// Create database service
	conn, err := services.NewDatabaseConn(config)
	if err != nil {
		return nil, err
	}

	// Create task service
	ts, err := services.NewTaskService(config)
	if err != nil {
		return nil, err
	}

	// Create and setup logger
	var logFile string
	if strings.HasSuffix(config.LogsPath, "/") {
		logFile = config.LogsPath + "mask.log"
	} else {
		logFile = config.LogsPath + "/mask.log"
	}

	var file *os.File
	if file, err = os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666); err != nil {
		return nil, errors.New("unable to open or create log file: " + logFile)
	}

	logger := log.New(file, "[mask] ", 0)

	// Map config as *Config
	m.Map(config)

	// Map conn as *Connection
	m.Map(conn)

	// Mask ts as *TaskService
	m.Map(ts)

	// Map logger as *log.Logger
	m.Map(logger)

	// Setup CORS
	m.Use(cors.Allow(&cors.Options{
		AllowOrigins:     []string{"https://*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "X-User-Token", "X-Access-Token"},
		ExposeHeaders:    []string{"Content-Length", "Content-Encoding", "Content-Type"},
		AllowCredentials: false,
	}))

	// Setup render middleware
	m.Use(render.Renderer())

	// Setup store paths
	m.Use(martini.Static(config.StorePath))
	m.Use(martini.Static(config.ThumbnailStorePath))
	m.Use(martini.Static(config.StaticContentPath))

	// Setup sessions
	store := sessions.NewCookieStore([]byte(config.SecretKey), []byte(config.EncriptionKey))
	store.Options(sessions.Options{
		MaxAge:   models.UserTokenExpirationDays * 86400,
		Secure:   config.SecureCookies,
		HttpOnly: true,
	})
	m.Use(sessions.Sessions(config.SessionName, store))

	// Add context middleware
	m.Use(middleware.CreateContext)

	// Add routes
	addRoutes(m)

	// Add NotFound handler
	m.Router.NotFound(strict.MethodNotAllowed, strict.NotFound)

	return &App{m, config, conn, file}, nil
}

// addRoutes adds all necessary routes to a martini instance
func addRoutes(m *martini.ClassicMartini) {
	m.Group("/api", func(r martini.Router) {
		// Post routes
		r.Group("/posts", func(r martini.Router) {
			r.Get("/show/:id", handlers.ShowPost)
			r.Post("/create", handlers.CreatePost)
			r.Delete("/destroy/:id", handlers.DeletePost)
			r.Put("/like/:id", handlers.LikePost)
			r.Put("/change_privacy/:id", handlers.ChangePostPrivacy)
		}, middleware.LoginRequired)

		r.Group("/auth", func(r martini.Router) {
			r.Get("/access_token", handlers.GetAccessToken)
			r.Post("/user_token", handlers.GetUserToken)
			r.Post("/login", handlers.Login)
		}, middleware.LoginForbidden)

		// Comment routes
		r.Group("/comments", func(r martini.Router) {
			r.Post("/create", handlers.CreateComment)
			r.Get("/for_post/:post_id", handlers.CommentsForPost)
			r.Delete("/destroy/:comment_id", handlers.RemoveComment)
		}, middleware.LoginRequired)

		// Account routes
		r.Group("/account", func(r martini.Router) {
			r.Post("/signup", handlers.CreateAccount)
			r.Get("/info", handlers.GetAccountInfo)
			r.Put("/info", handlers.UpdateAccountInfo)
			r.Get("/settings", handlers.GetAccountSettings)
			r.Put("/settings", handlers.UpdateAccountSettings)
		}, middleware.WebOnly, middleware.LoginRequired)

		// Logout
		r.Get("/account/logout", handlers.DestroyUserToken)

		// Block routes
		r.Group("/blocks", func(r martini.Router) {
			r.Post("/create", handlers.BlockHandler)
			r.Delete("/destroy", handlers.Unblock)
			r.Get("/show", handlers.ListBlocks)
		}, middleware.LoginRequired)

		// User routes
		r.Group("/users", func(r martini.Router) {
			r.Post("/follow", handlers.SendFollowRequest)
			r.Delete("/unfollow", handlers.Unfollow)
			r.Get("/follow_requests", handlers.ListFollowRequests)
			r.Post("/reply_follow_request", handlers.ReplyFollowRequest)

			r.Get("/followers", handlers.ListFollowers)
			r.Get("/following", handlers.ListFollowing)
		}, middleware.LoginRequired)

		// Notification routes
		r.Group("/notifications", func(r martini.Router) {
			r.Get("/list", handlers.ListNotifications)
			r.Put("/seen/:id", handlers.MarkNotificationRead)
		}, middleware.LoginRequired)

		// Show user profile
		r.Get("/u/:username", middleware.LoginRequired, handlers.ShowUserProfile)

		// Search for users
		r.Get("/search", middleware.LoginRequired, handlers.Search)

		// Get user timeline
		r.Get("/timeline", middleware.LoginRequired, handlers.GetUserTimeline)
	})

	// Render the layout
	m.Get("/", func(c middleware.Context) {
		if c.User == nil {
			// Render not logged in home
			// TODO
		} else {
			// Render logged in home
			// TODO
		}
	})
}
