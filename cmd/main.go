package main

import (
	"context"
	"log"
	"net/http"

	"github.com/DukeRupert/haven/config"
	"github.com/DukeRupert/haven/db"
	"github.com/DukeRupert/haven/store"

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

	// Initialize session store
	ctx := context.Background()
	store, err := store.NewPGStoreFromPool(ctx, pool, []byte("tU0bNcAgjeUNHIUYCvdyL7EsSeT6W4bo"), []byte("e9EamH55rqvPHoPtEam3GbeW7HE5DIpY"))
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	e.Use(middleware.SessionMiddleware(store))

	// Add test routes
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
