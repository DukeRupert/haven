package store

import (
	"context"
	"encoding/base32"
	"encoding/gob"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/DukeRupert/haven/db"
	"github.com/DukeRupert/haven/models"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
)

// HTTPSession represents a session stored in the database
type HTTPSession struct {
	ID         int64      `db:"id" json:"id"`
	Key        string     `db:"key" json:"key"`
	Data       []byte     `db:"data" json:"data"`
	CreatedOn  time.Time  `db:"created_on" json:"created_on"`
	ModifiedOn *time.Time `db:"modified_on" json:"modified_on,omitempty"`
	ExpiresOn  *time.Time `db:"expires_on" json:"expires_on,omitempty"`
}

// PgxStore represents the session store backed by PostgreSQL
type PgxStore struct {
	db      *db.DB
	Codecs  []securecookie.Codec
	Options *sessions.Options
}

const (
	DefaultSessionName = "session" // Use this consistently across your application
)

var logger zerolog.Logger

func init() {
	// Initialize zerolog
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger = zerolog.New(os.Stdout).With().Timestamp().Caller().Logger()

	// Register types for session storage
	gob.Register(time.Time{})
	gob.Register(models.UserRole(""))
	logger.Debug().Msg("registered time.Time with gob encoder")
}

// NewPgxStore creates a new PgxStore instance
func NewPgxStore(db *db.DB, keyPairs ...[]byte) (*PgxStore, error) {
	store := &PgxStore{
		db:     db,
		Codecs: securecookie.CodecsFromPairs(keyPairs...),
		Options: &sessions.Options{
			Path:     "/",
			MaxAge:   86400 * 7, // 7 days
			HttpOnly: true,
			Secure:   true, // Enable for HTTPS
			SameSite: http.SameSiteStrictMode,
		},
	}

	err := store.createSessionsTable()
	if err != nil {
		return nil, err
	}

	return store, nil
}

// Get fetches a session for a given name
func (s *PgxStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(s, name)
}

// New returns a new session for the given name
func (s *PgxStore) New(r *http.Request, name string) (*sessions.Session, error) {
	session := sessions.NewSession(s, name)
	if session == nil {
		return nil, nil
	}

	opts := *s.Options
	session.Options = &opts
	session.IsNew = true

	if c, errCookie := r.Cookie(name); errCookie == nil {
		err := securecookie.DecodeMulti(name, c.Value, &session.ID, s.Codecs...)
		if err == nil {
			err = s.load(r.Context(), session)
			if err == nil {
				session.IsNew = false
			} else if errors.Is(err, pgx.ErrNoRows) {
				err = nil
			}
		}
	}

	s.MaxAge(s.Options.MaxAge)
	return session, nil
}

// Save saves the session to the database
func (s *PgxStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	logger := zerolog.Ctx(r.Context()).With().
		Str("method", "Save").
		Str("session_id", session.ID).
		Logger()

	if session.Options.MaxAge < 0 {
		logger.Debug().Msg("deleting session due to negative MaxAge")

		if session.ID != "" {
			if err := s.destroy(r.Context(), session); err != nil {
				logger.Error().
					Err(err).
					Msg("failed to destroy session")
				return err
			}
			logger.Debug().Msg("session destroyed successfully")
		} else {
			logger.Warn().Msg("attempted to delete session with empty ID")
		}

		http.SetCookie(w, sessions.NewCookie(session.Name(), "", session.Options))
		return nil
	}

	if session.ID == "" {
		session.ID = strings.TrimRight(
			base32.StdEncoding.EncodeToString(
				securecookie.GenerateRandomKey(32),
			), "=")
		logger.Debug().
			Str("session_id", session.ID).
			Msg("generated new session ID")
	}

	if err := s.save(r.Context(), session); err != nil {
		return err
	}

	encoded, err := securecookie.EncodeMulti(session.Name(), session.ID, s.Codecs...)
	if err != nil {
		return err
	}

	http.SetCookie(w, sessions.NewCookie(session.Name(), encoded, session.Options))
	return nil
}

// MaxAge sets the maximum age for the store and the underlying cookie
func (s *PgxStore) MaxAge(age int) {
	s.Options.MaxAge = age
	for _, codec := range s.Codecs {
		if sc, ok := codec.(*securecookie.SecureCookie); ok {
			sc.MaxAge(age)
		}
	}
}

func (s *PgxStore) save(ctx context.Context, session *sessions.Session) error {
	log := logger.With().
		Str("method", "PgxStore.save").
		Str("session_id", session.ID).
		Logger()

	encoded, err := securecookie.EncodeMulti(session.Name(), session.Values, s.Codecs...)
	if err != nil {
		log.Error().Err(err).Interface("values", session.Values).Msg("failed to encode session values")
		return err
	}

	now := time.Now()
	expiresOn := now.Add(time.Second * time.Duration(session.Options.MaxAge))

	var query string
	var args []interface{}

	if session.IsNew {
		query = `INSERT INTO http_sessions (key, data, created_on, modified_on, expires_on)
                VALUES ($1, $2, $3, $4, $5)`
		args = []interface{}{session.ID, []byte(encoded), now, &now, &expiresOn}
		log.Debug().Msg("inserting new session")
	} else {
		query = `UPDATE http_sessions 
                SET data = $1, modified_on = $2, expires_on = $3 
                WHERE key = $4`
		args = []interface{}{[]byte(encoded), &now, &expiresOn, session.ID}
		log.Debug().Msg("updating existing session")
	}

	_, err = s.db.Exec(ctx, query, args...)
	if err != nil {
		log.Error().Err(err).Msg("database operation failed")
		return err
	}

	return nil
}

func (s *PgxStore) load(ctx context.Context, session *sessions.Session) error {
	log := logger.With().
		Str("method", "PgxStore.load").
		Str("session_id", session.ID).
		Logger()

	var sess HTTPSession
	err := s.db.QueryRow(ctx,
		`SELECT id, key, data, created_on, modified_on, expires_on 
         FROM http_sessions WHERE key = $1`,
		session.ID).Scan(&sess.ID, &sess.Key, &sess.Data, &sess.CreatedOn, &sess.ModifiedOn, &sess.ExpiresOn)
	if err != nil {
		log.Error().Err(err).Msg("failed to load session from database")
		return err
	}

	err = securecookie.DecodeMulti(session.Name(), string(sess.Data), &session.Values, s.Codecs...)
	if err != nil {
		log.Error().Err(err).Msg("failed to decode session data")
		return err
	}

	log.Debug().
		Msg("session loaded successfully")

	return nil
}

func (s *PgxStore) destroy(ctx context.Context, session *sessions.Session) error {
	logger := zerolog.Ctx(ctx).With().
		Str("method", "destroy").
		Str("session_id", session.ID).
		Logger()

	logger.Debug().Msg("attempting to destroy session")

	result, err := s.db.Exec(ctx, "DELETE FROM http_sessions WHERE key = $1", session.ID)
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

func (s *PgxStore) createSessionsTable() error {
	ctx := context.Background()
	_, err := s.db.Exec(ctx, `
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
