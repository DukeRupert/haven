package context

import (
	"fmt"

	"github.com/DukeRupert/haven/internal/model/dto"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

// RouteContextMiddleware holds the middleware configuration
type RouteContextMiddleware struct {
	logger zerolog.Logger
}

// NewRouteContextMiddleware creates a new middleware instance
func NewRouteContextMiddleware(logger zerolog.Logger) *RouteContextMiddleware {
	return &RouteContextMiddleware{
		logger: logger,
	}
}

// WithRouteContext middleware ensures route context is available
func (m *RouteContextMiddleware) WithRouteContext() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Create route context with only path information
			routeCtx := &dto.RouteContext{
				CurrentPath: getRoutePattern(c),
				FullPath:    c.Request().URL.Path,
			}

			// Set base path if we're in a facility context
			if facilityCode := c.Param("facility_id"); facilityCode != "" {
				routeCtx.BasePath = fmt.Sprintf("/facility/%s", facilityCode)
			}

			c.Set("routeCtx", routeCtx)
			return next(c)
		}
	}
}

// getRoutePattern returns the route pattern with parameter placeholders
func getRoutePattern(c echo.Context) string {
	route := c.Echo().Reverse(c.Path())
	if route == "" {
		return c.Path() // Fallback to actual path if reverse lookup fails
	}
	return route
}
