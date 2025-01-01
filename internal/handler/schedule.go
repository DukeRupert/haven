package handler

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/DukeRupert/haven/internal/middleware"
	"github.com/DukeRupert/haven/internal/model/dto"
	"github.com/DukeRupert/haven/internal/model/entity"
	"github.com/DukeRupert/haven/internal/model/params"
	"github.com/DukeRupert/haven/internal/model/types"
	"github.com/DukeRupert/haven/internal/repository/schedule"
	"github.com/DukeRupert/haven/internal/response"
	"github.com/DukeRupert/haven/web/view/component"
	"github.com/DukeRupert/haven/web/view/page"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

func (h *Handler) HandleUpdatePublishedThrough(c echo.Context) error {
	logger := h.logger.With().
		Str("handler", "HandleUpdatePublishedThrough").
		Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
		Logger()

	// Get auth context
	auth, err := middleware.GetAuthContext(c)
	if err != nil {
		logger.Error().Msg("missing auth context")
		return response.System(c)
	}

	// Verify admin role
	if auth.Role != types.UserRoleAdmin {
		return response.Error(c,
			http.StatusForbidden,
			"Access Denied",
			[]string{"Only administrators can update schedule publication dates"},
		)
	}

	// Parse date from request body
	var req struct {
		PublishedThrough time.Time `json:"published_through"`
	}
	if err := c.Bind(&req); err != nil {
		return response.Error(c,
			http.StatusBadRequest,
			"Invalid Request",
			[]string{"Invalid date format"},
		)
	}

	// Update publication date
	pub, err := h.repos.Publication.Update(
		c.Request().Context(),
		auth.FacilityID,
		req.PublishedThrough,
	)
	if err != nil {
		logger.Error().
			Err(err).
			Int("facility_id", auth.FacilityID).
			Time("published_through", req.PublishedThrough).
			Msg("failed to update publication date")
		return response.System(c)
	}

	logger.Info().
		Int("facility_id", auth.FacilityID).
		Time("published_through", pub.PublishedThrough).
		Msg("publication date updated successfully")

	return response.Success(c, "Success", "Schedule publication date has been updated")
}

func (h *Handler) HandleCreateSchedule(c echo.Context) error {
	logger := h.logger.With().
		Str("handler", "HandleCreateSchedule").
		Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
		Logger()
	
	// Get auth context
	auth, err := middleware.GetAuthContext(c)
	if err != nil {
		logger.Error().Msg("missing auth context")
		return response.System(c)
	}

	// Get route context
	route, err := middleware.GetRouteContext(c)
	if err != nil {
		logger.Error().Msg("missing route context")
		return response.System(c)
	}

	if err := ensureRouteParams(route); err != nil {
        return err
    }

	// Parse form data
	var formData params.CreateScheduleRequest
	if err := c.Bind(&formData); err != nil {
		return response.Error(c,
			http.StatusBadRequest,
			"Invalid Form Data",
			[]string{"Please check your input and try again"},
		)
	}

	// Parse and validate date
	startDate, err := time.Parse("2006-01-02", formData.StartDate)
	if err != nil {
		return response.Error(c,
			http.StatusBadRequest,
			"Invalid Date",
			[]string{"Please provide a valid start date (YYYY-MM-DD)"},
		)
	}

	createData := &scheduleCreateData{
		FacilityCode:  route.FacilityCode,
		UserInitials:  route.UserInitials,
		FirstWeekday:  time.Weekday(formData.FirstWeekday),
		SecondWeekday: time.Weekday(formData.SecondWeekday),
		StartDate:     startDate,
	}

	// Create schedule
	schedule, err := h.repos.Schedule.Create(c.Request().Context(), createData.toParams())
	if err != nil {
		logger.Error().
			Err(err).
			Str("facility_code", createData.FacilityCode).
			Str("user_initials", createData.UserInitials).
			Msg("failed to create schedule")
		return response.System(c)
	}

	logger.Info().
		Int("schedule_id", schedule.ID).
		Str("facility_code", createData.FacilityCode).
		Str("user_initials", createData.UserInitials).
		Time("start_date", schedule.StartDate).
		Str("first_weekday", schedule.FirstWeekday.String()).
		Str("second_weekday", schedule.SecondWeekday.String()).
		Msg("schedule created successfully")

	return render(c, page.ScheduleCard(*auth, *route, *schedule))
}

