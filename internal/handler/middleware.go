package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/DukeRupert/haven/internal/model/types"
	"github.com/DukeRupert/haven/internal/response"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
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

// Combined middleware that handles role checks, facility access, and redirects
func RouteMiddleware(path string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			logger := log.With().
				Str("middleware", "RouteMiddleware").
				Str("path", c.Request().URL.Path).
				Logger()

			// Get route config
			config, exists := RouteConfigs[c.Path()]
			if !exists {
				return next(c) // No specific config, continue
			}

			// Get auth context
			auth, err := GetAuthContext(c)
			if err != nil {
				logger.Error().Err(err).Msg("failed to get auth context")
				return echo.NewHTTPError(http.StatusInternalServerError, "auth context error")
			}

			// Check role requirements
			if !IsAtLeastRole(string(auth.Role), string(config.MinRole)) {
				return echo.NewHTTPError(http.StatusForbidden, "insufficient permissions")
			}

			// Handle facility-specific logic
			facilityCode := c.Param("facility")

			// Handle redirects for non-facility routes
			if config.RedirectToFacility && facilityCode == "" {
				// Skip redirect for exempt roles
				for _, role := range config.ExceptRoles {
					if auth.Role == role {
						return next(c)
					}
				}

				// Redirect to facility-specific route if user has a facility
				if auth.FacilityCode != "" {
					newPath := fmt.Sprintf("/%s%s", auth.FacilityCode, c.Request().URL.Path)
					if c.QueryString() != "" {
						newPath = fmt.Sprintf("%s?%s", newPath, c.QueryString())
					}
					return c.Redirect(http.StatusTemporaryRedirect, newPath)
				}
			}

			// Check facility access for facility-specific routes
			if config.RequiresFacility && facilityCode != "" {
				// Super users can access any facility
				if auth.Role == "super" {
					return next(c)
				}

				// Other users must match their assigned facility
				if auth.FacilityCode != facilityCode {
					return echo.NewHTTPError(http.StatusForbidden, "unauthorized access to facility")
				}
			}

			return next(c)
		}
	}
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

// RequireMinRole creates middleware that ensures users have at least the specified role level
func RequireMinRole(minimumRole types.UserRole) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth, err := GetAuthContext(c)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "auth context error")
			}

			if !IsAtLeastRole(string(auth.Role), string(minimumRole)) {
				return echo.NewHTTPError(http.StatusForbidden, "insufficient permissions")
			}

			return next(c)
		}
	}
}

// RequireFacilityAccess creates middleware that ensures users have access to the specified facility
func RequireFacilityAccess() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get auth context
			auth, err := GetAuthContext(c)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "auth context error")
			}

			// Super users can access any facility
			if auth.Role == "super" {
				return next(c)
			}

			// Get facility from URL parameter
			facilityCode := c.Param("facility")
			if facilityCode == "" {
				return echo.NewHTTPError(http.StatusBadRequest, "facility code is required")
			}

			// For admin users, check if they're assigned to this facility
			if auth.Role == "admin" {
				if auth.FacilityCode == facilityCode {
					return next(c)
				}
				return echo.NewHTTPError(http.StatusForbidden, "unauthorized access to facility")
			}

			// For regular users (or unknown roles), deny access
			return echo.NewHTTPError(http.StatusForbidden, "insufficient permissions")
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

type APIRouteConfig struct {
	MinRole types.UserRole
}

var APIRouteConfigs = map[string]APIRouteConfig{
	"/api/facility": {
		MinRole: "super",
	},
	"/api/schedule": {
		MinRole: "admin",
	},
	"/api/user": {
		MinRole: "user",
	},
	"/api/available": {
		MinRole: "user",
	},
}

func API_Role_Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := getBasePath(c.Path())
			config, exists := APIRouteConfigs[path]
			if !exists {
				return response.System(c)
			}

			auth, err := GetAuthContext(c)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "auth context error")
			}

			// Check minimum required role
			if !IsAtLeastRole(string(auth.Role), string(config.MinRole)) {
				return echo.NewHTTPError(http.StatusForbidden, "insufficient permissions")
			}

			return next(c)
		}
	}
}

func getBasePath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		return path
	}

	// Handle paths with facility parameter
	if parts[2] == ":facility" || len(parts[2]) == 3 { // Assuming facility codes are 3 characters
		if len(parts) >= 4 {
			return fmt.Sprintf("/api/%s", parts[3])
		}
		return "/api"
	}

	return fmt.Sprintf("/api/%s", parts[2])
}

func FacilityAPIMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth, err := GetAuthContext(c)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "auth context error")
			}

			// Super users bypass facility check
			if auth.Role == "super" {
				return next(c)
			}

			facilityCode := c.Param("facility")
			if facilityCode != auth.FacilityCode {
				return echo.NewHTTPError(http.StatusForbidden, "unauthorized facility access")
			}

			return next(c)
		}
	}
}
