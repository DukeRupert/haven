package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

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
		MaxConns:     25,         // Maximum number of connections
		MinConns:     5,          // Minimum number of connections
		MaxConnIdle:  30 * time.Minute,
		MaxConnLife:  24 * time.Hour,
		MaxConnTries: 5,          // Number of connection attempts
		RetryDelay:   time.Second, // Delay between retries
	}
}

// InitDB initializes and returns a connection pool
func InitDB(databaseURL string) (*pgxpool.Pool, error) {
	return InitDBWithConfig(databaseURL, DefaultConfig())
}

// InitDBWithConfig initializes and returns a connection pool with custom configuration
func InitDBWithConfig(databaseURL string, config Config) (*pgxpool.Pool, error) {
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
				return pool, nil
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

// HealthCheck performs a health check on the database
func HealthCheck(pool *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return pool.Ping(ctx)
}