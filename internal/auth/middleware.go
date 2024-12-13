// internal/auth/middleware.go
package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/DukeRupert/haven/internal/model/dto"
	"github.com/DukeRupert/haven/internal/model/entity"
	"github.com/DukeRupert/haven/internal/model/types"
	facilityRepo "github.com/DukeRupert/haven/internal/repository/facility"
	"github.com/DukeRupert/haven/internal/response"
	"github.com/DukeRupert/haven/internal/store"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

type RouteType int

const (
	RoutePublic RouteType = iota
	RouteAuthenticated
)

// AuthConfig holds the configuration for auth middleware
type AuthConfig struct {
	RouteType       RouteType
	RedirectToLogin bool
}

// AuthContext combines session data and optional full user data
type AuthContext struct {
	// Quick access session data
	UserID       int            `json:"user_id"`
	Role         types.UserRole `json:"role"`
	Initials     string         `json:"initials,omitempty"`
	FacilityID   int            `json:"facility_id,omitempty"`
	FacilityCode string         `json:"facility_code,omitempty"`

	// Lazy loaded data
	user     *entity.User
	facility *entity.Facility
	loader   func() (*entity.User, *entity.Facility, error)
}

// PublicRoute represents a route that can be accessed without authentication
type PublicRoute struct {
	Path   string
	Method string
}

// Middleware structure
type Middleware struct {
	service *Service
	logger  zerolog.Logger
	public  map[PublicRoute]bool
}

// NewMiddleware creates a new auth middleware instance
func NewMiddleware(service *Service, logger zerolog.Logger) *Middleware {
	m := &Middleware{
		service: service,
		logger:  logger,
		public:  make(map[PublicRoute]bool),
	}

	// Register default public routes
	m.registerPublicRoutes()
	return m
}

// RegisterPublicRoute adds a route to the public routes map
func (m *Middleware) RegisterPublicRoute(path, method string) {
	route := PublicRoute{
		Path:   path,
		Method: method,
	}
	m.public[route] = true
}

// registerPublicRoutes sets up default public routes
func (m *Middleware) registerPublicRoutes() {
	// Auth routes
	m.RegisterPublicRoute("/login", http.MethodGet)
	m.RegisterPublicRoute("/login", http.MethodPost)
	m.RegisterPublicRoute("/logout", http.MethodPost)
	m.RegisterPublicRoute("/register", http.MethodGet)
	m.RegisterPublicRoute("/register", http.MethodPost)
	m.RegisterPublicRoute("/set-password", http.MethodGet)
	m.RegisterPublicRoute("/set-password", http.MethodPost)

	// Add any other public routes your application needs
}

// Auth middleware that sets up the context
func (m *Middleware) Auth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			logger := m.logger.With().
				Str("path", c.Path()).
				Str("method", c.Request().Method).
				Logger()

			// Skip auth for static assets
			if strings.HasPrefix(c.Request().RequestURI, "/static/") {
				logger.Debug().Msg("allowing static asset access")
				return next(c)
			}

			// Check if it's a public route
			route := PublicRoute{
				Path:   c.Path(),
				Method: c.Request().Method,
			}
			if m.public[route] {
				logger.Debug().Msg("allowing public route access")
				return next(c)
			}

			// Get session
			sess, err := session.Get("session", c)
			if err != nil {
				logger.Error().Err(err).Msg("session error")
				return response.System(c)
			}

			// Create auth context from session
			auth := &AuthContext{}

			// Required values
			userID, ok := sess.Values["user_id"].(int)
			if !ok || userID == 0 {
				logger.Debug().Msg("no valid user_id in session")
				return redirectToLogin(c)
			}
			auth.UserID = userID

			role, ok := sess.Values["role"].(types.UserRole)
			if !ok {
				logger.Debug().Msg("no valid role in session")
				return redirectToLogin(c)
			}
			auth.Role = role

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

			// Set up lazy loader
			auth.loader = func() (*entity.User, *entity.Facility, error) {
				user, err := m.service.repos.User.GetByID(c.Request().Context(), userID)
				if err != nil {
					return nil, nil, err
				}

				var facility *entity.Facility
				if user.FacilityID != 0 {
					facility, err = m.service.repos.Facility.GetByID(c.Request().Context(), user.FacilityID)
					if err != nil && !errors.Is(err, facilityRepo.ErrNotFound) {
						return nil, nil, err
					}
				}

				return user, facility, nil
			}

			// Store in context
			c.Set("auth", auth)

			return next(c)
		}
	}
}

// GetUser returns the full user object, loading it if necessary
func (ac *AuthContext) GetUser() (*entity.User, error) {
	if ac.user != nil {
		return ac.user, nil
	}
	if ac.loader == nil {
		return nil, fmt.Errorf("no user loader available")
	}

	user, facility, err := ac.loader()
	if err != nil {
		return nil, err
	}

	ac.user = user
	ac.facility = facility
	return user, nil
}

// GetFacility returns the facility object, loading it if necessary
func (ac *AuthContext) GetFacility() (*entity.Facility, error) {
	if ac.facility != nil {
		return ac.facility, nil
	}
	if ac.loader == nil {
		return nil, fmt.Errorf("no facility loader available")
	}

	user, facility, err := ac.loader()
	if err != nil {
		return nil, err
	}

	ac.user = user
	ac.facility = facility
	return facility, nil
}

// Helper to get auth context
func GetAuthContext(c echo.Context) (*AuthContext, error) {
	auth, ok := c.Get("auth").(*AuthContext)
	if !ok {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "no auth context found")
	}
	return auth, nil
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
				if err != nil && !errors.Is(err, facilityRepo.ErrNotFound) {
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
