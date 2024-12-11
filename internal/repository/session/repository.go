// internal/repository/session/repository.go
package session

import (
	"context"
	"fmt"
	"time"

	"github.com/DukeRupert/haven/internal/model/entity"
	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type CreateParams struct {
    Key       string
    Data      []byte
    ExpiresOn time.Time
    IsNew     bool
}

type Repository struct {
    pool   *pgxpool.Pool
    logger zerolog.Logger
}

// Common errors
var (
    ErrNotFound         = fmt.Errorf("session not found")
)

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

func (r *Repository) CreateSessionsTable() error {
	ctx := context.Background()
	_, err := r.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS http_sessions (
			id BIGSERIAL PRIMARY KEY,
			key TEXT NOT NULL,
			data BYTEA NOT NULL,
			created_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			modified_on TIMESTAMPTZ,
			expires_on TIMESTAMPTZ,
			CONSTRAINT http_sessions_key_key UNIQUE (key)
		);
		CREATE INDEX IF NOT EXISTS http_sessions_expiry_idx ON http_sessions (expires_on);
		CREATE INDEX IF NOT EXISTS http_sessions_key_idx ON http_sessions (key);
	`)
	if err != nil {
		return fmt.Errorf("failed to create http_sessions table: %w", err)
	}
	return nil
}

func (r *Repository) Get(ctx context.Context, key string) (*entity.HTTPSession, error) {
	var session entity.HTTPSession
	err := r.pool.QueryRow(ctx, `
        SELECT id, key, data, created_on, modified_on, expires_on
        FROM sessions
        WHERE key = $1
    `, key).Scan(
		&session.ID,
		&session.Key,
		&session.Data,
		&session.CreatedOn,
		&session.ModifiedOn,
		&session.ExpiresOn,
	)
	if err != nil {
		return nil, fmt.Errorf("getting session: %w", err)
	}
	return &session, nil
}

func (r *Repository) Save(ctx context.Context, params CreateParams) error {
    log := r.logger.With().
        Str("method", "session.Save").
        Str("session_key", params.Key).
        Logger()

    now := time.Now()
    var query string
    var args []interface{}

    if params.IsNew {
        query = `
            INSERT INTO http_sessions (key, data, created_on, modified_on, expires_on)
            VALUES ($1, $2, $3, $4, $5)
        `
        args = []interface{}{params.Key, params.Data, now, now, params.ExpiresOn}
        log.Debug().Msg("inserting new session")
    } else {
        query = `
            UPDATE http_sessions 
            SET data = $1, modified_on = $2, expires_on = $3 
            WHERE key = $4
        `
        args = []interface{}{params.Data, now, params.ExpiresOn, params.Key}
        log.Debug().Msg("updating existing session")
    }

    _, err := r.pool.Exec(ctx, query, args...)
    if err != nil {
        log.Error().Err(err).Msg("database operation failed")
        return fmt.Errorf("saving session to database: %w", err)
    }

    return nil
}

func (r *Repository) Load(ctx context.Context, key string) (*entity.HTTPSession, error) {
    log := r.logger.With().
        Str("method", "session.Load").
        Str("session_key", key).
        Logger()

    var sess entity.HTTPSession
    err := r.pool.QueryRow(ctx, `
        SELECT id, key, data, created_on, modified_on, expires_on 
        FROM http_sessions 
        WHERE key = $1
    `, key).Scan(
        &sess.ID,
        &sess.Key,
        &sess.Data,
        &sess.CreatedOn,
        &sess.ModifiedOn,
        &sess.ExpiresOn,
    )
    
    if err == pgx.ErrNoRows {
        log.Debug().Msg("session not found")
        return nil, ErrNotFound
    }
    if err != nil {
        log.Error().Err(err).Msg("failed to load session from database")
        return nil, fmt.Errorf("loading session from database: %w", err)
    }

    log.Debug().Msg("session loaded successfully")
    return &sess, nil
}

func (r *Repository) destroy(ctx context.Context, session *sessions.Session) error {
	logger := zerolog.Ctx(ctx).With().
		Str("method", "destroy").
		Str("session_id", session.ID).
		Logger()

	logger.Debug().Msg("attempting to destroy session")

	result, err := r.pool.Exec(ctx, "DELETE FROM http_sessions WHERE key = $1", session.ID)
	if err != nil {
		logger.Error().
			Err(err).
			Str("session_id", session.ID).
			Msg("failed to delete session from database")
		return err
	}

	rowsAffected := result.RowsAffected()
	logger.Debug().
		Str("session_id", session.ID).
		Int64("rows_affected", rowsAffected).
		Msg("session destroy completed")

	return nil
}
