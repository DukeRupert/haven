package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/DukeRupert/haven/internal/model/entity"
	"github.com/DukeRupert/haven/internal/model/params"

	"github.com/DukeRupert/haven/web/view/alert"
	"github.com/DukeRupert/haven/web/view/page"
	"github.com/labstack/echo/v4"
)

type RegistrationParams struct {
	FacilityCode string `form:"facility_code"`
	Initials     string `form:"initials"`
	Email        string `form:"email"`
}

type EmailVerificationRequest struct {
	Email string `json:"email" form:"email"`
}

// InitiateEmailVerification handles the initial email verification request
func (h *Handler) InitiateEmailVerification(c echo.Context) error {
	logger := h.logger.With().
		Str("component", "verification").
		Str("handler", "InitiateEmailVerification").
		Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
		Logger()

	var req EmailVerificationRequest
	if err := c.Bind(&req); err != nil {
		logger.Debug().Err(err).Msg("Invalid request format")
		return alert.Error(
			"Invalid Request",
			[]string{"Please provide a valid email address."},
		).Render(c.Request().Context(), c.Response().Writer)
	}

	// Validate email format
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return alert.Error(
			"Invalid Email",
			[]string{"Please provide a valid email address."},
		).Render(c.Request().Context(), c.Response().Writer)
	}

	// Check if user exists with this email
	user, err := h.repos.User.GetByEmail(c.Request().Context(), req.Email)
	if err != nil {
		logger.Error().Err(err).Str("email", req.Email).Msg("Database error when checking user")
		return alert.Error(
			"System Error",
			[]string{"Unable to process request. Please try again."},
		).Render(c.Request().Context(), c.Response().Writer)
	}
	if user == nil {
		// Don't reveal if email exists or not for security
		return alert.Success(
			"Verification Email Sent",
			"If an account exists with this email, you will receive a verification link shortly.",
		).Render(c.Request().Context(), c.Response().Writer)
	}

	// Generate verification token
	token, err := generateSecureToken()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to generate verification token")
		return alert.Error(
			"System Error",
			[]string{"Unable to process request. Please try again."},
		).Render(c.Request().Context(), c.Response().Writer)
	}

	// Store verification token
	verificationToken := &entity.VerificationToken{
		UserID:    user.ID,
		Token:     token,
		Email:     req.Email,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := h.repos.Token.StoreVerification(c.Request().Context(), verificationToken); err != nil {
		logger.Error().Err(err).Msg("Failed to store verification token")
		return alert.Error(
			"System Error",
			[]string{"Unable to process request. Please try again."},
		).Render(c.Request().Context(), c.Response().Writer)
	}

	// Send verification email
	if err := h.sendVerificationEmail(c.Request().Context(), user, token); err != nil {
		logger.Error().Err(err).Msg("Failed to send verification email")
		return alert.Error(
			"System Error",
			[]string{"Unable to send verification email. Please try again."},
		).Render(c.Request().Context(), c.Response().Writer)
	}

	return alert.Success(
		"Verification Email Sent",
		"Please check your email for a verification link.",
	).Render(c.Request().Context(), c.Response().Writer)
}

func (h *Handler) sendVerificationEmail(ctx context.Context, user *entity.User, token string) error {
	verificationURL := fmt.Sprintf("%s/register?token=%s", h.config.BaseURL, token)

	data := map[string]interface{}{
		"VerificationURL": verificationURL,
		"ExpiresIn":       "24 hours",
		"FromName":        "Haven Support",
		"Subject":         "Complete Your Registration",
	}

	return h.mailer.SendTemplate(ctx, "verification", user.Email, data)
}

