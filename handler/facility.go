package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/DukeRupert/haven/db"
	"github.com/DukeRupert/haven/validation"
	"github.com/DukeRupert/haven/view/alert"
	"github.com/DukeRupert/haven/view/page"
	"github.com/DukeRupert/haven/view/super"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

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