// db/db.go
package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DB represents a database instance with its connection pool
type DB struct {
	pool   *pgxpool.Pool
	config Config
}

// Config holds database configuration
type Config struct {
	MaxConns     int32
	MinConns     int32
	MaxConnIdle  time.Duration
	MaxConnLife  time.Duration
	MaxConnTries int
	RetryDelay   time.Duration
}

// DefaultConfig returns sensible default configuration
func DefaultConfig() Config {
	return Config{
		MaxConns:     25,
		MinConns:     5,
		MaxConnIdle:  30 * time.Minute,
		MaxConnLife:  24 * time.Hour,
		MaxConnTries: 5,
		RetryDelay:   time.Second,
	}
}

// New creates a new DB instance with the given configuration
func New(databaseURL string, config Config) (*DB, error) {
	ctx := context.Background()

	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database URL: %w", err)
	}

	// Configure the connection pool
	poolConfig.MaxConns = config.MaxConns
	poolConfig.MinConns = config.MinConns
	poolConfig.MaxConnIdleTime = config.MaxConnIdle
	poolConfig.MaxConnLifetime = config.MaxConnLife

	// Attempt to connect with retries
	var pool *pgxpool.Pool
	var lastErr error

	for tries := 0; tries < config.MaxConnTries; tries++ {
		pool, err = pgxpool.NewWithConfig(ctx, poolConfig)
		if err == nil {
			// Test the connection
			err = pool.Ping(ctx)
			if err == nil {
				log.Printf("Successfully connected to database (attempt %d/%d)",
					tries+1, config.MaxConnTries)
				return &DB{
					pool:   pool,
					config: config,
				}, nil
			}
		}

		lastErr = err
		log.Printf("Failed to connect to database (attempt %d/%d): %v",
			tries+1, config.MaxConnTries, err)

		// If this isn't our last try, wait before retrying
		if tries+1 < config.MaxConnTries {
			time.Sleep(config.RetryDelay)
		}
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts: %w",
		config.MaxConnTries, lastErr)
}

// Close closes the database connection pool
func (db *DB) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
}

// Pool returns the underlying connection pool
func (db *DB) Pool() *pgxpool.Pool {
	return db.pool
}

// QueryRow executes a query that is expected to return at most one row
func (db *DB) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	return db.pool.QueryRow(ctx, query, args...)
}

// Query executes a query that returns rows
func (db *DB) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	return db.pool.Query(ctx, query, args...)
}

// Exec executes a query that doesn't return rows
func (db *DB) Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	return db.pool.Exec(ctx, query, args...)
}