// HandleGetSchedule retrieves and displays a schedule
func (h *Handler) HandleGetSchedule(c echo.Context) error {
	logger := h.logger.With().
		Str("handler", "HandleGetSchedule").
		Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
		Logger()
	
	// Get auth context
	auth, err := middleware.GetAuthContext(c)
	if err != nil {
		logger.Error().Msg("missing auth context")
		return response.System(c)
	}

	// Get route context
	route, err := middleware.GetRouteContext(c)
	if err != nil {
		logger.Error().Msg("missing route context")
		return response.System(c)
	}

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

	// Check if user can view this schedule
	if !canViewSchedule(auth, schedule) {
		logger.Warn().
			Int("schedule_id", scheduleID).
			Int("user_id", auth.UserID).
			Str("role", string(auth.Role)).
			Msg("unauthorized schedule view attempt")
		return response.Error(c,
			http.StatusForbidden,
			"Access Denied",
			[]string{"You don't have permission to view this schedule"},
		)
	}

	logger.Debug().
		Int("schedule_id", scheduleID).
		Int("user_id", schedule.UserID).
		Str("viewer_role", string(auth.Role)).
		Msg("rendering schedule card")

	return render(c, page.ScheduleCard(*auth, *route, *schedule))
}

// HandleAvailabilityToggle processes requests to toggle protected date availability
func (h *Handler) HandleAvailabilityToggle(c echo.Context) error {
	logger := h.logger.With().
		Str("handler", "HandleAvailabilityToggle").
		Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
		Logger()

	// Get auth context
	auth, err := middleware.GetAuthContext(c)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get auth context")
		return response.System(c)
	}

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

	// Get protected date with its associated schedule
	protectedDate, err := h.repos.Schedule.GetProtectedDateByID(c.Request().Context(), dateID)
	if err != nil {
		logger.Error().
			Err(err).
			Int("date_id", dateID).
			Msg("failed to fetch protected date")
		return response.System(c)
	}

	// Log attempt details
	logger.Debug().
		Int("date_id", dateID).
		Int("user_id", auth.UserID).
		Int("date_user_id", protectedDate.UserID).
		Int("facility_id", auth.FacilityID).
		Int("date_facility_id", protectedDate.FacilityID).
		Str("role", string(auth.Role)).
		Msg("attempting to toggle availability")

	// Verify authorization
	if !canToggleAvailability(auth, &protectedDate) {
		logger.Warn().
			Int("date_id", dateID).
			Int("requesting_user_id", auth.UserID).
			Int("date_user_id", protectedDate.UserID).
			Int("requesting_facility_id", auth.FacilityID).
			Int("date_facility_id", protectedDate.FacilityID).
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
		if errors.Is(err, schedule.ErrSchedulePublished) {
			return response.Error(c,
				http.StatusBadRequest,
				"Schedule Published",
				[]string{"Cannot modify availability for dates in published schedule"},
			)
		}
		logger.Error().
			Err(err).
			Int("date_id", dateID).
			Msg("failed to toggle availability")
		return response.System(c)
	}

	logger.Debug().
		Int("date_id", dateID).
		Int("user_id", auth.UserID).
		Int("date_user_id", updatedDate.UserID).
		Bool("is_available", updatedDate.Available).
		Msg("availability toggled successfully")

	return render(c, component.ProtectedDay(
		auth.UserID,
		updatedDate,
	))
}

// Types for validation
type scheduleRequestContext struct {
	FacilityCode string
	UserInitials string
	Facility     *entity.Facility
	User         *entity.User
	Auth         *dto.AuthContext
}

