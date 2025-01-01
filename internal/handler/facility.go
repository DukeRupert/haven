package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/DukeRupert/haven/internal/middleware"
	"github.com/DukeRupert/haven/internal/model/dto"
	"github.com/DukeRupert/haven/internal/model/params"
	"github.com/DukeRupert/haven/internal/response"
	"github.com/DukeRupert/haven/internal/validation"
	"github.com/DukeRupert/haven/web/view/alert"
	"github.com/DukeRupert/haven/web/view/page"
	"github.com/DukeRupert/haven/web/view/super"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

func (h *Handler) HandleGetFacilities(c echo.Context) error {
	logger := h.logger.With().
		Str("handler", "HandleFacilities").
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

	route, err := middleware.GetRouteContext(c)
	if err != nil {
		logger.Error().Msg("missing route context")
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"Missing route context",
		)
	}

	// Get facilities from repository
	facilities, err := h.repos.Facility.List(c.Request().Context())
	if err != nil {
		logger.Error().Err(err).Msg("failed to retrieve facilities")
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"Unable to load facilities. Please try again later.",
		)
	}

	// Build nav items
	navItems := BuildNav(route, auth, c.Request().URL.Path)

	// Build props
	props := dto.FacilityPageProps {
		Title:       "Facilities",
		Description: "A list of all facilities including their name and code.",
		NavItems: navItems,
		AuthCtx: *auth,
		RouteCtx: *route,
		Facilities: facilities,
	}

	logger.Debug().
		Int("facility_count", len(facilities)).
		Msg("rendering facilities page")

	// Render the page
	return page.Facilities(
		props,
	).Render(c.Request().Context(), c.Response().Writer)
}

func (h *Handler) HandleCreateFacility(c echo.Context) error {
	return render(c, page.CreateFacilityForm())
}

func (h *Handler) HandleUpdateFacility(c echo.Context) error {
	logger := zerolog.Ctx(c.Request().Context())
	id, err := strconv.Atoi(c.Param("fid"))
	if err != nil {
		logger.Error().
			Err(err).
			Str("facility_id", c.Param("fid")).
			Msg("invalid facility ID format")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid facility ID")
	}

	var params params.UpdateFacilityParams
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

	facility, err := h.repos.Facility.Update(c.Request().Context(), id, params)
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

func (h *Handler) HandleDeleteFacility(c echo.Context) error {
	return render(c, page.CreateFacilityForm())
}

// GET /app/facilities/create
func (h *Handler) GetCreateFacilityForm(c echo.Context) error {
	return render(c, page.CreateFacilityForm())
}

// Get /app/facilities/update
func (h *Handler) GetUpdateFacilityForm(c echo.Context) error {
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

	facility, err := h.repos.Facility.GetByID(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Error().
				Int("facility_id", id).
				Msg("facility not found")
			return response.Error(c, http.StatusInternalServerError, "Not found",
				[]string{"The requested facility does not exist"})
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
