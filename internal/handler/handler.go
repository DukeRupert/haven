// internal/handler/handler.go
package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/DukeRupert/haven/internal/auth"
	"github.com/DukeRupert/haven/internal/model/dto"
	"github.com/DukeRupert/haven/internal/model/types"
	"github.com/DukeRupert/haven/internal/repository"
	"github.com/DukeRupert/haven/internal/store"
	"github.com/DukeRupert/haven/web/view/component"
	"github.com/DukeRupert/haven/web/view/page"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
)

// Handler encapsulates dependencies for all handlers
type Handler struct {
    repos    *repository.Repositories
    sessions *store.PgxStore
    logger   zerolog.Logger
    RouteCtx dto.RouteContext
}

// HandlerFunc is the type for your page handlers
type HandlerFunc func(c echo.Context, routeCtx *dto.RouteContext, navItems []dto.NavItem) error

// Config holds handler configuration
type Config struct {
    Repos    *repository.Repositories
    Sessions *store.PgxStore
    Logger   zerolog.Logger
}

// New creates a new handler with all dependencies
func New(cfg Config) *Handler {
    return &Handler{
        repos:    cfg.Repos,
        sessions: cfg.Sessions,
        logger:   cfg.Logger.With().Str("component", "handler").Logger(),
    }
}

// SetupRoutes configures all application routes
func SetupRoutes(e *echo.Echo, h *Handler, auth *auth.AuthHandler) {
    // Middleware setup
    e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
        AllowOrigins: []string{"https://sturdy-train-vq455j4p4rwf666v-8080.app.github.dev"},
        AllowMethods: []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
        AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
    }))
    e.Use(middleware.Recover())
    e.Use(middleware.RequestID())
    e.Use(middleware.Logger())
    e.Use(session.Middleware(h.sessions))
    e.Use(auth.AuthMiddleware())
    e.Use(auth.WithRouteContext())

    // Group routes by function
    setupPublicRoutes(e, h, auth)
    setupProtectedRoutes(e, h)
    setupAPIRoutes(e, h)
}

// setupPublicRoutes configures public access routes
func setupPublicRoutes(e *echo.Echo, h *Handler, auth *auth.AuthHandler) {
    e.GET("/", h.ShowHome)
    e.GET("/login", h.GetLogin, auth.RedirectIfAuthenticated())
    e.POST("/login", auth.LoginHandler())
    e.POST("/logout", auth.LogoutHandler())
    e.GET("/register", h.GetRegister)
    e.POST("/register", h.HandleRegistration)
    e.GET("/set-password", h.GetSetPassword)
    e.POST("/set-password", h.HandleSetPassword)
}

// setupProtectedRoutes configures authenticated routes
func setupProtectedRoutes(e *echo.Echo, h *Handler) {
    // Super admin routes
    super := e.Group("/", RoleMiddleware(types.RoleSuper))
    super.GET("/facilities", h.WithNav(h.handleFacilities))

    // Admin routes
    admin := e.Group("/", RoleMiddleware(types.RoleAdmin))
    admin.GET("/controllers", h.WithNav(h.handleUsers))
    admin.GET("/:facility/controllers", h.WithNav(h.handleUsers))

    // User routes
    user := e.Group("/", RoleMiddleware(types.RoleUser))
    user.GET("/calendar", h.WithNav(h.handleCalendar))
    user.GET("/:facility/calendar", h.WithNav(h.handleCalendar))
    user.GET("/profile", h.WithNav(h.handleProfile))
    user.GET("/:facility/profile", h.WithNav(h.handleProfile))
    user.GET("/:facility/:initials", h.WithNav(h.handleProfile))
}

// setupAPIRoutes configures API endpoints
func setupAPIRoutes(e *echo.Echo, h *Handler) {
    api := e.Group("/api", API_Role_Middleware())
    
    // User management
    api.POST("/user/:user_id", h.handleUpdateUser)
    api.DELETE("/user/:user_id", h.WithNav(h.DeleteUser))
    api.GET("/user/:user_id/update", h.updateUserForm)
    api.GET("/user/:user_id/password", h.updatePasswordForm)
    api.POST("/user/:user_id/password", h.handleUpdatePassword)

    // Facility specific endpoints
    facility := api.Group("/:facility", FacilityAPIMiddleware())
    setupFacilityRoutes(facility, h)
}

// setupFacilityRoutes configures facility-specific endpoints
func setupFacilityRoutes(g *echo.Group, h *Handler) {
    g.POST("/available/:id", h.handleAvailabilityToggle)
    g.GET("/schedule/:facility/:initials", h.createScheduleForm)
    g.POST("/schedule/:facility/:initials", h.handleCreateSchedule)
    g.GET("/schedule/:id", h.handleGetSchedule)
    g.POST("/schedule/:id", h.handleUpdateSchedule)
    g.GET("/schedule/update/:id", h.updateScheduleForm)
    g.GET("/user/:facility", h.createUserForm)
    g.POST("/user", h.handleCreateUser)
}

func (h *Handler) handleFacilities(c echo.Context, routeCtx *dto.RouteContext, navItems []dto.NavItem) error {
	facs, err := h.repos.Facility.List(c.Request().Context())
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
func (h *Handler) handleCalendar(c echo.Context, routeCtx *dto.RouteContext, navItems []dto.NavItem) error {
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
	protectedDates, err := h.repos.Schedule.GetProtectedDatesByFacilityCode(c.Request().Context(), facilityCode)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "error fetching protected dates")
	}

	auth, err := GetAuthContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "auth context error")
	}

	props := dto.CalendarProps{
		CurrentMonth:   viewDate,
		FacilityCode:   facilityCode,
		ProtectedDates: protectedDates,
		UserRole:       auth.Role,
		CurrentUserID:  auth.UserID,
	}

	pageProps := dto.CalendarPageProps{
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
	protectedDate, err := h.repos.Schedule.GetProtectedDateByID(c.Request().Context(), dateID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "error fetching protected date")
	}

	// Check authorization
	if !isAuthorizedToToggle(auth.UserID, auth.Role, protectedDate) {
		return echo.NewHTTPError(http.StatusForbidden, "unauthorized to modify this protected date")
	}

	// Toggle the availability
	protectedDate, err = h.repos.Schedule.ToggleProtectedDateAvailability(c.Request().Context(), dateID)
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

func GetAuthContext(c echo.Context) (*dto.AuthContext, error) {
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

	auth := &dto.AuthContext{
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

func LogAuthContext(logger zerolog.Logger, auth *dto.AuthContext) {
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
