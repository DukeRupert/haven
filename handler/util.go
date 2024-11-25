package handler

import (
	"fmt"
	"strings"

	"github.com/DukeRupert/haven/types"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func render(c echo.Context, component templ.Component) error {
	c.Response().Header().Set("Content-Type", "text/html")
	return component.Render(c.Request().Context(), c.Response())
}

func BuildNav(routeCtx *types.RouteContext, currentPath string) []types.NavItem {
	strippedPath := strings.TrimPrefix(currentPath, "/"+routeCtx.FacilityCode)

	navItems := []types.NavItem{}

	// Add nav items based on role access
	for path, config := range RouteConfigs {
		if IsAtLeastRole(string(routeCtx.UserRole), string(config.MinRole)) {
			navPath := path
			if config.RequiresFacility && routeCtx.FacilityCode != "" {
				navPath = fmt.Sprintf("/%s%s", routeCtx.FacilityCode, path)
			}

			title := cases.Title(language.English)

			navItems = append(navItems, types.NavItem{
				Path:    navPath,
				Name:    title.String(strings.TrimPrefix(path, "/")),
				Active:  strippedPath == path,
				Visible: true,
			})
		}
	}

	return navItems
}

// Helper function to build correct paths
func buildPath(facilityCode string, path string) string {
	if facilityCode == "" {
		return path
	}
	return fmt.Sprintf("/%s%s", facilityCode, path)
}
