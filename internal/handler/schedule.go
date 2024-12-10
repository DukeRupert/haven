package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/DukeRupert/haven/types"
	"github.com/DukeRupert/haven/view/alert"
	"github.com/DukeRupert/haven/view/component"
	"github.com/DukeRupert/haven/view/page"
	"github.com/labstack/echo/v4"
)

func (h *Handler) createScheduleForm(c echo.Context) error {
	// Get facility code and user initials from params
	facility := c.Param("facility")
	if facility == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing facility parameter")
	}

	initials := c.Param("initials")
	if initials == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing initials parameter")
	}

	return render(c, component.CreateScheduleForm(facility, initials))
}

func (h *Handler) updateScheduleForm(c echo.Context) error {
	h.logger.Info().
		Str("Schedule ID", c.Param("id")).
		Msg("updateScheduleForm() executing")

	// Parse the schedule ID from the URL
	scheduleID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid protected date ID")
	}

	// Get user from context
	auth, err := GetAuthContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "auth context error")
	}

	// Query database for record
	schedule, err := h.db.GetScheduleByID(
		c.Request().Context(),
		scheduleID,
	)
	if err != nil {
		h.logger.Error().Err(err).
			Msg("Failed to retrieve schedule")
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve schedule")
	}

	// If no schedule found, return 404
	if schedule == nil {
		return echo.NewHTTPError(http.StatusNotFound, "No schedule found for this user")
	}

	// Check authorization
	if !isAuthorized(auth.UserID, auth.Role, schedule.UserID) {
		return echo.NewHTTPError(http.StatusForbidden, "unauthorized to modify this protected date")
	}

	// Build and return component
	component := component.UpdateScheduleForm(*schedule)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

func (h *Handler) handleGetSchedule(c echo.Context) error {
	logger := h.logger

	// Get route parameters
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		logger.Error().
			Err(err).
			Str("facility_id", c.Param("id")).
			Msg("invalid facility ID format")
		return render(c, alert.Error(
			"Invalid request",
			[]string{"Invalid schedule ID provided"},
		))
	}

	// Get schedule from database
	schedule, err := h.db.GetScheduleByID(
		c.Request().Context(),
		id,
	)
	if err != nil {
		logger.Error().Err(err).
			Int("schedule id", id).
			Msg("Failed to get schedule")
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get schedule")
	}

	// Get session data for rendering
	auth, err := GetAuthContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "auth context error")
	}

	// Return the schedule card
	return render(c, page.ScheduleCard(auth.Role, *schedule))
}

func (h *Handler) handleCreateSchedule(c echo.Context) error {
	logger := h.logger

	// Get facility code and user initials from params
	facility := c.Param("facility")
	if facility == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing facility parameter")
	}

	initials := c.Param("initials")
	if initials == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing initials parameter")
	}

	// Parse form data
	var formData types.CreateScheduleRequest
	if err := c.Bind(&formData); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Parse the start date
	startDate, err := time.Parse("2006-01-02", formData.StartDate)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid date format")
	}

	// Create parameters with converted values
	params := types.CreateScheduleByCodeParams{
		FacilityCode:  facility,
		UserInitials:  initials,
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

	logger.Info().
		Int("schedule_id", schedule.ID).
		Str("facility_code", params.FacilityCode).
		Str("user_initials", params.UserInitials).
		Time("start_date", schedule.StartDate).
		Msgf("schedule created successfully with weekdays %s and %s",
			schedule.FirstWeekday, schedule.SecondWeekday)

	// Return the updated schedule card
	return render(c, page.ScheduleCard(auth.Role, *schedule))
}

func (h *Handler) handleUpdateSchedule(c echo.Context) error {
	logger := h.logger

	// Parse the schedule ID from the URL
	scheduleID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid protected date ID")
	}

	// Get user from context
	auth, err := GetAuthContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "auth context error")
	}

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
	params := types.UpdateScheduleParams{
		FirstWeekday:  time.Weekday(formData.FirstWeekday),
		SecondWeekday: time.Weekday(formData.SecondWeekday),
		StartDate:     startDate,
	}

	// Fetch the protected date to check ownership
	schedule, err := h.db.GetScheduleByID(c.Request().Context(), scheduleID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "error fetching protected date")
	}

	// Check authorization
	if !isAuthorized(auth.UserID, auth.Role, schedule.UserID) {
		return echo.NewHTTPError(http.StatusForbidden, "unauthorized to modify this protected date")
	}

	// Update schedule in database
	schedule, err = h.db.UpdateSchedule(
		c.Request().Context(),
		scheduleID,
		params,
	)
	if err != nil {
		logger.Error().Err(err).
			Msg("Failed to update schedule")
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update schedule")
	}

	component := page.ScheduleCard(
		auth.Role,
		*schedule,
	)

	return component.Render(c.Request().Context(), c.Response().Writer)
}

// isAuthorized checks if a user is authorized to toggle a protected date's availability
func isAuthorized(userID int, role types.UserRole, recordID int) bool {
	// Allow access if user is admin or super
	if role == "admin" || role == "super" {
		return true
	}

	// Allow access if user owns the protected date
	if role == "user" && userID == recordID {
		return true
	}

	return false
}
