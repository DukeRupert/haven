package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/DukeRupert/haven/types"
	"github.com/DukeRupert/haven/view/alert"
	"github.com/DukeRupert/haven/view/component"
	"github.com/DukeRupert/haven/view/page"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

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

// Updated handler function
func (h *Handler) handleProfile(c echo.Context, routeCtx *types.RouteContext, navItems []types.NavItem) error {
	logger := h.logger.With().
		Str("handler", "handleProfile").
		Str("facility", c.Param("facility")).
		Logger()

	auth, err := GetAuthContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "auth context error")
	}

	// Check for initials query param
	initials := c.Param("initials")

	// If no initials provided, show the current user's profile
	if initials == "" {
		initials = auth.Initials // Assuming this exists in your auth context
		logger.Debug().
			Str("initials", initials).
			Msg("no initials provided, showing own profile")
	} else {
		// Log if we're viewing someone else's profile
		logger.Debug().
			Str("initials", initials).
			Str("viewer", auth.Initials).
			Str("route_initials", routeCtx.UserInitials).
			Msg("viewing other user's profile")

		// Optional: Check if user has permission to view other profiles
		if !canViewOtherProfiles(auth) {
			logger.Warn().
				Str("initials", initials).
				Str("viewer", auth.Initials).
				Msg("unauthorized attempt to view other profile")
			return echo.NewHTTPError(http.StatusForbidden, "Not authorized to view other profiles")
		}
	}

	details, err := h.db.GetUserDetails(
		c.Request().Context(),
		initials,
		auth.FacilityID,
	)
	if err != nil {
		logger.Error().
			Err(err).
			Str("initials", initials).
			Msg("failed to find user details")
		return c.String(http.StatusNotFound, "User not found")
	}

	title := "Profile"
	description := "Manage your account information and schedule"

	// Optionally modify title/description when viewing other profiles
	if initials != auth.Initials {
		title = fmt.Sprintf("%s's Profile", details.User.Initials) // Adjust based on your user details structure
		description = fmt.Sprintf("View %s's information and schedule", details.User.Initials)
	}

	component := page.Profile(
		*routeCtx,
		navItems,
		title,
		description,
		auth,
		details,
	)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

// Helper function to check if user can view other profiles
func canViewOtherProfiles(auth *types.AuthContext) bool {
	// Implement your permission logic here
	// For example:
	switch auth.Role {
	case types.UserRoleAdmin, types.UserRoleSuper:
		return true
	default:
		return false
	}
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

func (h *Handler) handleUsers(c echo.Context, routeCtx *types.RouteContext, navItems []types.NavItem) error {
	// Get facility code from route parameter
	code := c.Param("facility")
	if code == "" {
		h.logger.Error().Msg("facility code is missing from request")
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "facility code is required",
		})
	}

	// Get users from database
	// // Track database query duration specifically
	users, err := h.db.GetUsersByFacilityCode(c.Request().Context(), code)
	if err != nil {
		h.logger.Error().
			Err(err).
			Msg("failed to retrieve users from database")

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to retrieve users",
		})
	}

	// If no users found, return empty array instead of null
	if users == nil {
		users = []types.User{}
	}

	title := "Controllers"
	description := "A list of all controllers assigned to the facility."
	auth, err := GetAuthContext(c)

	component := page.ShowUsers(
		*routeCtx,
		navItems,
		title,
		description,
		auth.Role,
		users,
	)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

// DeleteUser handles DELETE /api/user/:id
func (h *Handler) DeleteUser(c echo.Context) error {
	logger := h.logger.With().
		Str("component", "handler").
		Str("handler", "DeleteUser").
		Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
		Logger()

	// Parse user ID from path
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		logger.Debug().
			Err(err).
			Str("user_id", c.Param("id")).
			Msg("Invalid user ID format")
		return ErrorResponse(c, http.StatusBadRequest, "Invalid Request",
			[]string{"Invalid user ID format"})
	}

	// Get the user to check if exists and for logging
	user, err := h.db.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		logger.Error().
			Err(err).
			Int("user_id", userID).
			Msg("Failed to fetch user")
		return SystemError(c)
	}
	if user == nil {
		return ErrorResponse(c, http.StatusNotFound, "Not Found",
			[]string{"User not found"})
	}

	// Delete the user
	err = h.db.DeleteUser(c.Request().Context(), userID)
	if err != nil {
		logger.Error().
			Err(err).
			Int("user_id", userID).
			Msg("Failed to delete user")
		return SystemError(c)
	}

	// Log success
	logger.Info().
		Int("user_id", user.ID).
		Str("email", user.Email).
		Int("facility_id", user.FacilityID).
		Msg("User deleted successfully")

	// For HTMX requests, return a success message
	if c.Request().Header.Get("HX-Request") == "true" {
		return render(c, alert.Success(
			"User Deleted",
			fmt.Sprintf("Successfully deleted user %s %s", user.FirstName, user.LastName),
		))
	}

	// For non-HTMX requests, return a 204 No Content
	return c.NoContent(http.StatusNoContent)
}

// handles /api/user/:id/password
func (h *Handler) updatePasswordForm(c echo.Context) error {
	// Get user id from params
	user_id, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		h.logger.Error().Err(err).Msg("Missing user_id parameter")
		return SystemError(c)
	}

	return render(c, component.Update_Password_Form(user_id))
}

type UpdatePasswordParams struct {
	Password string `form:"password"`
	Confirm  string `form:"confirm"`
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
	var formData UpdatePasswordParams
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
	err = h.db.UpdateUserPassword(c.Request().Context(), userID, string(hashedPassword))
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
