package handler

import (

    "github.com/DukeRupert/haven/types"
	"github.com/labstack/echo/v4"
)

// middleware/route_context.go
func SetRouteContext(h *Handler) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            h.RouteCtx = types.RouteContext{
                FacilityCode: c.Param("code"),
                UserInitials: c.Param("initials"),
                BasePath:    "/app",
            }
            return next(c)
        }
    }
}