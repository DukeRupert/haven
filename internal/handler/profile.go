// internal/handler/profile.go
package handler

import (
	"fmt"
	"net/http"

	"github.com/DukeRupert/haven/internal/model/dto"
	"github.com/DukeRupert/haven/internal/model/types"
	"github.com/DukeRupert/haven/web/view/page"
	"github.com/labstack/echo/v4"
)

// HandleProfile renders the user profile page
func (h *Handler) HandleProfile(c echo.Context, routeCtx *dto.RouteContext, navItems []dto.NavItem) error {
	logger := h.logger.With().
		Str("handler", "HandleProfile").
		Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
		Str("facility", c.Param("facility")).
		Logger()

	// Get auth context
	auth, err := GetAuthContext(c)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get auth context")
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"Unable to verify permissions",
		)
	}

	// Determine which profile to show
	initials := determineProfileInitials(c.Param("initials"), auth.Initials)

	// Check permissions if viewing other profile
	if initials != auth.Initials && !canViewOtherProfiles(auth) {
		logger.Warn().
			Str("viewer_initials", auth.Initials).
			Str("target_initials", initials).
			Msg("unauthorized attempt to view other profile")
		return echo.NewHTTPError(
			http.StatusForbidden,
			"You don't have permission to view other profiles",
		)
	}

	// Get user details
	details, err := h.repos.User.GetDetails(
		c.Request().Context(),
		initials,
		auth.FacilityID,
	)
	if err != nil {
		logger.Error().
			Err(err).
			Str("initials", initials).
			Int("facility_id", auth.FacilityID).
			Msg("failed to get user details")
		return echo.NewHTTPError(
			http.StatusNotFound,
			"Profile not found",
		)
	}

	// Build page props
	pageProps := buildProfilePageProps(initials, auth, details, routeCtx, navItems)

	logger.Debug().
		Str("viewer_initials", auth.Initials).
		Str("profile_initials", initials).
		Str("role", string(auth.Role)).
		Bool("is_own_profile", initials == auth.Initials).
		Msg("rendering profile page")

	return page.Profile(
		pageProps.RouteCtx,
		pageProps.NavItems,
		pageProps.Title,
		pageProps.Description,
		pageProps.Auth,
		pageProps.Details,
	).Render(c.Request().Context(), c.Response().Writer)
}

// Helper functions
func determineProfileInitials(paramInitials, authInitials string) string {
	if paramInitials == "" {
		return authInitials
	}
	return paramInitials
}

func canViewOtherProfiles(auth *dto.AuthContext) bool {
	return auth.Role == types.UserRoleAdmin || auth.Role == types.UserRoleSuper
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
