package handler

import (
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

func render(c echo.Context, component templ.Component) error {
	c.Response().Header().Set("Content-Type", "text/html")
	return component.Render(c.Request().Context(), c.Response())
}
