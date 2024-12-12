package handler

import (
    "fmt"

    "github.com/DukeRupert/haven/internal/model/types"
    "github.com/DukeRupert/haven/internal/model/dto"
	"github.com/labstack/echo/v4"
)

var AppRoutes = []dto.Route{
    {
        Path: "/dashboard",
        Name: "Dashboard",
        Icon: "dashboard",
        MinRole: types.UserRoleUser,
        NeedsFacility: true,
    },
    {
        Path: "/calendar",
        Name: "Calendar",
        Icon: "calendar",
        MinRole: types.UserRoleUser,
        NeedsFacility: true,
    },
    {
        Path: "/controllers",
        Name: "Controllers",
        Icon: "settings",
        MinRole: types.UserRoleAdmin,
        NeedsFacility: true,
    },
    {
        Path: "/profile",
        Name: "Profile",
        Icon: "user",
        MinRole: types.UserRoleUser,
        NeedsFacility: false,
    },
}

func GetRouteContext(c echo.Context) (*dto.RouteContext, error) {
    routeCtx, ok := c.Get("routeCtx").(*dto.RouteContext)
    if !ok {
        return nil, fmt.Errorf("route context not found")
    }
    return routeCtx, nil
}

func BuildRoutePath(routeCtx *dto.RouteContext, path string) string {
    if routeCtx.BasePath == "" {
        return path
    }
    return fmt.Sprintf("/%s%s", routeCtx.BasePath, path)
}