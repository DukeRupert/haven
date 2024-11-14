package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/DukeRupert/haven/db"
	"github.com/DukeRupert/haven/view/alert"
	"github.com/DukeRupert/haven/view/super"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

type SuperHandler struct {
	db     *db.DB
	logger zerolog.Logger
}

// NewSuperHandler creates a new handler with both pool and store
func NewSuperHandler(db *db.DB, logger zerolog.Logger) *SuperHandler {
	return &SuperHandler{
		db:     db,
		logger: logger.With().Str("component", "superHandler").Logger(),
	}
}

// GET /super/facilities
func (h *SuperHandler) GetFacilities(c echo.Context) error {
	facs, err := h.db.ListFacilities(c.Request().Context())
	if err != nil {
		// You might want to implement a custom error handler
		return echo.NewHTTPError(http.StatusInternalServerError,
			"Failed to retrieve facilities")
	}
	return render(c, super.ShowFacilities("facilities", "Add a description here", facs))
}

// Create handles POST /super/facilities
func (h *SuperHandler) Create(c echo.Context) error {
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

// Update handles PUT /super/facilities/:id
func (h *SuperHandler) Update(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid facility ID")
	}

	var params db.UpdateFacilityParams
	if err := c.Bind(&params); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	// Validate input
	if params.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Name is required")
	}
	if params.Code == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Code is required")
	}

	facility, err := h.db.UpdateFacility(c.Request().Context(), id, params)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update facility")
	}

	return c.JSON(http.StatusOK, facility)
}

// GET /super/facilities/create
func (h *SuperHandler) CreateFacilityForm(c echo.Context) error {
	fmt.Println("CreateFacilityForm()")
	return render(c, super.CreateFacilityForm())
}
