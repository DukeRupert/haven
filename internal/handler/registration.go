// handlers/registration.go
package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/DukeRupert/haven/view/alert"
	"github.com/DukeRupert/haven/view/page"
	"github.com/labstack/echo/v4"
)

type RegistrationParams struct {
	FacilityCode string `form:"facility_code"`
	Initials     string `form:"initials"`
	Email        string `form:"email"`
}

// HandleRegistration processes the registration form submission
func (h *Handler) HandleRegistration(c echo.Context) error {
	logger := h.logger.With().
		Str("component", "registration").
		Str("handler", "HandleRegistration").
		Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
		Logger()

	logger.Debug().Msg("Processing registration request")

	// Bind form data
	params := new(RegistrationParams)
	if err := c.Bind(params); err != nil {
		logger.Debug().
			Err(err).
			Interface("params", params).
			Msg("Invalid form data submitted")
		return alert.Error(
			"Invalid Request",
			[]string{"Please check your input and try again."},
		).Render(c.Request().Context(), c.Response().Writer)
	}

	// Validate input format
	if err := validateRegistrationParams(params); err != nil {
		logger.Debug().
			Err(err).
			Interface("params", params).
			Msg("Validation failed")
		return alert.Error(
			"Invalid Input",
			[]string{err.Error()},
		).Render(c.Request().Context(), c.Response().Writer)
	}

	// Add 'K' prefix to facility code
	fullFacilityCode := "K" + strings.ToUpper(params.FacilityCode)

	// Check if facility exists
	facility, err := h.db.GetFacilityByCode(c.Request().Context(), fullFacilityCode)
	if err != nil {
		logger.Error().
			Err(err).
			Str("facility_code", fullFacilityCode).
			Msg("Database error when checking facility")
		return alert.Error(
			"System Error",
			[]string{"Unable to process registration. Please try again."},
		).Render(c.Request().Context(), c.Response().Writer)
	}
	if facility == nil {
		logger.Debug().
			Str("facility_code", fullFacilityCode).
			Msg("Invalid facility code")
		return alert.Error(
			"Invalid Facility",
			[]string{"The facility code you entered is not valid."},
		).Render(c.Request().Context(), c.Response().Writer)
	}

	// Verify user credentials
	user, err := h.db.VerifyUserCredentials(
		c.Request().Context(),
		facility.ID,
		strings.ToUpper(params.Initials),
		strings.ToLower(params.Email),
	)
	if err != nil {
		logger.Error().
			Err(err).
			Str("facility_code", fullFacilityCode).
			Str("email", params.Email).
			Msg("Database error when verifying credentials")
		return alert.Error(
			"System Error",
			[]string{"Unable to verify credentials. Please try again."},
		).Render(c.Request().Context(), c.Response().Writer)
	}
	if user == nil {
		logger.Debug().
			Str("facility_code", fullFacilityCode).
			Str("email", params.Email).
			Msg("No matching user found")
		return alert.Error(
			"Invalid Credentials",
			[]string{"No matching user found with these credentials."},
		).Render(c.Request().Context(), c.Response().Writer)
	}

	// Generate and store registration token
	token, err := generateSecureToken()
	if err != nil {
		logger.Error().
			Err(err).
			Int("user_id", user.ID).
			Msg("Failed to generate security token")
		return alert.Error(
			"System Error",
			[]string{"Unable to complete registration. Please try again."},
		).Render(c.Request().Context(), c.Response().Writer)
	}

	// Store token with 24-hour expiration
	err = h.db.StoreRegistrationToken(
		c.Request().Context(),
		user.ID,
		token,
		time.Now().Add(24*time.Hour),
	)
	if err != nil {
		logger.Error().
			Err(err).
			Int("user_id", user.ID).
			Msg("Failed to store registration token")
		return alert.Error(
			"System Error",
			[]string{"Unable to complete registration. Please try again."},
		).Render(c.Request().Context(), c.Response().Writer)
	}

	logger.Info().
		Int("user_id", user.ID).
		Str("email", params.Email).
		Str("facility_code", fullFacilityCode).
		Msg("Registration successful")

	if c.Request().Header.Get("HX-Request") == "true" {
		// Return both success alert and redirect
		c.Response().Header().Set("HX-Redirect", fmt.Sprintf("/set-password?token=%s", token))
		return render(c, ComponentGroup(
			alert.Success(
				"Registration Successful",
				"Please set your password to complete registration.",
			),
		))
	}

	// Non-HTMX fallback
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/set-password?token=%s", token))
}

