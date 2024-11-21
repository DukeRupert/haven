package main

import (
	"encoding/gob"
	"log"
	"net/http"
	"os"

	"github.com/DukeRupert/haven/auth"
	"github.com/DukeRupert/haven/config"
	"github.com/DukeRupert/haven/db"
	"github.com/DukeRupert/haven/types"
	"github.com/DukeRupert/haven/handler"
	"github.com/DukeRupert/haven/store"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
)

func init() {
	// Register your custom types with gob
	var userRole types.UserRole
	gob.Register(userRole)
}

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

	// Global middleware that should apply to all routes
    e.Use(middleware.Recover())
    e.Use(middleware.RequestID())
    e.Use(middleware.Logger())
    e.Use(session.Middleware(store))

    // Initialize handlers
    h := handler.NewHandler(database, logger)
    authHandler := auth.NewAuthHandler(database, store, logger)

    // Setup all routes
    handler.SetupRoutes(e, h, authHandler)

	// Start server
	logger.Info().Msg("Starting server on :8080")
	e.Logger.Fatal(e.Start(":8080"))
}
