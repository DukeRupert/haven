package middleware

import (
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

// Global singleton
var store *sessions.CookieStore

// Initialize store with secret key

func InitStore(secret []byte) {
	store = sessions.NewCookieStore(secret)
}

type SessionConfig struct {
	Name     string
	MaxAge   int
	Path     string
	HttpOnly bool
	Secure   bool
	Domain   string
}

// DefaultSessionConfig returns default configuration
func DefaultSessionConfig() SessionConfig {
	return SessionConfig{
		Name:     "session",
		MaxAge:   86400 * 7, // 7 days
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
	}
}

func NewSessionMiddleware(config SessionConfig) echo.MiddlewareFunc {
	if store == nil {
		log.Fatal().Msg("Session store not initialized. Call InitStore() first")
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			logger := log.With().
				Str("middleware", "session").
				Str("session_name", config.Name).
				Logger()

			// Try to get the session
			sess, err := store.Get(c.Request(), config.Name)
			if err != nil {
				// Log the error
				logger.Warn().
					Err(err).
					Msg("Failed to get existing session, creating new one")

				// Clear the invalid cookie
				c.SetCookie(&http.Cookie{
					Name:   config.Name,
					Value:  "",
					Path:   "/",
					MaxAge: -1,
				})

				// Create a new session
				sess, err = store.New(c.Request(), config.Name)
				if err != nil {
					logger.Error().
						Err(err).
						Msg("Failed to create new session")
					return c.NoContent(http.StatusInternalServerError)
				}
			}

			// Configure session options
			sess.Options = &sessions.Options{
				Path:     config.Path,
				MaxAge:   config.MaxAge,
				HttpOnly: config.HttpOnly,
				Secure:   config.Secure,
				Domain:   config.Domain,
			}

			// Add the session to the context
			c.Set("session", sess)

			// Continue with the next handler
			err = next(c)

			// Save the session after processing
			if err := sess.Save(c.Request(), c.Response().Writer); err != nil {
				logger.Error().
					Err(err).
					Msg("Failed to save session")
				return err
			}

			logger.Debug().
				Dur("duration", time.Since(start)).
				Bool("is_new_session", sess.IsNew).
				Msg("Session middleware completed")

			return err
		}
	}
}
