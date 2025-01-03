package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/DukeRupert/haven/internal/middleware"
	"github.com/DukeRupert/haven/internal/model/dto"
	"github.com/DukeRupert/haven/internal/model/entity"
	"github.com/DukeRupert/haven/internal/model/params"
	"github.com/DukeRupert/haven/internal/response"
	"github.com/DukeRupert/haven/web/view/alert"
	"github.com/DukeRupert/haven/web/view/component"
	"github.com/DukeRupert/haven/web/view/page"

	"github.com/labstack/echo/v4"
)

// HandleUsers renders the users list page for a specific facility
func (h *Handler) HandleUsers(c echo.Context) error {
	logger := h.logger.With().
		Str("handler", "HandleUsers").
		Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
		Logger()

	// Get auth context
	auth, err := middleware.GetAuthContext(c)
	if err != nil {
		logger.Error().Msg("missing auth context")
		return response.System(c)
	}

	// Get route context
	route, err := middleware.GetRouteContext(c)
	if err != nil {
		logger.Error().Msg("missing route context")
		return response.System(c)
	}

	// Get and validate facility code
	if route.FacilityCode == "" {
		logger.Error().Msg("missing facility code in request")
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"Facility code is required",
		)
	}

	// Get users from repository
	users, err := h.repos.User.GetByFacilityCode(c.Request().Context(), route.FacilityCode)
	if err != nil {
		logger.Error().
			Err(err).
			Str("facility_code", route.FacilityCode).
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

	// Build nav items
	navItems := BuildNav(route, auth, c.Request().URL.Path)

	// Build props
	props := dto.UsersPageProps{
		Title:       "Controllers",
		Description: "A list of all controllers assigned to the facility.",
		NavItems:    navItems,
		AuthCtx:     *auth,
		RouteCtx:    *route,
		Users:       users,
	}

	logger.Debug().
		Str("facility_code", route.FacilityCode).
		Int("user_count", len(users)).
		Str("user_role", string(auth.Role)).
		Msg("rendering users page")

	// Render the page
	return page.ShowUsers(
		props,
	).Render(c.Request().Context(), c.Response().Writer)
}

// HandleCreateUser processes new user creation requests
func (h *Handler) HandleCreateUser(c echo.Context) error {
	logger := h.logger.With().
		Str("handler", "HandleCreateUser").
		Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
		Logger()

	// Get route context
	route, err := middleware.GetRouteContext(c)
	if err != nil {
		logger.Error().Msg("missing route context")
		return response.System(c)
	}

	// Get and validate facility code
	if route.FacilityCode == "" {
		logger.Error().Msg("missing facility code in request")
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"Facility code is required",
		)
	}

	// Validate and parse create request
	createData, err := h.validateCreateUser(c)
	if err != nil {
		return err // validateCreateUser handles error responses
	}

	// Prepare create params
	params := params.CreateUserParams{
		FirstName:  createData.FirstName,
		LastName:   createData.LastName,
		Email:      createData.Email,
		Password:   "", // Password will be set by user during verification
		Initials:   createData.Initials,
		Role:       createData.Role,
		FacilityID: createData.FacilityID,
	}

	// Create user
	user, err := h.repos.User.Create(c.Request().Context(), params)
	if err != nil {
		logger.Error().
			Err(err).
			Str("email", params.Email).
			Msg("failed to create user")
		return response.System(c)
	}

	// Send verification email
	if err := h.SendVerificationEmail(c.Request().Context(), user, logger); err != nil {
		logger.Error().
			Err(err).
			Int("user_id", user.ID).
			Msg("failed to send verification email")
		// Continue execution since user was created successfully
	}

	logger.Info().
		Int("user_id", user.ID).
		Str("email", user.Email).
		Int("facility_id", user.FacilityID).
		Str("role", string(user.Role)).
		Msg("user created successfully and verification email sent")

	// Handle HTMX request
	if isHtmxRequest(c) {
		return render(c, ComponentGroup(
			alert.Success(
				"User Created",
				fmt.Sprintf("Successfully created user %s %s. A verification email has been sent to %s.",
					user.FirstName,
					user.LastName,
					user.Email,
				),
			),
			page.UserListItem(route.FacilityCode, *user),
		))
	}

	return render(c, page.UserListItem(route.FacilityCode, *user))
}

