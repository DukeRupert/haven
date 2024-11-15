package handler

import (
	"fmt"
	"strconv"
	"time"
	"net/http"

	"github.com/DukeRupert/haven/db"
	"github.com/DukeRupert/haven/view/alert"
	"github.com/DukeRupert/haven/view/page"
	"github.com/DukeRupert/haven/view/component"

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

func (h *Handler) ShowHome(c echo.Context) error {
		return render(c, page.Landing())
	}

// PlaceholderMessage handles rendering a simple string message
func (h *Handler) PlaceholderMessage(c echo.Context) error {

	// Here you would typically have your component.PlaceholderMessage
	// For this example, we'll return the raw message
	return c.String(http.StatusOK, "Fix me. I need some love.")
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
