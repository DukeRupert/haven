package handler

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/DukeRupert/haven/internal/model/dto"
	"github.com/DukeRupert/haven/internal/model/types"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

func RequestLogger(logger zerolog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()

			start := time.Now()

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			stop := time.Now()

			logger.Info().
				Str("method", req.Method).
				Str("path", req.URL.Path).
				Int("status", res.Status).
				Str("ip", c.RealIP()).
				Dur("latency", stop.Sub(start)).
				Msg("Request handled")

			return err
		}
	}
}

func ErrorLogger(logger zerolog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err != nil {
				logger.Error().
					Err(err).
					Str("path", c.Request().URL.Path).
					Str("method", c.Request().Method).
					Msg("Request error occurred")
			}
			return err
		}
	}
}

// RouteConfig defines the configuration for a route including access control
type RouteConfig struct {
	Title            string
	Icon             string
	MinRole          types.UserRole
	RequiresFacility bool
	Children         []string
}

// RouteConfigs maps paths to their configurations
var RouteConfigs = map[string]RouteConfig{
	"/calendar": {
		Title:            "Calendar",
		Icon:             "calendar",
		MinRole:          types.UserRoleUser,
		RequiresFacility: true,
		Children: []string{
			"/calendar/:date",
			"/calendar/:date/events",
		},
	},
	"/profile": {
		Title:    "Profile",
		Icon:     "user",
		MinRole:  types.UserRoleUser,
		Children: []string{"/profile/edit"},
	},
	"/users": {
		Title:            "Users",
		Icon:             "users",
		MinRole:          types.UserRoleAdmin,
		RequiresFacility: true,
		Children: []string{
			"/users/new",
			"/users/:id",
			"/users/:id/edit",
		},
	},
	"/facilities": {
		Title:   "Facilities",
		Icon:    "building",
		MinRole: types.UserRoleSuper,
		Children: []string{
			"/facilities/new",
			"/facilities/:id",
			"/facilities/:id/edit",
		},
	},
}

// Updated handler wrapper for cleaner context usage
func (h *Handler) WithNav(fn HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		routeCtx, err := getRouteContext(c)
		if err != nil {
			return err
		}

		auth, err := GetAuthContext(c)
		if err != nil {
			return err
		}

		navItems := BuildNav(routeCtx, auth, c.Request().URL.Path)

		pageCtx := &dto.PageContext{
			Route: routeCtx,
			Auth:  auth,
			Nav:   navItems,
		}

		return fn(c, pageCtx)
	}
}

// getRouteContext extracts and validates route context
func getRouteContext(c echo.Context) (*dto.RouteContext, error) {
	// Try to get existing route context
	ctx, ok := c.Get("routeCtx").(*dto.RouteContext)
	if !ok || ctx == nil {
		return nil, fmt.Errorf("route context not found or invalid")
	}

	// Always check URL param for facility code to ensure it's current
	if facilityCode := c.Param("facility_id"); facilityCode != "" {
		// Update base path with current facility code
		ctx.BasePath = fmt.Sprintf("/facility/%s", facilityCode)
	}

	// Ensure CurrentPath and FullPath are set
	if ctx.CurrentPath == "" {
		ctx.CurrentPath = c.Path()
	}
	if ctx.FullPath == "" {
		ctx.FullPath = c.Request().URL.Path
	}

	return ctx, nil
}

// Helper function to create a new route context if needed
func newRouteContext(c echo.Context) *dto.RouteContext {
	facilityCode := c.Param("facility_id")

	var basePath string
	if facilityCode != "" {
		basePath = fmt.Sprintf("/facility/%s", facilityCode)
	}

	return &dto.RouteContext{
		BasePath:    basePath,
		CurrentPath: c.Path(),             // Route pattern with parameter placeholders
		FullPath:    c.Request().URL.Path, // Actual URL path
	}
}

// getRouteContext extracts and validates route context
func BuildNav(routeCtx *dto.RouteContext, auth *dto.AuthContext, currentPath string) []dto.NavItem {
	navItems := []dto.NavItem{}

	for configPath, config := range RouteConfigs {
		// Skip if user doesn't have required role
		if !IsAtLeastRole(string(auth.Role), string(config.MinRole)) {
			continue
		}

		// Build full path using facility code from auth context
		fullPath := configPath
		if config.RequiresFacility && auth.FacilityCode != "" {
			fullPath = fmt.Sprintf("/facility/%s%s", auth.FacilityCode, configPath)
		}

		navItems = append(navItems, dto.NavItem{
			Path:    fullPath,
			Name:    config.Title,
			Icon:    config.Icon,
			Active:  isActiveRoute(routeCtx.CurrentPath, fullPath, RouteConfigs),
			Visible: true,
		})
	}

	return navItems
}

// isActiveRoute checks if the current path matches the nav item or its children
func isActiveRoute(currentPattern, navPath string, configs map[string]RouteConfig) bool {
	// Remove facility prefix for comparison
	normalizedCurrent := currentPattern
	if strings.HasPrefix(normalizedCurrent, "/facility/") {
		parts := strings.SplitN(normalizedCurrent, "/", 4)
		if len(parts) >= 4 {
			normalizedCurrent = "/" + parts[3]
		}
	}

	// Remove facility prefix from nav path
	normalizedNav := navPath
	if strings.HasPrefix(normalizedNav, "/facility/") {
		parts := strings.SplitN(normalizedNav, "/", 4)
		if len(parts) >= 4 {
			normalizedNav = "/" + parts[3]
		}
	}

	// Check direct match
	if normalizedCurrent == normalizedNav {
		return true
	}

	// Check if current path is a child of this nav item
	config, exists := configs[normalizedNav]
	if !exists {
		return false
	}

	for _, childPath := range config.Children {
		// Convert parameter placeholders to match current pattern
		pattern := convertPathToPattern(childPath)
		if pattern == normalizedCurrent {
			return true
		}
	}

	return false
}

// convertPathToPattern converts paths with parameters to match Echo's pattern
func convertPathToPattern(path string) string {
	// Replace :param with *
	re := regexp.MustCompile(`:[^/]+`)
	return re.ReplaceAllString(path, "*")
}

// Helper function to check role hierarchy
func IsAtLeastRole(userRole string, minRole string) bool {
	roleHierarchy := map[string]int{
		string(types.UserRoleSuper): 3,
		string(types.UserRoleAdmin): 2,
		string(types.UserRoleUser):  1,
	}

	userRoleLevel, ok1 := roleHierarchy[userRole]
	minRoleLevel, ok2 := roleHierarchy[minRole]

	if !ok1 || !ok2 {
		return false
	}

	return userRoleLevel >= minRoleLevel
}

