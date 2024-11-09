package middleware

import (
	"time"

	"github.com/DukeRupert/haven/store"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

// SessionMiddleware creates a new middleware for handling sessions with pgstore
func SessionMiddleware(store *store.PGStore) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			requestID := c.Request().Header.Get("X-Request-ID")
			path := c.Request().URL.Path
			method := c.Request().Method

			logger := log.With().
				Str("middleware", "session").
				Str("request_id", requestID).
				Str("path", path).
				Str("method", method).
				Logger()

			logger.Debug().Msg("Processing request with session middleware")

			// Get a session using the store
			session, err := store.Get(c.Request(), "session-key")
			if err != nil {
				logger.Error().
					Err(err).
					Msg("Failed to get session")
				return echo.NewHTTPError(500, "Error getting session")
			}

			logger.Debug().
				Bool("is_new_session", session.IsNew).
				Int("values_count", len(session.Values)).
				Msg("Session retrieved")

			// Add the session to the context
			c.Set("session", session)

			// Continue with the next handler
			err = next(c)
			if err != nil {
				logger.Error().
					Err(err).
					Msg("Handler error occurred")
				return err
			}

			// Save the session after the handler is done
			if err := session.Save(c.Request(), c.Response()); err != nil {
				logger.Error().
					Err(err).
					Msg("Failed to save session")
				return echo.NewHTTPError(500, "Error saving session")
			}

			// Log completion metrics
			logger.Info().
				Dur("duration_ms", time.Since(start)).
				Int("status", c.Response().Status).
				Bool("session_modified", len(session.Values) > 0).
				Msg("Session middleware completed")

			return nil
		}
	}
}
