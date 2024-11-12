package handler

import (
	"net/http"
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

func (h *SuperHandler) GetFacilities(c echo.Context) error {
	facs, err := h.db.ListFacilities(c.Request().Context())
	if err != nil {
        // You might want to implement a custom error handler
        return echo.NewHTTPError(http.StatusInternalServerError, 
            "Failed to retrieve facilities")
    }
	return render(c, super.ShowFacilities(facs))
}
