package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/DukeRupert/haven/db"
	"github.com/DukeRupert/haven/view/super"
	"github.com/labstack/echo/v4"
)

type SuperHandler struct {
	db *db.DB
}

// NewSuperHandler creates a new handler with both pool and store
func NewSuperHandler(db *db.DB) *SuperHandler {
	return &SuperHandler{
		db: db,
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
	var params db.CreateFacilityParams
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

	facility, err := h.db.CreateFacility(c.Request().Context(), params)
	if err != nil {
		// Log the specific error
        fmt.Printf("Database error: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create facility")
	}

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
