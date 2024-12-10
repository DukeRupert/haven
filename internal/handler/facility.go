package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/DukeRupert/haven/types"
	"github.com/DukeRupert/haven/validation"
	"github.com/DukeRupert/haven/view/alert"
	"github.com/DukeRupert/haven/view/page"
	"github.com/DukeRupert/haven/view/super"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

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

	var params types.UpdateFacilityParams
	if err := c.Bind(&params); err != nil {
		logger.Error().
			Err(err).
			Msg("failed to bind request payload")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	// Collect validation errors
	var errors []string

	// Validate facility name
	name, err := validation.ValidateFacilityName(params.Name)
	if err != nil {
		errors = append(errors, err.Error())
	} else {
		params.Name = string(name)
	}

	// Validate facility code
	code, err := validation.ValidateFacilityCode(params.Code)
	if err != nil {
		errors = append(errors, err.Error())
	} else {
		params.Code = string(code)
	}

	if len(errors) > 0 {
		logger.Error().
			Strs("validation_errors", errors).
			Interface("params", params).
			Msg("validation failed")
		
		// Join all errors with semicolons for better readability
		return echo.NewHTTPError(http.StatusBadRequest, strings.Join(errors, "; "))
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