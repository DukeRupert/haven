package handler

import (
	"fmt"
    "net/http"

	"github.com/DukeRupert/haven/view/alert"
	
    "github.com/labstack/echo/v4"
)

// ErrorResponse wraps error responses with proper status codes and HTMX handling
func (h *Handler) ErrorResponse(c echo.Context, status int, heading string, messages []string) error {
    // Set the status code
    c.Response().Status = status
    
    // Check if it's an HTMX request
    if c.Request().Header.Get("HX-Request") == "true" {
        // For HTMX requests, return just the alert component
        return render(c, alert.Error(heading, messages))
    }
    
    // For non-HTMX requests, you might want to render a full error page
    // or redirect with a flash message
    // TODO: Implement this based on your needs
    return c.JSON(status, map[string]interface{}{
        "error": heading,
        "messages": messages,
    })
}

// ValidationError handles validation errors specifically
func (h *Handler) ValidationError(c echo.Context, errors []string) error {
    heading := "There was 1 error with your submission"
    if len(errors) > 1 {
        heading = fmt.Sprintf("There were %d errors with your submission", len(errors))
    }
    return h.ErrorResponse(c, http.StatusUnprocessableEntity, heading, errors)
}

// SystemError handles internal server errors
func (h *Handler) SystemError(c echo.Context, err error, message string) error {
    h.logger.Error().Err(err).Msg(message)
    return h.ErrorResponse(c, http.StatusInternalServerError, "System Error", 
        []string{"An unexpected error occurred. Please try again later."})
}