// GetCreateScheduleForm renders the schedule creation form
func (h *Handler) GetCreateScheduleForm(c echo.Context) error {
	logger := h.logger.With().
		Str("handler", "GetCreateScheduleForm").
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
	
	// Get route context
	route, err := middleware.GetRouteContext(c)
	if err != nil {
		logger.Error().Msg("missing route context")
		return response.System(c)
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

	var userInitials string
    // If we're on a user-specific route, use those initials
    if route.UserInitials != "" {
        userInitials = route.UserInitials
    } else {
        userInitials = auth.Initials
    }

	logger.Debug().
		Str("facility_code",facilityCode).
		Str("user_initials", userInitials).
		Msg("rendering schedule creation form")

	return render(c, component.CreateScheduleForm(
		facilityCode,
		userInitials,
	))
}

// validateScheduleRequest handles all validation steps
func (h *Handler) validateScheduleRequest(c echo.Context) (*scheduleRequestContext, error) {
	// Get and validate basic params
	params, err := validateScheduleFormParams(c)
	if err != nil {
		return nil, err
	}

	// Get auth context
	auth, err := middleware.GetAuthContext(c)
	if err != nil {
		return nil, response.System(c)
	}

	reqCtx := &scheduleRequestContext{
		FacilityCode: params.FacilityCode,
		UserInitials: params.Initials,
		Auth:         auth,
	}

	// Verify facility and permissions
	if err := h.validateFacilityAccess(c, reqCtx); err != nil {
		return nil, err
	}

	// Verify user exists in facility
	if err := h.validateUserInFacility(c, reqCtx); err != nil {
		return nil, err
	}

	return reqCtx, nil
}

func (h *Handler) validateFacilityAccess(c echo.Context, ctx *scheduleRequestContext) error {
	facility, err := h.repos.Facility.GetByCode(c.Request().Context(), ctx.FacilityCode)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("facility_code", ctx.FacilityCode).
			Msg("facility not found")
		return response.Error(c,
			http.StatusNotFound,
			"Invalid Facility",
			[]string{"The specified facility does not exist"},
		)
	}

	// Verify admin has access to this facility
	if ctx.Auth.Role == types.UserRoleAdmin && ctx.Auth.FacilityID != facility.ID {
		h.logger.Warn().
			Int("auth_facility_id", ctx.Auth.FacilityID).
			Int("target_facility_id", facility.ID).
			Msg("admin attempted to access different facility")
		return response.Error(c,
			http.StatusForbidden,
			"Access Denied",
			[]string{"You don't have permission to access this facility"},
		)
	}

	ctx.Facility = facility
	return nil
}

func (h *Handler) validateUserInFacility(c echo.Context, ctx *scheduleRequestContext) error {
	user, err := h.repos.User.GetByInitialsAndFacility(
		c.Request().Context(),
		ctx.UserInitials,
		ctx.Facility.Code,
	)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("initials", ctx.UserInitials).
			Int("facility_id", ctx.Facility.ID).
			Msg("user not found")
		return response.Error(c,
			http.StatusNotFound,
			"Invalid User",
			[]string{"The specified user does not exist in this facility"},
		)
	}

	ctx.User = user
	return nil
}

type scheduleUpdateContext struct {
	ScheduleID int
	Schedule   *entity.Schedule
	User       *entity.User
	Facility   *entity.Facility
	Auth       *dto.AuthContext
}

// GetUpdateScheduleForm renders the schedule update form
func (h *Handler) GetUpdateScheduleForm(c echo.Context) error {
	logger := h.logger.With().
		Str("handler", "GetUpdateScheduleForm").
		Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
		Logger()

	logger.Debug().Msg("GetUpdateScheduleForm()")

	// Validate request and get context
	reqCtx, err := h.validateScheduleUpdateRequest(c)
	if err != nil {
		return err
	}

	logger.Debug().
		Int("schedule_id", reqCtx.ScheduleID).
		Int("user_id", reqCtx.Schedule.UserID).
		Int("facility_id", reqCtx.Facility.ID).
		Msg("rendering schedule update form")

	return render(c, component.UpdateScheduleForm(*reqCtx.Schedule, reqCtx.Facility.Code))
}

