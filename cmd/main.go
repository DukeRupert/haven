package main

import (
	"embed"
	"encoding/gob"
	"flag"
	"log"
	"os"
	"time"

	"github.com/DukeRupert/haven/internal/auth"
	"github.com/DukeRupert/haven/internal/config"
	"github.com/DukeRupert/haven/internal/context"
	"github.com/DukeRupert/haven/internal/handler"
	"github.com/DukeRupert/haven/internal/model/types"
	"github.com/DukeRupert/haven/internal/repository"
	"github.com/DukeRupert/haven/internal/store"
	"github.com/DukeRupert/haven/internal/worker"
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
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

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
	e.Pre(middleware.RemoveTrailingSlash())
	e.Static("/static", "assets")

	// Initialize database and repositories
	database, err := repository.New(config.DatabaseURL, repository.DefaultConfig())
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer database.Close()

	repos := repository.NewRepositories(database)

	// Initialize session store
	sessionStore, err := store.NewPgxStore(repos.Session, []byte(config.SessionKey))
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create session store")
	}

	// Initialize auth components
	authService := auth.NewService(auth.Config{
		Repos:  repos,
		Store:  sessionStore,
		Logger: logger,
	})

	authMiddleware := auth.NewMiddleware(authService, logger)

    authHandler := auth.NewHandler(auth.HandlerConfig{
        Service: authService,
        Store:   sessionStore,
        Logger:  logger,
    })

	// Initialize other middleware
	routeCtxMiddleware := context.NewRouteContextMiddleware(logger)

	// Initialize main application handler
	appHandler := handler.New(handler.Config{
        Repos:    repos,
        Auth:     authMiddleware,
        Sessions: sessionStore,
        Logger:   logger,
    })

	// Create and start token cleaner
	tokenCleaner := worker.NewTokenCleaner(
		repos.Token,
		logger,
		15*time.Minute,
	)
	tokenCleaner.Start()
	defer tokenCleaner.Stop()

	// Setup all routes
	handler.SetupRoutes(e, appHandler, authMiddleware, authHandler, routeCtxMiddleware)

	// Start server
	logger.Info().Msg("Starting server on :8080")
	e.Logger.Fatal(e.Start(":8080"))
}
