package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/DukeRupert/haven/auth"
	"github.com/DukeRupert/haven/db"
	"github.com/DukeRupert/haven/store"
	"github.com/DukeRupert/haven/types"
	"github.com/DukeRupert/haven/view/component"
	"github.com/DukeRupert/haven/view/page"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
)

type Handler struct {
	db       *db.DB
	logger   zerolog.Logger
	RouteCtx types.RouteContext
}

// HandlerFunc is the type for your page handlers
type HandlerFunc func(c echo.Context, routeCtx *types.RouteContext, navItems []types.NavItem) error

// NewHandler creates a new handler with both pool and store
func NewHandler(db *db.DB, logger zerolog.Logger) *Handler {
	return &Handler{
		db:     db,
		logger: logger.With().Str("component", "Handler").Logger(),
	}
}

func SetupRoutes(e *echo.Echo, h *Handler, auth *auth.AuthHandler, store *store.PgxStore) {
	// Define static assets
	e.Static("/static", "assets")

	// Apply global middleware
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.Logger())
	e.Use(session.Middleware(store))
	e.Use(auth.AuthMiddleware())
	e.Use(auth.WithRouteContext())

	// Public routes
	e.GET("/", h.ShowHome)
	e.GET("/login", h.GetLogin, auth.RedirectIfAuthenticated())
	e.POST("/login", auth.LoginHandler())
	e.POST("/logout", auth.LogoutHandler())
	e.GET("/register", h.GetRegister)
	e.POST("/register", h.HandleRegistration)
	e.GET("/set-password", h.GetSetPassword)
	e.POST("/set-password", h.HandleSetPassword)

	// Protected routes
	// Super-only routes
	e.GET("/facilities", h.WithNav(h.handleFacilities), RouteMiddleware("/facilities"))

	// Admin and above routes
	e.GET("/controllers", h.WithNav(h.handleUsers), RouteMiddleware("/controllers"))
	e.GET("/:facility/controllers", h.WithNav(h.handleUsers), RouteMiddleware("/controllers"))

	// User and above routes
	e.GET("/calendar", h.WithNav(h.handleCalendar), RouteMiddleware("/calendar"))
	e.GET("/:facility/calendar", h.WithNav(h.handleCalendar), RouteMiddleware("/calendar"))
	e.GET("/profile", h.WithNav(h.handleProfile), RouteMiddleware("/profile"))
	e.GET("/:facility/profile", h.WithNav(h.handleProfile), RouteMiddleware("/profile"))
	e.GET("/:facility/:initials", h.WithNav(h.handleProfile), RouteMiddleware("/profile"))

	// API routes
	api := e.Group("/api", API_Role_Middleware())
	// Todo: Super only facility endpoints
	// api.POST("/facility", h.handleCreateFacility)
	// api.PUT("/facility/:id", h.handleUpdateFacility)
	api.GET("/facility/create", h.CreateFacilityForm)
	// api.GET("/facility/update", h.updateFacilityForm)
	api.POST("/user/:user_id", h.handleUpdateUser)
	api.GET("/user/:user_id/update", h.updateUserForm)
	api.GET("/user/:user_id/password", h.updatePasswordForm)
	api.POST("/user/:user_id/password", h.handleUpdatePassword)

	// Facility specific endpoints go here
	facility := api.Group("/:facility", FacilityAPIMiddleware())
	facility.POST("/available/:id", h.handleAvailabilityToggle)
	facility.GET("/schedule/:facility/:initials", h.createScheduleForm)
	facility.POST("/schedule/:facility/:initials", h.handleCreateSchedule)
	facility.GET("/schedule/:id", h.handleGetSchedule)
	facility.POST("/schedule/:id", h.handleUpdateSchedule)
	facility.GET("/schedule/update/:id", h.updateScheduleForm)
	facility.GET("/user/:facility", h.createUserForm)
	facility.POST("/user", h.handleCreateUser)
	facility.DELETE("/user/:id", h.DeleteUser)
}

