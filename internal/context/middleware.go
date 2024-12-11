// internal/context/middleware.go
package context

import (
    "github.com/DukeRupert/haven/internal/model/types"
	"github.com/DukeRupert/haven/internal/model/dto"
    "github.com/labstack/echo/v4"
    "github.com/rs/zerolog"
)

// RouteContextMiddleware handles setting up template context
type RouteContextMiddleware struct {
    logger zerolog.Logger
}

// New creates a new route context middleware
func NewRouteContextMiddleware(logger zerolog.Logger) *RouteContextMiddleware {
    return &RouteContextMiddleware{
        logger: logger.With().Str("component", "route_context_middleware").Logger(),
    }
}

// WithRouteContext adds template rendering context to requests
func (m *RouteContextMiddleware) WithRouteContext() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            logger := m.logger.With().
                Str("path", c.Path()).
                Str("method", c.Request().Method).
                Logger()

            // Extract user context
            routeCtx, err := m.buildRouteContext(c)
            if err != nil {
                logger.Error().Err(err).Msg("failed to build route context")
                return next(c)
            }

            // Log context creation
            logger.Debug().
                Str("base_path", routeCtx.BasePath).
                Str("user_role", string(routeCtx.UserRole)).
                Str("facility_code", routeCtx.FacilityCode).
                Msg("route context created")

            // Store in context
            c.Set("routeCtx", routeCtx)

            return next(c)
        }
    }
}

// buildRouteContext constructs the template context from request data
func (m *RouteContextMiddleware) buildRouteContext(c echo.Context) (*dto.RouteContext, error) {
    // Get auth context values
    userRole, _ := c.Get("user_role").(types.UserRole)
    userInitials, _ := c.Get("user_initials").(string)
    facilityID, _ := c.Get("facility_id").(int)
    facilityCode, _ := c.Get("facility_code").(string)

    // Get path parameters
    pathFacility := c.Param("facility")
    pathInitials := c.Param("initials")

    // Determine base path
    basePath := determineBasePath(facilityCode)

    return &dto.RouteContext{
        BasePath:     basePath,
        UserRole:     userRole,
        UserInitials: userInitials,
        FacilityID:   facilityID,
        FacilityCode: facilityCode,
        PathFacility: pathFacility,
        PathInitials: pathInitials,
    }, nil
}

// determineBasePath generates the base path for templates
func determineBasePath(facilityCode string) string {
    if facilityCode != "" {
        return facilityCode
    }
    return ""
}