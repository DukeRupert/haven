// internal/handler/handler.go
package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/DukeRupert/haven/internal/auth"
	"github.com/DukeRupert/haven/internal/context"
	"github.com/DukeRupert/haven/internal/model/dto"
	"github.com/DukeRupert/haven/internal/model/types"
	"github.com/DukeRupert/haven/internal/repository"
	"github.com/DukeRupert/haven/web/view/component"
	"github.com/DukeRupert/haven/web/view/page"
	"github.com/gorilla/sessions"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
)

// HandlerFunc is the type for your page handlers
type HandlerFunc func(c echo.Context, routeCtx *dto.RouteContext, navItems []dto.NavItem) error

type Config struct {
	Repos    *repository.Repositories
	Sessions sessions.Store
	Logger   zerolog.Logger
}

type Handler struct {
	repos    *repository.Repositories
	sessions sessions.Store
	logger   zerolog.Logger
	RouteCtx dto.RouteContext
}

func New(cfg Config) *Handler {
	return &Handler{
		repos:    cfg.Repos,
		sessions: cfg.Sessions,
		logger:   cfg.Logger.With().Str("component", "handler").Logger(),
	}
}

// SetupRoutes configures all application routes
func SetupRoutes(e *echo.Echo, h *Handler, auth *auth.Middleware, authHandler *auth.Handler, ctx *context.RouteContextMiddleware) {
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
	e.Use(auth.Authenticate())
	e.Use(ctx.WithRouteContext())

	// Group routes by function
	setupPublicRoutes(e, h, auth, authHandler)
	setupProtectedRoutes(e, h, auth)
	setupAPIRoutes(e, h, auth)
}

// setupPublicRoutes configures public access routes
func setupPublicRoutes(e *echo.Echo, h *Handler, auth *auth.Middleware, authHandler *auth.Handler) {
	public := e.Group("")
	public.Use(auth.EnsurePublic())

	public.GET("/", h.ShowHome)
	public.GET("/login", h.GetLogin, auth.RedirectAuthenticated())
	public.POST("/login", authHandler.LoginHandler())
	public.POST("/logout", authHandler.LogoutHandler())
	public.GET("/register", h.GetRegister)
	public.POST("/register", h.HandleRegistration)
	public.GET("/set-password", h.GetSetPassword)
	public.POST("/set-password", h.HandleSetPassword)
}

// setupProtectedRoutes configures authenticated routes
func setupProtectedRoutes(e *echo.Echo, h *Handler, auth *auth.Middleware) {
	// Super admin routes
	super := e.Group("/super")
	super.Use(auth.RequireRole(types.UserRoleSuper))
	super.GET("/facilities", h.WithNav(h.HandleFacilities))

	// Admin routes
	admin := e.Group("/admin")
	admin.Use(auth.RequireRole(types.UserRoleAdmin))
	admin.GET("/controllers", h.WithNav(h.HandleUsers))
	admin.GET("/:facility/controllers", h.WithNav(h.HandleUsers))

	// User routes
	user := e.Group("/user")
	user.Use(auth.RequireRole(types.UserRoleUser))
	user.GET("/calendar", h.WithNav(h.HandleCalendar))
	user.GET("/:facility/calendar", h.WithNav(h.HandleCalendar))
	user.GET("/profile", h.WithNav(h.HandleProfile))
	user.GET("/:facility/profile", h.WithNav(h.HandleProfile))
	user.GET("/:facility/:initials", h.WithNav(h.HandleProfile))
}

// setupAPIRoutes configures API endpoints
func setupAPIRoutes(e *echo.Echo, h *Handler, auth *auth.Middleware) {
	api := e.Group("/api")
	api.Use(auth.RequireRole(types.UserRoleUser)) // Base API authentication

	// User management endpoints - require admin role
	users := api.Group("/users")
	users.Use(auth.RequireRole(types.UserRoleAdmin))
	{
		users.POST("/:user_id", h.HandleUpdateUser)
		users.DELETE("/:user_id", h.DeleteUser)
		users.GET("/:user_id/update", h.GetUpdateUserForm)
		users.GET("/:user_id/password", h.GetUpdatePasswordForm)
		users.POST("/:user_id/password", h.HandleUpdatePassword)
	}

	// Facility specific endpoints
	facilities := api.Group("/facilities")
	facilities.Use(auth.ValidateFacility())
	setupFacilityRoutes(facilities, h)
}

// setupFacilityRoutes configures facility-specific endpoints
func setupFacilityRoutes(g *echo.Group, h *Handler) {
	g.POST("/:facility/available/:id", h.HandleAvailabilityToggle)
	g.GET("/:facility/schedule/:initials", h.GetCreateScheduleForm)
	g.POST("/:facility/schedule/:initials", h.HandleCreateSchedule)
	g.GET("/schedule/:id", h.HandleGetSchedule)
	g.POST("/schedule/:id", h.HandleUpdateSchedule)
	g.GET("/schedule/:id/update", h.GetUpdateScheduleForm)
	g.GET("/:facility/users/create", h.GetCreateUserForm)
	g.POST("/:facility/users", h.HandleCreateUser)
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