// Modify the existing registration handler to work with verification
func (h *Handler) HandleRegistration(c echo.Context) error {
	logger := h.logger.With().
		Str("component", "registration").
		Str("handler", "HandleRegistration").
		Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
		Logger()

	// Get token from query params
	token := c.QueryParam("token")
	if token == "" {
		return c.Redirect(http.StatusSeeOther, "/verify")
	}

	// Verify token exists and is valid
	verificationToken, err := h.repos.Token.GetVerificationToken(c.Request().Context(), token)
	if err != nil {
		logger.Error().Err(err).Msg("Invalid verification token")
		return c.Redirect(http.StatusSeeOther, "/verify")
	}

	// If it's a GET request, render the registration form with pre-filled email
	if c.Request().Method == http.MethodGet {
		return render(c, page.Register(params.RegisterParams{
			Email: verificationToken.Email,
			Token: token,
		}))
	}

	// Handle POST request
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
		return alert.Error(
			"Invalid Input",
			[]string{err.Error()},
		).Render(c.Request().Context(), c.Response().Writer)
	}

	// Verify email matches the one in the token
	if params.Email != verificationToken.Email {
		return alert.Error(
			"Invalid Email",
			[]string{"The email address doesn't match the verification token."},
		).Render(c.Request().Context(), c.Response().Writer)
	}

	// Add 'K' prefix to facility code
	fullFacilityCode := "K" + strings.ToUpper(params.FacilityCode)

	// Check if facility exists
	facility, err := h.repos.Facility.GetByCode(c.Request().Context(), fullFacilityCode)
	if err != nil || facility == nil {
		return alert.Error(
			"Invalid Facility",
			[]string{"The facility code you entered is not valid."},
		).Render(c.Request().Context(), c.Response().Writer)
	}

	// Verify user credentials
	user, err := h.repos.User.VerifyCredentials(
		c.Request().Context(),
		facility.ID,
		strings.ToUpper(params.Initials),
		strings.ToLower(params.Email),
	)
	if err != nil || user == nil {
		return alert.Error(
			"Invalid Credentials",
			[]string{"No matching user found with these credentials."},
		).Render(c.Request().Context(), c.Response().Writer)
	}

	// Mark verification token as used
	if err := h.repos.Token.MarkAsUsed(c.Request().Context(), token); err != nil {
		logger.Error().Err(err).Msg("Failed to mark token as used")
		// Continue anyway since verification was successful
	}

	// Generate registration token for password setup
	registrationToken, err := generateSecureToken()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to generate registration token")
		return alert.Error(
			"System Error",
			[]string{"Unable to complete registration. Please try again."},
		).Render(c.Request().Context(), c.Response().Writer)
	}

	// Store registration token
	if err := h.repos.Token.Store(
		c.Request().Context(),
		user.ID,
		registrationToken,
		time.Now().Add(24*time.Hour),
	); err != nil {
		logger.Error().Err(err).Msg("Failed to store registration token")
		return alert.Error(
			"System Error",
			[]string{"Unable to complete registration. Please try again."},
		).Render(c.Request().Context(), c.Response().Writer)
	}

	// Redirect to password setup
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/set-password?token=%s", registrationToken))
}

func validateRegistrationParams(params *RegistrationParams) error {
	// Validate facility code (3 letters)
	if len(params.FacilityCode) == 0 {
		return fmt.Errorf("facility code is required")
	}
	if len(params.FacilityCode) != 3 {
		return fmt.Errorf("facility code must be exactly 3 letters")
	}
	if !isAlpha(params.FacilityCode) {
		return fmt.Errorf("facility code must contain only letters")
	}

	// Validate initials (2 letters)
	if len(params.Initials) == 0 {
		return fmt.Errorf("initials are required")
	}
	if len(params.Initials) != 2 {
		return fmt.Errorf("initials must be exactly 2 letters")
	}
	if !isAlpha(params.Initials) {
		return fmt.Errorf("initials must contain only letters")
	}

	// Validate email
	if params.Email == "" {
		return fmt.Errorf("email is required")
	}

	// Use mail.ParseAddress for robust email validation
	if _, err := mail.ParseAddress(params.Email); err != nil {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

// Helper function to check if a string contains only letters
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
	_, err := h.repos.Token.Verify(ctx, token)
	if err != nil {
		h.logger.Error().Err(err).Str("token", token).Msg("Invalid or expired registration token")
		return c.Redirect(302, "/register")
	}

	// Optionally, you could pass the token to the template
	// if you need it for the form submission
	return render(c, page.SetPassword())
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
	userID, err := h.repos.Token.Verify(ctx, req.Token)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid or expired token"})
	}

	// Hash password using bcrypt
	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not process password"})
	}

	// Update user password and registration status
	err = h.repos.User.SetPassword(ctx, userID, string(hashedPassword))
	if err != nil {
		return alert.Error(
			"System Error",
			[]string{"Unable to set password. Please try again."},
		).Render(c.Request().Context(), c.Response().Writer)
	}

	// Clean up used token
	err = h.repos.Token.Delete(ctx, req.Token)
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
