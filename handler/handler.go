package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/DukeRupert/haven/db"
	"github.com/DukeRupert/haven/view/alert"
	"github.com/DukeRupert/haven/view/component"
	"github.com/DukeRupert/haven/view/page"
	"github.com/DukeRupert/haven/view/super"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

type Handler struct {
	db     *db.DB
	logger zerolog.Logger
}

// NewSuperHandler creates a new handler with both pool and store
func NewHandler(db *db.DB, logger zerolog.Logger) *Handler {
	return &Handler{
		db:     db,
		logger: logger.With().Str("component", "superHandler").Logger(),
	}
}

// GET /app/facilities
func (h *Handler) GetFacilities(c echo.Context) error {
	facs, err := h.db.ListFacilities(c.Request().Context())
	if err != nil {
		// You might want to implement a custom error handler
		return echo.NewHTTPError(http.StatusInternalServerError,
			"Failed to retrieve facilities")
	}

	title := "Facilities"
	description := "A list of all facilities including their name and code."
	return render(c, page.Facilities(title, description, facs))
}

// POST /app/facilities
func (h *Handler) PostFacilities(c echo.Context) error {
	logger := zerolog.Ctx(c.Request().Context())

	var params db.CreateFacilityParams
	if err := c.Bind(&params); err != nil {
		logger.Error().
			Err(err).
			Msg("failed to bind request payload")

		return render(c, alert.Error(
			"Invalid request",
			[]string{"The submitted form data was invalid"},
		))
	}

	// Collect validation errors
	var errors []string
	if params.Name == "" {
		errors = append(errors, "Facility name is required")
	}
	if params.Code == "" {
		errors = append(errors, "Facility code is required")
	}
	if len(errors) > 0 {
		logger.Error().
			Strs("validation_errors", errors).
			Msg("validation failed")

		heading := "There was 1 error with your submission"
		if len(errors) > 1 {
			heading = fmt.Sprintf("There were %d errors with your submission", len(errors))
		}

		return render(c, alert.Error(heading, errors))
	}

	facility, err := h.db.CreateFacility(c.Request().Context(), params)
	if err != nil {
		logger.Error().
			Err(err).
			Interface("params", params).
			Msg("failed to create facility in database")
		return render(c, alert.Error("System error",
			[]string{"Failed to create facility. Please try again later"}))
	}

	logger.Info().
		Int("facility_id", facility.ID).
		Str("name", facility.Name).
		Str("code", facility.Code).
		Msg("facility created successfully")

	return render(c, super.FacilityListItem(*facility))
}

// GET /app/facilities/create
func (h *Handler) CreateFacilityForm(c echo.Context) error {
	return render(c, page.CreateFacilityForm())
}

// Get /app/facilities/update
func (h *Handler) UpdateFacilityForm(c echo.Context) error {
	logger := zerolog.Ctx(c.Request().Context())
	id, err := strconv.Atoi(c.Param("fid"))
	if err != nil {
		logger.Error().
			Err(err).
			Str("facility_id", c.Param("fid")).
			Msg("invalid facility ID format")
		return render(c, alert.Error(
			"Invalid request",
			[]string{"Invalid facility ID provided"},
		))
	}

	facility, err := h.db.GetFacilityByID(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Error().
				Int("facility_id", id).
				Msg("facility not found")
			return render(c, alert.Error(
				"Not found",
				[]string{"The requested facility does not exist"},
			))
		}
		logger.Error().
			Err(err).
			Int("facility_id", id).
			Msg("failed to retrieve facility")
		return render(c, alert.Error(
			"System error",
			[]string{"Unable to load facility. Please try again later"},
		))
	}

	return render(c, page.UpdateFacilityForm(*facility))
}

