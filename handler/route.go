package handler

import (
    "fmt"

    "github.com/DukeRupert/haven/types"
	"github.com/labstack/echo/v4"
)

var AppRoutes = []types.Route{
    {
        Path: "/dashboard",
        Name: "Dashboard",
        Icon: "dashboard",
        MinRole: types.UserRole("user"),
        NeedsFacility: true,
    },
    {
        Path: "/calendar",
        Name: "Calendar",
        Icon: "calendar",
        MinRole: types.UserRole("user"),
        NeedsFacility: true,
    },
    {
        Path: "/controllers",
        Name: "Controllers",
        Icon: "settings",
        MinRole: types.UserRole("admin"),
        NeedsFacility: true,
    },
    {
        Path: "/profile",
        Name: "Profile",
        Icon: "user",
        MinRole: types.UserRole("user"),
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