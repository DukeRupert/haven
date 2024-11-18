package handler

import (
	"net/http"
	"time"

	"github.com/DukeRupert/haven/db"
	"github.com/DukeRupert/haven/view/component"
	"github.com/DukeRupert/haven/view/page"
	"github.com/labstack/echo/v4"
)

func (h *Handler) CreateScheduleForm(c echo.Context) error {
	route := h.RouteCtx
	return render(c, component.CreateScheduleForm(route))
}

func (h *Handler) UpdateScheduleForm(c echo.Context) error {
	h.logger.Info().
		Str("facility_code", h.RouteCtx.FacilityCode).
		Str("user_initials", h.RouteCtx.UserInitials).
		Msg("UpdateScheduleForm() executing")

	schedule, err := h.db.GetScheduleByCode(
		c.Request().Context(),
		h.RouteCtx.FacilityCode,
		h.RouteCtx.UserInitials,
	)
	if err != nil {
		h.logger.Error().Err(err).
			Str("facility_code", h.RouteCtx.FacilityCode).
			Str("user_initials", h.RouteCtx.UserInitials).
			Msg("Failed to retrieve schedule")
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve schedule")
	}

	// If no schedule found, return 404
	if schedule == nil {
		return echo.NewHTTPError(http.StatusNotFound, "No schedule found for this user")
	}

	h.logger.Debug().
		Str("facility_code", h.RouteCtx.FacilityCode).
		Str("user_initials", h.RouteCtx.UserInitials).
		Int("schedule_id", schedule.ID).
		Msg("Schedule retrieved successfully")

	route := h.RouteCtx

	return render(c, component.UpdateScheduleForm(route, *schedule))
}

func (h *Handler) GetScheduleHandler(c echo.Context) error {
	logger := h.logger

	// Get route parameters
	code := c.Param("code")
	initials := c.Param("initials")

	// Get schedule from database
	schedule, err := h.db.GetScheduleByCode(
		c.Request().Context(),
		code,
		initials,
	)
	if err != nil {
		logger.Error().Err(err).
			Str("facility_code", code).
			Str("user_initials", initials).
			Msg("Failed to get schedule")
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get schedule")
	}

	// Get session data for rendering
	auth, err := GetAuthContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "auth context error")
	}

	// Log successful retrieval
	logger.Info().
		Int("schedule_id", schedule.ID).
		Str("facility_code", code).
		Str("user_initials", initials).
		Time("start_date", schedule.StartDate).
		Msgf("schedule retrieved successfully with weekdays %s and %s",
			schedule.FirstWeekday, schedule.SecondWeekday)

	// Return the schedule card
	return render(c, page.ScheduleCard(h.RouteCtx, *auth, *schedule))
}

type CreateScheduleRequest struct {
	FirstWeekday  int    `form:"first_weekday" validate:"required,min=0,max=6"`
	SecondWeekday int    `form:"second_weekday" validate:"required,min=0,max=6"`
	StartDate     string `form:"start_date" validate:"required"`
}

func (h *Handler) CreateScheduleHandler(c echo.Context) error {
	logger := h.logger

	// Parse form data
	var formData CreateScheduleRequest
	if err := c.Bind(&formData); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Parse the start date
	startDate, err := time.Parse("2006-01-02", formData.StartDate)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid date format")
	}

	// Create parameters with converted values
	params := db.CreateScheduleByCodeParams{
		FacilityCode:  h.RouteCtx.FacilityCode,
		UserInitials:  h.RouteCtx.UserInitials,
		FirstWeekday:  time.Weekday(formData.FirstWeekday),
		SecondWeekday: time.Weekday(formData.SecondWeekday),
		StartDate:     startDate,
	}

	schedule, err := h.db.CreateScheduleByCode(c.Request().Context(), params)
	if err != nil {
		logger.Error().Err(err).
			Str("facility_code", params.FacilityCode).
			Str("user_initials", params.UserInitials).
			Msg("Failed to create schedule")
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create schedule")
	}

	// Get session data
	auth, err := GetAuthContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "auth context error")
	}

	route := h.RouteCtx
	logger.Info().
		Int("schedule_id", schedule.ID).
		Str("facility_code", params.FacilityCode).
		Str("user_initials", params.UserInitials).
		Time("start_date", schedule.StartDate).
		Msgf("schedule created successfully with weekdays %s and %s",
			schedule.FirstWeekday, schedule.SecondWeekday)

	// Return the updated schedule card
	return render(c, page.ScheduleCard(route, *auth, *schedule))
}

func (h *Handler) UpdateScheduleHandler(c echo.Context) error {
	logger := h.logger

	// Get route parameters
	code := c.Param("code")
	initials := c.Param("initials")

	// Parse form data with correct form binding
	var formData struct {
		FirstWeekday  int    `form:"first_weekday"`
		SecondWeekday int    `form:"second_weekday"`
		StartDate     string `form:"start_date"`
	}

	if err := c.Bind(&formData); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Parse the start date
	startDate, err := time.Parse("2006-01-02", formData.StartDate)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid date format")
	}

	// Create update parameters with converted values
	params := db.UpdateScheduleParams{
		FirstWeekday:  time.Weekday(formData.FirstWeekday),
		SecondWeekday: time.Weekday(formData.SecondWeekday),
		StartDate:     startDate,
	}

	// Update schedule in database
	schedule, err := h.db.UpdateScheduleByCode(
		c.Request().Context(),
		code,
		initials,
		params,
	)
	if err != nil {
		logger.Error().Err(err).
			Str("facility_code", code).
			Str("user_initials", initials).
			Msg("Failed to update schedule")
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update schedule")
	}

	// Get session data for rendering
	auth, err := GetAuthContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "auth context error")
	}

	// Log successful update
	logger.Info().
		Int("schedule_id", schedule.ID).
		Str("facility_code", code).
		Str("user_initials", initials).
		Time("start_date", schedule.StartDate).
		Msgf("schedule updated successfully with weekdays %s and %s",
			schedule.FirstWeekday, schedule.SecondWeekday)

	// Return the updated schedule card
	return render(c, page.ScheduleCard(h.RouteCtx, *auth, *schedule))
}
