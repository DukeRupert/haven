package main

import (
	"embed"
	"encoding/gob"
	"flag"
	"log"
	"os"
	"time"

	"github.com/DukeRupert/haven/auth"
	"github.com/DukeRupert/haven/config"
	"github.com/DukeRupert/haven/db"
	"github.com/DukeRupert/haven/handler"
	"github.com/DukeRupert/haven/store"
	"github.com/DukeRupert/haven/types"
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
	// Parse flags, but don't set a default value
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

	// Only run migrations if explicitly requested
	if *migrateCmd != "" {
		if err := runMigrations(config.DatabaseURL, *migrateCmd); err != nil {
			l.Fatal().Err(err).Str("command", *migrateCmd).Msg("Migration failed")
		}
		l.Info().Msg("Migrations completed successfully")
		// Only exit if migrations were explicitly requested
		if flag.Lookup("migrate").Value.String() != "" {
			return
		}
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

	// Initialize session s
	s, err := store.NewPgxStore(db, []byte(config.SessionKey))
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to create session store")
	}

	// Initialize handlers
	h := handler.NewHandler(db, l)
	authHandler := auth.NewAuthHandler(db, s, l)

	// Setup all routes
	handler.SetupRoutes(e, h, authHandler, s)

	// Start server
	l.Info().Msg("Starting server on :8080")
	e.Logger.Fatal(e.Start(":8080"))
}
