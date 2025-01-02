// internal/handler/calendar.go
package handler

import (
	"net/http"
	"time"

	"github.com/DukeRupert/haven/internal/middleware"
	"github.com/DukeRupert/haven/internal/model/dto"
	"github.com/DukeRupert/haven/web/view/component"
	"github.com/DukeRupert/haven/web/view/page"
	"github.com/labstack/echo/v4"
)

// HandleCalendar renders the calendar view
func (h *Handler) HandleCalendar(c echo.Context) error {
	logger := h.logger.With().
        Str("handler", "HandleCalendar").
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

	// Get view date from query params or default to current month
    monthParam := c.QueryParam("month")
    logger.Debug().
        Str("raw_month_param", monthParam).
        Bool("has_month_param", monthParam != "").
        Msg("Month parameter check")

    viewDate, err := getViewDate(monthParam)
    if err != nil {
        logger.Error().
            Err(err).
            Str("month_param", monthParam).
            Msg("invalid month format")
        return echo.NewHTTPError(
            http.StatusBadRequest,
            "Invalid month format. Please use YYYY-MM",
        )
    }

	var facilityCode string
    // If we're on a facility-specific route, use that facility code
    if route.FacilityCode != "" {
        facilityCode = route.FacilityCode
    } else {
        // Default to user's facility code for the general calendar route
        if auth.FacilityCode == "" {
            return echo.NewHTTPError(http.StatusBadRequest, "no facility available")
        }
        facilityCode = auth.FacilityCode
    }

	// Get protected dates
	protectedDates, err := h.repos.Schedule.GetProtectedDatesByFacilityCode(
		c.Request().Context(),
		facilityCode,
	)
	if err != nil {
		logger.Error().
			Err(err).
			Str("facility_code", facilityCode).
			Msg("failed to fetch protected dates")
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"Unable to load calendar data",
		)
	}

	// Build nav items
	navItems := BuildNav(route, auth, c.Request().URL.Path)

	 // Build calendar props
    calendarProps := dto.CalendarProps{
        CurrentMonth:    viewDate,
        ProtectedDates: protectedDates,
        AuthCtx:       *auth,
        RouteCtx:  *route,
    }

	// Build page props
	pageProps := dto.CalendarPageProps{
		Title:       "Calendar",
		Description: "View the facility calendar",
		NavItems:	navItems,
		AuthCtx: 	*auth,
		RouteCtx: 	*route,
		Calendar:    calendarProps,
	}

	// Handle HTMX requests
	if isHtmxRequest(c) {
		return component.Calendar(calendarProps).Render(
			c.Request().Context(),
			c.Response().Writer,
		)
	}

	// Render full page
	return page.CalendarPage(pageProps).Render(
		c.Request().Context(),
		c.Response().Writer,
	)
}

// Helper functions
func getViewDate(monthStr string) (time.Time, error) {
	if monthStr != "" {
		return time.Parse("2006-01", monthStr)
	}
	now := time.Now()
	return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local), nil
}