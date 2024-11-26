package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// StoreRegistrationToken stores a registration token for a user with expiration
// Note: You'll need to create a registration_tokens table first
func (db *DB) StoreRegistrationToken(ctx context.Context, userID int, token string, expiresAt time.Time) error {
	_, err := db.pool.Exec(ctx, `
        INSERT INTO registration_tokens (user_id, token, expires_at)
        VALUES ($1, $2, $3)
        ON CONFLICT (user_id) 
        DO UPDATE SET 
            token = EXCLUDED.token,
            expires_at = EXCLUDED.expires_at
    `, userID, token, expiresAt)
	if err != nil {
		return fmt.Errorf("error storing registration token: %w", err)
	}
	return nil
}

// VerifyRegistrationToken checks if a token is valid and not expired
func (db *DB) VerifyRegistrationToken(ctx context.Context, token string) (int, error) {
	var userID int
	err := db.pool.QueryRow(ctx, `
        SELECT user_id 
        FROM registration_tokens
        WHERE token = $1
        AND expires_at > NOW()
    `, token).Scan(&userID)
	if err == pgx.ErrNoRows {
		return 0, fmt.Errorf("invalid or expired token")
	}
	if err != nil {
		return 0, fmt.Errorf("error verifying registration token: %w", err)
	}
	return userID, nil
}

// DeleteRegistrationToken removes a used token from the database
func (db *DB) DeleteRegistrationToken(ctx context.Context, token string) error {
	_, err := db.pool.Exec(ctx, `
        DELETE FROM registration_tokens
        WHERE token = $1
    `, token)
	if err != nil {
		return fmt.Errorf("error deleting registration token: %w", err)
	}
	return nil
}

// DeleteExpiredTokens removes expired registration tokens and returns the number of tokens deleted
func (db *DB) DeleteExpiredTokens(ctx context.Context) (int64, error) {
	result, err := db.pool.Exec(ctx, `
        DELETE FROM registration_tokens 
        WHERE expires_at < CURRENT_TIMESTAMP
    `)
	if err != nil {
		return 0, fmt.Errorf("error deleting expired tokens: %w", err)
	}

	return result.RowsAffected(), nil
}