func (h *Handler) validateScheduleUpdateRequest(c echo.Context) (*scheduleUpdateContext, error) {
	// Get schedule ID
	scheduleID, err := getScheduleID(c)
	if err != nil {
		h.logger.Debug().
			Err(err).
			Str("schedule_id_param", c.Param("id")).
			Msg("invalid schedule ID")
		return nil, response.Error(c,
			http.StatusBadRequest,
			"Invalid Request",
			[]string{"Please provide a valid schedule ID"},
		)
	}

	// Get auth context
	auth, err := middleware.GetAuthContext(c)
	if err != nil {
		return nil, response.System(c)
	}

	reqCtx := &scheduleUpdateContext{
		ScheduleID: scheduleID,
		Auth:       auth,
	}

	// Get and validate schedule
	if err := h.validateScheduleAccess(c, reqCtx); err != nil {
		return nil, err
	}

	// Get and validate user and facility
	if err := h.validateScheduleUserAndFacility(c, reqCtx); err != nil {
		return nil, err
	}

	return reqCtx, nil
}

func (h *Handler) validateScheduleAccess(c echo.Context, ctx *scheduleUpdateContext) error {
	schedule, err := h.repos.Schedule.GetByID(c.Request().Context(), ctx.ScheduleID)
	if err != nil {
		h.logger.Error().
			Err(err).
			Int("schedule_id", ctx.ScheduleID).
			Msg("failed to retrieve schedule")
		return response.System(c)
	}

	if schedule == nil {
		h.logger.Debug().
			Int("schedule_id", ctx.ScheduleID).
			Msg("schedule not found")
		return response.Error(c,
			http.StatusNotFound,
			"Not Found",
			[]string{"The requested schedule does not exist"},
		)
	}

	ctx.Schedule = schedule
	return nil
}

func (h *Handler) validateScheduleUserAndFacility(c echo.Context, ctx *scheduleUpdateContext) error {
	// Get user associated with schedule
	user, err := h.repos.User.GetByID(c.Request().Context(), ctx.Schedule.UserID)
	if err != nil {
		h.logger.Error().
			Err(err).
			Int("user_id", ctx.Schedule.UserID).
			Msg("failed to retrieve user")
		return response.System(c)
	}
	ctx.User = user

	// Get facility
	facility, err := h.repos.Facility.GetByID(c.Request().Context(), user.FacilityID)
	if err != nil {
		h.logger.Error().
			Err(err).
			Int("facility_id", user.FacilityID).
			Msg("failed to retrieve facility")
		return response.System(c)
	}
	ctx.Facility = facility

	// Check permissions
	if !canModifySchedule(ctx.Auth, ctx.Schedule) {
		h.logger.Warn().
			Int("schedule_id", ctx.ScheduleID).
			Int("user_id", ctx.Auth.UserID).
			Int("facility_id", facility.ID).
			Str("role", string(ctx.Auth.Role)).
			Msg("unauthorized schedule modification attempt")
		return response.Error(c,
			http.StatusForbidden,
			"Access Denied",
			[]string{"You don't have permission to modify this schedule"},
		)
	}

	return nil
}

// HandleUpdateSchedule processes schedule update requests
func (h *Handler) HandleUpdateSchedule(c echo.Context) error {
	logger := h.logger.With().
		Str("handler", "HandleUpdateSchedule").
		Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
		Logger()

	// Get auth context
	auth, err := middleware.GetAuthContext(c)
	if err != nil {
		logger.Error().Msg("missing auth context")
		return response.System(c)
	}

	// Get route context
	route, err := middleware.GetRouteContext(c)
	if err != nil {
		logger.Error().Msg("missing route context")
		return response.System(c)
	}

	// Validate input and permissions
	updateData, auth, err := h.validateScheduleUpdate(c)
	if err != nil {
		return err // validateScheduleUpdate handles error responses
	}

	// Update schedule
	schedule, err := h.repos.Schedule.Update(
		c.Request().Context(),
		updateData.ScheduleID,
		updateData.toParams(),
	)
	if err != nil {
		logger.Error().
			Err(err).
			Int("schedule_id", updateData.ScheduleID).
			Msg("failed to update schedule")
		return response.System(c)
	}

	logger.Info().
		Int("schedule_id", schedule.ID).
		Int("user_id", schedule.UserID).
		Time("start_date", schedule.StartDate).
		Str("first_weekday", schedule.FirstWeekday.String()).
		Str("second_weekday", schedule.SecondWeekday.String()).
		Msg("schedule updated successfully")

	return render(c, page.ScheduleCard(*auth, *route, *schedule))
}

