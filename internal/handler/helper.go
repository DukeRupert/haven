package handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/DukeRupert/haven/internal/middleware"
	"github.com/DukeRupert/haven/internal/model/dto"
	"github.com/DukeRupert/haven/internal/model/entity"
	"github.com/DukeRupert/haven/internal/model/params"
	"github.com/DukeRupert/haven/internal/model/types"
	"github.com/DukeRupert/haven/internal/response"
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
)

func canUpdateUser(auth *dto.AuthContext, targetUserID int) bool {
	if auth.Role == types.UserRoleSuper {
		return true
	}
	if auth.Role == types.UserRoleAdmin {
		return true
	}
	return auth.UserID == targetUserID
}

func (h *Handler) validateRoleChange(ctx context.Context, auth *dto.AuthContext,
	userID int, newRole types.UserRole,
) error {
	// Get existing user
	existingUser, err := h.repos.User.GetByID(ctx, userID)
	if err != nil {
		return response.System(echo.New().AcquireContext())
	}

	// No role change, no validation needed
	if existingUser.Role == newRole {
		return nil
	}

	// Validate role change permissions
	if auth.Role != types.UserRoleSuper && newRole == types.UserRoleSuper {
		return response.Error(echo.New().AcquireContext(),
			http.StatusForbidden,
			"Invalid Role",
			[]string{"Only super admins can assign super admin role"})
	}

	if auth.Role == types.UserRoleUser {
		return response.Error(echo.New().AcquireContext(),
			http.StatusForbidden,
			"Invalid Role",
			[]string{"Users cannot change roles"})
	}

	return nil
}

func getUserID(c echo.Context) (int, error) {
	return strconv.Atoi(c.Param("user_id"))
}

func canAccessUserForm(auth *dto.AuthContext, targetUserID int) bool {
	// Super admins can edit anyone
	if auth.Role == types.UserRoleSuper {
		return true
	}

	// Admins can edit non-super users
	if auth.Role == types.UserRoleAdmin {
		return true
	}

	// Users can only edit themselves
	return auth.UserID == targetUserID
}

func canDeleteUser(auth *dto.AuthContext, targetUser *entity.User) bool {
	switch auth.Role {
	case types.UserRoleSuper:
		return true
	case types.UserRoleAdmin:
		// Admin can delete users in their facility except themselves
		return auth.UserID != targetUser.ID &&
			auth.FacilityID == targetUser.FacilityID
	default:
		return false
	}
}

func canUpdatePassword(auth *dto.AuthContext, targetUserID int) bool {
	switch auth.Role {
	case types.UserRoleSuper:
		return true
	case types.UserRoleAdmin:
		return true
	default:
		return auth.UserID == targetUserID
	}
}

type passwordUpdateData struct {
	UserID   int
	Password string
}

func (h *Handler) validatePasswordUpdate(c echo.Context) (*passwordUpdateData, *dto.AuthContext, error) {
	// Parse user ID
	userID, err := getUserID(c)
	if err != nil {
		return nil, nil, response.Error(c,
			http.StatusBadRequest,
			"Invalid Request",
			[]string{"Invalid user ID provided"},
		)
	}

	// Get auth context
	auth, err := middleware.GetAuthContext(c)
	if err != nil {
		return nil, nil, response.Error(c,
			http.StatusInternalServerError,
			"System Error",
			[]string{"Authentication error occurred"},
		)
	}

	// Check authorization
	if !canUpdatePassword(auth, userID) {
		return nil, nil, response.Error(c,
			http.StatusForbidden,
			"Unauthorized",
			[]string{"You don't have permission to update this password"},
		)
	}

	// Parse and validate form data
	var formData params.UpdatePasswordParams
	if err := c.Bind(&formData); err != nil {
		return nil, nil, response.Error(c,
			http.StatusBadRequest,
			"Invalid Request",
			[]string{"Please check your input and try again"},
		)
	}

	// Validate password
	if err := validatePassword(formData); err != nil {
		return nil, nil, response.Error(c,
			http.StatusBadRequest,
			"Validation Error",
			[]string{err.Error()},
		)
	}

	return &passwordUpdateData{
		UserID:   userID,
		Password: formData.Password,
	}, auth, nil
}

