package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/DukeRupert/haven/db"
	"github.com/DukeRupert/haven/types"
	"github.com/DukeRupert/haven/view/alert"
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
	database *db.DB
	store    sessions.Store // Add store to the handler
	logger   zerolog.Logger
}

const (
	DefaultSessionName = "session"
)

func NewAuthHandler(pool *db.DB, store sessions.Store, logger zerolog.Logger) *AuthHandler {
	return &AuthHandler{
		database: pool,
		store:    store,
		logger:   logger.With().Str("component", "auth").Logger(),
	}
}

// LoginParams represents the expected login request body
type LoginParams struct {
	Email    string `json:"email" form:"email" validate:"required,email"`
	Password string `json:"password" form:"password" validate:"required"`
}

// LoginResponse handles both the alert and potential redirect for login attempts
func (h *AuthHandler) LoginResponse(c echo.Context, status int, heading string, messages []string, redirectURL string) error {
    c.Response().Status = status
    
    // For successful login with redirect
    if status == http.StatusOK && redirectURL != "" {
        c.Response().Header().Set("HX-Redirect", redirectURL)
        return c.String(http.StatusOK, "")
    }

    // For errors, return the alert component
    return alert.Error(heading, messages).Render(c.Request().Context(), c.Response().Writer)
}

// LoginHandler handles the login form submission
func (h *AuthHandler) LoginHandler() echo.HandlerFunc {
    return func(c echo.Context) error {
        logger := h.logger.With().
            Str("component", "auth").
            Str("handler", "LoginHandler").
            Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
            Logger()

        logger.Debug().Msg("Processing login request")

        // Get session
        sess, err := session.Get(DefaultSessionName, c)
        if err != nil {
            logger.Error().
                Err(err).
                Str("session_name", DefaultSessionName).
                Msg("Failed to get session")
            return h.LoginResponse(c, 
                http.StatusInternalServerError,
                "System Error",
                []string{"Unable to process login request. Please try again."},
                "")
        }

        // Bind form data
        params := new(LoginParams)
        if err := c.Bind(params); err != nil {
            logger.Debug().
                Err(err).
                Str("email", params.Email).
                Msg("Invalid form data submitted")
            return h.LoginResponse(c,
                http.StatusUnauthorized,
                "Invalid Request",
                []string{"Please check your email and password and try again."},
                "")
        }

        logger.Debug().Str("email", params.Email).Msg("Attempting user authentication")

        // Authenticate user
        user, err := authenticateUser(c.Request().Context(), h.database, params.Email, params.Password)
        if err != nil {
            logger.Debug().
                Err(err).
                Str("email", params.Email).
                Msg("Authentication failed")
            return h.LoginResponse(c,
                http.StatusUnauthorized,
                "Login Failed",
                []string{"Invalid email or password. Please try again."},
                "")
        }

        // Get user's facility
        facility, err := h.database.GetFacilityByID(c.Request().Context(), user.FacilityID)
        if err != nil {
            logger.Error().
                Err(err).
                Int("user_id", user.ID).
                Int("facility_id", user.FacilityID).
                Msg("Failed to get user's facility")
            return h.LoginResponse(c,
                http.StatusInternalServerError,
                "System Error",
                []string{"Unable to complete login process. Please try again."},
                "")
        }

        // Set session values
        logger.Debug().
            Int("user_id", user.ID).
            Str("role", string(user.Role)).
            Msg("Setting session values")

        sess.Values["user_id"] = user.ID
        sess.Values["user_role"] = user.Role
        sess.Values["user_initials"] = user.Initials
        sess.Values["facility_code"] = facility.Code
        sess.Values["facility_id"] = facility.ID
        sess.Values["last_access"] = time.Now()

        // Save session
        if err := sess.Save(c.Request(), c.Response()); err != nil {
            logger.Error().
                Err(err).
                Int("user_id", user.ID).
                Str("session_id", sess.ID).
                Msg("Failed to save session")
            return h.LoginResponse(c,
                http.StatusInternalServerError,
                "System Error",
                []string{"Unable to complete login process. Please try again."},
                "")
        }

        redirectURL := fmt.Sprintf("/%s/calendar", facility.Code)
        logger.Info().
            Int("user_id", user.ID).
            Str("email", params.Email).
            Str("facility_code", facility.Code).
            Str("redirect_url", redirectURL).
            Msg("Login successful")

        // On success, redirect to the calendar
        return h.LoginResponse(c, http.StatusOK, "", nil, redirectURL)
    }
}

