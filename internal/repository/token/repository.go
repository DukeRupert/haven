// internal/repository/token/repository.go
package token

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles token-related database operations
type Repository struct {
	pool *pgxpool.Pool
}

// New creates a new token repository
func New(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

// Common errors
var (
	ErrInvalidToken = fmt.Errorf("invalid or expired token")
)

// Store saves or updates a registration token for a user
func (r *Repository) Store(ctx context.Context, userID int, token string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
        INSERT INTO registration_tokens (user_id, token, expires_at)
        VALUES ($1, $2, $3)
        ON CONFLICT (user_id) 
        DO UPDATE SET 
            token = EXCLUDED.token,
            expires_at = EXCLUDED.expires_at
    `, userID, token, expiresAt)
	if err != nil {
		return fmt.Errorf("storing registration token: %w", err)
	}
	return nil
}

// Verify checks if a token is valid and not expired, returns associated user ID
func (r *Repository) Verify(ctx context.Context, token string) (int, error) {
	var userID int
	err := r.pool.QueryRow(ctx, `
        SELECT user_id 
        FROM registration_tokens
        WHERE token = $1
        AND expires_at > NOW()
    `, token).Scan(&userID)
	if err == pgx.ErrNoRows {
		return 0, ErrInvalidToken
	}
	if err != nil {
		return 0, fmt.Errorf("verifying token: %w", err)
	}
	return userID, nil
}

// Delete removes a specific token from the database
func (r *Repository) Delete(ctx context.Context, token string) error {
	result, err := r.pool.Exec(ctx, `
        DELETE FROM registration_tokens
        WHERE token = $1
    `, token)
	if err != nil {
		return fmt.Errorf("deleting token: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrInvalidToken
	}
	return nil
}

// DeleteExpired removes all expired tokens and returns the count of deleted tokens
func (r *Repository) DeleteExpired(ctx context.Context) (int64, error) {
	result, err := r.pool.Exec(ctx, `
        DELETE FROM registration_tokens 
        WHERE expires_at < CURRENT_TIMESTAMP
    `)
	if err != nil {
		return 0, fmt.Errorf("deleting expired tokens: %w", err)
	}
	return result.RowsAffected(), nil
}
