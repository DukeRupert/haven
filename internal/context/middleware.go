package context

import (
	"fmt"
	"strings"

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

// RouteContext middleware ensures route context is available
func (m *RouteContextMiddleware) RouteContext() echo.MiddlewareFunc {
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

// getRoutePattern returns the route pattern with parameter placeholders
func getRoutePattern(c echo.Context) string {
	route := c.Echo().Reverse(c.Path())
	if route == "" {
		return c.Path() // Fallback to actual path if reverse lookup fails
	}
	return route
}