var ErrInvalidCredentials = errors.New("invalid credentials")

// authenticateUser verifies the email and password combination
func authenticateUser(ctx context.Context, database *db.DB, email, password string) (*types.User, error) {
	log := zerolog.Ctx(ctx).With().Str("method", "authenticateUser").Logger()

	var user types.User
	err := database.QueryRow(ctx,
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

        // Get the session
        sess, err := session.Get(DefaultSessionName, c)
        if err != nil {
            logger.Error().Err(err).Msg("failed to get session")
        }

        // Log initial state
        logger.Debug().
            Bool("is_new", sess.IsNew).
            Str("session_id", sess.ID).
            Interface("values", sess.Values).
            Msg("starting logout process")

        // Clear the session regardless of whether it's new or not
        sess.Values = make(map[interface{}]interface{})
        sess.Options.MaxAge = -1

        // Force save the cleared session
        if err = sess.Save(c.Request(), c.Response()); err != nil {
            logger.Error().Err(err).Msg("failed to save cleared session")
            // Continue with redirect anyway
        }

        // Add Cache-Control header to prevent browser caching
        c.Response().Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0, pre-check=0")
        
        logger.Debug().Msg("logout completed, redirecting to login page")
        
        // Use StatusSeeOther (303) to ensure GET request
        return c.Redirect(http.StatusSeeOther, "/login")
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

type PublicRoute struct {
	Path   string
	Method string
}

// AuthMiddleware ensures requests are authenticated and maintains session state.
// It performs the following:
// - Validates session exists and contains valid user_id
// - Fetches fresh user data from database
// - Fetches associated facility data if applicable
// - Updates session with current user/facility information
// - Sets user/facility data in request context for handlers
// - Redirects unauthenticated requests to login
//
// Context values set:
// - user_id: int
// - user_role: types.UserRole
// - user_initials: string
// - facility_id: int (if user has facility)
// - facility_code: string (if user has facility)
func (h *AuthHandler) AuthMiddleware() echo.MiddlewareFunc {
	// Define public paths that should skip authentication
	publicRoutes := map[PublicRoute]bool{
		// Core auth routes
		{Path: "/login", Method: "GET"}:   true,
		{Path: "/login", Method: "POST"}:  true,
		{Path: "/logout", Method: "POST"}: true,

		// Public pages
		{Path: "/", Method: "GET"}: true,

		// Add more public routes here as needed, for example:
		// {Path: "/about", Method: "GET"}: true,
		// {Path: "/contact", Method: "GET"}: true,
		// {Path: "/health", Method: "GET"}: true,
		// {Path: "/api/public/status", Method: "GET"}: true,
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			logger := h.logger.With().
				Str("middleware", "AuthMiddleware()").
				Str("path", c.Path()).
				Logger()

			// Allow static assets using the RequestURI instead of Path
			if strings.HasPrefix(c.Request().RequestURI, "/static/") {
				logger.Debug().
					Str("uri", c.Request().RequestURI).
					Msg("allowing static asset access")
				return next(c)
			}

			// Check if this is a public route
			route := PublicRoute{
				Path:   c.Path(),
				Method: c.Request().Method,
			}

			if publicRoutes[route] {
				logger.Debug().
					Str("path", route.Path).
					Str("method", route.Method).
					Msg("allowing public route access")
				return next(c)
			}

			sess, err := session.Get(DefaultSessionName, c)
			if err != nil {
				logger.Error().Err(err).Msg("session error")
				// Clear any partial/corrupted session
				sess.Values = make(map[interface{}]interface{})
				sess.Options.MaxAge = -1
				sess.Save(c.Request(), c.Response())
				return redirectToLogin(c)
			}

			// Validate user authentication
			userID, ok := sess.Values["user_id"].(int)
			if !ok || userID == 0 {
				logger.Debug().Msg("no valid user_id in session, silly")
				return redirectToLogin(c)
			}

			// Fetch fresh user data
			user, err := h.database.GetUserByID(c.Request().Context(), userID)
			if err != nil {
				logger.Error().Err(err).Int("user_id", userID).Msg("failed to fetch user data")
				return redirectToLogin(c)
			}

			// Fetch facility data if user has facility association
			var facility *types.Facility
			if user.FacilityID != 0 {
				facility, err = h.database.GetFacilityByID(c.Request().Context(), user.FacilityID)
				if err != nil && !errors.Is(err, pgx.ErrNoRows) {
					logger.Error().Err(err).Int("facility_id", user.FacilityID).Msg("failed to fetch facility")
					return echo.NewHTTPError(http.StatusInternalServerError, "database error")
				}
			}

			// Update session with fresh data
			sess.Values["user_id"] = user.ID
			sess.Values["role"] = user.Role
			sess.Values["initials"] = user.Initials
			sess.Values["last_access"] = time.Now()

			if facility != nil {
				sess.Values["facility_id"] = facility.ID
				sess.Values["facility_code"] = facility.Code
			}

			if err := sess.Save(c.Request(), c.Response()); err != nil {
				logger.Error().Err(err).Msg("failed to save session")
				return echo.NewHTTPError(http.StatusInternalServerError, "session error")
			}

			// Set context values for handlers
			c.Set("user_id", user.ID)
			c.Set("user_role", user.Role)
			c.Set("user_initials", user.Initials)
			if facility != nil {
				c.Set("facility_id", facility.ID)
				c.Set("facility_code", facility.Code)
			}

			return next(c)
		}
	}
}

// If you need to check public routes elsewhere in your code,
// you might want to add this helper method
func (h *AuthHandler) IsPublicRoute(path, method string) bool {
	publicRoutes := map[PublicRoute]bool{
		{Path: "/login", Method: "GET"}:   true,
		{Path: "/login", Method: "POST"}:  true,
		{Path: "/logout", Method: "POST"}: true,
		{Path: "/", Method: "GET"}:        true,
		// Add the same routes as above
	}

	return publicRoutes[PublicRoute{Path: path, Method: method}]
}

// RoleAuthMiddleware protects routes by checking if the user has sufficient role permissions.
// It verifies session auth, validates user_id and role, and enforces minimum role requirements.
// Returns 403 Forbidden for insufficient permissions or redirects to login for auth failures.
// Usage: e.GET("/admin", handler, h.RoleAuthMiddleware("super"))
func (h *AuthHandler) RoleAuthMiddleware(minimumRole types.UserRole) echo.MiddlewareFunc {
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
			role, ok := sess.Values["role"].(types.UserRole)
			if !ok {
				logger.Debug().Str("role", string(role)).Int("user_id", userID).Msg("no valid role in session")
				return redirectToLogin(c)
			}

			// Check if user's role meets minimum required role
			if !IsAtLeastRole(string(role), string(minimumRole)) {
				logger.Debug().
					Str("user_role", string(role)).
					Msg("insufficient role permissions")
				return echo.NewHTTPError(http.StatusForbidden, "Insufficient permissions")
			}
			return next(c)
		}
	}
}

