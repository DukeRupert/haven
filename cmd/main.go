package main

import (
	"embed"
	"encoding/gob"
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"github.com/DukeRupert/haven/internal/config"
	"github.com/DukeRupert/haven/internal/handler"
	"github.com/DukeRupert/haven/internal/middleware"
	"github.com/DukeRupert/haven/internal/model/types"
	"github.com/DukeRupert/haven/internal/repository"
	"github.com/DukeRupert/haven/internal/store"
	"github.com/DukeRupert/haven/internal/worker"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

// runMigrations handles database migrations using goose
func runMigrations(dbURL string, command string) error {
	// Parse the connection config
	config, err := pgx.ParseConfig(dbURL)
	if err != nil {
		return err
	}

	// Convert to *sql.DB
	db := stdlib.OpenDB(*config)
	defer db.Close()

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	switch command {
	case "status":
		return goose.Status(db, "migrations")
	case "reset":
		if err := goose.Reset(db, "migrations"); err != nil {
			return err
		}
		return goose.Up(db, "migrations")
	case "down":
		return goose.Down(db, "migrations")
	default:
		return goose.Up(db, "migrations")
	}
}

// Helper function to set log level based on environment
func getLogLevel() zerolog.Level {
	level := os.Getenv("LOG_LEVEL")
	switch strings.ToLower(level) {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	default:
		// Default to INFO in production, DEBUG in development
		if os.Getenv("ENVIRONMENT") == "production" {
			return zerolog.InfoLevel
		}
		return zerolog.DebugLevel
	}
}

func init() {
	gob.Register(types.UserRole(""))
	gob.Register(time.Time{})
}

func main() {
	// parse db migrate flags
	migrateCmd := flag.String("migrate", "", "Migration command (up/down/reset/status)")
	flag.Parse()

	// Initialize logger
	zerolog.TimeFieldFormat = time.RFC3339
	logger := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Caller(). // Adds file name and line number
		Int("pid", os.Getpid()).
		Logger().
		Level(getLogLevel()) // We'll define this function below

	// Load configuration
	config, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Run up migrations regardless of flag
	if err := runMigrations(config.DatabaseURL, "up"); err != nil {
		logger.Fatal().Err(err).Msg("Initial migration failed")
	}

	// If migration command was explicitly provided, exit after running it
	if *migrateCmd != "" {
		logger.Info().Msg("Migrations completed successfully")
		return
	}

	// Initialize Echo instance
	e := echo.New()
	e.Static("/static", "web/assets")
	// Add Prometheus middleware early in the chain
	e.Use(echoprometheus.NewMiddleware("mirandashift"))
	// Add metrics endpoint - consider protecting this in production
	e.GET("/metrics", echoprometheus.NewHandler())

	// Initialize database and repositories
	database, err := repository.New(config.DatabaseURL, repository.DefaultConfig())
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer database.Close()

	repos := repository.NewRepositories(database)

	// Initialize session store
	sessionStore, err := store.NewPgxStore(repos.Session, []byte("your-secret-key"))
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create session store")
	}
	// Add session middleware at application level
	e.Use(session.Middleware(sessionStore))

	middleware := middleware.NewMiddleware(middleware.Config{
		Repos:  repos,
		Logger: logger,
	})

	// Initialize main application handler
	appHandler, err := handler.New(handler.Config{
		Repos:   repos,
		Logger:  logger,
		BaseURL: config.BaseURL,
		MailerConfig: handler.MailerConfig{
			ServerToken: config.PostmarkServerToken,
			FromEmail:   config.FromEmail,
			FromName:    "MirandaShift Support",
		},
	})
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to initialize handler")
	}

	// Setup routes with all components
	handler.SetupRoutes(e, appHandler, middleware)

	// Create and start token cleaner
	tokenCleaner := worker.NewTokenCleaner(
		repos.Token,
		logger,
		15*time.Minute,
	)
	tokenCleaner.Start()
	defer tokenCleaner.Stop()

	// Start server
	logger.Info().Msg("Starting server on :8080")
	e.Logger.Fatal(e.Start(":8080"))
}