// General helpers
func getScheduleID(c echo.Context) (int, error) {
	return strconv.Atoi(c.Param("schedule_id"))
}

// HandleCreateSchedule helpers
type scheduleCreateData struct {
	FacilityCode  string
	UserInitials  string
	FirstWeekday  time.Weekday
	SecondWeekday time.Weekday
	StartDate     time.Time
}

func (data *scheduleCreateData) toParams() params.CreateScheduleByCodeParams {
	return params.CreateScheduleByCodeParams{
		FacilityCode:  data.FacilityCode,
		UserInitials:  data.UserInitials,
		FirstWeekday:  data.FirstWeekday,
		SecondWeekday: data.SecondWeekday,
		StartDate:     data.StartDate,
	}
}

func (h *Handler) validateCreateSchedule(c echo.Context) (*scheduleCreateData, error) {
	// Validate path parameters
	facility := c.Param("facility")
	if facility == "" {
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

	// Parse form data
	var formData params.CreateScheduleRequest
	if err := c.Bind(&formData); err != nil {
		return nil, response.Error(c,
			http.StatusBadRequest,
			"Invalid Form Data",
			[]string{"Please check your input and try again"},
		)
	}

	// Parse and validate date
	startDate, err := time.Parse("2006-01-02", formData.StartDate)
	if err != nil {
		return nil, response.Error(c,
			http.StatusBadRequest,
			"Invalid Date",
			[]string{"Please provide a valid start date (YYYY-MM-DD)"},
		)
	}

	return &scheduleCreateData{
		FacilityCode:  facility,
		UserInitials:  initials,
		FirstWeekday:  time.Weekday(formData.FirstWeekday),
		SecondWeekday: time.Weekday(formData.SecondWeekday),
		StartDate:     startDate,
	}, nil
}

func canCreateSchedule(auth *dto.AuthContext, facilityCode string) bool {
	switch auth.Role {
	case types.UserRoleSuper:
		return true
	case types.UserRoleAdmin:
		return auth.FacilityCode == facilityCode
	default:
		return false
	}
}

// HandleAvailabilityToggle helpers
func getProtectedDateID(c echo.Context) (int, error) {
	return strconv.Atoi(c.Param("id"))
}

// Authorization helper function
func canToggleAvailability(auth *dto.AuthContext, protectedDate *entity.ProtectedDate) bool {
	// Super admins can toggle any date
	if auth.Role == types.UserRoleSuper {
		return true
	}

	// Facility admins can toggle dates within their facility
	if auth.Role == types.UserRoleAdmin && auth.FacilityID == protectedDate.FacilityID {
		return true
	}

	// Regular users can only toggle their own dates
	if auth.Role == types.UserRoleUser {
		return auth.UserID == protectedDate.UserID && auth.FacilityID == protectedDate.FacilityID
	}

	return false
}

// GetUpdateScheduleForm helpers
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

// HandleUpdateSchedule helpers
type scheduleUpdateData struct {
	ScheduleID    int
	FirstWeekday  time.Weekday
	SecondWeekday time.Weekday
	StartDate     time.Time
}

func (data *scheduleUpdateData) toParams() params.UpdateScheduleParams {
	return params.UpdateScheduleParams{
		FirstWeekday:  data.FirstWeekday,
		SecondWeekday: data.SecondWeekday,
		StartDate:     data.StartDate,
	}
}

func (h *Handler) validateScheduleUpdate(c echo.Context) (*scheduleUpdateData, *dto.AuthContext, error) {
	logger := h.logger.With().Str("method", "validateScheduleUpdate").Logger()

	// Get schedule ID
	scheduleID, err := getScheduleID(c)
	if err != nil {
		return nil, nil, response.Error(c,
			http.StatusBadRequest,
			"Invalid Request",
			[]string{"Please provide a valid schedule ID"},
		)
	}

	// Get auth context
	auth, err := middleware.GetAuthContext(c)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get auth context")
		return nil, nil, response.System(c)
	}

	// Get existing schedule
	schedule, err := h.repos.Schedule.GetByID(c.Request().Context(), scheduleID)
	if err != nil {
		logger.Error().
			Err(err).
			Int("schedule_id", scheduleID).
			Msg("failed to fetch schedule")
		return nil, nil, response.System(c)
	}

	// Check authorization
	if !canModifySchedule(auth, schedule) {
		logger.Warn().
			Int("schedule_id", scheduleID).
			Int("user_id", auth.UserID).
			Str("role", string(auth.Role)).
			Msg("unauthorized schedule modification attempt")
		return nil, nil, response.Error(c,
			http.StatusForbidden,
			"Access Denied",
			[]string{"You don't have permission to modify this schedule"},
		)
	}

	// Parse form data
	var formData struct {
		FirstWeekday  int    `form:"first_weekday"`
		SecondWeekday int    `form:"second_weekday"`
		StartDate     string `form:"start_date"`
	}

	if err := c.Bind(&formData); err != nil {
		return nil, nil, response.Error(c,
			http.StatusBadRequest,
			"Invalid Form Data",
			[]string{"Please check your input and try again"},
		)
	}

	// Parse and validate date
	startDate, err := time.Parse("2006-01-02", formData.StartDate)
	if err != nil {
		return nil, nil, response.Error(c,
			http.StatusBadRequest,
			"Invalid Date",
			[]string{"Please provide a valid start date (YYYY-MM-DD)"},
		)
	}

	// Validate weekdays
	if err := validateWeekdays(formData.FirstWeekday, formData.SecondWeekday); err != nil {
		return nil, nil, response.Error(c,
			http.StatusBadRequest,
			"Invalid Weekdays",
			[]string{err.Error()},
		)
	}

	return &scheduleUpdateData{
		ScheduleID:    scheduleID,
		FirstWeekday:  time.Weekday(formData.FirstWeekday),
		SecondWeekday: time.Weekday(formData.SecondWeekday),
		StartDate:     startDate,
	}, auth, nil
}

