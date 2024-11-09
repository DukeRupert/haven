package main

import (
	"log"
	"net/http"

	"github.com/DukeRupert/haven/config"
	"github.com/DukeRupert/haven/db"

	// "github.com/DukeRupert/haven/logger"
	"github.com/DukeRupert/haven/middleware"

	"github.com/labstack/echo/v4"
)

func main() {
	// Load configuration
	config, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Initialize logger
	// logger := logger.Initialize(config.Environment)

	// Initialize database with custom configuration
	dbConfig := db.DefaultConfig()
	pool, err := db.InitDBWithConfig(config.DatabaseURL, dbConfig)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer pool.Close()

	// Initialize Echo instance
	e := echo.New()
	e.Static("/static", "assets")
	// Apply the middleware
	middleware.InitStore([]byte("supersecret"))
	cfg := middleware.DefaultSessionConfig()
	e.Use(middleware.NewSessionMiddleware(cfg))
	// Initialize Middleware
	// e.Use(logger.Middleware())
	// FIXME e.Use(middleware.CORS())

	// userHandler := handler.UserHandler{}
	// authHandler := handler.AuthHandler{}
	// e.GET("/login", authHandler.HandleLogin)
	// e.GET("/user", userHandler.HandleUserShow)
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World! Welcome to Haven.")
	})

	e.Logger.Fatal(e.Start(":8080"))
}
