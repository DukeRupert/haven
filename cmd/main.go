package main

import (
	"log"
	"net/http"
	"os"

	"github.com/DukeRupert/haven/auth"
	"github.com/DukeRupert/haven/config"
	"github.com/DukeRupert/haven/db"
	"github.com/DukeRupert/haven/handler"
	"github.com/DukeRupert/haven/store"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
)

func main() {
	// Initialize logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	// Load configuration
	config, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Initialize Echo instance
	e := echo.New()
	e.Pre(middleware.RemoveTrailingSlash())
	e.Static("/static", "assets")

	// Initialize database
	dbConfig := db.DefaultConfig()
	database, err := db.New(config.DatabaseURL, dbConfig)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Initialize session store
	store, err := store.NewPgxStore(database, []byte(config.SessionKey))
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create session store")
	}

	// Configure session store
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   true, // true in production
		SameSite: http.SameSiteLaxMode,
	}

	// Add basic middleware
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	// e.Use(middleware.Secure())
	e.Use(middleware.Logger())

	// Initialize session middleware first
	e.Use(session.Middleware(store))

	// Initialize handlers after session middleware
	h := handler.NewHandler(database, logger)
	authHandler := auth.NewAuthHandler(database, store, logger)

	// Public routes
	e.GET("/", h.ShowHome)
	e.GET("/login", h.GetLogin, authHandler.RedirectIfAuthenticated())
	e.POST("/login", authHandler.LoginHandler())
	e.POST("/logout", authHandler.LogoutHandler())

	// Api
	api := e.Group("/api")
	api.Use(authHandler.AuthMiddleware())
	api.GET("/:fid/:uid/schedule/new", h.CreateScheduleForm)
	api.POST("/:fid/:uid/schedule", h.CreateSchedule)

	// Protected routes
	app := e.Group("/app")
	app.Use(authHandler.AuthMiddleware())
	app.Use(handler.SetRouteContext(h))
	app.GET("/", h.PlaceholderMessage)
	app.GET("/:code/calendar", h.PlaceholderMessage)
	app.GET("/:code", h.GetUsersByFacility)
	app.GET("/:code/:initials", h.GetUser)

	admin := app.Group("", authHandler.RoleAuthMiddleware("admin"))
	admin.POST("/:code", h.CreateUser)
	admin.GET("/:code/create", h.CreateUserForm)
	admin.PUT("/:code/:initials", h.PlaceholderMessage)
	admin.GET("/:code/:initials/update", h.CreateUserForm)
	admin.POST("/:code/:initials/schedule", h.CreateSchedule)
	admin.GET("/:code/:initials/schedule/create", h.CreateScheduleForm)
	admin.GET("/:code/:initials/schedule/update", h.UpdateScheduleForm)
	
	

	super := app.Group("", authHandler.RoleAuthMiddleware("super"))
	super.GET("/facilities", h.GetFacilities)
	super.POST("/facilities", h.PostFacilities)
	super.GET("/facilities/create", h.CreateFacilityForm)
	super.GET("/facilities/:fid/update", h.UpdateFacilityForm)
	super.PUT("/facilities/:fid", h.UpdateFacility)

	// Start server
	logger.Info().Msg("Starting server on :8080")
	e.Logger.Fatal(e.Start(":8080"))
}
