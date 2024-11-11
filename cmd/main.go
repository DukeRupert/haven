package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/DukeRupert/haven/auth"
	"github.com/DukeRupert/haven/config"
	"github.com/DukeRupert/haven/db"
	"github.com/DukeRupert/haven/handler"
	"github.com/DukeRupert/haven/store"

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
	e.Static("/static", "assets")

	// Initialize database
	dbConfig := db.DefaultConfig()
	pool, err := db.InitDBWithConfig(config.DatabaseURL, dbConfig)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer pool.Close()

	// Initialize session store
	store, err := store.NewPgxStore(pool, []byte("tU0bNcAgjeUNHIUYCvdyL7EsSeT6W4bo"))
	if err != nil {
		log.Fatal(err)
	}

	// Add middleware
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.Secure())
	e.Use(middleware.Logger())

	// Initialize session middleware
	authHandler := auth.NewAuthHandler(pool, store, logger)
	// e.Use(authHandler.SessionMiddleware())
	e.Use(session.MiddlewareWithConfig(session.Config{
		Skipper: middleware.DefaultSkipper,
		Store:   store,
	}))

	e.POST("/login", authHandler.LoginHandler())
	// Example routes
	e.GET("/create-session", createSession)
	e.GET("/read-session", readSession)
	e.GET("/update-session", updateSession)
	e.GET("/delete-session", deleteSession)

	// Initialize Handlers
	userHandler := handler.NewUserHandler(pool)

	// Public routes
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World! Welcome to Haven.")
	})
	e.GET("/login", userHandler.GetLogin)
	// e.POST("/login", authHandler.Login)
	// e.POST("logout", authHandler.Logout)

	// Protected routes
	app := e.Group("/app")
	app.Use(authHandler.AuthMiddleware())
	// app.Use(authHandler.RequireAuth())
	// app.Use(authHandler.RoleBasedMiddleware())

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

	e.Logger.Fatal(e.Start(":8080"))
}

func createSession(c echo.Context) error {
	sess, err := session.Get("session", c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Set session values
	sess.Values["user_id"] = 12345
	sess.Values["username"] = "johndoe"
	sess.Values["last_access"] = time.Now()

	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Session created successfully",
		"data":    sess.Values,
	})
}

func readSession(c echo.Context) error {
	sess, err := session.Get("session", c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if sess.IsNew {
		return echo.NewHTTPError(http.StatusNotFound, "No session found")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user_id":     sess.Values["user_id"],
		"username":    sess.Values["username"],
		"last_access": sess.Values["last_access"],
	})
}

func updateSession(c echo.Context) error {
	log.Println("Starting updateSession handler")

	sess, err := session.Get("session", c)
	if err != nil {
		log.Printf("Error getting session: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to get session: %v", err))
	}

	log.Printf("Session retrieved - IsNew: %v, ID: %v", sess.IsNew, sess.ID)
	log.Printf("Current session values: %+v", sess.Values)

	if sess.IsNew {
		log.Println("Session is new, returning not found error")
		return echo.NewHTTPError(http.StatusNotFound, "No session found")
	}

	// Update session values
	now := time.Now()
	log.Printf("Updating last_access to: %v", now)

	// Convert old values to map[string]interface{}
	oldValues := make(map[string]interface{})
	for k, v := range sess.Values {
		if key, ok := k.(string); ok {
			oldValues[key] = v
		}
	}

	sess.Values["last_access"] = now
	sess.Values["updated"] = true

	// Convert new values to map[string]interface{}
	newValues := make(map[string]interface{})
	for k, v := range sess.Values {
		if key, ok := k.(string); ok {
			newValues[key] = v
		}
	}

	log.Printf("Session values before save - Old: %+v, New: %+v", oldValues, newValues)

	if err := sess.Save(c.Request(), c.Response()); err != nil {
		log.Printf("Error saving session: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to save session: %v", err))
	}

	log.Println("Session successfully saved")

	response := map[string]interface{}{
		"message": "Session updated successfully",
		"data": map[string]interface{}{
			"old_values": oldValues,
			"new_values": newValues,
			"session_id": sess.ID,
			"is_new":     sess.IsNew,
		},
	}

	return c.JSON(http.StatusOK, response)
}

func deleteSession(c echo.Context) error {
	sess, err := session.Get("session", c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Set MaxAge to -1 to delete the session
	sess.Options.MaxAge = -1
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Session deleted successfully",
	})
}
