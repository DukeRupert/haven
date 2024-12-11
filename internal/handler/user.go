package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/DukeRupert/haven/internal/model/dto"
	"github.com/DukeRupert/haven/internal/model/entity"
	"github.com/DukeRupert/haven/internal/model/params"
	"github.com/DukeRupert/haven/internal/model/types"
	"github.com/DukeRupert/haven/view/alert"
	"github.com/DukeRupert/haven/view/component"
	"github.com/DukeRupert/haven/web/view/page"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// HandleUsers renders the users list page for a specific facility
func (h *Handler) HandleUsers(c echo.Context, routeCtx *dto.RouteContext, navItems []dto.NavItem) error {
	logger := h.logger.With().
		Str("handler", "HandleUsers").
		Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
		Logger()

	// Get and validate facility code
	code := c.Param("facility")
	if code == "" {
		logger.Error().Msg("missing facility code in request")
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"Facility code is required",
		)
	}

	// Get users from repository
	users, err := h.repos.User.GetByFacilityCode(c.Request().Context(), code)
	if err != nil {
		logger.Error().
			Err(err).
			Str("facility_code", code).
			Msg("failed to retrieve users")

		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"Unable to load controllers. Please try again later.",
		)
	}

	// Ensure users is never nil
	if users == nil {
		users = []entity.User{}
	}

	// Get auth context for role information
	auth, err := GetAuthContext(c)
	if err != nil {
		logger.Error().
			Err(err).
			Msg("failed to get auth context")
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"Unable to verify permissions. Please try again later.",
		)
	}

	// Page metadata
	pageData := struct {
		Title       string
		Description string
	}{
		Title:       "Controllers",
		Description: "A list of all controllers assigned to the facility.",
	}

	logger.Debug().
		Str("facility_code", code).
		Int("user_count", len(users)).
		Str("user_role", string(auth.Role)).
		Msg("rendering users page")

	// Render the page
	return page.ShowUsers(
		*routeCtx,
		navItems,
		pageData.Title,
		pageData.Description,
		auth.Role,
		users,
	).Render(c.Request().Context(), c.Response().Writer)
}

// Helper function to validate role
func isValidRole(role types.UserRole) bool {
	validRoles := []types.UserRole{"super", "admin", "user"}
	for _, r := range validRoles {
		if role == r {
			return true
		}
	}
	return false
}

func (h *Handler) HandleUpdateUser(c echo.Context) error {
	logger := h.logger.With().
		Str("handler", "HandleUpdateUser").
		Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
		Logger()

	// Validate user ID
	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		logger.Debug().Err(err).Msg("invalid user ID format")
		return ErrorResponse(c, http.StatusBadRequest,
			"Invalid User ID",
			[]string{"Please provide a valid user ID"})
	}

	// Get auth context
	auth, err := GetAuthContext(c)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get auth context")
		return SystemError(c)
	}

	// Check basic permissions
	if !canUpdateUser(auth, userID) {
		logger.Warn().
			Int("user_id", userID).
			Str("role", string(auth.Role)).
			Msg("unauthorized update attempt")
		return ErrorResponse(c, http.StatusForbidden,
			"Forbidden",
			[]string{"You don't have permission to update this user"})
	}

	// Bind and validate update params
	var params types.UpdateUserParams
	if err := c.Bind(&params); err != nil {
		logger.Debug().Err(err).Msg("invalid form data")
		return ErrorResponse(c, http.StatusBadRequest,
			"Invalid Request",
			[]string{"Please check the form data and try again"})
	}

	logger.Debug().Interface("update_params", params).Msg("processing update")

	// Validate role changes
	if err := h.validateRoleChange(c.Request().Context(), auth, userID, params.Role); err != nil {
		return err
	}

	// Perform update
	updatedUser, err := h.repos.User.Update(c.Request().Context(), userID, params)
	if err != nil {
		if strings.Contains(err.Error(), "email already exists") {
			return ValidationError(c, []string{"This email address is already in use"})
		}
		logger.Error().Err(err).Int("user_id", userID).Msg("failed to update user")
		return SystemError(c)
	}

	logger.Info().
		Int("user_id", updatedUser.ID).
		Str("email", updatedUser.Email).
		Str("role", string(updatedUser.Role)).
		Msg("user updated successfully")

	return render(c, ComponentGroup(
		alert.Success("User Updated",
			fmt.Sprintf("Successfully updated user %s %s",
				updatedUser.FirstName, updatedUser.LastName)),
		page.ProfileUserCard(*updatedUser, *auth),
	))
}