// PUT /app/facilities/:id
func (h *Handler) UpdateFacility(c echo.Context) error {
	logger := zerolog.Ctx(c.Request().Context())
	id, err := strconv.Atoi(c.Param("fid"))
	if err != nil {
		logger.Error().
			Err(err).
			Str("facility_id", c.Param("fid")).
			Msg("invalid facility ID format")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid facility ID")
	}

	var params db.UpdateFacilityParams
	if err := c.Bind(&params); err != nil {
		logger.Error().
			Err(err).
			Msg("failed to bind request payload")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	// Collect validation errors
	var errors []string
	if params.Name == "" {
		errors = append(errors, "Name is required")
	}
	if params.Code == "" {
		errors = append(errors, "Code is required")
	}
	if len(errors) > 0 {
		logger.Error().
			Strs("validation_errors", errors).
			Interface("params", params).
			Msg("validation failed")
		return echo.NewHTTPError(http.StatusBadRequest, strings.Join(errors, ", "))
	}

	facility, err := h.db.UpdateFacility(c.Request().Context(), id, params)
	if err != nil {
		logger.Error().
			Err(err).
			Int("facility_id", id).
			Interface("params", params).
			Msg("failed to update facility in database")
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update facility")
	}

	logger.Info().
		Int("facility_id", id).
		Str("name", facility.Name).
		Str("code", facility.Code).
		Msg("facility updated successfully")

	return render(c, super.FacilityListItem(*facility))
}

func (h *Handler) CreateScheduleForm(c echo.Context) error {
	logger := h.logger.With().
		Str("handler", "CreateScheduleForm").
		Str("method", "GET").
		Str("path", "/:fid/:uid/schedule/new").
		Logger()

	logger.Debug().
		Str("raw_fid", c.Param("fid")).
		Str("raw_uid", c.Param("uid")).
		Str("full_path", c.Path()).
		Str("request_uri", c.Request().RequestURI).
		Interface("params", c.ParamNames()).
		Msg("received schedule form request")

	// Parse and validate path parameters
	fid, err := strconv.Atoi(c.Param("fid"))
	if err != nil {
		logger.Error().
			Err(err).
			Str("fid", c.Param("fid")).
			Msg("invalid facility id parameter")
		return render(c, alert.Error(
			"Invalid request",
			[]string{"Invalid facility ID provided"},
		))
	}

	uid, err := strconv.Atoi(c.Param("uid"))
	if err != nil {
		logger.Error().
			Err(err).
			Str("uid", c.Param("uid")).
			Msg("invalid user id parameter")
		return render(c, alert.Error(
			"Invalid request",
			[]string{"Invalid user ID provided"},
		))
	}

	return render(c, component.CreateScheduleForm(fid, uid))
}

type CreateScheduleRequest struct {
	FirstWeekday  int    `json:"first_weekday" form:"first_weekday"`
	SecondWeekday int    `json:"second_weekday" form:"second_weekday"`
	StartDate     string `json:"start_date" form:"start_date"`
}

func (h *Handler) CreateSchedule(c echo.Context) error {
	logger := h.logger

	// Parse and validate path parameters
	fid, err := strconv.Atoi(c.Param("fid"))
	if err != nil {
		logger.Error().
			Err(err).
			Str("fid", c.Param("fid")).
			Msg("invalid facility id parameter")
		return render(c, alert.Error(
			"Invalid request",
			[]string{"Invalid facility ID provided"},
		))
	}

	uid, err := strconv.Atoi(c.Param("uid"))
	if err != nil {
		logger.Error().
			Err(err).
			Str("uid", c.Param("uid")).
			Msg("invalid user id parameter")
		return render(c, alert.Error(
			"Invalid request",
			[]string{"Invalid user ID provided"},
		))
	}

	// Bind request data
	var req CreateScheduleRequest
	if err := c.Bind(&req); err != nil {
		logger.Error().
			Err(err).
			Msg("failed to bind request payload")

		return render(c, alert.Error(
			"Invalid request",
			[]string{"The submitted form data was invalid"},
		))
	}

	// Parse and validate start date
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		logger.Error().
			Err(err).
			Str("start_date", req.StartDate).
			Msg("invalid start date format")
		return render(c, alert.Error(
			"Invalid date",
			[]string{"Start date must be in YYYY-MM-DD format"},
		))
	}

	// Collect validation errors
	var errors []string

	// Validate weekdays range
	if req.FirstWeekday < 0 || req.FirstWeekday > 6 {
		errors = append(errors, "First weekday must be between 0 (Sunday) and 6 (Saturday)")
	}
	if req.SecondWeekday < 0 || req.SecondWeekday > 6 {
		errors = append(errors, "Second weekday must be between 0 (Sunday) and 6 (Saturday)")
	}

	// Validate weekdays are different
	if req.FirstWeekday == req.SecondWeekday {
		errors = append(errors, "First and second weekdays must be different")
	}

	// Return validation errors if any
	if len(errors) > 0 {
		logger.Error().
			Strs("validation_errors", errors).
			Interface("request", req).
			Msg("schedule validation failed")

		heading := "There was 1 error with your submission"
		if len(errors) > 1 {
			heading = fmt.Sprintf("There were %d errors with your submission", len(errors))
		}

		return render(c, alert.Error(heading, errors))
	}

	// Create schedule params
	params := db.CreateScheduleParams{
		UserID:    uid,
		FirstDay:  time.Weekday(req.FirstWeekday),
		SecondDay: time.Weekday(req.SecondWeekday),
		StartDate: startDate,
	}

	// Create schedule in database
	schedule, err := h.db.CreateSchedule(c.Request().Context(), params)
	if err != nil {
		logger.Error().
			Err(err).
			Interface("params", params).
			Msg("failed to create schedule in database")

		return render(c, alert.Error(
			"System error",
			[]string{"Failed to create schedule. Please try again later"},
		))
	}

	logger.Info().
		Int("schedule_id", schedule.ID).
		Int("facility_id", fid).
		Int("user_id", uid).
		Time("start_date", schedule.StartDate).
		Msgf("schedule created successfully with weekdays %s and %s",
			schedule.FirstWeekday, schedule.SecondWeekday)

	// Return the updated schedule card
	return render(c, component.ScheduleCard(*schedule))
}
