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

func (h *Handler) HandleGetUser(c echo.Context, ctx *dto.PageContext) error {
    logger := h.logger.With().
        Str("handler", "HandleGetUser").
        Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
        Str("facility", ctx.Auth.FacilityCode).
        Logger()
    
    auth, err := h.auth.GetAuthContext(c)
	if err != nil {
		logger.Error().
			Err(err).
			Msg("failed to get auth context")
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"Unable to verify permissions. Please try again later.",
		)
	}

    // Determine which profile to show
    initials := determineProfileInitials(c.Param("user_id"), ctx.Auth.Initials)

    // Check permissions if viewing other profile
    if initials != ctx.Auth.Initials && !canViewOtherProfiles(ctx.Auth) {
        logger.Warn().
            Str("viewer_initials", ctx.Auth.Initials).
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
        ctx.Auth.FacilityID,
    )
    if err != nil {
        logger.Error().
            Err(err).
            Str("initials", initials).
            Int("facility_id", ctx.Auth.FacilityID).
            Msg("failed to get user details")
        return echo.NewHTTPError(
            http.StatusNotFound,
            "Profile not found",
        )
    }

    // Build page props
    pageProps := dto.UserPageProps{
        PageCtx: 	 *ctx,
        Title:       fmt.Sprintf("Profile - %s %s", details.User.FirstName, details.User.LastName),
        Description: "View user profile and schedule information",
        Details:     details,
    }

    logger.Debug().
        Str("viewer_initials", ctx.Auth.Initials).
        Str("profile_initials", initials).
        Str("role", string(ctx.Auth.Role)).
        Bool("is_own_profile", initials == ctx.Auth.Initials).
        Msg("rendering profile page")

    // Handle HTMX requests if needed
    if isHtmxRequest(c) {
        return page.UserDetails(details.User, *auth ).Render(
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
