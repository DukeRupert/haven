// internal/auth/middleware.go
package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/DukeRupert/haven/internal/response"
	"github.com/DukeRupert/haven/internal/store"
	"github.com/DukeRupert/haven/internal/model/dto"
	"github.com/DukeRupert/haven/internal/model/entity"
	"github.com/DukeRupert/haven/internal/model/types"
	"github.com/DukeRupert/haven/internal/repository/facility"
	
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

// PublicRoute represents a route that doesn't require authentication
type PublicRoute struct {
	Path   string
	Method string
}

// Middleware handles authentication and session management
type Middleware struct {
	service *Service
	logger  zerolog.Logger
	public  map[PublicRoute]bool
}

// NewMiddleware creates a new auth middleware instance
func NewMiddleware(service *Service, logger zerolog.Logger) *Middleware {
	return &Middleware{
		service: service,
		logger:  logger.With().Str("component", "auth_middleware").Logger(),
		public: map[PublicRoute]bool{
			{Path: "/login", Method: "GET"}:         true,
			{Path: "/login", Method: "POST"}:        true,
			{Path: "/logout", Method: "POST"}:       true,
			{Path: "/register", Method: "GET"}:      true,
			{Path: "/register", Method: "POST"}:     true,
			{Path: "/set-password", Method: "GET"}:  true,
			{Path: "/set-password", Method: "POST"}: true,
			{Path: "/", Method: "GET"}:              true,
		},
	}
}

// Authenticate middleware ensures requests are authenticated
func (m *Middleware) Authenticate() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			logger := m.logger.With().
				Str("path", c.Path()).
				Str("method", c.Request().Method).
				Logger()

			// Check static assets
			if strings.HasPrefix(c.Request().RequestURI, "/static/") {
				logger.Debug().Msg("allowing static asset access")
				return next(c)
			}

			// Check public routes
			route := PublicRoute{
				Path:   c.Path(),
				Method: c.Request().Method,
			}
			if m.public[route] {
				logger.Debug().Msg("allowing public route access")
				return next(c)
			}

			// Get session
			sess, err := session.Get(store.DefaultSessionName, c)
			if err != nil {
				logger.Error().Err(err).Msg("session error")
				return m.clearSessionAndRedirect(c, sess)
			}

			// Validate user ID
			userID, ok := sess.Values["user_id"].(int)
			if !ok || userID == 0 {
				logger.Debug().Msg("no valid user_id in session")
				return redirectToLogin(c)
			}

			// Get fresh user data using repository
			user, err := m.service.repos.User.GetByID(c.Request().Context(), userID)
			if err != nil {
				logger.Error().Err(err).Int("user_id", userID).Msg("failed to fetch user")
				return redirectToLogin(c)
			}

			// Get fac data if needed
			var fac *entity.Facility
			if user.FacilityID != 0 {
				fac, err = m.service.repos.Facility.GetByID(c.Request().Context(), user.FacilityID)
				if err != nil && !errors.Is(err, facility.ErrNotFound) {
					logger.Error().Err(err).Msg("failed to fetch facility")
					return echo.NewHTTPError(http.StatusInternalServerError, "database error")
				}
			}

			// Update session
			if err := m.updateSession(c, sess, user, fac); err != nil {
				logger.Error().Err(err).Msg("failed to update session")
				return echo.NewHTTPError(http.StatusInternalServerError, "session error")
			}

			// Set context values
			m.setContextValues(c, user, fac)

			return next(c)
		}
	}
}

// RequireRole ensures users have sufficient role permissions
func (m *Middleware) RequireRole(minimumRole types.UserRole) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			logger := m.logger.With().
				Str("middleware", "RequireRole").
				Str("path", c.Path()).
				Str("minimum_role", string(minimumRole)).
				Logger()

			logger.Debug().Msg("checking role requirements")

			// Get session
			sess, err := session.Get(store.DefaultSessionName, c)
			if err != nil {
				logger.Error().Err(err).Msg("no session found")
				return redirectToLogin(c)
			}

			// Validate user authentication
			userID, ok := sess.Values["user_id"].(int)
			if !ok || userID == 0 {
				logger.Debug().Msg("invalid user ID in session")
				return redirectToLogin(c)
			}

			// Check role
			role, ok := sess.Values["role"].(types.UserRole)
			if !ok {
				return redirectToLogin(c)
			}

			if !HasMinimumRole(role, minimumRole) {
				return echo.NewHTTPError(http.StatusForbidden, "Insufficient permissions")
			}

			return next(c)
		}
	}
}

// RedirectAuthenticated redirects logged-in users away from auth pages
func (m *Middleware) RedirectAuthenticated() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			logger := m.logger.With().
				Str("middleware", "RedirectAuthenticated").
				Str("path", c.Path()).
				Str("method", c.Request().Method).
				Logger()

			logger.Debug().Msg("checking authenticated status")

			// Check authentication status
			redirect, url, err := m.shouldRedirect(c)
			if err != nil {
				logger.Error().Err(err).Msg("error checking redirect status")
				return next(c)
			}

			if redirect {
				logger.Debug().
					Str("redirect_url", url).
					Msg("redirecting authenticated user")
				return c.Redirect(http.StatusTemporaryRedirect, url)
			}

			return next(c)
		}
	}
}

// EnsurePublic middleware handles public route access and error states
func (m *Middleware) EnsurePublic() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			logger := m.logger.With().
				Str("middleware", "EnsurePublic").
				Str("path", c.Path()).
				Str("method", c.Request().Method).
				Logger()

			// Clear any existing error state
			c.Set("error", nil)

			// Log public route access
			logger.Debug().Msg("accessing public route")

			// Handle rate limiting or other public route protections here
			if err := m.validatePublicAccess(c); err != nil {
				logger.Warn().
					Err(err).
					Msg("public access denied")
				return err
			}

			return next(c)
		}
	}
}

