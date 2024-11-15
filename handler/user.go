package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/DukeRupert/haven/db"
	"github.com/DukeRupert/haven/validation"
	"github.com/DukeRupert/haven/view/alert"
	"github.com/DukeRupert/haven/view/auth"
	"github.com/DukeRupert/haven/view/component"
	"github.com/DukeRupert/haven/view/page"
	"github.com/DukeRupert/haven/view/user"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	db     *db.DB
	logger zerolog.Logger
}

// NewUserHandler creates a new handler with both pool and store
func NewUserHandler(db *db.DB, logger zerolog.Logger) *UserHandler {
	return &UserHandler{
		db:     db,
		logger: logger.With().Str("component", "userHandler").Logger(),
	}
}

// Create handles POST app/admin/:code/users
func (h *UserHandler) CreateUser(c echo.Context) error {
	logger := h.logger

	var params db.CreateUserParams
	if err := c.Bind(&params); err != nil {
		logger.Error().
			Err(err).
			Msg("failed to bind request payload")

		return render(c, alert.Error(
			"Invalid request",
			[]string{"The submitted form data was invalid"},
		))
	}

	// Get facility code from request
	facilityCode := c.FormValue("facility_code")
	if facilityCode == "" {
		return render(c, alert.Error(
			"Invalid request",
			[]string{"Facility code is required"},
		))
	}

	// Look up facility ID by code
	facility, err := h.db.GetFacilityByCode(c.Request().Context(), facilityCode)
	if err != nil {
		logger.Error().
			Err(err).
			Str("facility_code", facilityCode).
			Msg("failed to find facility by code")
		return render(c, alert.Error(
			"Invalid facility",
			[]string{"The specified facility code was not found"},
		))
	}

	// Set the facility ID in the params
	params.FacilityID = facility.ID

	// Collect validation errors
	var errors []string

	// Validate first name
	firstName, err := validation.ValidateUserName(params.FirstName, "First name")
	if err != nil {
		errors = append(errors, err.Error())
	} else {
		params.FirstName = string(firstName)
	}

	// Validate last name
	lastName, err := validation.ValidateUserName(params.LastName, "Last name")
	if err != nil {
		errors = append(errors, err.Error())
	} else {
		params.LastName = string(lastName)
	}

	// Validate initials
	initials, err := validation.ValidateUserInitials(params.Initials)
	if err != nil {
		errors = append(errors, err.Error())
	} else {
		params.Initials = string(initials)
	}

	// Validate email
	email, err := validation.ValidateUserEmail(params.Email)
	if err != nil {
		errors = append(errors, err.Error())
	} else {
		params.Email = string(email)
	}

	// Validate password
	password, err := validation.ValidateUserPassword(params.Password)
	if err != nil {
		errors = append(errors, err.Error())
	} else {
		params.Password = string(password)
	}

	// Validate role
	validatedRole, err := validation.ValidateUserRole(string(params.Role))
	if err != nil {
		errors = append(errors, err.Error())
	} else {
		params.Role = validatedRole
	}

	if len(errors) > 0 {
		logger.Error().
			Strs("validation_errors", errors).
			Msg("validation failed")

		heading := "There was 1 error with your submission"
		if len(errors) > 1 {
			heading = fmt.Sprintf("There were %d errors with your submission", len(errors))
		}

		return render(c, alert.Error(heading, errors))
	}

	// Hash the password before storing
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error().
			Err(err).
			Msg("failed to hash password")
		return render(c, alert.Error("System error",
			[]string{"Failed to create user. Please try again later"}))
	}
	params.Password = string(hashedPassword)

	// Create the user
	user, err := h.db.CreateUser(c.Request().Context(), params)
	if err != nil {
		logger.Error().
			Err(err).
			Interface("params", params).
			Msg("failed to create user in database")
		return render(c, alert.Error("System error",
			[]string{"Failed to create user. Please try again later"}))
	}

	logger.Info().
		Int("user_id", user.ID).
		Str("email", user.Email).
		Int("facility_id", user.FacilityID).
		Str("role", string(user.Role)).
		Msg("user created successfully")

	return render(c, page.UserListItem(*user))
}

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
func (h *UserHandler) GetUsersByFacility(c echo.Context) error {
	startTime := time.Now()
	logger := h.logger.With().
		Str("handler", "GetUsersByFacility").
		Str("method", "GET").
		Str("path", "/app/:code").
		Logger()

	auth, err := GetAuthContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "auth context error")
	}
	LogAuthContext(h.logger, auth)

	// Get facility code from route parameter
	code := c.Param("code")
	if code == "" {
		logger.Error().Msg("facility code is missing from request")
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "facility code is required",
		})
	}

	logger = logger.With().Str("facility_code", code).Logger()
	logger.Info().Msg("processing request to get users by facility code")

	// Get users from database
	// // Track database query duration specifically
	queryStartTime := time.Now()
	users, err := h.db.GetUsersByFacilityCode(c.Request().Context(), code)
	queryDuration := time.Since(queryStartTime)
	if err != nil {
		logger.Error().
			Err(err).
			Msg("failed to retrieve users from database")

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to retrieve users",
		})
	}

	// If no users found, return empty array instead of null
	if users == nil {
		users = []db.User{}
	}

	// Get session
	sess, err := session.Get("session", c)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get session")
		return echo.NewHTTPError(http.StatusInternalServerError, "session error")
	}

	// Check role if present
	role, ok := sess.Values["role"].(db.UserRole)
	if !ok {
		logger.Debug().Str("role", string(role)).Msg("no valid role in session")
		role = "user"
	}

	// Track handler duration
	handlerDuration := time.Since(startTime)

	logger.Info().
		Int("user_count", len(users)).
		Dur("query_duration_ms", queryDuration).
		Dur("handler_duration_ms", handlerDuration).
		Msg("successfully retrieved users")

	title := "Controllers"
	description := "A list of all controllers assigned to the facility."
	return render(c, page.ShowFacilities(title, description, role, users))
}

