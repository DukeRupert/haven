package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/DukeRupert/haven/internal/model/dto"
	"github.com/DukeRupert/haven/internal/model/entity"
	"github.com/DukeRupert/haven/internal/model/types"
	"github.com/DukeRupert/haven/internal/repository"
	"github.com/DukeRupert/haven/internal/repository/facility"
	"github.com/DukeRupert/haven/internal/store"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
)

// Common errors
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// SessionKeys defines constants for session value keys
const (
	SessionKeyUserID       = "user_id"
	SessionKeyRole         = "role"
	SessionKeyInitials     = "initials"
	SessionKeyFacilityID   = "facility_id"
	SessionKeyFacilityCode = "facility_code"
	SessionKeyLastAccess   = "last_access"
)

// CtxKey define context value keys
const (
	CtxKeyAuth  = "auth"
	CtxKeyRoute = "routeCtx"
)

// RoleLevel represents the hierarchy level of a role
type RoleLevel int

const (
	UserLevel  RoleLevel = 1
	AdminLevel RoleLevel = 2
	SuperLevel RoleLevel = 3
)

// Middleware holds all middleware configuration
type Middleware struct {
	repos  *repository.Repositories
	logger zerolog.Logger
}

// Config holds middleware configuration
type Config struct {
	Repos  *repository.Repositories
	Logger zerolog.Logger
}

// NewMiddleware creates a new middleware instance
func NewMiddleware(cfg Config) *Middleware {
	return &Middleware{
		repos:  cfg.Repos,
		logger: cfg.Logger.With().Str("component", "middleware").Logger(),
	}
}

// Auth middleware ensures authentication context is available
func (m *Middleware) Auth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			logger := m.logger.With().
				Str("path", c.Path()).
				Str("method", c.Request().Method).
				Logger()

			// Get and validate session
			sess, err := m.getAndValidateSession(c, logger)
			if err != nil {
				return redirectToLogin(c)
			}

			// Get user and facility data
			user, facility, err := m.getUserAndFacility(c.Request().Context(), sess, logger)
			if err != nil {
				return err
			}

			// Update session with fresh data
			if err := m.updateSession(c, sess, user, facility); err != nil {
				return err
			}

			// Create the data provider
			provider := &authDataProvider{
				repos:      m.repos,
				user:       user,
				facility:   facility,
				userID:     user.ID,
				facilityID: user.FacilityID,
			}

			// Create and set auth context
			authContext := &dto.AuthContext{
				AuthContextData: dto.AuthContextData{
					UserID:       user.ID,
					Role:         user.Role,
					Initials:     user.Initials,
					FacilityID:   user.FacilityID,
					FacilityCode: facility.Code,
				},
				Provider: provider,
			}
			c.Set("auth", authContext)

			logger.Debug().
				Int("user_id", user.ID).
				Str("role", string(user.Role)).
				Str("facility_code", facility.Code).
				Msg("authentication successful")

			return next(c)
		}
	}
}

// RouteContext middleware ensures route context is available
func (m *Middleware) RouteContext() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get the full path
			fullPath := c.Request().URL.Path

			// Create route context
			routeCtx := &dto.RouteContext{
				CurrentPath: getRoutePattern(c),
				FullPath:    fullPath,
			}

			// Check if we're in an /app route
			if strings.HasPrefix(fullPath, "/app") {
				// Extract facility code and user initials from path params
				facilityCode := c.Param("facility_code")
				userInitials := c.Param("user_initials")

				// Store the values even if empty
				routeCtx.FacilityCode = facilityCode
				routeCtx.UserInitials = userInitials

				// Set base path if we have a facility code
				if facilityCode != "" {
					routeCtx.BasePath = fmt.Sprintf("/app/%s", facilityCode)

					// If we also have user initials, append them to base path
					if userInitials != "" {
						routeCtx.BasePath = fmt.Sprintf("/app/%s/%s", facilityCode, userInitials)
					}
				}

				// Log the extracted parameters for debugging
				m.logger.Debug().
					Str("path", fullPath).
					Str("facility_code", facilityCode).
					Str("user_initials", userInitials).
					Msg("Route context parameters extracted")
			}

			// Store the context
			c.Set("routeCtx", routeCtx)
			return next(c)
		}
	}
}

