// internal/response/response.go
package response

import (
    "net/http"
    "github.com/DukeRupert/haven/web/view/alert"
    "github.com/labstack/echo/v4"
)

// Error handles returning error alerts via HTMX
func Error(c echo.Context, status int, heading string, messages []string) error {
    c.Response().Status = status
    c.Response().Header().Set("HX-Reswap", "none")
    return alert.Error(heading, messages).Render(c.Request().Context(), c.Response().Writer)
}

// Redirect handles HTMX redirects
func Redirect(c echo.Context, redirectURL string) error {
    c.Response().Header().Set("HX-Redirect", redirectURL)
    return c.String(http.StatusOK, "")
}

// System returns a generic system error response
func System(c echo.Context) error {
    return Error(c,
        http.StatusInternalServerError,
        "System Error",
        []string{"An unexpected error occurred. Please try again."},
    )
}

// Unauthorized returns an unauthorized error response
func Unauthorized(c echo.Context) error {
    return Error(c,
        http.StatusUnauthorized,
        "Unauthorized",
        []string{"You are not authorized to perform this action."},
    )
}

// Validation returns a validation error response
func Validation(c echo.Context, messages []string) error {
    return Error(c,
        http.StatusUnprocessableEntity,
        "Validation Error",
        messages,
    )
}

// Success returns a success response
func Success(c echo.Context, heading string, message string) error {
    c.Response().Status = http.StatusOK
    return alert.Success(heading, message).Render(c.Request().Context(), c.Response().Writer)
}