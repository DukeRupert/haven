package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/DukeRupert/haven/internal/model/dto"
	"github.com/DukeRupert/haven/internal/model/entity"
	"github.com/DukeRupert/haven/internal/model/types"
	"github.com/DukeRupert/haven/internal/response"
	"github.com/DukeRupert/haven/web/view/alert"
	"github.com/DukeRupert/haven/web/view/component"
	"github.com/DukeRupert/haven/web/view/page"
	"github.com/labstack/echo/v4"
)

// HandleAvailabilityToggle processes requests to toggle protected date availability
func (h *Handler) HandleAvailabilityToggle(c echo.Context) error {
	logger := h.logger.With().
		Str("handler", "HandleAvailabilityToggle").
		Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
		Logger()

	// Get and validate protected date ID
	dateID, err := getProtectedDateID(c)
	if err != nil {
		logger.Debug().
			Err(err).
			Str("date_id_param", c.Param("id")).
			Msg("invalid protected date ID")
		return response.Error(c,
			http.StatusBadRequest,
			"Invalid Request",
			[]string{"Invalid protected date ID provided"},
		)
	}

	// Get auth context
	auth, err := h.auth.GetAuthContext(c)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get auth context")
		return response.System(c)
	}

	// Get protected date
	protectedDate, err := h.repos.Schedule.GetProtectedDateByID(c.Request().Context(), dateID)
	if err != nil {
		logger.Error().
			Err(err).
			Int("date_id", dateID).
			Msg("failed to fetch protected date")
		return response.System(c)
	}

	// Verify authorization
	if !canToggleAvailability(auth, protectedDate) {
		logger.Warn().
			Int("date_id", dateID).
			Int("user_id", auth.UserID).
			Str("role", string(auth.Role)).
			Msg("unauthorized toggle attempt")
		return response.Error(c,
			http.StatusForbidden,
			"Access Denied",
			[]string{"You don't have permission to modify this date"},
		)
	}

	// Toggle availability
	updatedDate, err := h.repos.Schedule.ToggleProtectedDateAvailability(
		c.Request().Context(),
		dateID,
	)
	if err != nil {
		logger.Error().
			Err(err).
			Int("date_id", dateID).
			Msg("failed to toggle availability")
		return response.System(c)
	}

	logger.Debug().
		Int("date_id", dateID).
		Bool("is_available", updatedDate.Available).
		Msg("availability toggled successfully")

	return render(c, component.ProtectedDay(
		auth.UserID,
		auth.FacilityCode,
		updatedDate,
	))
}

// Helper functions
func getProtectedDateID(c echo.Context) (int, error) {
	return strconv.Atoi(c.Param("id"))
}

func canToggleAvailability(auth *dto.AuthContext, date *entity.ProtectedDate) bool {
	// Super users can modify any date
	if auth.Role == types.RoleSuper {
		return true
	}

	// Admin users can modify dates in their facility
	if auth.Role == types.RoleAdmin && auth.FacilityID == date.FacilityID {
		return true
	}

	// Users can only modify their own dates
	return auth.UserID == date.UserID
}

// GetCreateScheduleForm renders the schedule creation form
func (h *Handler) GetCreateScheduleForm(c echo.Context) error {
	logger := h.logger.With().
		Str("handler", "GetCreateScheduleForm").
		Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
		Logger()

	// Get and validate params
	params, err := validateScheduleFormParams(c)
	if err != nil {
		return err // validateScheduleFormParams handles error responses
	}

	// Get auth context
	auth, err := h.auth.GetAuthContext(c)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get auth context")
		return response.System(c)
	}

	// Verify facility exists
	_, err = h.repos.Facility.GetByCode(c.Request().Context(), params.FacilityCode)
	if err != nil {
		logger.Error().
			Err(err).
			Str("facility_code", params.FacilityCode).
			Msg("facility not found")
		return response.Error(c,
			http.StatusNotFound,
			"Invalid Facility",
			[]string{"The specified facility does not exist"},
		)
	}

	// Verify user exists
	_, err = h.repos.User.GetByInitialsAndFacility(
		c.Request().Context(),
		params.Initials,
		auth.FacilityID,
	)
	if err != nil {
		logger.Error().
			Err(err).
			Str("initials", params.Initials).
			Int("facility_id", auth.FacilityID).
			Msg("user not found")
		return response.Error(c,
			http.StatusNotFound,
			"Invalid User",
			[]string{"The specified user does not exist in this facility"},
		)
	}

	logger.Debug().
		Str("facility_code", params.FacilityCode).
		Str("user_initials", params.Initials).
		Msg("rendering schedule creation form")

	return render(c, component.CreateScheduleForm(
		params.FacilityCode,
		params.Initials,
	))
}

