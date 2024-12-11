// internal/handler/calendar.go
package handler

import (
	"net/http"
	"time"

	"github.com/DukeRupert/haven/internal/model/dto"
	"github.com/DukeRupert/haven/web/view/component"
	"github.com/DukeRupert/haven/web/view/page"
	"github.com/labstack/echo/v4"
)

// HandleCalendar renders the calendar view
func (h *Handler) HandleCalendar(c echo.Context, routeCtx *dto.RouteContext, navItems []dto.NavItem) error {
	logger := h.logger.With().
		Str("handler", "HandleCalendar").
		Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
		Logger()

	// Get and validate facility code
	facilityCode := c.Param("facility")
	if facilityCode == "" {
		logger.Error().Msg("missing facility code")
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"Facility code is required",
		)
	}

	// Get view date from query params or default to current month
	viewDate, err := getViewDate(c.QueryParam("month"))
	if err != nil {
		logger.Error().
			Err(err).
			Str("month_param", c.QueryParam("month")).
			Msg("invalid month format")
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"Invalid month format. Please use YYYY-MM",
		)
	}

	// Get auth context
	auth, err := GetAuthContext(c)
	if err != nil {
		logger.Error().
			Err(err).
			Msg("failed to get auth context")
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"Unable to verify permissions",
		)
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

	// Build calendar props
	calendarProps := dto.CalendarProps{
		CurrentMonth:   viewDate,
		FacilityCode:   facilityCode,
		ProtectedDates: protectedDates,
		UserRole:       auth.Role,
		CurrentUserID:  auth.UserID,
	}

	// Build page props
	pageProps := dto.CalendarPageProps{
		Route:       *routeCtx,
		NavItems:    navItems,
		Auth:        *auth,
		Title:       "Calendar",
		Description: "View the facility calendar",
		Calendar:    calendarProps,
	}

	logger.Debug().
		Str("facility_code", facilityCode).
		Time("view_date", viewDate).
		Int("protected_dates_count", len(protectedDates)).
		Bool("is_htmx", isHtmxRequest(c)).
		Msg("rendering calendar")

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

func isHtmxRequest(c echo.Context) bool {
	return c.Request().Header.Get("HX-Request") == "true"
}