func validatePassword(data params.UpdatePasswordParams) error {
	if data.Password != data.Confirm {
		return errors.New("passwords do not match")
	}

	if len(data.Password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	// Add additional password requirements here
	return nil
}

func hashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

type createUserData struct {
	FirstName  string
	LastName   string
	Email      string
	Password   string
	Initials   string
	Role       types.UserRole
	FacilityID int
}

func (h *Handler) validateCreateUser(c echo.Context) (*createUserData, error) {
	logger := h.logger.With().
		Str("method", "validateCreateUser").
		Logger()

	// Bind form data
	var formParams params.CreateUserParams
	if err := c.Bind(&formParams); err != nil {
		return nil, response.Error(c,
			http.StatusBadRequest,
			"Invalid Request",
			[]string{"Please check your input and try again"},
		)
	}

	// Collect validation errors
	var errors []string

	// Validate facility
	var facilityID int
	if formParams.FacilityCode == "" {
		errors = append(errors, "Facility code is required")
	} else {
		facility, err := h.repos.Facility.GetByCode(
			c.Request().Context(),
			formParams.FacilityCode,
		)
		if err != nil {
			logger.Error().
				Err(err).
				Str("facility_code", formParams.FacilityCode).
				Msg("failed to find facility")
			errors = append(errors, "Invalid facility code")
		} else {
			facilityID = facility.ID
		}
	}

	// Validate email
	if err := validateEmail(formParams.Email); err != nil {
		errors = append(errors, err.Error())
	}

	// Return all validation errors
	if len(errors) > 0 {
		return nil, response.Validation(c, errors)
	}

	return &createUserData{
		FirstName:  formParams.FirstName,
		LastName:   formParams.LastName,
		Email:      formParams.Email,
		Password:   formParams.Password,
		Initials:   formParams.Initials,
		Role:       formParams.Role,
		FacilityID: facilityID,
	}, nil
}

func validateEmail(email string) error {
	// Add email validation logic
	if email == "" {
		return errors.New("email is required")
	}
	// Add more email validation as needed
	return nil
}

func canCreateUsers(auth *dto.AuthContext, facilityID int) bool {
	switch auth.Role {
	case types.UserRoleSuper:
		return true
	case types.UserRoleAdmin:
		return auth.FacilityID == facilityID
	default:
		return false
	}
}

// Reduce boilerplate for simple templ component renders
func render(c echo.Context, component templ.Component) error {
	c.Response().Header().Set("Content-Type", "text/html")
	return component.Render(c.Request().Context(), c.Response())
}

// ComponentGroup combines multiple templ components into a single component
func ComponentGroup(components ...templ.Component) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		for _, component := range components {
			if err := component.Render(ctx, w); err != nil {
				return err
			}
		}
		return nil
	})
}

// generateSecureToken creates a cryptographically secure token for registration
func generateSecureToken() (string, error) {
	// Generate 32 bytes of random data
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("error generating secure token: %w", err)
	}

	// Encode to base64URL (URL-safe version of base64)
	// Use RawURLEncoding to avoid padding characters
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// ValidateFacilityCode extracts and validates the facility code from the request context.
// Returns the facility code if valid, or an error that can be directly returned from the handler.
func ValidateFacilityCode(c echo.Context, logger zerolog.Logger) (string, error) {
	facilityCode := c.Param("facility_id")
	if facilityCode == "" {
		logger.Error().Msg("missing facility code parameter")
		return "", response.Error(c,
			http.StatusBadRequest,
			"Invalid Request",
			[]string{"Facility code is required"},
		)
	}
	return facilityCode, nil
}
