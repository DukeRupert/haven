package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/DukeRupert/haven/db"
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
	return render(c, super.ShowFacilities("facilities", facs))
}

// Create handles POST /super/facilities
func (h *SuperHandler) Create(c echo.Context) error {
	logger := h.logger

	// Log request details
	logger.Debug().
		Str("content_type", c.Request().Header.Get("Content-Type")).
		Str("method", c.Request().Method).
		Interface("headers", c.Request().Header).
		Msg("received request")

	// Parse form data
	if err := c.Request().ParseForm(); err != nil {
		logger.Error().
			Err(err).
			Msg("failed to parse form data")
	}

	formData := c.Request().Form
	logger.Debug().
		Interface("raw_form_data", formData).
		Interface("post_form", c.Request().PostForm).
		Msg("received form data")

	var params db.CreateFacilityParams
	if err := c.Bind(&params); err != nil {
		logger.Error().
			Err(err).
			Interface("raw_form_data", formData).
			Interface("post_form", c.Request().PostForm).
			Type("params_type", params).
			Msg("failed to bind request payload")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	logger.Debug().
		Str("name", params.Name).
		Str("code", params.Code).
		Interface("bound_params", params).
		Msg("bound parameters")

	// Validate input with more context
	if params.Name == "" {
		logger.Error().
			Interface("raw_form_data", formData).
			Interface("post_form", c.Request().PostForm).
			Interface("bound_params", params).
			Msg("facility name is required but was empty after binding")
		return echo.NewHTTPError(http.StatusBadRequest, "Name is required")
	}
	if params.Code == "" {
		logger.Error().
			Interface("raw_form_data", formData).
			Interface("post_form", c.Request().PostForm).
			Interface("bound_params", params).
			Msg("facility code is required but was empty after binding")
		return echo.NewHTTPError(http.StatusBadRequest, "Code is required")
	}

	facility, err := h.db.CreateFacility(c.Request().Context(), params)
	if err != nil {
		logger.Error().
			Err(err).
			Interface("params", params).
			Msg("failed to create facility in database")
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create facility")
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
