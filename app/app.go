package app

import (
	"encoding/json"
	"errors"
	"github.com/go-martini/martini"
	"github.com/gorilla/sessions"
	"github.com/martini-contrib/cors"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/strict"
	"github.com/mvader/sunglasses/handlers"
	"github.com/mvader/sunglasses/middleware"
	"github.com/mvader/sunglasses/services"
	"github.com/mvader/sunglasses/util"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

// App represents the application, it contains the martini instance and the
// instances of the Config and Connection services along with the LogFile
// to be closed after running the application
type App struct {
	Martini    *martini.ClassicMartini
	Config     *services.Config
	Connection *services.Connection
	LogFile    *os.File
}

// NewApp creates a new application given a path to a config file
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
		logFile = config.LogsPath + "sunglasses.log"
	} else {
		logFile = config.LogsPath + "/sunglasses.log"
	}

	var file *os.File
	if file, err = os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666); err != nil {
		return nil, errors.New("unable to open or create log file: " + logFile)
	}

	logger := log.New(file, "[sunglasses] ", 0)

	// Map config as *Config
	m.Map(config)

	// Map conn as *Connection
	m.Map(conn)

	// sunglasses ts as *TaskService
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
	store := sessions.NewCookieStore([]byte(config.SecretKey))
	m.Map(store)

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

		// Auth routes
		r.Group("/auth", func(r martini.Router) {
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
			r.Get("/info", handlers.GetAccountInfo)
			r.Put("/info", handlers.UpdateAccountInfo)
			r.Get("/settings", handlers.GetAccountSettings)
			r.Put("/settings", handlers.UpdateAccountSettings)
		}, middleware.WebOnly, middleware.LoginRequired)
		r.Get("/account/username_taken", middleware.WebOnly, middleware.LoginForbidden, handlers.IsUsernameTaken)
		r.Post("/account/signup", middleware.WebOnly, middleware.LoginForbidden, handlers.CreateAccount)

		// Destroy user token
		m.Delete("/auth/destroy_user_token", middleware.LoginRequired, handlers.DestroyUserToken)

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
	}, middleware.RequiresValidSignature)

	// Get access token
	m.Get("/api/auth/access_token", middleware.LoginForbidden, handlers.GetAccessToken)

	// Logout
	m.Get("/account/logout", middleware.WebOnly, handlers.Logout)

	// Render the layout
	m.Get("/", func(c middleware.Context) string {
		var (
			content    []byte
			strContent string
			err        error
			token      string
		)

		// Read app.html located on the static content path (public)
		if content, err = ioutil.ReadFile(c.Config.StaticContentPath + "app.html"); err != nil {
			panic(err)
		}

		strContent = string(content)

		// If the user not logged in we will replace the default userData with the
		// actual user data
		if c.User != nil {
			b, err := json.Marshal(*c.User)
			if err != nil {
				panic(err)
			}

			strContent = strings.Replace(strContent, "userData = undefined;", "userData = "+string(b)+";", 1)
		}

		if c.Session.Values["csrf_time"] == nil {
			c.Session.Values["csrf_time"] = int64(0)
		}

		if c.Session.Values["csrf_token"] == nil || time.Now().Unix()-c.Session.Values["csrf_time"].(int64) > 10 {
			token = util.NewRandomHash()
			c.Session.Values["csrf_token"] = token
			c.Session.Values["csrf_time"] = time.Now().Unix()
			c.Session.Save(c.Request, c.ResponseWriter)
		} else {
			token = c.Session.Values["csrf_token"].(string)
		}

		// Set new csrf_token and replace it on the HTML page
		strContent = strings.Replace(strContent, "csrfToken = undefined;", "csrfToken = '"+token+"';", 1)

		// Return page content
		return strContent
	})
}