func (h *Handler) HandleUpdateUser(c echo.Context) error {
	logger := h.logger.With().
		Str("handler", "HandleUpdateUser").
		Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
		Logger()

	// Get auth context
	auth, err := middleware.GetAuthContext(c)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get auth context")
		return response.System(c)
	}

	// Get route context
	route, err := middleware.GetRouteContext(c)
	if err != nil {
		logger.Error().Msg("missing route context")
		return response.System(c)
	}

	if err := ensureRouteParams(route); err != nil {
		return err
	}

	// Get existing user
	existingUser, err := h.repos.User.GetByInitialsAndFacility(c.Request().Context(), route.UserInitials, route.FacilityCode)
	if err != nil {
		return err
	}

	// Bind and validate update params
	var params params.UpdateUserParams
	if err := c.Bind(&params); err != nil {
		logger.Debug().Err(err).Msg("invalid form data")
		return response.Error(c, http.StatusBadRequest,
			"Invalid Request",
			[]string{"Please check the form data and try again"})
	}

	// Validate role changes
	if err := h.verifyRoleChange(auth, existingUser.Role, params.Role); err != nil {
		return err
	}

	// Perform update
	updatedUser, err := h.repos.User.Update(c.Request().Context(), existingUser.ID, params)
	if err != nil {
		if strings.Contains(err.Error(), "email already exists") {
			return response.Validation(c, []string{"This email address is already in use"})
		}
		logger.Error().Err(err).Int("user_id", existingUser.ID).Msg("failed to update user")
		return response.System(c)
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
		page.UserDetails(*updatedUser, route.FacilityCode, *auth),
	))
}

func (h *Handler) HandleDeleteUser(c echo.Context) error {
	logger := h.logger.With().
		Str("handler", "HandleDeleteUser").
		Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
		Logger()

	// Get auth context
	auth, err := middleware.GetAuthContext(c)
	if err != nil {
		logger.Error().Msg("missing auth context")
		return response.System(c)
	}

	// Get route context
	route, err := middleware.GetRouteContext(c)
	if err != nil {
		logger.Error().Msg("missing route context")
		return response.System(c)
	}

	// Verify params exist
	if route.UserInitials == "" || route.FacilityCode == "" {
		logger.Error().Msg("missing necessary params")
		return response.System(c)
	}

	// Verify user exists and get details for logging
	user, err := h.repos.User.GetByInitialsAndFacility(c.Request().Context(), route.UserInitials, route.FacilityCode)
	if err != nil {
		logger.Error().
			Err(err).
			Str("facility", route.FacilityCode).
			Str("initials", route.UserInitials).
			Msg("failed to fetch user")
		return response.System(c)
	}

	// Check permissions using auth context from PageContext
	if !canDeleteUser(auth, user) {
		logger.Warn().
			Int("target_user_id", user.ID).
			Int("requesting_user_id", auth.UserID).
			Str("role", string(auth.Role)).
			Msg("unauthorized deletion attempt")
		return response.Error(c, http.StatusForbidden,
			"Access Denied",
			[]string{"You don't have permission to delete this user"})
	}

	// Delete user
	if err := h.repos.User.Delete(c.Request().Context(), user.ID); err != nil {
		logger.Error().
			Err(err).
			Int("user_id", user.ID).
			Msg("failed to delete user")
		return response.System(c)
	}

	logger.Info().
		Int("user_id", user.ID).
		Str("email", user.Email).
		Int("facility_id", user.FacilityID).
		Msg("user deleted successfully")

	// Handle HTMX response using facility code from route context
	return handleDeleteResponse(c, route.FacilityCode)
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

	// Get auth context
	auth, err := middleware.GetAuthContext(c)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get auth context")
		return response.System(c)
	}

	// Get route context
	route, err := middleware.GetRouteContext(c)
	if err != nil {
		logger.Error().Msg("missing route context")
		return response.System(c)
	}

	if err := ensureRouteParams(route); err != nil {
		return err
	}

	// Get existing user
	existingUser, err := h.repos.User.GetByInitialsAndFacility(c.Request().Context(), route.UserInitials, route.FacilityCode)
	if err != nil {
		return err
	}

	// Check if user can update this password
	if !canUpdatePassword(auth, existingUser.ID) {
		logger.Warn().
			Int("target_user_id", existingUser.ID).
			Int("requesting_user_id", auth.UserID).
			Str("role", string(auth.Role)).
			Msg("unauthorized password form access attempt")
		return response.Error(c,
			http.StatusForbidden,
			"Access Denied",
			[]string{"You don't have permission to update this password"},
		)
	}

	logger.Debug().
		Int("user_id", existingUser.ID).
		Msg("rendering password update form")

	return render(c, component.Update_Password_Form(route.FacilityCode, route.UserInitials))
}

