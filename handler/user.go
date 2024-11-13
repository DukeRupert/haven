package handler

import (
	"net/http"
	"time"
	"fmt"

	"github.com/DukeRupert/haven/db"
	"github.com/DukeRupert/haven/view/alert"
	"github.com/DukeRupert/haven/view/auth"
	"github.com/DukeRupert/haven/view/user"
	"github.com/DukeRupert/haven/view/admin"
	"github.com/DukeRupert/haven/view/component"
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
    if params.FirstName == "" {
        errors = append(errors, "First name is required")
    }
    if params.LastName == "" {
        errors = append(errors, "Last name is required")
    }
    if params.Initials == "" {
        errors = append(errors, "Initials are required")
    }
    if len(params.Initials) > 10 {
        errors = append(errors, "Initials must not exceed 10 characters")
    }
    if params.Email == "" {
        errors = append(errors, "Email is required")
    }
    if params.Password == "" {
        errors = append(errors, "Password is required")
    }
    if len(params.Password) < 8 {
        errors = append(errors, "Password must be at least 8 characters")
    }
    if !isValidRole(params.Role) {
        errors = append(errors, "Invalid role specified")
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

    return render(c, component.UserListItem(*user))
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

// GetUsersByFacility handles the GET /app/admin/:code endpoint
func (h *UserHandler) GetUsersByFacility(c echo.Context) error {
	startTime := time.Now()
	logger := h.logger.With().
		Str("handler", "GetUsersByFacility").
		Str("method", "GET").
		Str("path", "/app/admin/:code").
		Logger()

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

	// Track handler duration
	handlerDuration := time.Since(startTime)

	logger.Info().
		Int("user_count", len(users)).
		Dur("query_duration_ms", queryDuration).
		Dur("handler_duration_ms", handlerDuration).
		Msg("successfully retrieved users")

	return render(c, admin.ShowFacilities("Controllers", users))
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