func (h *Handler) handleFacilities(c echo.Context, routeCtx *types.RouteContext, navItems []types.NavItem) error {
	facs, err := h.db.ListFacilities(c.Request().Context())
	if err != nil {
		// You might want to implement a custom error handler
		return echo.NewHTTPError(http.StatusInternalServerError,
			"Failed to retrieve facilities")
	}

	title := "Facilities"
	description := "A list of all facilities including their name and code."

	component := page.Facilities(
		*routeCtx,
		navItems,
		title,
		description,
		facs,
	)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

func (h *Handler) GetLogin(c echo.Context) error {
	return render(c, page.Login())
}

// handleCalendar handles GET requests for the calendar view
func (h *Handler) handleCalendar(c echo.Context, routeCtx *types.RouteContext, navItems []types.NavItem) error {
	// Get facility code from path parameter
	facilityCode := c.Param("facility")
	if facilityCode == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "facility code is required")
	}

	// Parse the month query parameter
	monthStr := c.QueryParam("month")
	var viewDate time.Time
	var err error

	if monthStr != "" {
		// Parse YYYY-MM format
		viewDate, err = time.Parse("2006-01", monthStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid month format")
		}
	} else {
		// Default to current month
		now := time.Now()
		viewDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	}

	// Get protected dates for the month
	protectedDates, err := h.db.GetProtectedDatesByFacilityCode(c.Request().Context(), facilityCode)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "error fetching protected dates")
	}

	auth, err := GetAuthContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "auth context error")
	}

	props := types.CalendarProps{
		CurrentMonth:   viewDate,
		FacilityCode:   facilityCode,
		ProtectedDates: protectedDates,
		UserRole:       auth.Role,
		CurrentUserID:  auth.UserID,
	}

	pageProps := types.CalendarPageProps{
		Route:       *routeCtx,
		NavItems:    navItems,
		Auth:        *auth,
		Title:       "Calendar",
		Description: "View the facility calendar",
		Calendar:    props,
	}

	// Render only the calendar component for HTMX requests
	if c.Request().Header.Get("HX-Request") == "true" {
		return component.Calendar(props).Render(c.Request().Context(), c.Response().Writer)
	}

	// For regular requests, render the full page (assuming you have a layout)
	return page.CalendarPage(pageProps).Render(c.Request().Context(), c.Response().Writer)
}

// Updated handler function
func (h *Handler) handleAvailabilityToggle(c echo.Context) error {
	// Parse the protected date ID from the URL
	dateID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid protected date ID")
	}

	// Get user from context
	auth, err := GetAuthContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "auth context error")
	}

	// Fetch the protected date to check ownership
	protectedDate, err := h.db.GetProtectedDate(c.Request().Context(), dateID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "error fetching protected date")
	}

	// Check authorization
	if !isAuthorizedToToggle(auth.UserID, auth.Role, protectedDate) {
		return echo.NewHTTPError(http.StatusForbidden, "unauthorized to modify this protected date")
	}

	// Toggle the availability
	protectedDate, err = h.db.ToggleProtectedDateAvailability(c.Request().Context(), dateID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "error toggling availability")
	}

	component := component.ProtectedDay(
		auth.UserID,
		auth.FacilityCode,
		protectedDate,
	)

	return component.Render(c.Request().Context(), c.Response().Writer)
}

func GetAuthContext(c echo.Context) (*types.AuthContext, error) {
	sess, err := session.Get("session", c)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	userID, ok := sess.Values["user_id"].(int)
	if !ok {
		return nil, fmt.Errorf("invalid user_id in session")
	}

	role, ok := sess.Values["role"].(types.UserRole)
	if !ok {
		return nil, fmt.Errorf("invalid role in session")
	}

	auth := &types.AuthContext{
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

func LogAuthContext(logger zerolog.Logger, auth *types.AuthContext) {
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