func (h *AuthHandler) WithRouteContext() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get values set by AuthMiddleware
			userRole, _ := c.Get("user_role").(types.UserRole)
			userInitials, _ := c.Get("user_initials").(string)
			facilityID, _ := c.Get("facility_id").(int)
			facilityCode, _ := c.Get("facility_code").(string)

			// Get path parameters if they exist
			pathFacility := c.Param("facility")
			pathInitials := c.Param("initials")

			// Determine base path based on role and facility
			var basePath string
			if facilityCode != "" {
				basePath = facilityCode
			}

			// Create route context
			routeCtx := &types.RouteContext{
				BasePath:     basePath,
				UserRole:     userRole,
				UserInitials: userInitials,
				FacilityID:   facilityID,
				FacilityCode: facilityCode,
				PathFacility: pathFacility,
				PathInitials: pathInitials,
			}

			// Store in context
			c.Set("routeCtx", routeCtx)

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

			// User is authenticated and has facility, redirect to facility dashboard
			logger.Debug().
				Int("user_id", userID).
				Str("facility_code", facilityCode).
				Msg("redirecting authenticated user from login page to facility calendar")

			// Changed to match the path pattern seen in logs
			redirectURL := fmt.Sprintf("/%s/calendar", facilityCode)
			return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
		}
	}
}

// You might also want to add this helper middleware for public routes
func (h *AuthHandler) PublicRouteMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Clear any existing error state from the context
			c.Set("error", nil)

			// For public routes, always allow GET requests
			if c.Request().Method == "GET" {
				return next(c)
			}

			// For POST requests to public routes, handle normally
			return next(c)
		}
	}
}

// IsAtLeastRole checks if the current role meets or exceeds the minimum required role
func IsAtLeastRole(currentRole string, minimumRole string) bool {
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
