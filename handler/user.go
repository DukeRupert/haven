package handler

import (
	"net/http"
	"time"

	"github.com/DukeRupert/haven/db"

	"github.com/DukeRupert/haven/view/auth"
	"github.com/DukeRupert/haven/view/component"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

// Create handles POST app/admin/:code/users
// func (h *Handler) CreateUserHandler(c echo.Context) error {
// 	logger := h.logger

// 	// Create a struct specifically for form binding
// 	var formData struct {
// 		FirstName    string `form:"first_name"`
// 		LastName     string `form:"last_name"`
// 		Initials     string `form:"initials"`
// 		Email        string `form:"email"`
// 		Password     string `form:"password"`
// 		Role         string `form:"role"` // Start as string for form binding
// 		FacilityCode string `form:"facility_code"`
// 	}

// 	// Bind the form data first
// 	if err := c.Bind(&formData); err != nil {
// 		logger.Error().
// 			Err(err).
// 			Msg("failed to bind request payload")
// 		return render(c, alert.Error(
// 			"Invalid request",
// 			[]string{"The submitted form data was invalid"},
// 		))
// 	}

// 	// Validate and collect all validation errors
// 	var errors []string
// 	var params db.CreateUserParams

// 	// Validate facility code and get facility
// 	if formData.FacilityCode == "" {
// 		errors = append(errors, "Facility code is required")
// 	} else {
// 		facility, err := h.db.GetFacilityByCode(c.Request().Context(), formData.FacilityCode)
// 		if err != nil {
// 			logger.Error().
// 				Err(err).
// 				Str("facility_code", formData.FacilityCode).
// 				Msg("failed to find facility by code")
// 			errors = append(errors, "The specified facility code was not found")
// 		} else {
// 			params.FacilityID = facility.ID
// 		}
// 	}

// 	// Validate first name
// 	if firstName, err := validation.ValidateUserName(formData.FirstName, "First name"); err != nil {
// 		errors = append(errors, err.Error())
// 	} else {
// 		params.FirstName = string(firstName)
// 	}

// 	// Validate last name
// 	if lastName, err := validation.ValidateUserName(formData.LastName, "Last name"); err != nil {
// 		errors = append(errors, err.Error())
// 	} else {
// 		params.LastName = string(lastName)
// 	}

// 	// Validate initials
// 	if initials, err := validation.ValidateUserInitials(formData.Initials); err != nil {
// 		errors = append(errors, err.Error())
// 	} else {
// 		params.Initials = string(initials)
// 	}

// 	// Validate email
// 	if email, err := validation.ValidateUserEmail(formData.Email); err != nil {
// 		errors = append(errors, err.Error())
// 	} else {
// 		params.Email = string(email)
// 	}

// 	// Validate password
// 	if password, err := validation.ValidateUserPassword(formData.Password); err != nil {
// 		errors = append(errors, err.Error())
// 	} else {
// 		params.Password = string(password)
// 	}

// 	// Validate role
// 	if validatedRole, err := validation.ValidateUserRole(formData.Role); err != nil {
// 		errors = append(errors, err.Error())
// 	} else {
// 		params.Role = validatedRole
// 	}

// 	// Return all validation errors if any exist
// 	if len(errors) > 0 {
// 		logger.Error().
// 			Strs("validation_errors", errors).
// 			Msg("validation failed")
// 		heading := "There was 1 error with your submission"
// 		if len(errors) > 1 {
// 			heading = fmt.Sprintf("There were %d errors with your submission", len(errors))
// 		}
// 		c.Response().Status = http.StatusUnprocessableEntity // Code 422
// 		return render(c, alert.Error(heading, errors))
// 	}

// 	// Hash the password
// 	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
// 	if err != nil {
// 		logger.Error().
// 			Err(err).
// 			Msg("failed to hash password")
// 		return render(c, alert.Error(
// 			"System error",
// 			[]string{"Failed to create user. Please try again later"},
// 		))
// 	}
// 	params.Password = string(hashedPassword)

// 	logger.Debug().
// 		Str("first_name", params.FirstName).
// 		Str("last_name", params.LastName).
// 		Str("initials", params.Initials).
// 		Str("email", params.Email).
// 		Str("role", string(params.Role)).
// 		Int("facility_id", params.FacilityID).
// 		Msg("attempting to create user with params")

// 	// Create the user
// 	user, err := h.db.CreateUser(c.Request().Context(), params)
// 	if err != nil {
// 		logger.Error().
// 			Err(err).
// 			Interface("params", params).
// 			Msg("failed to create user in database")
// 		c.Response().Status = http.StatusInternalServerError // Code 500
// 		return render(c, alert.Error(
// 			"System error",
// 			[]string{"Failed to create user. Please try again later"},
// 		))
// 	}

// 	route := h.RouteCtx
// 	logger.Info().
// 		Int("user_id", user.ID).
// 		Str("email", user.Email).
// 		Int("facility_id", user.FacilityID).
// 		Str("role", string(user.Role)).
// 		Msg("user created successfully")

// 	return render(c, page.UserListItem(route, *user))
// }

// Helper function to validate role
func isValidRole(role db.UserRole) bool {
	validRoles := []db.UserRole{"super", "admin", "user"}
	for _, r := range validRoles {
		if role == r {
			return true
		}
	}
	return false
}

// GetUsersByFacility handles the GET /app/:code endpoint
// func (h *Handler) GetUsersByFacility(c echo.Context) error {
// 	startTime := time.Now()
// 	logger := h.logger.With().
// 		Str("handler", "GetUsersByFacility").
// 		Str("method", "GET").
// 		Str("path", "/app/:code").
// 		Logger()

// 	auth, err := GetAuthContext(c)
// 	if err != nil {
// 		return echo.NewHTTPError(http.StatusInternalServerError, "auth context error")
// 	}
// 	LogAuthContext(h.logger, auth)

// 	// Get facility code from route parameter
// 	code := c.Param("code")
// 	if code == "" {
// 		logger.Error().Msg("facility code is missing from request")
// 		return c.JSON(http.StatusBadRequest, map[string]string{
// 			"error": "facility code is required",
// 		})
// 	}

// 	logger = logger.With().Str("facility_code", code).Logger()
// 	logger.Info().Msg("processing request to get users by facility code")

// 	// Get users from database
// 	// // Track database query duration specifically
// 	queryStartTime := time.Now()
// 	users, err := h.db.GetUsersByFacilityCode(c.Request().Context(), code)
// 	queryDuration := time.Since(queryStartTime)
// 	if err != nil {
// 		logger.Error().
// 			Err(err).
// 			Msg("failed to retrieve users from database")

// 		return c.JSON(http.StatusInternalServerError, map[string]string{
// 			"error": "failed to retrieve users",
// 		})
// 	}

// 	// If no users found, return empty array instead of null
// 	if users == nil {
// 		users = []db.User{}
// 	}

// 	// Track handler duration
// 	handlerDuration := time.Since(startTime)

// 	logger.Info().
// 		Int("user_count", len(users)).
// 		Dur("query_duration_ms", queryDuration).
// 		Dur("handler_duration_ms", handlerDuration).
// 		Msg("successfully retrieved users")

// 	route := h.RouteCtx
// 	title := "Controllers"
// 	description := "A list of all controllers assigned to the facility."
// 	return render(c, page.ShowFacilities(route, title, description, auth.Role, users))
// }

// UserPageData contains all data needed for user page rendering
type UserPageData struct {
	Title       string
	Description string
	Auth        *db.AuthContext
	User        *db.UserDetails
}

// func (h *Handler) GetUser(c echo.Context) error {
// 	auth, err := GetAuthContext(c)
// 	if err != nil {
// 		return echo.NewHTTPError(http.StatusInternalServerError, "auth context error")
// 	}

// 	details, err := h.db.GetUserDetails(
// 		c.Request().Context(),
// 		c.Param("initials"),
// 		auth.FacilityID,
// 	)
// 	if err != nil {
// 		h.logger.Error().Err(err).Msg("failed to find user details")
// 		return c.String(http.StatusNotFound, "User not found")
// 	}

// 	route := h.RouteCtx

// 	return render(c, page.UserPage(route, "Profile", "Manage your account information and schedule", auth, details))
// }

func logPageLoadSuccess(logger zerolog.Logger, details *db.UserDetails, startTime time.Time) {
	duration := time.Since(startTime)
	logger.Info().
		Int("user_id", details.User.ID).
		Str("initials", details.User.Initials).
		Int("facility_id", details.User.FacilityID).
		Bool("has_schedule", details.Schedule.ID != 0).
		Dur("duration_ms", duration).
		Float64("duration_seconds", duration.Seconds()).
		Str("performance_category", "page_load").
		Msg("user page rendered successfully")
}

func (h *Handler) CreateUserForm(c echo.Context) error {
	// Get facility code from route parameter
	code := c.Param("code")
	if code == "" {
		h.logger.Error().Msg("facility code is missing from request")
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "facility code is required",
		})
	}
	return render(c, component.CreateUserForm(code))
}

func (h *Handler) GetLogin(c echo.Context) error {
	return render(c, auth.Login())
}
