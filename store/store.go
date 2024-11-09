package store

import (
	"context"
	"encoding/base32"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PGStore represents the currently configured session store.
type PGStore struct {
	Codecs  []securecookie.Codec
	Options *sessions.Options
	Path    string
	DbPool  *pgxpool.Pool
}

// PGSession type
type PGSession struct {
	ID         int64
	Key        string
	Data       string
	CreatedOn  time.Time
	ModifiedOn time.Time
	ExpiresOn  time.Time
}

// NewPGStore creates a new PGStore instance and a new pgxpool.Pool.
func NewPGStore(ctx context.Context, dbURL string, keyPairs ...[]byte) (*PGStore, error) {
	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing database URL: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("error creating connection pool: %w", err)
	}

	return NewPGStoreFromPool(ctx, pool, keyPairs...)
}

// NewPGStoreFromPool creates a new PGStore instance from an existing pool.
func NewPGStoreFromPool(ctx context.Context, pool *pgxpool.Pool, keyPairs ...[]byte) (*PGStore, error) {
	dbStore := &PGStore{
		Codecs: securecookie.CodecsFromPairs(keyPairs...),
		Options: &sessions.Options{
			Path:   "/",
			MaxAge: 86400 * 30,
		},
		DbPool: pool,
	}

	if err := dbStore.createSessionsTable(ctx); err != nil {
		return nil, err
	}

	return dbStore, nil
}

// Close closes the database connection pool.
func (db *PGStore) Close() {
	db.DbPool.Close()
}

// Get fetches a session for a given name after it has been added to the registry.
func (db *PGStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(db, name)
}

// New returns a new session for the given name without adding it to the registry.
func (db *PGStore) New(r *http.Request, name string) (*sessions.Session, error) {
	session := sessions.NewSession(db, name)
	if session == nil {
		return nil, nil
	}

	opts := *db.Options
	session.Options = &(opts)
	session.IsNew = true

	var err error
	if c, errCookie := r.Cookie(name); errCookie == nil {
		err = securecookie.DecodeMulti(name, c.Value, &session.ID, db.Codecs...)
		if err == nil {
			err = db.load(r.Context(), session)
			if err == nil {
				session.IsNew = false
			} else if err == pgx.ErrNoRows {
				err = nil
			}
		}
	}

	db.MaxAge(db.Options.MaxAge)
	return session, err
}

// Save saves the given session into the database and deletes cookies if needed
func (db *PGStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	if session.Options.MaxAge < 0 {
		if err := db.destroy(r.Context(), session); err != nil {
			return err
		}
		http.SetCookie(w, sessions.NewCookie(session.Name(), "", session.Options))
		return nil
	}

	if session.ID == "" {
		session.ID = strings.TrimRight(
			base32.StdEncoding.EncodeToString(
				securecookie.GenerateRandomKey(32),
			), "=")
	}

	if err := db.save(r.Context(), session); err != nil {
		return err
	}

	encoded, err := securecookie.EncodeMulti(session.Name(), session.ID, db.Codecs...)
	if err != nil {
		return err
	}

	http.SetCookie(w, sessions.NewCookie(session.Name(), encoded, session.Options))
	return nil
}

func (db *PGStore) load(ctx context.Context, session *sessions.Session) error {
	var s PGSession

	err := db.selectOne(ctx, &s, session.ID)
	if err != nil {
		return err
	}

	return securecookie.DecodeMulti(session.Name(), string(s.Data), &session.Values, db.Codecs...)
}