// Create handles POST /api/user
func (h *Handler) handleCreateUser(c echo.Context) error {
	logger := h.logger

	// Create a struct specifically for form binding
	var params types.CreateUserParams

	// Bind the form data directly to params
	if err := c.Bind(&params); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "Invalid Request",
			[]string{"The submitted form data was invalid"})
	}

	// Validate and collect all validation errors
	var errors []string

	// Validate facility code and get facility
	if params.FacilityCode == "" {
		errors = append(errors, "Facility code is required")
	} else {
		facility, err := h.db.GetFacilityByCode(c.Request().Context(), params.FacilityCode)
		if err != nil {
			logger.Error().
				Err(err).
				Str("facility_code", params.FacilityCode).
				Msg("failed to find facility by code")
			errors = append(errors, "The specified facility code was not found")
		} else {
			params.FacilityID = facility.ID
		}
	}

	// Return all validation errors if any exist
	if len(errors) > 0 {
		return ValidationError(c, errors)
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error().Msg("Failed to hash password")
		return SystemError(c)
	}
	params.Password = string(hashedPassword)

	// Create the user
	user, err := h.db.CreateUser(c.Request().Context(), params)
	if err != nil {
		logger.Error().Msg("Failed to create user in database")
		return SystemError(c)
	}

	// Log success
	logger.Info().
		Int("user_id", user.ID).
		Str("email", user.Email).
		Int("facility_id", user.FacilityID).
		Str("role", string(user.Role)).
		Msg("user created successfully")

	// For successful creation, you might want to show a success message
	if c.Request().Header.Get("HX-Request") == "true" {
		// Return both the success alert and the new user list item
		return render(c, ComponentGroup(
			alert.Success("User Created", fmt.Sprintf("Successfully created user %s %s", user.FirstName, user.LastName)),
			page.UserListItem(h.RouteCtx, *user),
		))
	}

	return render(c, page.UserListItem(h.RouteCtx, *user))
}

func (h *Handler) createUserForm(c echo.Context) error {
	// Get facility code from route parameter
	code := c.Param("facility")

	if code == "" {
		h.logger.Error().Msg("facility code is missing from request")
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "facility code is required",
		})
	}

	auth, err := GetAuthContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "auth context error")
	}

	return render(c, component.CreateUserForm(code, string(auth.Role)))
}

func (h *Handler) HandleDeleteUser(c echo.Context, routeCtx *dto.RouteContext, navItems []dto.NavItem) error {
	logger := h.logger.With().
		Str("handler", "HandleDeleteUser").
		Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
		Logger()

	// Validate user ID
	userID, err := getUserID(c)
	if err != nil {
		logger.Debug().
			Err(err).
			Str("user_id_param", c.Param("user_id")).
			Msg("invalid user ID format")
		return ErrorResponse(c, http.StatusBadRequest,
			"Invalid Request",
			[]string{"Please provide a valid user ID"})
	}

	// Check permissions
	auth, err := GetAuthContext(c)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get auth context")
		return SystemError(c)
	}

	if !canDeleteUser(auth, userID) {
		logger.Warn().
			Int("target_user_id", userID).
			Int("requesting_user_id", auth.UserID).
			Str("role", string(auth.Role)).
			Msg("unauthorized deletion attempt")
		return ErrorResponse(c, http.StatusForbidden,
			"Access Denied",
			[]string{"You don't have permission to delete this user"})
	}

	// Verify user exists and get details for logging
	user, err := h.repos.User.GetByID(c.Request().Context(), userID)
	if err != nil {
		logger.Error().
			Err(err).
			Int("user_id", userID).
			Msg("failed to fetch user")
		return SystemError(c)
	}

	// Delete user
	if err := h.repos.User.Delete(c.Request().Context(), userID); err != nil {
		logger.Error().
			Err(err).
			Int("user_id", userID).
			Msg("failed to delete user")
		return SystemError(c)
	}

	logger.Info().
		Int("user_id", user.ID).
		Str("email", user.Email).
		Int("facility_id", user.FacilityID).
		Msg("user deleted successfully")

	// Handle HTMX response
	return handleDeleteResponse(c, routeCtx.FacilityCode)
}

func handleDeleteResponse(c echo.Context, facilityCode string) error {
	redirectURL := fmt.Sprintf("/%s/controllers", facilityCode)
	c.Response().Header().Set("HX-Redirect", redirectURL)
	return c.NoContent(http.StatusOK)
}

// GetUpdatePasswordForm renders the password update form
func (h *Handler) GetUpdatePasswordForm(c echo.Context) error {
	logger := h.logger.With().
		Str("handler", "GetUpdatePasswordForm").
		Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
		Logger()

	// Validate user ID
	userID, err := getUserID(c)
	if err != nil {
		logger.Error().
			Err(err).
			Str("user_id_param", c.Param("user_id")).
			Msg("invalid user ID parameter")
		return SystemError(c)
	}

	// Get auth context for permission check
	auth, err := GetAuthContext(c)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get auth context")
		return ErrorResponse(c,
			http.StatusInternalServerError,
			"Authentication Error",
			[]string{"Unable to verify permissions"},
		)
	}

	// Check if user can update this password
	if !canUpdatePassword(auth, userID) {
		logger.Warn().
			Int("target_user_id", userID).
			Int("requesting_user_id", auth.UserID).
			Str("role", string(auth.Role)).
			Msg("unauthorized password form access attempt")
		return ErrorResponse(c,
			http.StatusForbidden,
			"Access Denied",
			[]string{"You don't have permission to update this password"},
		)
	}

	logger.Debug().
		Int("user_id", userID).
		Msg("rendering password update form")

	return render(c, component.UpdatePasswordForm(userID))
}

