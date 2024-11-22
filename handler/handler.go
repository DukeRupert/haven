package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/DukeRupert/haven/auth"
	"github.com/DukeRupert/haven/store"
	"github.com/DukeRupert/haven/db"
	"github.com/DukeRupert/haven/types"
	"github.com/DukeRupert/haven/view/component"
	"github.com/DukeRupert/haven/view/page"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	db       *db.DB
	logger   zerolog.Logger
	RouteCtx types.RouteContext
}

// HandlerFunc is the type for your page handlers
type HandlerFunc func(c echo.Context, routeCtx *types.RouteContext, navItems []types.NavItem) error

// WithNav wraps your handler functions with common navigation logic
func (h *Handler) WithNav(fn HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		logger := h.logger.With().
			Str("middleware", "WithNav").
			Str("path", c.Request().URL.Path).
			Logger()

		// Get route context
		routeCtx, err := GetRouteContext(c)
		if err != nil {
			logger.Error().Err(err).Msg("failed to get route context")
			return err
		}

		// Build navigation
		currentPath := c.Request().URL.Path
		navItems := BuildNav(routeCtx, currentPath)

		// Call the wrapped handler
		return fn(c, routeCtx, navItems)
	}
}

// NewSuperHandler creates a new handler with both pool and store
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

	// Define all routes - public and protected
	// Public routes (these match the publicRoutes map in the middleware)
	e.GET("/", h.ShowHome)
	e.GET("/login", h.GetLogin, auth.RedirectIfAuthenticated())
	e.POST("/login", auth.LoginHandler())
	e.POST("/logout", auth.LogoutHandler())

	// Protected routes
	e.GET("/dashboard", h.WithNav(h.handleDashboard))
	e.GET("/:facility/dashboard", h.WithNav(h.handleDashboard))
	e.GET("/controllers", h.WithNav(h.handleUsers))
	e.GET("/:facility/controllers", h.WithNav(h.handleUsers))
	e.GET("/profile", h.WithNav(h.handleProfile))
	e.GET("/:facility/profile", h.WithNav(h.handleProfile))
	e.GET("/:facility/:initials", h.WithNav(h.handleProfile))
	e.GET("/calendar", h.WithNav(h.handleCalendar))
	e.GET("/:facility/calendar", h.WithNav(h.handleCalendar))

	// Api routes
	api := e.Group("/api")
	api.POST("/available/:id", h.handleAvailabilityToggle)
	api.GET("/schedule/:facility/:initials", h.createScheduleForm)
	api.POST("/schedule/:facility/:initials", h.handleCreateSchedule)
	api.POST("/schedule/:id", h.handleUpdateSchedule)
	api.GET("/schedule/update/:id", h.updateScheduleForm)
	api.GET("/user/:facility", h.createUserForm)
	api.POST("/user", h.handleCreateUser)

	// protected.GET("/calendar", h.WithNav(h.handleCalendar))
	// protected.GET("/:facility/calendar", h.WithNav(h.handleCalendar))
}

func wrapHandler(route types.Route, handler echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		routeCtx, err := GetRouteContext(c)
		if err != nil {
			return err
		}

		// If route needs facility and we don't have one, redirect to non-facility version
		if route.NeedsFacility && routeCtx.FacilityCode == "" {
			return c.Redirect(http.StatusTemporaryRedirect, route.Path)
		}

		// If we're on non-facility route but have a facility, redirect to facility version
		if route.NeedsFacility && c.Param("facility") == "" && routeCtx.FacilityCode != "" {
			return c.Redirect(http.StatusTemporaryRedirect,
				BuildRoutePath(routeCtx, route.Path))
		}

		return handler(c)
	}
}