// Types and validation
type scheduleFormParams struct {
	FacilityCode string
	Initials     string
}

func validateScheduleFormParams(c echo.Context) (*scheduleFormParams, error) {
	facilityCode := c.Param("facility")
	if facilityCode == "" {
		return nil, response.Error(c,
			http.StatusBadRequest,
			"Missing Parameter",
			[]string{"Facility code is required"},
		)
	}

	initials := c.Param("initials")
	if initials == "" {
		return nil, response.Error(c,
			http.StatusBadRequest,
			"Missing Parameter",
			[]string{"User initials are required"},
		)
	}

	return &scheduleFormParams{
		FacilityCode: facilityCode,
		Initials:     initials,
	}, nil
}

// GetUpdateScheduleForm renders the schedule update form
func (h *Handler) GetUpdateScheduleForm(c echo.Context) error {
	logger := h.logger.With().
		Str("handler", "GetUpdateScheduleForm").
		Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
		Logger()
 
	// Get and validate schedule ID
	scheduleID, err := getScheduleID(c)
	if err != nil {
		logger.Debug().
			Err(err).
			Str("schedule_id_param", c.Param("id")).
			Msg("invalid schedule ID")
		return response.Error(c,
			http.StatusBadRequest,
			"Invalid Request",
			[]string{"Please provide a valid schedule ID"},
		)
	}
 
	// Get auth context
	auth, err := h.auth.GetAuthContext(c)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get auth context")
		return response.System(c)
	}
 
	// Get schedule
	schedule, err := h.repos.Schedule.GetByID(c.Request().Context(), scheduleID)
	if err != nil {
		logger.Error().
			Err(err).
			Int("schedule_id", scheduleID).
			Msg("failed to retrieve schedule")
		return response.System(c)
	}
 
	if schedule == nil {
		logger.Debug().
			Int("schedule_id", scheduleID).
			Msg("schedule not found")
		return response.Error(c,
			http.StatusNotFound,
			"Not Found",
			[]string{"The requested schedule does not exist"},
		)
	}
 
	// Check authorization
	if !canModifySchedule(auth, schedule) {
		logger.Warn().
			Int("schedule_id", scheduleID).
			Int("user_id", auth.UserID).
			Str("role", string(auth.Role)).
			Msg("unauthorized schedule modification attempt")
		return response.Error(c,
			http.StatusForbidden,
			"Access Denied",
			[]string{"You don't have permission to modify this schedule"},
		)
	}
 
	logger.Debug().
		Int("schedule_id", scheduleID).
		Int("user_id", schedule.UserID).
		Msg("rendering schedule update form")
 
	return render(c, component.UpdateScheduleForm(*schedule))
 }
 
 // Helper functions
 func getScheduleID(c echo.Context) (int, error) {
	return strconv.Atoi(c.Param("id"))
 }
 
 func canModifySchedule(auth *dto.AuthContext, schedule *entity.Schedule) bool {
	switch auth.Role {
	case types.UserRoleSuper:
		return true
	case types.UserRoleAdmin:
		return auth.FacilityID == schedule.FacilityID
	default:
		return auth.UserID == schedule.UserID
	}
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
	auth, err := h.auth.GetAuthContext(c)
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
	auth, err := h.auth.GetAuthContext(c)
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
	auth, err := h.auth.GetAuthContext(c)
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
