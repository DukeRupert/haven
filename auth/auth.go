package auth

import (
	"context"
	"encoding/gob"
	"errors"
	"net/http"
	"time"

	"github.com/DukeRupert/haven/models"
	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID         int       `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Initials   string    `json:"initials"`
	Email      string    `json:"email"`
	Password   string    `json:"-"` // Hashed password
	FacilityID int       `json:"facility_id"`
	Role       string    `json:"role" validate:"required,oneof=super admin user"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginResponse struct {
	User    User   `json:"user"`
	Message string `json:"message"`
}

type AuthHandler struct {
	db     *pgxpool.Pool
	store  sessions.Store // Add store to the handler
	logger zerolog.Logger
}

func init() {
	// Register types that will be stored in sessions
	gob.Register(time.Time{})
}

func NewAuthHandler(pool *pgxpool.Pool, store sessions.Store, logger zerolog.Logger) *AuthHandler {
	return &AuthHandler{
		db:     pool,
		store:  store,
		logger: logger.With().Str("component", "auth").Logger(),
	}
}

// LoginParams represents the expected login request body
type LoginParams struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginHandler handles user login requests
func (h *AuthHandler) LoginHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get session using echo-contrib/session
		sess, err := session.Get("session", c)
		if err != nil {
			h.logger.Error().Err(err).Msg("Failed to get session")
			return echo.NewHTTPError(http.StatusInternalServerError, "session error")
		}

		// Parse login params
		params := new(LoginParams)
		if err := c.Bind(params); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
		}

		// Authenticate user
		user, err := authenticateUser(c.Request().Context(), h.db, params.Email, params.Password)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
		}

		log.Debug().Int("user_id", user.ID).Str("role", string(user.Role)).Msg("Setting session values")

		// Set session values
		sess.Values["user_id"] = user.ID
		sess.Values["role"] = user.Role
		sess.Values["last_access"] = time.Now()

		// Save session
		if err := sess.Save(c.Request(), c.Response()); err != nil {
			h.logger.Error().Err(err).Msg("Failed to save session")
			return echo.NewHTTPError(http.StatusInternalServerError, "session error")
		}

		return c.JSON(http.StatusOK, user)
	}
}

var ErrInvalidCredentials = errors.New("invalid credentials")

// authenticateUser verifies the email and password combination
func authenticateUser(ctx context.Context, db *pgxpool.Pool, email, password string) (*models.User, error) {
	log := zerolog.Ctx(ctx).With().Str("method", "authenticateUser").Logger()

	var user models.User
	err := db.QueryRow(ctx,
		`SELECT id, created_at, updated_at, first_name, last_name, 
                initials, email, password, facility_id, role 
         FROM users 
         WHERE email = $1`,
		email).Scan(
		&user.ID, &user.CreatedAt, &user.UpdatedAt, &user.FirstName,
		&user.LastName, &user.Initials, &user.Email, &user.Password,
		&user.FacilityID, &user.Role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Debug().Str("email", email).Msg("user not found")
			return nil, ErrInvalidCredentials
		}
		log.Error().Err(err).Msg("database error")
		return nil, err
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		log.Debug().Str("email", email).Msg("invalid password")
		return nil, ErrInvalidCredentials
	}

	return &user, nil
}

// SessionMiddleware handles session management for all HTTP requests.
// It performs the following tasks:
// 1. Retrieves or creates a session for each request
// 2. Logs session and cookie information for debugging
// 3. Configures default session options for new sessions
// 4. Makes the session available to subsequent handlers via context
// 5. Automatically saves any session changes after the handler completes
//
// Usage:
//
//	e.Use(authHandler.SessionMiddleware())
func (h *AuthHandler) SessionMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			logger := h.logger.With().
				Str("middleware", "session").
				Str("path", c.Path()).
				Str("method", c.Request().Method).
				Logger()

			// Log all cookies for debugging
			cookies := c.Request().Cookies()
			for _, cookie := range cookies {
				logger.Debug().
					Str("cookie_name", cookie.Name).
					Str("cookie_value", cookie.Value).
					Msg("Request cookie found")
			}

			// Get or create session from store
			sess, err := session.Get("session", c)
			// sess, err := h.store.Get(c.Request(), "session-key")
			if err != nil {
				logger.Error().
					Err(err).
					Msg("Failed to get/create session")

				// Try to create a new session
				sess, err = h.store.New(c.Request(), "session-key")
				if err != nil {
					logger.Error().
						Err(err).
						Msg("Failed to create new session")
					return echo.NewHTTPError(http.StatusInternalServerError, "Session error")
				}

				// Configure new session
				sess.Options = &sessions.Options{
					Path:     "/",
					MaxAge:   86400 * 7,
					HttpOnly: true,
					Secure:   false, // Set to true in production with HTTPS
					SameSite: http.SameSiteLaxMode,
				}
			}

			// Add session to context
			c.Set("session", sess)

			// Execute next handler
			if err := next(c); err != nil {
				return err
			}

			// Save any changes made to the session
			if err := sess.Save(c.Request(), c.Response()); err != nil {
				logger.Error().
					Err(err).
					Interface("session_values", sess.Values).
					Msg("Failed to save session")
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save session")
			}

			return nil
		}
	}
}

// AuthMiddleware protects routes by verifying user authentication.
// It expects a session to be present in the context (set by SessionMiddleware).
// Unauthenticated requests are redirected to the login page or receive 401 for API requests.
func (h *AuthHandler) AuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			logger := h.logger.With().
				Str("middleware", "AuthMiddleware()").
				Str("path", c.Path()).
				Logger()

			// Get session from context (previously set by SessionMiddleware)
			sess, err := session.Get("session", c)
			// sess, ok := c.Get("session").(*sessions.Session)
			if err != nil {
				logger.Error().Msg("no session found in context")
				return redirectToLogin(c)
			}

			// Check if user is authenticated
			userID, ok := sess.Values["user_id"].(int)
			if !ok || userID == 0 {
				logger.Debug().Msg("no valid user_id in session")
				return redirectToLogin(c)
			}

			// Check role if present
			role, ok := sess.Values["role"].(models.UserRole)
			if !ok {
				logger.Debug().Str("role", string(role)).Int("user_id", userID).Msg("no valid role in session")
				return redirectToLogin(c)
			}

			// Add user info to context for use in handlers
			c.Set("user_id", userID)
			c.Set("user_role", role)

			return next(c)
		}
	}
}

func redirectToLogin(c echo.Context) error {
	// If it's an API request, return 401
	if isAPIRequest(c) {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}
	// For regular requests, redirect to login
	return c.Redirect(http.StatusTemporaryRedirect, "/login")
}

func isAPIRequest(c echo.Context) bool {
	// Check Accept header
	if c.Request().Header.Get("Accept") == "application/json" {
		return true
	}
	// Check if it's an XHR request
	if c.Request().Header.Get("X-Requested-With") == "XMLHttpRequest" {
		return true
	}
	return false
}
