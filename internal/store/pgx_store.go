// internal/store/pgx_store.go
package store

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/DukeRupert/haven/internal/repository/session"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/rs/zerolog"
)

type PgxStore struct {
	sessions *session.Repository
	Codecs   []securecookie.Codec
	Options  *sessions.Options
	logger   zerolog.Logger
}

// SessionKeys defines constants for session value keys
const (
	DefaultSessionName     = "session"
	SessionKeyUserID       = "user_id"
	SessionKeyRole         = "role"
	SessionKeyInitials     = "initials"
	SessionKeyFacilityID   = "facility_id"
	SessionKeyFacilityCode = "facility_code"
	SessionKeyLastAccess   = "last_access"
)

func NewPgxStore(repo *session.Repository, keyPairs ...[]byte) (*PgxStore, error) {
	store := &PgxStore{
		sessions: repo,
		Codecs:   securecookie.CodecsFromPairs(keyPairs...),
		Options: &sessions.Options{
			Path:     "/",
			MaxAge:   86400 * 7,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		},
	}

	// Create table if needed
	err := repo.CreateSessionsTable()
	if err != nil {
		return nil, err
	}

	return store, nil
}

func (s *PgxStore) New(r *http.Request, name string) (*sessions.Session, error) {
	session := sessions.NewSession(s, name)
	opts := *s.Options
	session.Options = &opts
	session.IsNew = true

	// Generate a unique session ID
	session.ID = generateSessionID() // Add this line

	return session, nil
}

// Add this helper function
func generateSessionID() string {
	// Generate a random 32 byte (256 bit) value
	b := securecookie.GenerateRandomKey(32)
	return fmt.Sprintf("%x", b)
}

func (s *PgxStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	session := sessions.NewSession(s, name)
	opts := *s.Options
	session.Options = &opts
	session.IsNew = true

	// Get cookie
	cookie, err := r.Cookie(name)
	if err != nil {
		// No cookie found, return new session
		return session, nil
	}

	// Decode session ID from cookie
	var id string
	if err := securecookie.DecodeMulti(name, cookie.Value, &id, s.Codecs...); err != nil {
		// Invalid cookie, return new session
		return session, nil
	}

	// Set session ID
	session.ID = id
	session.IsNew = false

	// Load session data
	if err := s.Load(r.Context(), session); err != nil {
		// Failed to load session, return new one
		return sessions.NewSession(s, name), nil
	}

	return session, nil
}

// encodeSession handles the session-specific encoding
func (s *PgxStore) encodeSession(session *sessions.Session) ([]byte, error) {
	encoded, err := securecookie.EncodeMulti(session.Name(), session.Values, s.Codecs...)
	if err != nil {
		s.logger.Error().
			Err(err).
			Interface("values", session.Values).
			Msg("failed to encode session values")
		return nil, fmt.Errorf("encoding session values: %w", err)
	}
	return []byte(encoded), nil
}

func (s *PgxStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	// Skip saving if session values are empty
	if len(session.Values) == 0 {
		return nil
	}

	// Set cookie before saving to database
	if session.ID == "" {
		session.ID = generateSessionID()
	}

	// Set cookie
	encoded, err := securecookie.EncodeMulti(session.Name(), session.ID, s.Codecs...)
	if err != nil {
		return err
	}

	cookie := sessions.NewCookie(session.Name(), encoded, session.Options)
	http.SetCookie(w, cookie)

	// Save to database
	return s.save(r.Context(), session)
}

// save is your internal method that handles the actual saving
func (s *PgxStore) save(ctx context.Context, sess *sessions.Session) error {
	encoded, err := s.encodeSession(sess)
	if err != nil {
		return err
	}

	expiresOn := time.Now().Add(time.Second * time.Duration(sess.Options.MaxAge))

	params := session.CreateParams{
		Key:       sess.ID,
		Data:      encoded,
		ExpiresOn: expiresOn,
		IsNew:     sess.IsNew,
	}

	return s.sessions.Save(ctx, params)
}

// decodeSession handles the session-specific decoding
func (s *PgxStore) decodeSession(name string, data []byte) (map[interface{}]interface{}, error) {
	var values map[interface{}]interface{}
	err := securecookie.DecodeMulti(name, string(data), &values, s.Codecs...)
	if err != nil {
		s.logger.Error().
			Err(err).
			Msg("failed to decode session data")
		return nil, fmt.Errorf("decoding session values: %w", err)
	}
	return values, nil
}

// Load handles the high-level session loading logic
func (s *PgxStore) Load(ctx context.Context, session *sessions.Session) error {
	sess, err := s.sessions.Load(ctx, session.ID)
	if err != nil {
		return fmt.Errorf("loading session: %w", err)
	}

	values, err := s.decodeSession(session.Name(), sess.Data)
	if err != nil {
		return err
	}

	session.Values = values
	session.IsNew = false

	s.logger.Debug().
		Str("session_id", session.ID).
		Msg("session loaded and decoded successfully")

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
