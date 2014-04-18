package app

import (
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	"github.com/mvader/mask/handlers"
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
	addRoutes(m)

	return m, config.Port, nil
}

// addRoutes adds all necessary routes to a martini instance
func addRoutes(m *martini.ClassicMartini) {
	// Post routes
	m.Group("/posts", func(r martini.Router) {
		r.Get("/show/:id", handlers.ShowPost)
		r.Post("/create", handlers.CreatePost)
		r.Delete("/destroy/:id", handlers.DeletePost)
		r.Put("/like/:id", handlers.LikePost)
		r.Put("/change_privacy/:id", handlers.ChangePostPrivacy)
	}, middleware.LoginRequired)

	m.Group("/auth", func(r martini.Router) {
		r.Get("/access_token", handlers.GetAccessToken)
		r.Post("/user_token", handlers.GetUserToken)
		r.Post("/login", handlers.Login)
	}, middleware.LoginForbidden)

	// Comment routes
	m.Group("/comments", func(r martini.Router) {
		r.Post("/create", handlers.CreateComment)
		r.Get("/for_post/:post_id", handlers.CommentsForPost)
		r.Delete("/destroy/:comment_id", handlers.RemoveComment)
	}, middleware.LoginRequired)

	// Account routes
	m.Group("/account", func(r martini.Router) {
		r.Post("/signup", handlers.CreateAccount)
		r.Get("/info", handlers.GetAccountInfo)
		r.Put("/info", handlers.UpdateAccountInfo)
		r.Get("/settings", handlers.GetAccountSettings)
		r.Put("/settings", handlers.UpdateAccountSettings)
	}, middleware.WebOnly, middleware.LoginRequired)

	// Logout
	m.Get("/account/logout", handlers.DestroyUserToken)

	// Block routes
	m.Group("/blocks", func(r martini.Router) {
		r.Post("/create", handlers.BlockHandler)
		r.Delete("/destroy", handlers.Unblock)
		r.Get("/show", handlers.ListBlocks)
	}, middleware.LoginRequired)

	// User routes
	m.Group("/users", func(r martini.Router) {
		r.Post("/follow", handlers.SendFollowRequest)
		r.Delete("/unfollow", handlers.Unfollow)
		r.Get("/follow_requests", handlers.ListFollowRequests)
		r.Post("/reply_follow_request", handlers.ReplyFollowRequest)

		r.Get("/followers", handlers.ListFollowers)
		r.Get("/following", handlers.ListFollowing)
	}, middleware.LoginRequired)

	// Notification routes
	m.Group("/notifications", func(r martini.Router) {
		r.Get("/list", handlers.ListNotifications)
		r.Put("/seen/:id", handlers.MarkNotificationRead)
	}, middleware.LoginRequired)

	// Show user profile
	m.Get("/u/:username", middleware.LoginRequired, handlers.ShowUserProfile)

	// Search for users
	m.Get("/search", middleware.LoginRequired, handlers.Search)

	// Get user timeline
	m.Get("/timeline", middleware.LoginRequired, handlers.GetUserTimeline)

	// Render the layout
	m.Get("/", func(c middleware.Context) {
		if c.User == nil {
			// Render not logged in home
		} else {
			// Render logged in home
		}
	})
}
