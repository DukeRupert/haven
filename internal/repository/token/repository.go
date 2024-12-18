// internal/repository/token/repository.go
package token

import (
	"context"
	"fmt"
	"time"

	"github.com/DukeRupert/haven/internal/model/entity"

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
    ErrTokenUsed    = fmt.Errorf("token has already been used")
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

// StoreVerification saves a new verification token
func (r *Repository) StoreVerification(ctx context.Context, vt *entity.VerificationToken) error {
    _, err := r.pool.Exec(ctx, `
        INSERT INTO verification_tokens (user_id, token, email, expires_at)
        VALUES ($1, $2, $3, $4)
    `, vt.UserID, vt.Token, vt.Email, vt.ExpiresAt)
    if err != nil {
        return fmt.Errorf("storing verification token: %w", err)
    }
    return nil
}

// GetVerificationToken retrieves a verification token and its details
func (r *Repository) GetVerificationToken(ctx context.Context, token string) (*entity.VerificationToken, error) {
    vt := &entity.VerificationToken{}
    err := r.pool.QueryRow(ctx, `
        SELECT user_id, token, email, created_at, expires_at, used
        FROM verification_tokens
        WHERE token = $1
    `, token).Scan(&vt.UserID, &vt.Token, &vt.Email, &vt.CreatedAt, &vt.ExpiresAt, &vt.Used)
    
    if err == pgx.ErrNoRows {
        return nil, ErrInvalidToken
    }
    if err != nil {
        return nil, fmt.Errorf("getting verification token: %w", err)
    }
    
    if vt.Used {
        return nil, ErrTokenUsed
    }
    
    if time.Now().After(vt.ExpiresAt) {
        return nil, ErrInvalidToken
    }
    
    return vt, nil
}

// MarkAsUsed marks a verification token as used
func (r *Repository) MarkAsUsed(ctx context.Context, token string) error {
    result, err := r.pool.Exec(ctx, `
        UPDATE verification_tokens
        SET used = true
        WHERE token = $1 AND NOT used
    `, token)
    if err != nil {
        return fmt.Errorf("marking token as used: %w", err)
    }
    
    if result.RowsAffected() == 0 {
        return ErrInvalidToken
    }
    return nil
}

// Verify checks if a registration token is valid and not expired, returns associated user ID
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

// DeleteExpired removes all expired tokens from both tables
func (r *Repository) DeleteExpired(ctx context.Context) (int64, error) {
    tx, err := r.pool.Begin(ctx)
    if err != nil {
        return 0, fmt.Errorf("beginning transaction: %w", err)
    }
    defer tx.Rollback(ctx)

    // Delete from registration_tokens
    regResult, err := tx.Exec(ctx, `
        DELETE FROM registration_tokens 
        WHERE expires_at < CURRENT_TIMESTAMP
    `)
    if err != nil {
        return 0, fmt.Errorf("deleting expired registration tokens: %w", err)
    }

    // Delete from verification_tokens
    verResult, err := tx.Exec(ctx, `
        DELETE FROM verification_tokens 
        WHERE expires_at < CURRENT_TIMESTAMP OR used = true
    `)
    if err != nil {
        return 0, fmt.Errorf("deleting expired verification tokens: %w", err)
    }

    if err := tx.Commit(ctx); err != nil {
        return 0, fmt.Errorf("committing transaction: %w", err)
    }

    return regResult.RowsAffected() + verResult.RowsAffected(), nil
}