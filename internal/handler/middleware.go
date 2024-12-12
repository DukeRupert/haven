package handler

import (
	"github.com/DukeRupert/haven/internal/model/types"
	"github.com/labstack/echo/v4"
)

// RouteConfig defines access and redirection rules for each route
type RouteConfig struct {
	MinRole            types.UserRole   // Minimum role required to access this route
	RequiresFacility   bool             // Whether this route needs facility access
	RedirectToFacility bool             // Whether to redirect to facility-specific route
	ExceptRoles        []types.UserRole // Roles exempt from facility redirect
}

// RouteConfigs defines the access configuration for all routes
var RouteConfigs = map[string]RouteConfig{
	"/facilities": {
		MinRole:            "super",
		RequiresFacility:   false,
		RedirectToFacility: false,
	},
	"/controllers": {
		MinRole:            "admin",
		RequiresFacility:   true,
		RedirectToFacility: true,
		ExceptRoles:        []types.UserRole{"super"},
	},
	"/calendar": {
		MinRole:            "user",
		RequiresFacility:   true,
		RedirectToFacility: true,
		ExceptRoles:        []types.UserRole{"super"},
	},
	"/profile": {
		MinRole:            "user",
		RequiresFacility:   true,
		RedirectToFacility: true,
		ExceptRoles:        []types.UserRole{"super"},
	},
}

// WithNav wraps your handler functions with common navigation logic
func (h *Handler) WithNav(fn HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		logger := h.logger.With().
			Str("middleware", "WithNav").
			Str("path", c.Request().URL.Path).
			Logger()

		// Get route context
		routeCtx, err := GetRouteContext(c)
		if err != nil {
			logger.Error().Err(err).Msg("failed to get route context")
			return err
		}

		// Build navigation
		currentPath := c.Request().URL.Path
		navItems := BuildNav(routeCtx, currentPath)

		// Call the wrapped handler
		return fn(c, routeCtx, navItems)
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
