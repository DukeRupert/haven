package main

import (
	"context"
	"embed"
	"encoding/gob"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DukeRupert/haven/auth"
	"github.com/DukeRupert/haven/config"
	"github.com/DukeRupert/haven/db"
	"github.com/DukeRupert/haven/handler"
	"github.com/DukeRupert/haven/internal/repository"
	"github.com/DukeRupert/haven/store"
	"github.com/DukeRupert/haven/types"
	"github.com/DukeRupert/haven/worker"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

func init() {
	gob.Register(types.UserRole(""))
	gob.Register(time.Time{})
}

func main() {
	// parse db migrate flags
	migrateCmd := flag.String("migrate", "", "Migration command (up/down/reset/status)")
	flag.Parse()

	// Initialize logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	l := zerolog.New(os.Stdout).With().Timestamp().Logger()

	// Load configuration
	config, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Run up migrations regardless of flag
	if err := runMigrations(config.DatabaseURL, "up"); err != nil {
		l.Fatal().Err(err).Msg("Initial migration failed")
	}

	// If migration command was explicitly provided, exit after running it
	if *migrateCmd != "" {
		l.Info().Msg("Migrations completed successfully")
		return
	}

	// Initialize Echo instance
	e := echo.New()
	e.Pre(middleware.RemoveTrailingSlash())
	e.Static("/static", "assets")

	// Initialize database
	dbConfig := db.DefaultConfig()
	db, err := db.New(config.DatabaseURL, dbConfig)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
    repos := repository.NewRepositories(db)

    // Create and start token cleaner
    tokenCleaner := worker.NewTokenCleaner(
        repos.Token,
        l,
        15*time.Minute,
    )
    tokenCleaner.Start()
    defer tokenCleaner.Stop()

	// Initialize session store
	s, err := store.NewPgxStore(db, []byte(config.SessionKey))
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to create session store")
	}

	// Initialize handlers
	h := handler.NewHandler(db, l)
	authHandler := auth.NewAuthHandler(db, s, l)

	// Setup all routes
	handler.SetupRoutes(e, h, authHandler, s)

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		l.Info().Msg("Shutting down server")

		// Stop the token cleaner
		cleaner.Stop()

		// Shutdown Echo server
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := e.Shutdown(ctx); err != nil {
			l.Fatal().Err(err).Msg("Failed to shutdown server gracefully")
		}
	}()

	// Start server
	l.Info().Msg("Starting server on :8080")
	e.Logger.Fatal(e.Start(":8080"))
}
