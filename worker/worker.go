package workers

import (
	"context"
	"fmt"
	"time"

	"github.com/DukeRupert/haven/db"
	"github.com/rs/zerolog"
)

// TokenCleaner handles periodic cleanup of expired registration tokens
type TokenCleaner struct {
	db       *db.DB
	logger   zerolog.Logger
	interval time.Duration
	done     chan struct{}
}

// NewTokenCleaner creates a new TokenCleaner instance
func NewTokenCleaner(db *db.DB, logger zerolog.Logger, interval time.Duration) *TokenCleaner {
	if interval < time.Minute {
		interval = 15 * time.Minute
	}

	return &TokenCleaner{
		db:       db,
		logger:   logger.With().Str("component", "token_cleaner").Logger(),
		interval: interval,
		done:     make(chan struct{}),
	}
}

// Start begins the periodic cleanup process
func (tc *TokenCleaner) Start() {
	tc.logger.Info().
		Dur("interval", tc.interval).
		Msg("Starting token cleanup worker")

	go func() {
		ticker := time.NewTicker(tc.interval)
		defer ticker.Stop()

		// Perform initial cleanup
		if err := tc.cleanup(); err != nil {
			tc.logger.Error().Err(err).Msg("Initial cleanup failed")
		}

		for {
			select {
			case <-ticker.C:
				if err := tc.cleanup(); err != nil {
					tc.logger.Error().Err(err).Msg("Periodic cleanup failed")
				}
			case <-tc.done:
				tc.logger.Info().Msg("Token cleanup worker stopped")
				return
			}
		}
	}()
}

// Stop gracefully stops the cleanup process
func (tc *TokenCleaner) Stop() {
	tc.logger.Info().Msg("Stopping token cleanup worker")
	close(tc.done)
}

// cleanup performs a single cleanup operation
func (tc *TokenCleaner) cleanup() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	startTime := time.Now()
	count, err := tc.db.DeleteExpiredTokens(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete expired tokens: %w", err)
	}

	tc.logger.Info().
		Int64("deleted_count", count).
		Dur("duration", time.Since(startTime)).
		Msg("Completed token cleanup")

	return nil
}