func BuildNav(routeCtx *types.RouteContext, currentPath string) []types.NavItem {
	logger := log.With().
		Str("function", "BuildNav").
		Str("current_path", currentPath).
		Str("facility_code", routeCtx.FacilityCode).
		Logger()

	// Strip facility prefix for comparison
	strippedPath := currentPath
	if routeCtx.FacilityCode != "" {
		prefix := "/" + routeCtx.FacilityCode
		strippedPath = strings.TrimPrefix(currentPath, prefix)
	}

	logger.Debug().
		Str("stripped_path", strippedPath).
		Msg("building navigation items")

	navItems := []types.NavItem{
		{
			Path:    buildPath(routeCtx.FacilityCode, "/dashboard"),
			Name:    "Dashboard",
			Active:  strippedPath == "/dashboard",
			Visible: true,
		},
		{
			Path:    buildPath(routeCtx.FacilityCode, "/calendar"),
			Name:    "Calendar",
			Active:  strippedPath == "/calendar",
			Visible: true,
		},
		{
			Path:    buildPath(routeCtx.FacilityCode, "/controllers"),
			Name:    "Controllers",
			Active:  strippedPath == "/controllers",
			Visible: true,
		},
		{
			Path:    buildPath(routeCtx.FacilityCode, "/profile"),
			Name:    "Profile",
			Active:  strippedPath == "/profile",
			Visible: true,
		},
	}

	// Debug logging
	// for _, item := range navItems {
	// 	logger.Debug().
	// 		Str("name", item.Name).
	// 		Str("path", item.Path).
	// 		Bool("active", item.Active).
	// 		Str("compare_path", strippedPath).
	// 		Msg("nav item state")
	// }

	return navItems
}

// Helper function to build correct paths
func buildPath(facilityCode string, path string) string {
	if facilityCode == "" {
		return path
	}
	return fmt.Sprintf("/%s%s", facilityCode, path)
}

func (h *Handler) handleDashboard(c echo.Context, routeCtx *types.RouteContext, navItems []types.NavItem) error {
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
		protectedDate,
	)

	return component.Render(c.Request().Context(), c.Response().Writer)
}

// isAuthorizedToToggle checks if a user is authorized to toggle a protected date's availability
func isAuthorizedToToggle(userID int, role types.UserRole, protectedDate types.ProtectedDate) bool {
	// Allow access if user is admin or super
	if role == "admin" || role == "super" {
		return true
	}

	// Allow access if user owns the protected date
	if role == "user" && userID == protectedDate.UserID {
		return true
	}

	return false
}

// func (h *Handler) handleProfile(c echo.Context) error {
//     routeCtx, err := GetRouteContext(c)
//     if err != nil {
//         return err
//     }

//     // Your calendar logic here
//     return c.Render(http.StatusOK, "calendar.html", map[string]interface{}{
//         "routeCtx": routeCtx,
//         "navItems": BuildNav(*routeCtx, c.Path()),
//     })
// }

// Main route handler that dispatches to specific handlers
func (h *Handler) handleRoute(c echo.Context) error {
	logger := h.logger.With().
		Str("handler", "handleRoute").
		Str("request_path", c.Request().URL.Path).
		Str("echo_path", c.Path()).
		Logger()

	logger.Debug().Msg("starting route handler")

	// Use Request().URL.Path instead of c.Path() to get actual path
	path := c.Request().URL.Path

	// Strip facility prefix if present
	if facility := c.Param("facility"); facility != "" {
		logger.Debug().
			Str("facility", facility).
			Str("original_path", path).
			Msg("stripping facility prefix from path")

		path = strings.TrimPrefix(path, "/"+facility)

		logger.Debug().
			Str("stripped_path", path).
			Msg("facility prefix stripped")
	}

	logger.Debug().
		Str("final_path", path).
		Msg("routing request")

	var handler string
	var handlerFunc HandlerFunc

	switch path {
	case "/dashboard":
		handler = "handleDashboard"
		logger.Debug().Msg("routing to dashboard handler")
		handlerFunc = h.handleDashboard

	case "/controllers":
		handler = "handleControllers"
		logger.Debug().Msg("routing to controller handler")
		handlerFunc = h.handleUsers

	default:
		logger.Warn().
			Str("path", path).
			Msg("no route match found")
		return echo.NewHTTPError(http.StatusNotFound, "Page not found")
	}

	// Wrap the handler with navigation and execute it
	err := h.WithNav(handlerFunc)(c)
	if err != nil {
		logger.Error().
			Err(err).
			Str("handler", handler).
			Str("path", path).
			Msg("handler returned error")
		return err
	}

	logger.Debug().
		Str("handler", handler).
		Str("path", path).
		Msg("handler completed successfully")

	return nil
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

// PlaceholderMessage handles rendering a simple string message
func (h *Handler) PlaceholderMessage(c echo.Context) error {
	// Here you would typically have your component.PlaceholderMessage
	// For this example, we'll return the raw message
	return c.String(http.StatusOK, "Fix me. I need some love.")
}
