package middleware

import (
	"context"
	"encoding/gob"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
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
	pool  *pgxpool.Pool
	store sessions.Store // Add store to the handler
}

func init() {
	// Register types that will be stored in sessions
	gob.Register(time.Time{})
}

// NewAuthHandler creates a new auth handler with both pool and store
func NewAuthHandler(pool *pgxpool.Pool, store sessions.Store) *AuthHandler {
	return &AuthHandler{
		pool:  pool,
		store: store,
	}
}

func (h *AuthHandler) SessionMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			logger := log.With().
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
			sess, err := h.store.Get(c.Request(), "session-key")
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

			logger.Debug().
				Bool("is_new", sess.IsNew).
				Interface("session_values", sess.Values).
				Interface("session_options", sess.Options).
				Msg("Session state")

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

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
	RoleSuper = "super"
)

// RoleLevel maps roles to their hierarchy level
var RoleLevel = map[string]int{
	RoleUser:  1,
	RoleAdmin: 2,
	RoleSuper: 3,
}

// RequireAuth ensures a valid session exists
func (h *AuthHandler) RequireAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			logger := log.With().
				Str("middleware", "auth").
				Str("path", c.Path()).
				Str("method", c.Request().Method).
				Logger()

				// Get session from context - this should use c.Get()
			sess, ok := c.Get("session").(*sessions.Session)
			if !ok {
				logger.Warn().Msg("No session in context")
				return c.Redirect(http.StatusSeeOther, "/login")
			}

			userID, ok := sess.Values["user_id"]
			if !ok || userID == nil {
				logger.Warn().Msg("No user ID in session")
				return c.Redirect(http.StatusSeeOther, "/login")
			}

			logger.Debug().
				Interface("user_id", userID).
				Interface("session_values", sess.Values).
				Msg("User authenticated")

			return next(c)
		}
	}
}

// RequireRole ensures the user has the required role or higher
func (h *AuthHandler) RequireRole(requiredRole string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			logger := log.With().
				Str("middleware", "role").
				Str("path", c.Path()).
				Str("required_role", requiredRole).
				Logger()

			sess, ok := c.Get("session").(*sessions.Session)
			if !ok {
				logger.Warn().Msg("No session found")
				return c.Redirect(http.StatusSeeOther, "/login")
			}

			role, ok := sess.Values["user_role"].(string)
			if !ok {
				logger.Warn().Msg("No role found in context")
				return c.Redirect(http.StatusSeeOther, "/login")
			}

			// Check if user's role level meets the required level
			userLevel := RoleLevel[role]
			requiredLevel := RoleLevel[requiredRole]

			if userLevel < requiredLevel {
				logger.Warn().
					Str("user_role", role).
					Msg("Insufficient permissions")

				// If user is authenticated but lacks permission, redirect to appropriate page
				if userLevel >= RoleLevel[RoleUser] {
					return c.Redirect(http.StatusSeeOther, "/app")
				}
				return c.Redirect(http.StatusSeeOther, "/login")
			}

			return next(c)
		}
	}
}

// RoleBasedMiddleware determines the required role based on the URL path
func (h *AuthHandler) RoleBasedMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path

			switch {
			case strings.HasPrefix(path, "/app/super"):
				return h.RequireRole(RoleSuper)(next)(c)
			case strings.HasPrefix(path, "/app/admin"):
				return h.RequireRole(RoleAdmin)(next)(c)
			case strings.HasPrefix(path, "/app"):
				return h.RequireRole(RoleUser)(next)(c)
			default:
				return next(c)
			}
		}
	}
}

