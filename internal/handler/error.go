// handler/error.go
package handler

import (
	"net/http"

	"github.com/DukeRupert/haven/web/view/page"

	"github.com/labstack/echo/v4"
)

// Custom error handler for Echo
func CustomHTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	title := "Server Error"
	message := "Something went wrong"
	returnURL := "/app/calendar"
	returnText := "Return Home"

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code

		switch code {
		case http.StatusNotFound:
			title = "Page Not Found"
			message = "The page you're looking for doesn't exist or has been moved."
		case http.StatusForbidden:
			title = "Access Denied"
			message = "You don't have permission to access this resource."
			returnURL = "/login"
			returnText = "Log In"
		case http.StatusUnauthorized:
			title = "Authentication Required"
			message = "Please log in to continue."
			returnURL = "/login"
			returnText = "Log In"
		}

		if m, ok := he.Message.(string); ok {
			message = m
		}
	}

	// Handle API requests
	if isAPIRequest(c) || isHTMXRequest(c) {
		_ = c.JSON(code, map[string]interface{}{
			"error": message,
		})
		return
	}

	// Render error page for normal requests
	_ = render(c, page.ErrorPage(page.ErrorPageParams{
		Title:      title,
		Message:    message,
		StatusCode: code,
		ReturnURL:  returnURL,
		ReturnText: returnText,
	}))
}

// Helper functions to determine request type
func isAPIRequest(c echo.Context) bool {
	return c.Request().Header.Get("Accept") == "application/json" ||
		c.Request().Header.Get("Content-Type") == "application/json"
}

func isHTMXRequest(c echo.Context) bool {
	return c.Request().Header.Get("HX-Request") == "true"
}

