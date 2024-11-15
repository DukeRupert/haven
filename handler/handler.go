package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/DukeRupert/haven/db"
	"github.com/DukeRupert/haven/types"
	"github.com/DukeRupert/haven/view/page"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

type Handler struct {
	db     *db.DB
	logger zerolog.Logger
	RouteCtx types.RouteContext
}

// NewSuperHandler creates a new handler with both pool and store
func NewHandler(db *db.DB, logger zerolog.Logger) *Handler {
	return &Handler{
		db:     db,
		logger: logger.With().Str("component", "superHandler").Logger(),
	}
}

func GetAuthContext(c echo.Context) (*db.AuthContext, error) {
    sess, err := session.Get("session", c)
    if err != nil {
        return nil, fmt.Errorf("failed to get session: %w", err)
    }

    userID, ok := sess.Values["user_id"].(int)
    if !ok {
        return nil, fmt.Errorf("invalid user_id in session")
    }

    role, ok := sess.Values["role"].(db.UserRole)
    if !ok {
        return nil, fmt.Errorf("invalid role in session")
    }

    auth := &db.AuthContext{
        UserID: userID,
        Role:   role,
    }

    // Optional values
    if initials, ok := sess.Values["initials"].(string); ok {
        auth.Initials = initials
    }
    if facilityID, ok := sess.Values["facility_id"].(int); ok {
        auth.FacilityID = facilityID
    }
    if facilityCode, ok := sess.Values["facility_code"].(string); ok {
        auth.FacilityCode = facilityCode
    }

    return auth, nil
}

func LogAuthContext(logger zerolog.Logger, auth *db.AuthContext) {
    logEvent := logger.Debug().
        Int("user_id", auth.UserID).
        Str("role", string(auth.Role))

    if auth.Initials != "" {
        logEvent.Str("initials", auth.Initials)
    }
    if auth.FacilityID != 0 {
        logEvent.Int("facility_id", auth.FacilityID)
    }
    if auth.FacilityCode != "" {
        logEvent.Str("facility_code", auth.FacilityCode)
    }

    logEvent.Msg("auth context retrieved")
}

func (h *Handler) ShowHome(c echo.Context) error {
	return render(c, page.Landing())
}

// PlaceholderMessage handles rendering a simple string message
func (h *Handler) PlaceholderMessage(c echo.Context) error {
	// Here you would typically have your component.PlaceholderMessage
	// For this example, we'll return the raw message
	return c.String(http.StatusOK, "Fix me. I need some love.")
}


type CreateScheduleRequest struct {
	FirstWeekday  int    `json:"first_weekday" form:"first_weekday"`
	SecondWeekday int    `json:"second_weekday" form:"second_weekday"`
	StartDate     string `json:"start_date" form:"start_date"`
}

func (h *Handler) CreateSchedule(c echo.Context) error {
	logger := h.logger
	
	var params db.CreateScheduleByCodeParams
    params.FacilityCode = h.RouteCtx.FacilityCode
    params.UserInitials = h.RouteCtx.UserInitials
    
    // Parse other parameters from request body
    var requestBody struct {
        FirstDay    time.Weekday `json:"first_day" validate:"required,min=0,max=6"`
        SecondDay   time.Weekday `json:"second_day" validate:"required,min=0,max=6"`
        StartDate   time.Time    `json:"start_date" validate:"required"`
    }
    
    if err := c.Bind(&requestBody); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, err.Error())
    }
    
    params.FirstDay = requestBody.FirstDay
    params.SecondDay = requestBody.SecondDay
    params.StartDate = requestBody.StartDate
    
    schedule, err := h.db.CreateScheduleByCode(c.Request().Context(), params)
    if err != nil {
        h.logger.Error().Err(err).
            Str("facility_code", params.FacilityCode).
            Str("user_initials", params.UserInitials).
            Msg("Failed to create schedule")
        return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create schedule")
    }

	// Get session data
	auth, err := GetAuthContext(c)
    if err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, "auth context error")
    }

	route := h.RouteCtx

	logger.Info().
		Int("schedule_id", schedule.ID).
		Str("facility_code", params.FacilityCode).
		Str("user_initials", params.UserInitials).
		Time("start_date", schedule.StartDate).
		Msgf("schedule created successfully with weekdays %s and %s",
			schedule.FirstWeekday, schedule.SecondWeekday)

	// Return the updated schedule card
	return render(c, page.ScheduleCard(route, *auth, *schedule))	
}