// RequireRole creates middleware that checks if user has sufficient role access
func (m *Middleware) RequireRole(minimumRole types.UserRole) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            logger := m.logger.With().
                Str("path", c.Path()).
                Str("required_role", string(minimumRole)).
                Logger()

            // Get auth context
            auth, ok := c.Get("auth").(*dto.AuthContext)
            if !ok {
                logger.Error().Msg("No auth context found")
                return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
            }

            // Check if user has sufficient role
            if !HasMinimumRole(auth.Role, minimumRole) {
                logger.Warn().
                    Str("user_role", string(auth.Role)).
                    Str("required_role", string(minimumRole)).
                    Msg("Insufficient role")
                return echo.NewHTTPError(http.StatusForbidden, "insufficient permissions")
            }

            logger.Debug().
                Str("user_role", string(auth.Role)).
                Msg("Role check passed")

            return next(c)
        }
    }
}

// RequireFacilityAccess creates middleware that checks if user has access to the facility
func (m *Middleware) RequireFacilityAccess() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            logger := m.logger.With().
                Str("path", c.Path()).
                Logger()

            // Get auth context
            auth, ok := c.Get("auth").(*dto.AuthContext)
            if !ok {
                logger.Error().Msg("No auth context found")
                return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
            }

            // Get route parameter
            facilityCode := c.Param("facility_code")
            if facilityCode == "" {
                logger.Error().Msg("No facility code in route")
                return echo.NewHTTPError(http.StatusBadRequest, "facility code required")
            }

            // Super users have access to all facilities
            if auth.Role == types.UserRoleSuper {
                logger.Debug().
                    Str("user_role", string(auth.Role)).
                    Str("facility_code", facilityCode).
                    Msg("Super user granted facility access")
                return next(c)
            }

            // For admin and user roles, check if facility code matches
            if auth.FacilityCode != facilityCode {
                logger.Warn().
                    Str("user_role", string(auth.Role)).
                    Str("user_facility", auth.FacilityCode).
                    Str("requested_facility", facilityCode).
                    Msg("Facility access denied")
                return echo.NewHTTPError(http.StatusForbidden, "insufficient facility permissions")
            }

            logger.Debug().
                Str("user_role", string(auth.Role)).
                Str("facility_code", facilityCode).
                Msg("Facility access granted")

            return next(c)
        }
    }
}

// RequireProfileAccess checks if user has permission to view the requested profile
func (m *Middleware) RequireProfileAccess() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            logger := m.logger.With().
                Str("path", c.Path()).
                Logger()

            // Get auth and route contexts
            auth, err := GetAuthContext(c)
            if err != nil {
                logger.Error().Msg("No auth context found")
                return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
            }

            // Get requested user initials from route parameter
            requestedInitials := c.Param("user_initials")
            if requestedInitials == "" {
                logger.Error().Msg("No user initials in route")
                return echo.NewHTTPError(http.StatusBadRequest, "user initials required")
            }

            // Super and Admin users have access to all profiles within their facility scope
            // (facility scope is already checked by previous middleware)
            if auth.Role == types.UserRoleSuper || auth.Role == types.UserRoleAdmin {
                logger.Debug().
                    Str("user_role", string(auth.Role)).
                    Str("requested_initials", requestedInitials).
                    Msg("Access granted based on role")
                return next(c)
            }

            // Regular users can only view their own profile
            if auth.Initials != requestedInitials {
                logger.Warn().
                    Str("user_initials", auth.Initials).
                    Str("requested_initials", requestedInitials).
                    Msg("Profile access denied")
                return echo.NewHTTPError(http.StatusForbidden, "insufficient permissions")
            }

            logger.Debug().
                Str("user_initials", auth.Initials).
                Msg("Profile access granted")

            return next(c)
        }
    }
}

// Helper functions
type authDataProvider struct {
	repos      *repository.Repositories
	user       *entity.User
	facility   *entity.Facility
	userID     int
	facilityID int
}

func (p *authDataProvider) GetUser() (*entity.User, error) {
	if p.user != nil {
		return p.user, nil
	}
	user, err := p.repos.User.GetByID(context.Background(), p.userID)
	if err != nil {
		return nil, err
	}
	p.user = user
	return user, nil
}

func (p *authDataProvider) GetFacility() (*entity.Facility, error) {
	if p.facility != nil {
		return p.facility, nil
	}
	facility, err := p.repos.Facility.GetByID(context.Background(), p.facilityID)
	if err != nil {
		return nil, err
	}
	p.facility = facility
	return facility, nil
}