func validateRegistrationParams(params *RegistrationParams) error {
	// Validate facility code (3 letters)
	if len(params.FacilityCode) != 3 {
		return fmt.Errorf("facility code must be exactly 3 letters")
	}
	if !isAlpha(params.FacilityCode) {
		return fmt.Errorf("facility code must contain only letters")
	}

	// Validate initials (2 letters)
	if len(params.Initials) != 2 {
		return fmt.Errorf("initials must be exactly 2 letters")
	}
	if !isAlpha(params.Initials) {
		return fmt.Errorf("initials must contain only letters")
	}

	// Basic email validation
	if params.Email == "" {
		return fmt.Errorf("email is required")
	}
	if !strings.Contains(params.Email, "@") || !strings.Contains(params.Email, ".") {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

func isAlpha(s string) bool {
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')) {
			return false
		}
	}
	return true
}

type SetPasswordRequest struct {
	Password        string `json:"password" form:"password"`
	ConfirmPassword string `json:"confirm_password" form:"confirm_password"`
	Token           string `json:"token" form:"token"`
}

func (h *Handler) GetSetPassword(c echo.Context) error {
	ctx := c.Request().Context()

	// Get and validate token
	token := c.QueryParam("token")
	if token == "" {
		h.logger.Warn().Msg("Attempted to access set-password page without token")
		return c.Redirect(302, "/register")
	}

	// Verify token is valid in database
	_, err := h.db.VerifyRegistrationToken(ctx, token)
	if err != nil {
		h.logger.Error().Err(err).Str("token", token).Msg("Invalid or expired registration token")
		return c.Redirect(302, "/register")
	}

	// Optionally, you could pass the token to the template
	// if you need it for the form submission
	return render(c, page.SetPassword())
}

// GetRegister renders the registration page
func (h *Handler) GetRegister(c echo.Context) error {
	return render(c, page.Register())
}

func (h *Handler) HandleSetPassword(c echo.Context) error {
	ctx := context.Background()
	var req SetPasswordRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	// Validate passwords match and meet requirements
	if err := validatePasswords(req.Password, req.ConfirmPassword); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Verify token and get user ID
	userID, err := h.db.VerifyRegistrationToken(ctx, req.Token)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid or expired token"})
	}

	// Hash password using bcrypt
	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not process password"})
	}

	// Update user password and registration status
	err = h.db.SetUserPassword(ctx, userID, hashedPassword)
	if err != nil {
		return alert.Error(
			"System Error",
			[]string{"Unable to set password. Please try again."},
		).Render(c.Request().Context(), c.Response().Writer)
	}

	// Clean up used token
	err = h.db.DeleteRegistrationToken(ctx, req.Token)
	if err != nil {
		// Log the error but don't fail the request
		h.logger.Error().Msg("failed to delete registration token")
	}

	// Redirect to login page
	return c.JSON(http.StatusOK, map[string]string{
		"redirect": "/login",
		"message":  "Password set successfully. Please log in.",
	})
}

func validateRegistrationRequest(req RegistrationParams) error {
	// Validate facility code (3 letters)
	if len(req.FacilityCode) != 3 || !isAlpha(req.FacilityCode) {
		return fmt.Errorf("facility code must be exactly 3 letters")
	}

	// Validate initials (2 letters)
	if len(req.Initials) != 2 || !isAlpha(req.Initials) {
		return fmt.Errorf("initials must be exactly 2 letters")
	}

	// Basic email validation
	if !strings.Contains(req.Email, "@") || !strings.Contains(req.Email, ".") {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

func validatePasswords(password, confirmPassword string) error {
	if password != confirmPassword {
		return fmt.Errorf("passwords do not match")
	}

	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	// Add any additional password strength requirements here
	// For example:
	hasUpper := false
	hasLower := false
	hasNumber := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasNumber = true
		}
	}

	if !hasUpper || !hasLower || !hasNumber {
		return fmt.Errorf("password must contain at least one uppercase letter, one lowercase letter, and one number")
	}

	return nil
}