func (db *PGStore) save(ctx context.Context, session *sessions.Session) error {
	encoded, err := securecookie.EncodeMulti(session.Name(), session.Values, db.Codecs...)
	if err != nil {
		return err
	}

	crOn := session.Values["created_on"]
	exOn := session.Values["expires_on"]

	var expiresOn time.Time

	createdOn, ok := crOn.(time.Time)
	if !ok {
		createdOn = time.Now()
	}

	if exOn == nil {
		expiresOn = time.Now().Add(time.Second * time.Duration(session.Options.MaxAge))
	} else {
		expiresOn = exOn.(time.Time)
		if expiresOn.Sub(time.Now().Add(time.Second*time.Duration(session.Options.MaxAge))) < 0 {
			expiresOn = time.Now().Add(time.Second * time.Duration(session.Options.MaxAge))
		}
	}

	s := PGSession{
		Key:        session.ID,
		Data:       encoded,
		CreatedOn:  createdOn,
		ExpiresOn:  expiresOn,
		ModifiedOn: time.Now(),
	}

	if session.IsNew {
		return db.insert(ctx, &s)
	}

	return db.update(ctx, &s)
}

func (db *PGStore) destroy(ctx context.Context, session *sessions.Session) error {
	_, err := db.DbPool.Exec(ctx, "DELETE FROM http_sessions WHERE key = $1", session.ID)
	return err
}

func (db *PGStore) createSessionsTable(ctx context.Context) error {
	stmt := `DO $$
              BEGIN
              CREATE TABLE IF NOT EXISTS http_sessions (
              id BIGSERIAL PRIMARY KEY,
              key BYTEA,
              data BYTEA,
              created_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
              modified_on TIMESTAMPTZ,
              expires_on TIMESTAMPTZ);
              CREATE INDEX IF NOT EXISTS http_sessions_expiry_idx ON http_sessions (expires_on);
              CREATE INDEX IF NOT EXISTS http_sessions_key_idx ON http_sessions (key);
              EXCEPTION WHEN insufficient_privilege THEN
                IF NOT EXISTS (SELECT FROM pg_catalog.pg_tables WHERE schemaname = current_schema() AND tablename = 'http_sessions') THEN
                  RAISE;
                END IF;
              WHEN others THEN RAISE;
              END;
              $$;`

	_, err := db.DbPool.Exec(ctx, stmt)
	if err != nil {
		return fmt.Errorf("unable to create http_sessions table: %w", err)
	}

	return nil
}

func (db *PGStore) selectOne(ctx context.Context, s *PGSession, key string) error {
	stmt := "SELECT id, key, data, created_on, modified_on, expires_on FROM http_sessions WHERE key = $1"
	err := db.DbPool.QueryRow(ctx, stmt, key).Scan(
		&s.ID,
		&s.Key,
		&s.Data,
		&s.CreatedOn,
		&s.ModifiedOn,
		&s.ExpiresOn,
	)
	if err != nil {
		return fmt.Errorf("unable to find session: %w", err)
	}

	return nil
}

func (db *PGStore) insert(ctx context.Context, s *PGSession) error {
	stmt := `INSERT INTO http_sessions (key, data, created_on, modified_on, expires_on)
           VALUES ($1, $2, $3, $4, $5)`
	_, err := db.DbPool.Exec(ctx, stmt, s.Key, s.Data, s.CreatedOn, s.ModifiedOn, s.ExpiresOn)
	return err
}

func (db *PGStore) update(ctx context.Context, s *PGSession) error {
	stmt := `UPDATE http_sessions SET data=$1, modified_on=$2, expires_on=$3 WHERE key=$4`
	_, err := db.DbPool.Exec(ctx, stmt, s.Data, s.ModifiedOn, s.ExpiresOn, s.Key)
	return err
}

// MaxLength and MaxAge methods remain unchanged as they don't interact with the database
func (db *PGStore) MaxLength(l int) {
	for _, c := range db.Codecs {
		if codec, ok := c.(*securecookie.SecureCookie); ok {
			codec.MaxLength(l)
		}
	}
}

func (db *PGStore) MaxAge(age int) {
	db.Options.MaxAge = age
	for _, codec := range db.Codecs {
		if sc, ok := codec.(*securecookie.SecureCookie); ok {
			sc.MaxAge(age)
		}
	}
}
