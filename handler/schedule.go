package handler

import (
	"net/http"
	"strings"

	"github.com/DukeRupert/haven/db"
	"github.com/DukeRupert/haven/view/component"
	"github.com/DukeRupert/haven/view/page"
	"github.com/labstack/echo/v4"
)

func (h *Handler) CreateScheduleForm(c echo.Context) error {
	route := h.RouteCtx
	return render(c, component.CreateScheduleForm(route))
}

func (h *Handler) UpdateScheduleHandler(c echo.Context) error {
	auth, err := GetAuthContext(c)
    if err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, "auth context error")
    }

	var params db.UpdateScheduleParams
	if err := c.Bind(&params); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Validate the params
	if err := c.Validate(params); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	schedule, err := h.db.UpdateScheduleByCode(
		c.Request().Context(),
		h.RouteCtx.FacilityCode,
		h.RouteCtx.UserInitials,
		params,
	)

	if err != nil {
		h.logger.Error().Err(err).
			Str("facility_code", h.RouteCtx.FacilityCode).
			Str("user_initials", h.RouteCtx.UserInitials).
			Interface("params", params).
			Msg("Failed to update schedule")

		if strings.Contains(err.Error(), "no schedule found") {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update schedule")
	}

	h.logger.Info().
		Str("facility_code", h.RouteCtx.FacilityCode).
		Str("user_initials", h.RouteCtx.UserInitials).
		Int("schedule_id", schedule.ID).
		Msg("Schedule updated successfully")

		route := h.RouteCtx
		return render(c, page.ScheduleCard(route, *auth, *schedule))
}