// HandleUpdatePassword processes password update requests
func (h *Handler) HandleUpdatePassword(c echo.Context) error {
	logger := h.logger.With().
		Str("handler", "HandleUpdatePassword").
		Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
		Logger()

	// Get auth context
	auth, err := middleware.GetAuthContext(c)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get auth context")
		return response.System(c)
	}

	// Get route context
	route, err := middleware.GetRouteContext(c)
	if err != nil {
		logger.Error().Msg("missing route context")
		return response.System(c)
	}

	if err := ensureRouteParams(route); err != nil {
		return err
	}

	// Get existing user
	existingUser, err := h.repos.User.GetByInitialsAndFacility(c.Request().Context(), route.UserInitials, route.FacilityCode)
	if err != nil {
		return err
	}

	// Get and validate form data
	formData, err := h.validatePasswordUpdate(c, *existingUser, auth)
	if err != nil {
		return err // validatePasswordUpdate handles error responses
	}

	// Hash password
	hashedPassword, err := hashPassword(formData.Password)
	if err != nil {
		logger.Error().Err(err).Msg("failed to hash password")
		return response.Error(c,
			http.StatusInternalServerError,
			"System Error",
			[]string{"Unable to process password update"},
		)
	}

	// Update password
	if err := h.repos.User.UpdatePassword(
		c.Request().Context(),
		formData.UserID,
		string(hashedPassword),
	); err != nil {
		logger.Error().
			Err(err).
			Int("user_id", formData.UserID).
			Msg("failed to update password")
		return response.Error(c,
			http.StatusInternalServerError,
			"System Error",
			[]string{"Failed to save new password"},
		)
	}

	logger.Info().
		Int("user_id", formData.UserID).
		Str("updater_role", string(auth.Role)).
		Msg("password updated successfully")

	return response.Success(c, "Success", "Password has been updated")
}

// GetCreateUserForm renders the user creation form
func (h *Handler) GetCreateUserForm(c echo.Context) error {
	logger := h.logger.With().
		Str("handler", "GetCreateUserForm").
		Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
		Logger()

	// Get and validate facility code
	facilityCode := c.Param("facility_code")
	if facilityCode == "" {
		logger.Error().Msg("missing facility code")
		return response.Error(c,
			http.StatusBadRequest,
			"Invalid Request",
			[]string{"Facility code is required"},
		)
	}

	// Verify facility exists
	facility, err := h.repos.Facility.GetByCode(c.Request().Context(), facilityCode)
	if err != nil {
		logger.Error().
			Err(err).
			Str("facility_code", facilityCode).
			Msg("failed to find facility")
		return response.Error(c,
			http.StatusNotFound,
			"Facility Not Found",
			[]string{"The specified facility does not exist"},
		)
	}

	// Get auth context
	auth, err := middleware.GetAuthContext(c)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get auth context")
		return response.Error(c,
			http.StatusInternalServerError,
			"Authentication Error",
			[]string{"Unable to verify permissions"},
		)
	}

	// Check if user can create users for this facility
	if !canCreateUsers(auth, facility.ID) {
		logger.Warn().
			Int("facility_id", facility.ID).
			Str("user_role", string(auth.Role)).
			Msg("unauthorized attempt to access user creation form")
		return response.Error(c,
			http.StatusForbidden,
			"Access Denied",
			[]string{"You don't have permission to create users"},
		)
	}

	logger.Debug().
		Str("facility_code", facilityCode).
		Str("creator_role", string(auth.Role)).
		Msg("rendering user creation form")

	return render(c, component.CreateUserForm(facilityCode, string(auth.Role)))
}

// GetUpdateUserForm renders the form for updating a user
func (h *Handler) GetUpdateUserForm(c echo.Context) error {
	logger := h.logger.With().
		Str("handler", "GetUpdateUserForm").
		Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
		Logger()

	auth, err := middleware.GetAuthContext(c)
	if err != nil {
		logger.Error().Msg("missing auth context")
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"Authentication is required",
		)
	}

	route, err := middleware.GetRouteContext(c)
	if err != nil {
		logger.Error().Msg("missing route context")
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"Missing route context",
		)
	}

	if err := ensureRouteParams(route); err != nil {
		return err
	}

	// Get user details
	user, err := h.repos.User.GetByInitialsAndFacility(c.Request().Context(), route.UserInitials, route.FacilityCode)
	if err != nil {
		logger.Error().
			Err(err).
			Str("user_initials", route.UserInitials).
			Msg("failed to retrieve user")
		return response.System(c)
	}

	logger.Debug().
		Int("user_id", user.ID).
		Str("email", user.Email).
		Str("viewer_role", string(auth.Role)).
		Msg("rendering update form")

	return render(c, component.UpdateUserForm(*user, route.FacilityCode, *auth))
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
		return nil, response.System(c)
	}

	// Get auth context
	auth, err := middleware.GetAuthContext(c)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get auth context")
		return nil, response.Error(c,
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
		return nil, response.Error(c,
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