func (h *Handler) handleUpdatePassword(c echo.Context) error {
	logger := h.logger.With().
		Str("handler", "handleUpdatePassword").
		Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
		Logger()

	// Parse the user ID from the URL
	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		logger.Error().Err(err).Msg("Invalid user ID")
		return ErrorResponse(c,
			http.StatusBadRequest,
			"Invalid Request",
			[]string{"Invalid user ID provided"},
		)
	}

	// Get authenticated user from context
	auth, err := GetAuthContext(c)
	if err != nil {
		logger.Error().Err(err).Msg("Auth context error")
		return ErrorResponse(c,
			http.StatusInternalServerError,
			"System Error",
			[]string{"Authentication error occurred"},
		)
	}

	// Parse form data
	var formData params.UpdatePasswordParams
	if err := c.Bind(&formData); err != nil {
		logger.Error().Err(err).Msg("Failed to bind form data")
		return ErrorResponse(c,
			http.StatusBadRequest,
			"Invalid Request",
			[]string{"Please check your input and try again"},
		)
	}

	// Validate passwords match
	if formData.Password != formData.Confirm {
		return ErrorResponse(c,
			http.StatusBadRequest,
			"Validation Error",
			[]string{"Passwords do not match"},
		)
	}

	// Validate password length
	if len(formData.Password) < 8 {
		return ErrorResponse(c,
			http.StatusBadRequest,
			"Validation Error",
			[]string{"Password must be at least 8 characters long"},
		)
	}

	// Check authorization
	if !isAuthorized(auth.UserID, auth.Role, userID) {
		logger.Warn().
			Int("requesting_user", auth.UserID).
			Int("target_user", userID).
			Msg("Unauthorized password update attempt")
		return ErrorResponse(c,
			http.StatusForbidden,
			"Unauthorized",
			[]string{"You are not authorized to modify this user's password"},
		)
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(formData.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to hash password")
		return ErrorResponse(c,
			http.StatusInternalServerError,
			"System Error",
			[]string{"Failed to process password update"},
		)
	}

	// Update password in database
	err = h.repos.User.UpdatePassword(c.Request().Context(), userID, string(hashedPassword))
	if err != nil {
		logger.Error().Err(err).
			Int("user_id", userID).
			Msg("Failed to update password")
		return ErrorResponse(c,
			http.StatusInternalServerError,
			"System Error",
			[]string{"Failed to update password"},
		)
	}

	// Return success component
	return SuccessResponse(c, "Success", "Password updated")
}

// GetUpdateUserForm renders the form for updating a user
func (h *Handler) GetUpdateUserForm(c echo.Context) error {
	logger := h.logger.With().
		Str("handler", "GetUpdateUserForm").
		Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
		Logger()

	// Validate input and permissions
	formData, err := h.validateFormAccess(c)
	if err != nil {
		return err // validateFormAccess handles error responses
	}

	// Get user details
	user, err := h.repos.User.GetByID(c.Request().Context(), formData.UserID)
	if err != nil {
		logger.Error().
			Err(err).
			Int("user_id", formData.UserID).
			Msg("failed to retrieve user")
		return SystemError(c)
	}

	logger.Debug().
		Int("user_id", user.ID).
		Str("email", user.Email).
		Str("viewer_role", string(formData.Auth.Role)).
		Msg("rendering update form")

	return render(c, component.UpdateUserForm(*user, *formData.Auth))
}

// Form validation types and helpers
type UpdateUserFormData struct {
	UserID int
	Auth   *dto.AuthContext
}

func (h *Handler) validateFormAccess(c echo.Context) (*UpdateUserFormData, error) {
	logger := h.logger.With().
		Str("method", "validateFormAccess").
		Logger()

	// Parse and validate user ID
	userID, err := getUserID(c)
	if err != nil {
		logger.Debug().Err(err).Msg("invalid user ID")
		return nil, SystemError(c)
	}

	// Get auth context
	auth, err := GetAuthContext(c)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get auth context")
		return nil, ErrorResponse(c,
			http.StatusInternalServerError,
			"Authentication Error",
			[]string{"Unable to verify permissions"},
		)
	}

	// Check permissions
	if !canAccessUserForm(auth, userID) {
		logger.Warn().
			Int("target_user_id", userID).
			Int("requesting_user_id", auth.UserID).
			Str("role", string(auth.Role)).
			Msg("unauthorized form access attempt")
		return nil, ErrorResponse(c,
			http.StatusForbidden,
			"Access Denied",
			[]string{"You don't have permission to edit this user"},
		)
	}

	return &UpdateUserFormData{
		UserID: userID,
		Auth:   auth,
	}, nil
}
