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

const (
	SessionSecret = "tU0bNcAgjeUNHIUYCvdyL7EsSeT6W4bo"
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
	e.Static("/static", "assets")

	// Initialize database
	dbConfig := db.DefaultConfig()
	pool, err := db.InitDBWithConfig(config.DatabaseURL, dbConfig)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer pool.Close()

	// Initialize session store
	store, err := store.NewPgxStore(pool, []byte(SessionSecret))
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
	e.Use(middleware.Secure())
	e.Use(middleware.Logger())

	// Initialize session middleware first
	e.Use(session.Middleware(store))

	// Initialize handlers after session middleware
	authHandler := auth.NewAuthHandler(pool, store, logger)
	userHandler := handler.NewUserHandler(pool)

	// Public routes
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World! Welcome to Haven.")
	})
	e.GET("/login", userHandler.GetLogin)
	e.POST("/login", authHandler.LoginHandler())
	e.POST("/logout", authHandler.LogoutHandler()) // Fixed route path to include leading slash

	// Protected routes
	app := e.Group("/app")
	app.Use(authHandler.AuthMiddleware())
	app.GET("", userHandler.HandleUserShow)
	app.GET("/dashboard", func(c echo.Context) error {
		return c.String(http.StatusOK, "You have access to User routes")
	})

	// Admin routes
	admin := app.Group("/admin")
	admin.GET("", func(c echo.Context) error {
		return c.String(http.StatusOK, "You have access to Admin routes")
	})

	// Super admin routes
	super := app.Group("/super")
	super.GET("", func(c echo.Context) error {
		return c.String(http.StatusOK, "You have access to Super routes")
	})

	// Start server
	logger.Info().Msg("Starting server on :8080")
	e.Logger.Fatal(e.Start(":8080"))
}
