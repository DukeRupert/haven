package auth

import (
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/DukeRupert/haven/db"
	"github.com/DukeRupert/haven/models"
	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
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
	db     *db.DB
	store  sessions.Store // Add store to the handler
	logger zerolog.Logger
}

const (
	DefaultSessionName = "session"
)

func init() {
	// Register types that will be stored in sessions
	gob.Register(time.Time{})
}

func NewAuthHandler(pool *db.DB, store sessions.Store, logger zerolog.Logger) *AuthHandler {
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
		// Get session
		sess, err := session.Get(DefaultSessionName, c)
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

		// Query facility information if user has an associated facility
		var facility *db.Facility

		facility, err = h.db.GetFacilityByID(c.Request().Context(), user.FacilityID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				h.logger.Warn().
					Int("facility_id", user.FacilityID).
					Msg("Facility not found for authenticated user")
			} else {
				h.logger.Error().
					Err(err).
					Int("facility_id", user.FacilityID).
					Msg("Failed to query facility information")
				return echo.NewHTTPError(http.StatusInternalServerError, "database error")
			}
		} else {
			// Add facility to session if found
			sess.Values["facility_id"] = facility.ID
			sess.Values["facility_name"] = facility.Name
			sess.Values["facility_code"] = facility.Code
		}

		// Set session values
		sess.Values["user_id"] = user.ID
		sess.Values["role"] = user.Role
		sess.Values["last_access"] = time.Now()

		// Save session
		if err := sess.Save(c.Request(), c.Response()); err != nil {
			h.logger.Error().Err(err).Msg("Failed to save session")
			return echo.NewHTTPError(http.StatusInternalServerError, "session error")
		}

		// Redirect to facility page
		redirectURL := fmt.Sprintf("/app/%s/", facility.Code)
		return c.Redirect(http.StatusSeeOther, redirectURL)
	}
}

var ErrInvalidCredentials = errors.New("invalid credentials")

// authenticateUser verifies the email and password combination
func authenticateUser(ctx context.Context, db *db.DB, email, password string) (*models.User, error) {
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

func (h *AuthHandler) LogoutHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		logger := h.logger.With().Str("handler", "LogoutHandler").Logger()

		// Try to get the existing session by name
		sess, err := session.Get(DefaultSessionName, c)
		if err != nil {
			logger.Error().Err(err).Msg("failed to get session")
			return c.Redirect(http.StatusTemporaryRedirect, "/login")
		}

		// Log full session state for debugging
		logger.Debug().
			Bool("is_new", sess.IsNew).
			Str("session_id", sess.ID).
			Interface("values", convertSessionValues(sess.Values)).
			Interface("options", sess.Options).
			Msg("current session state")

		// If we don't have a valid session, just redirect to login
		if sess.IsNew {
			logger.Debug().Msg("no existing session found")
			return c.Redirect(http.StatusTemporaryRedirect, "/login")
		}

		// We have a valid session, proceed with logout
		logger.Debug().
			Str("session_id", sess.ID).
			Msg("proceeding with session deletion")

		// Clear all session values
		sess.Values = make(map[interface{}]interface{})

		// Set the session to expire
		sess.Options.MaxAge = -1

		// Save the session (this will trigger deletion due to MaxAge = -1)
		if err = sess.Save(c.Request(), c.Response()); err != nil {
			logger.Error().
				Err(err).
				Str("session_id", sess.ID).
				Msg("failed to delete session")
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to logout")
		}

		logger.Debug().
			Str("session_id", sess.ID).
			Msg("logout successful")

		return c.Redirect(http.StatusOK, "/login")
	}
}

// Helper function to convert session values for logging
func convertSessionValues(values map[interface{}]interface{}) map[string]interface{} {
	converted := make(map[string]interface{})
	for k, v := range values {
		if key, ok := k.(string); ok {
			converted[key] = v
		}
	}
	return converted
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

			h.logger.Debug().
				Str("path", c.Path()).
				Str("method", c.Request().Method).
				Msg("auth middleware hit")

			// Get session from context (previously set by SessionMiddleware)
			sess, err := session.Get(DefaultSessionName, c)
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

// RoleAuthMiddleware protects routes by checking if the user has sufficient role permissions.
// It verifies session auth, validates user_id and role, and enforces minimum role requirements.
// Returns 403 Forbidden for insufficient permissions or redirects to login for auth failures.
// Usage: e.GET("/admin", handler, h.RoleAuthMiddleware("super"))
func (h *AuthHandler) RoleAuthMiddleware(minimumRole models.UserRole) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			logger := h.logger.With().
				Str("middleware", "RoleAuthMiddleware()").
				Str("path", c.Path()).
				Logger()

			logger.Debug().Msg("Executing RoleAuthMiddleware()")
			// Get session from context
			sess, err := session.Get(DefaultSessionName, c)
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

			// Check if user's role meets minimum required role
			if !isAtLeastRole(string(role), string(minimumRole)) {
				logger.Debug().
					Str("user_role", string(role)).
					Msg("insufficient role permissions")
				return echo.NewHTTPError(http.StatusForbidden, "Insufficient permissions")
			}
			return next(c)
		}
	}
}

// RedirectIfAuthenticated middleware checks if a user is already logged in when accessing the login page
// and redirects them to their facility page if they are.
func (h *AuthHandler) RedirectIfAuthenticated() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			logger := h.logger.With().
				Str("middleware", "RedirectIfAuthenticated()").
				Str("path", c.Path()).
				Logger()

			logger.Debug().
				Str("method", c.Request().Method).
				Msg("redirect middleware hit")

			// Get session
			sess, err := session.Get(DefaultSessionName, c)
			if err != nil {
				logger.Debug().Err(err).Msg("no valid session found")
				return next(c)
			}

			// Check if user is authenticated
			userID, ok := sess.Values["user_id"].(int)
			if !ok || userID == 0 {
				logger.Debug().Msg("no valid user_id in session")
				return next(c)
			}

			// Check if facility code exists in session
			facilityCode, ok := sess.Values["facility_code"].(string)
			if !ok || facilityCode == "" {
				logger.Debug().
					Int("user_id", userID).
					Msg("no valid facility code in session")
				return next(c)
			}

			// User is authenticated and has facility, redirect to facility page
			logger.Debug().
				Int("user_id", userID).
				Str("facility_id", facilityCode).
				Msg("redirecting authenticated user from login page to facility")

			redirectURL := fmt.Sprintf("/app/%s/", facilityCode)
			return c.Redirect(http.StatusSeeOther, redirectURL)
		}
	}
}

// isAtLeastRole checks if the current role meets or exceeds the minimum required role
func isAtLeastRole(currentRole string, minimumRole string) bool {
	roleValues := map[string]int{
		"user":  1,
		"admin": 2,
		"super": 3,
	}

	currentLevel, ok := roleValues[currentRole]
	if !ok {
		return false
	}

	requiredLevel, ok := roleValues[minimumRole]
	if !ok {
		return false
	}

	return currentLevel >= requiredLevel
}

func redirectToLogin(c echo.Context) error {
	// If it's an API request, return 401
	if isAPIRequest(c) {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}
	// For regular requests, redirect to login
	return c.Redirect(http.StatusOK, "/login")
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