func (h *AuthHandler) Login(c echo.Context) error {
	ctx := c.Request().Context()
	logger := log.With().
		Str("handler", "login").
		Str("path", c.Path()).
		Str("method", c.Request().Method).
		Logger()

	// Parse login request
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		logger.Error().Err(err).Msg("Failed to parse login request")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid login request")
	}

	logger.Info().Str("email", req.Email).Str("password", req.Password).Msg("Parsed formdata")

	// Create/Get session directly from store since this is the login flow
	sess, error := h.store.Get(c.Request(), "session-key")
	if error != nil {
		logger.Error().Err(error).Msg("Failed to create session")
		return echo.NewHTTPError(http.StatusInternalServerError, "Session error")
	}

	// Check if user is already logged in
	if sess.Values["user_id"] != nil {
		logger.Info().Msg("User already logged in")
		return echo.NewHTTPError(http.StatusBadRequest, "Already logged in")
	}

	// Get user from database
	var user User
	logger.Info().Str("user", req.Email).Msg("Querying database for user")
	err := h.pool.QueryRow(ctx, `
		SELECT 
			id, created_at, updated_at, first_name, last_name, 
			initials, email, password, facility_id, role
		FROM users 
		WHERE email = $1`,
		req.Email,
	).Scan(
		&user.ID, &user.CreatedAt, &user.UpdatedAt, &user.FirstName, &user.LastName,
		&user.Initials, &user.Email, &user.Password, &user.FacilityID, &user.Role,
	)

	if err == pgx.ErrNoRows {
		logger.Warn().Str("email", req.Email).Msg("User not found")
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid credentials")
	} else if err != nil {
		logger.Error().Err(err).Msg("Database error")
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	// Compare password with hash
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		logger.Warn().
			Str("email", req.Email).
			Msg("Invalid password attempt")
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid credentials")
	}

	// Set session values
	sess.Values["user_id"] = user.ID
	sess.Values["email"] = user.Email
	sess.Values["role"] = user.Role
	sess.Values["facility_id"] = user.FacilityID
	sess.Values["logged_in_at"] = time.Now()

	// Save session
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		logger.Error().
			Err(err).
			Interface("session_values", sess.Values).
			Msg("Failed to save session")
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create session")
	}

	logger.Info().
		Int("user_id", user.ID).
		Str("email", user.Email).
		Str("role", user.Role).
		Int("facility_id", user.FacilityID).
		Msg("User logged in successfully")

	return c.JSON(http.StatusOK, LoginResponse{
		User:    user,
		Message: "Login successful",
	})
}

func (h *AuthHandler) Logout(c echo.Context) error {
	logger := log.With().
		Str("handler", "logout").
		Str("path", c.Path()).
		Str("method", c.Request().Method).
		Logger()

	// Get session
	sess, err := h.store.Get(c.Request(), "session-key")
	if err != nil {
		logger.Error().Msg("Failed to get session from context")
		return echo.NewHTTPError(http.StatusInternalServerError, "Session not found")
	}

	// Check if user is logged in
	userID, ok := sess.Values["user_id"].(int)
	if !ok {
		logger.Warn().Msg("No user found in session")
		return echo.NewHTTPError(http.StatusBadRequest, "Not logged in")
	}

	// Clear session
	sess.Options.MaxAge = -1
	sess.Values = make(map[interface{}]interface{})

	logger.Info().
		Int("user_id", userID).
		Msg("User logged out successfully")

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Logout successful",
	})
}

// CreateUser creates a new user with a hashed password
func (h *AuthHandler) CreateUser(ctx context.Context, user *User) error {
	logger := log.With().
		Str("handler", "create_user").
		Str("email", user.Email).
		Logger()

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to hash password")
		return err
	}

	// Generate initials if not provided
	if user.Initials == "" {
		user.Initials = generateInitials(user.FirstName, user.LastName)
	}

	// Start a transaction
	tx, err := h.pool.Begin(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to start transaction")
		return err
	}
	defer tx.Rollback(ctx)

	// Check if email already exists
	var exists bool
	err = tx.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`,
		user.Email,
	).Scan(&exists)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to check email existence")
		return err
	}

	if exists {
		logger.Warn().Msg("Email already exists")
		return echo.NewHTTPError(http.StatusConflict, "Email already exists")
	}

	// Insert user into database
	err = tx.QueryRow(ctx, `
		INSERT INTO users (
			first_name, last_name, initials, email, 
			password, facility_id, role
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`,
		user.FirstName, user.LastName, user.Initials, user.Email,
		string(hashedPassword), user.FacilityID, user.Role,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to insert user")
		return err
	}

	// Commit the transaction
	if err = tx.Commit(ctx); err != nil {
		logger.Error().Err(err).Msg("Failed to commit transaction")
		return err
	}

	logger.Info().
		Int("user_id", user.ID).
		Str("email", user.Email).
		Str("role", user.Role).
		Msg("User created successfully")

	return nil
}

func generateInitials(firstName, lastName string) string {
	firstInitial := ""
	if len(firstName) > 0 {
		firstInitial = strings.ToUpper(firstName[:1])
	}

	lastInitial := ""
	if len(lastName) > 0 {
		lastInitial = strings.ToUpper(lastName[:1])
	}

	return firstInitial + lastInitial
}

// Add this middleware for debugging session issues
func DebugSession() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			logger := log.With().
				Str("middleware", "debug_session").
				Str("path", c.Path()).
				Str("method", c.Request().Method).
				Logger()

			sess, ok := c.Get("session").(*sessions.Session)
			if !ok {
				logger.Warn().Msg("No session in context")
				return next(c)
			}

			logger.Debug().
				Interface("session_values", sess.Values).
				Str("session_name", sess.Name()).
				Bool("is_new", sess.IsNew).
				Interface("session_options", sess.Options).
				Msg("Current session state")

			return next(c)
		}
	}
}