// ValidateFacility ensures users have appropriate facility access
func (m *Middleware) ValidateFacility() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			logger := m.logger.With().
                Str("middleware", "ValidateFacility").
                Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
                Logger()

            // Get auth context
            auth, err := m.GetAuthContext(c)
            if err != nil {
                logger.Error().Err(err).Msg("failed to get auth context")
                return response.Error(c,
                    http.StatusInternalServerError,
                    "Authentication Error",
                    []string{"Unable to verify permissions"},
                )
            }

			// Get and validate facility code
			facilityCode := c.Param("facility")
			if facilityCode == "" {
				logger.Error().Msg("missing facility code parameter")
				return response.Error(c,
					http.StatusBadRequest,
					"Invalid Request",
					[]string{"Facility code is required"},
				)
			}

			// Check facility access
			if canAccessFacility(auth, facilityCode) {
				logger.Debug().
					Str("facility_code", facilityCode).
					Str("user_role", string(auth.Role)).
					Msg("facility access granted")
				return next(c)
			}

			logger.Warn().
				Str("facility_code", facilityCode).
				Str("user_role", string(auth.Role)).
				Str("user_facility", auth.FacilityCode).
				Msg("unauthorized facility access attempt")

			return response.Error(c,
				http.StatusForbidden,
				"Access Denied",
				[]string{"You don't have permission to access this facility"},
			)
		}
	}
}

// GetContextFromSession retrieves auth context from session
func (m *Middleware) GetAuthContext(c echo.Context) (*dto.AuthContext, error) {
    sess, err := session.Get("session", c)
    if err != nil {
        return nil, fmt.Errorf("failed to get session: %w", err)
    }

    userID, ok := sess.Values["user_id"].(int)
    if !ok {
        return nil, fmt.Errorf("invalid user_id in session")
    }

    role, ok := sess.Values["role"].(types.UserRole)
    if !ok {
        return nil, fmt.Errorf("invalid role in session")
    }

    auth := &dto.AuthContext{
        UserID: userID,
        Role:   role,
    }

    // Optional values
    if initials, ok := sess.Values["initials"].(string); ok {
        auth.Initials = initials
    }
    if facilityID, ok := sess.Values["facility_id"].(int); ok {
        auth.FacilityID = facilityID
    }
    if facilityCode, ok := sess.Values["facility_code"].(string); ok {
        auth.FacilityCode = facilityCode
    }

    return auth, nil
}

// Helper methods
func canAccessFacility(auth *dto.AuthContext, facilityCode string) bool {
	switch auth.Role {
	case types.UserRoleSuper:
		return true
	case types.UserRoleAdmin:
		return auth.FacilityCode == facilityCode
	default:
		return false
	}
}

// validatePublicAccess handles any validation needed for public routes
func (m *Middleware) validatePublicAccess(c echo.Context) error {
	// Add any public route validations here, such as:
	// - Rate limiting
	// - Bot protection
	// - Request validation
	// - CORS checks
	// - etc.

	return nil
}

// hasMinimumRole checks if a role meets the minimum required level
func (m *Middleware) hasMinimumRole(userRole, minimumRole types.UserRole) bool {
	roleHierarchy := map[types.UserRole]int{
		types.UserRoleUser:  1,
		types.UserRoleAdmin: 2,
		types.UserRoleSuper: 3,
	}

	userLevel := roleHierarchy[userRole]
	requiredLevel := roleHierarchy[minimumRole]

	return userLevel >= requiredLevel
}

// shouldRedirect determines if user should be redirected and where
func (m *Middleware) shouldRedirect(c echo.Context) (bool, string, error) {
	sess, err := session.Get(store.DefaultSessionName, c)
	if err != nil {
		return false, "", fmt.Errorf("getting session: %w", err)
	}

	// Check user authentication
	userID, ok := sess.Values["user_id"].(int)
	if !ok || userID == 0 {
		return false, "", nil
	}

	// Check facility code
	facilityCode, ok := sess.Values["facility_code"].(string)
	if !ok || facilityCode == "" {
		return false, "", nil
	}

	// Build redirect URL
	redirectURL := fmt.Sprintf("/%s/calendar", facilityCode)
	return true, redirectURL, nil
}

func (m *Middleware) clearSessionAndRedirect(c echo.Context, sess *sessions.Session) error {
	sess.Values = make(map[interface{}]interface{})
	sess.Options.MaxAge = -1
	sess.Save(c.Request(), c.Response())
	return redirectToLogin(c)
}

func (m *Middleware) updateSession(c echo.Context, sess *sessions.Session, user *entity.User, facility *entity.Facility) error {
	sess.Values["user_id"] = user.ID
	sess.Values["role"] = user.Role
	sess.Values["initials"] = user.Initials
	sess.Values["last_access"] = time.Now()

	if facility != nil {
		sess.Values["facility_id"] = facility.ID
		sess.Values["facility_code"] = facility.Code
	}

	return sess.Save(c.Request(), c.Response())
}

func (m *Middleware) setContextValues(c echo.Context, user *entity.User, facility *entity.Facility) {
	c.Set("user_id", user.ID)
	c.Set("user_role", user.Role)
	c.Set("user_initials", user.Initials)
	if facility != nil {
		c.Set("facility_id", facility.ID)
		c.Set("facility_code", facility.Code)
	}
}

// IsPublicRoute checks if a route is public
func (m *Middleware) IsPublicRoute(path, method string) bool {
	return m.public[PublicRoute{Path: path, Method: method}]
}
