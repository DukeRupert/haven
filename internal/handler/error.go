package handler

import (
	"net/http"

	"github.com/DukeRupert/haven/web/view/alert"

	"github.com/labstack/echo/v4"
)

// ErrorResponse handles returning error alerts via HTMX
func ErrorResponse(c echo.Context, status int, heading string, messages []string) error {
	c.Response().Status = status
	c.Response().Header().Set("HX-Reswap", "none")
	return alert.Error(heading, messages).Render(c.Request().Context(), c.Response().Writer)
}

// RedirectResponse handles HTMX redirects
func RedirectResponse(c echo.Context, redirectURL string) error {
	c.Response().Header().Set("HX-Redirect", redirectURL)
	return c.String(http.StatusOK, "")
}

// Common error helpers
func SystemError(c echo.Context) error {
	return ErrorResponse(c,
		http.StatusInternalServerError,
		"System Error",
		[]string{"An unexpected error occurred. Please try again."},
	)
}

func UnauthorizedError(c echo.Context) error {
	return ErrorResponse(c,
		http.StatusUnauthorized,
		"Unauthorized",
		[]string{"You are not authorized to perform this action."},
	)
}

func ValidationError(c echo.Context, messages []string) error {
	return ErrorResponse(c,
		http.StatusUnprocessableEntity,
		"Validation Error",
		messages,
	)
}

func SuccessResponse(c echo.Context, heading string, message string) error {
	c.Response().Status = http.StatusOK
	return alert.Success(heading, message).Render(c.Request().Context(), c.Response().Writer)
}
