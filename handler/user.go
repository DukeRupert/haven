package handler

import (
	"net/http"
	"time"

	"github.com/DukeRupert/haven/db"
	"github.com/DukeRupert/haven/view/auth"
	"github.com/DukeRupert/haven/view/user"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
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

	return c.JSON(http.StatusOK, map[string]interface{}{
		"users": users,
	})
}

func (h *UserHandler) GetLogin(c echo.Context) error {
	return render(c, auth.Login())
}

func (h UserHandler) HandleUserShow(c echo.Context) error {
	return render(c, user.Show())
}