// HasMinimumRole checks if a role meets or exceeds the minimum required role
func HasMinimumRole(current, minimum types.UserRole) bool {
	roleValues := map[types.UserRole]RoleLevel{
		types.UserRoleUser:  UserLevel,
		types.UserRoleAdmin: AdminLevel,
		types.UserRoleSuper: SuperLevel,
	}

	currentLevel, ok := roleValues[current]
	if !ok {
		return false
	}

	requiredLevel, ok := roleValues[minimum]
	if !ok {
		return false
	}

	return currentLevel >= requiredLevel
}

// Authenticate verifies user credentials
func (m *Middleware) Authenticate(ctx context.Context, email, password string) (*entity.User, error) {
	log := m.logger.With().Str("method", "Authenticate").Logger()

	user, err := m.repos.User.GetByEmail(ctx, email)
	if err != nil {
		log.Debug().Str("email", email).Err(err).Msg("user lookup failed")
		return nil, ErrInvalidCredentials
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		log.Debug().Str("email", email).Msg("invalid password")
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

// getRoutePattern returns the route pattern with parameter placeholders
func getRoutePattern(c echo.Context) string {
	route := c.Echo().Reverse(c.Path())
	if route == "" {
		return c.Path() // Fallback to actual path if reverse lookup fails
	}
	return route
}

// Helper functions to get contexts
func GetAuthContext(c echo.Context) (*dto.AuthContext, error) {
	auth, ok := c.Get(CtxKeyAuth).(*dto.AuthContext)
	if !ok {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "no auth context found")
	}
	return auth, nil
}

func GetRouteContext(c echo.Context) (*dto.RouteContext, error) {
	routeCtx, ok := c.Get(CtxKeyRoute).(*dto.RouteContext)
	if !ok {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "no route context found")
	}
	return routeCtx, nil
}

// Helper methods to break down the middleware logic
func (m *Middleware) getAndValidateSession(c echo.Context, logger zerolog.Logger) (*sessions.Session, error) {
	sess, err := session.Get(store.DefaultSessionName, c)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get session")
		return nil, err
	}

	userID, ok := sess.Values[SessionKeyUserID].(int)
	if !ok || userID == 0 {
		logger.Debug().Msg("Invalid or missing user_id in session")
		return nil, fmt.Errorf("invalid user_id")
	}

	role, ok := sess.Values[SessionKeyRole].(types.UserRole)
	if !ok {
		logger.Debug().Msg("Invalid or missing role in session")
		return nil, fmt.Errorf("invalid role")
	}

	logger.Debug().
		Int("user_id", userID).
		Str("role", string(role)).
		Msg("Session validated")

	return sess, nil
}

func (m *Middleware) getUserAndFacility(ctx context.Context, sess *sessions.Session, logger zerolog.Logger) (*entity.User, *entity.Facility, error) {
	userID := sess.Values[SessionKeyUserID].(int)

	user, err := m.repos.User.GetByID(ctx, userID)
	if err != nil {
		logger.Error().Err(err).Int("user_id", userID).Msg("failed to fetch user")
		return nil, nil, err
	}

	var f *entity.Facility
	if user.FacilityID != 0 {
		f, err = m.repos.Facility.GetByID(ctx, user.FacilityID)
		if err != nil && !errors.Is(err, facility.ErrNotFound) {
			logger.Error().Err(err).Msg("failed to fetch facility")
			return nil, nil, echo.NewHTTPError(http.StatusInternalServerError, "database error")
		}
	}

	return user, f, nil
}

func (m *Middleware) updateSession(c echo.Context, sess *sessions.Session, user *entity.User, facility *entity.Facility) error {
	sess.Values[SessionKeyUserID] = user.ID
	sess.Values[SessionKeyRole] = user.Role
	sess.Values[SessionKeyInitials] = user.Initials
	if facility != nil {
		sess.Values[SessionKeyFacilityID] = facility.ID
		sess.Values[SessionKeyFacilityCode] = facility.Code
	}
	sess.Values[SessionKeyLastAccess] = time.Now()

	return sess.Save(c.Request(), c.Response())
}

func redirectToLogin(c echo.Context) error {
	// For regular requests, redirect to login
	return c.Redirect(http.StatusTemporaryRedirect, "/login")
}
