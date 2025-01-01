// internal/handler/profile.go
package handler

import (
	"fmt"
	"net/http"

	"github.com/DukeRupert/haven/internal/middleware"
	"github.com/DukeRupert/haven/internal/model/dto"
	"github.com/DukeRupert/haven/internal/model/types"
	"github.com/DukeRupert/haven/web/view/page"
	"github.com/labstack/echo/v4"
)

func (h *Handler) HandleGetUser(c echo.Context) error {
	logger := h.logger.With().
		Str("handler", "HandleGetUser").
		Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
		Logger()

	auth, err := middleware.GetAuthContext(c)
	if err != nil {
		logger.Error().Msg("missing auth context")
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"Authentication is required",
		)
	}

	route, err := middleware.GetRouteContext(c)
	if err != nil {
		logger.Error().Msg("missing route context")
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"Missing route context",
		)
	}

	// Get profile from parameters, default to user initials
	initials := c.Param("user_initials")
	if initials == "" {
		initials = auth.Initials
	}

	facility := c.Param("facility_code")
	if facility == "" {
		facility = auth.FacilityCode
	}

	// Check permissions if viewing other profile
	if initials != auth.Initials && !canViewOtherProfiles(auth, route) {
		logger.Warn().
			Str("viewer_initials", auth.Initials).
			Str("target_initials", initials).
			Msg("unauthorized attempt to view other profile")
		return echo.NewHTTPError(
			http.StatusForbidden,
			"You don't have permission to view this profile",
		)
	}

	// Get user details
	details, err := h.repos.User.GetDetails(
		c.Request().Context(),
		initials,
		facility,
	)
	if err != nil {
		logger.Error().
			Err(err).
			Str("initials", initials).
			Str("facility_id", facility).
			Msg("failed to get user details")
		return echo.NewHTTPError(
			http.StatusNotFound,
			"Profile not found",
		)
	}

	// Build nav items
	navItems := BuildNav(route, auth, c.Request().URL.Path)

	// Build page props
	pageProps := dto.ProfilePageProps{
		Title:       fmt.Sprintf("Profile - %s %s", details.User.FirstName, details.User.LastName),
		Description: "View user profile and schedule information",
		NavItems:    navItems,
		AuthCtx:     *auth,
		RouteCtx:    *route,
		Details:     details,
	}

	// Handle HTMX requests if needed
	if isHtmxRequest(c) {
		return page.UserDetails(details.User, route.FacilityCode, *auth).Render(
			c.Request().Context(),
			c.Response().Writer,
		)
	}

	// Render full page
	return page.UserPage(pageProps).Render(
		c.Request().Context(),
		c.Response().Writer,
	)
}

// Helper functions

func determineProfileInitials(paramInitials, authInitials string) string {
	if paramInitials == "" {
		return authInitials // Default to viewing own profile
	}
	return paramInitials
}

func canViewOtherProfiles(auth *dto.AuthContext, route *dto.RouteContext) bool {
	// Super users can view all profiles
	if auth.Role == types.UserRoleSuper {
		return true
	}

	// Admins can only view profiles within their facility
	if auth.Role == types.UserRoleAdmin {
		return auth.FacilityCode == route.FacilityCode
	}

	// Regular users cannot view other profiles
	return false
}

type profilePageProps struct {
	RouteCtx    dto.RouteContext
	NavItems    []dto.NavItem
	Title       string
	Description string
	Auth        *dto.AuthContext
	Details     *dto.UserDetails
}

func buildProfilePageProps(initials string, auth *dto.AuthContext, details *dto.UserDetails,
	routeCtx *dto.RouteContext, navItems []dto.NavItem,
) profilePageProps {
	title := "Profile"
	description := "Manage your account information and schedule"

	if initials != auth.Initials {
		title = fmt.Sprintf("%s's Profile", details.User.Initials)
		description = fmt.Sprintf("View %s's information and schedule", details.User.Initials)
	}

	return profilePageProps{
		RouteCtx:    *routeCtx,
		NavItems:    navItems,
		Title:       title,
		Description: description,
		Auth:        auth,
		Details:     details,
	}
}