type pageContext struct {
	facility *db.Facility
	details  *db.UserDetails
	userID   int
	role     db.UserRole // Changed from db.UserRole to db.UserRole
}

func (h *UserHandler) GetUser(c echo.Context) error {
	startTime := time.Now()
	ctx := c.Request().Context()

	// Setup structured logging
	logger := h.logger.With().
		Str("handler", "UserPage").
		Str("method", "GET").
		Str("path", "/app/facility/:code/user/:initials").
		Str("facility_code", c.Param("code")).
		Str("initials", c.Param("initials")).
		Logger()

	logger.Info().Msg("processing request to get user details")

	// Gather all required data
	pc, err := h.gatherPageData(ctx, c)
	if err != nil {
		return err
	}

	// Construct view data
	viewData := h.constructViewData(pc)

	// Log success with performance metrics
	logPageLoadSuccess(logger, pc.details, startTime)

	return render(c, page.UserPage(viewData.pageData, viewData.userData, *pc.details))
}

func (h *UserHandler) gatherPageData(ctx context.Context, c echo.Context) (*pageContext, error) {
	// Get facility
	facility, err := h.db.GetFacilityByCode(ctx, c.Param("code"))
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to find facility")
		return nil, c.String(http.StatusNotFound, "Facility not found")
	}

	// Get user details
	details, err := h.db.GetUserDetails(ctx, c.Param("initials"), facility.ID)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to find user details")
		return nil, c.String(http.StatusNotFound, "User not found")
	}

	// Get session data
	sess, err := session.Get("session", c)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get session")
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "session error")
	}

	return &pageContext{
		facility: facility,
		details:  details,
		userID:   sess.Values["user_id"].(int),
		role:     sess.Values["role"].(db.UserRole), // Changed to db.UserRole
	}, nil
}

type viewData struct {
	pageData page.PageData
	userData page.UserData
}

func (h *UserHandler) constructViewData(pc *pageContext) viewData {
	return viewData{
		pageData: page.PageData{
			Title:       "Profile",
			Description: "Manage your account information and schedule",
		},
		userData: page.UserData{
			ID:   pc.userID,
			Role: pc.role, // Make sure user.UserData expects db.UserRole
		},
	}
}

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

func (h *UserHandler) CreateUserForm(c echo.Context) error {
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

func (h *UserHandler) GetLogin(c echo.Context) error {
	return render(c, auth.Login())
}

func (h UserHandler) HandleUserShow(c echo.Context) error {
	return render(c, user.Show())
}
