package handler

import (
    "fmt"

    "github.com/DukeRupert/haven/db"
    "github.com/DukeRupert/haven/types"
	"github.com/labstack/echo/v4"
)

type Route struct {
    Path        string
    Name        string
    Icon        string
    MinRole     db.UserRole    // Minimum role required
    NeedsFacility bool  // Whether route requires facility context
}

var AppRoutes = []Route{
    {
        Path: "/dashboard",
        Name: "Dashboard",
        Icon: "dashboard",
        MinRole: db.UserRole("user"),
        NeedsFacility: true,
    },
    {
        Path: "/calendar",
        Name: "Calendar",
        Icon: "calendar",
        MinRole: db.UserRole("user"),
        NeedsFacility: true,
    },
    {
        Path: "/controllers",
        Name: "Controllers",
        Icon: "settings",
        MinRole: db.UserRole("admin"),
        NeedsFacility: true,
    },
    {
        Path: "/profile",
        Name: "Profile",
        Icon: "user",
        MinRole: db.UserRole("user"),
        NeedsFacility: false,
    },
}

func GetRouteContext(c echo.Context) (*types.RouteContext, error) {
    routeCtx, ok := c.Get("routeCtx").(*types.RouteContext)
    if !ok {
        return nil, fmt.Errorf("route context not found")
    }
    return routeCtx, nil
}

func BuildRoutePath(routeCtx *types.RouteContext, path string) string {
    if routeCtx.BasePath == "" {
        return path
    }
    return fmt.Sprintf("/%s%s", routeCtx.BasePath, path)
}

// middleware/route_context.go
// func SetRouteContext(h *Handler) echo.MiddlewareFunc {
//     return func(next echo.HandlerFunc) echo.HandlerFunc {
//         return func(c echo.Context) error {
//             h.RouteCtx = types.RouteContext{
//                 FacilityCode: c.Param("code"),
//                 UserInitials: c.Param("initials"),
//                 BasePath:    "/app",
//             }
//             return next(c)
//         }
//     }
// }