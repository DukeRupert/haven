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
	userHandler := handler.NewUserHandler(database, logger)

	// Public routes
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World! Welcome to Haven.")
	})
	e.GET("/login", userHandler.GetLogin)
	e.POST("/login", authHandler.LoginHandler())
	e.POST("/logout", authHandler.LogoutHandler()) // Fixed route path to include leading slash

	// Api
	api := e.Group("/api")
	api.Use(authHandler.AuthMiddleware())
	api.GET("/:fid/:uid/schedule/new", h.CreateScheduleForm)
	api.POST("/:fid/:uid/schedule", h.CreateSchedule)

	// Protected routes
	app := e.Group("/app")
	app.Use(authHandler.AuthMiddleware())
	app.GET("", userHandler.HandleUserShow)
	app.GET("/dashboard", func(c echo.Context) error {
		return c.String(http.StatusOK, "You have access to User routes")
	})

	super := app.Group("", authHandler.RoleAuthMiddleware("super"))
	super.GET("/facilities", h.GetFacilities)
	super.POST("/facilities", h.PostFacilities)
	super.GET("/facilities/create", h.CreateFacilityForm)
	super.GET("/facilities/:fid/update", h.UpdateFacilityForm)
	super.PUT("/facilities/:fid", h.UpdateFacility)
	// Admin routes
	admin := app.Group("/admin")
	admin.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "You have access to Admin routes")
	})
	admin.GET("/:code", userHandler.GetUsersByFacility)
	admin.GET("/:code/user/create", userHandler.CreateUserForm)
	admin.POST("/:code/user", userHandler.CreateUser)
	admin.GET("/:code/:initials", userHandler.UserPage)

	// Start server
	logger.Info().Msg("Starting server on :8080")
	e.Logger.Fatal(e.Start(":8080"))
}