func validateWeekdays(first, second int) error {
	if first < 0 || first > 6 || second < 0 || second > 6 {
		return fmt.Errorf("weekdays must be between 0 (Sunday) and 6 (Saturday)")
	}
	if first == second {
		return fmt.Errorf("weekdays must be different")
	}
	return nil
}

func canViewSchedule(auth *dto.AuthContext, schedule *entity.Schedule) bool {
	switch auth.Role {
	case types.UserRoleSuper:
		return true
	case types.UserRoleAdmin:
		return auth.FacilityID == schedule.FacilityID
	default:
		return auth.UserID == schedule.UserID
	}
}

func canModifySchedule(auth *dto.AuthContext, schedule *entity.Schedule) bool {
	logger := zerolog.New(os.Stdout).With().
		Str("method", "canModifySchedule").
		Int("auth_user_id", auth.UserID).
		Int("auth_facility_id", auth.FacilityID).
		Str("auth_role", string(auth.Role)).
		Int("schedule_user_id", schedule.UserID).
		Int("schedule_facility_id", schedule.FacilityID).
		Logger()

	switch auth.Role {
	case types.UserRoleSuper:
		logger.Debug().Msg("super user access granted")
		return true

	case types.UserRoleAdmin:
		canModify := auth.FacilityID == schedule.FacilityID
		if canModify {
			logger.Debug().Msg("admin access granted - same facility")
		} else {
			logger.Debug().Msg("admin access denied - different facility")
		}
		return canModify

	default:
		canModify := auth.UserID == schedule.UserID
		if canModify {
			logger.Debug().Msg("user access granted - own schedule")
		} else {
			logger.Debug().Msg("user access denied - not own schedule")
		}
		return canModify
	}
}
