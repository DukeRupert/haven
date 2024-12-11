// internal/handler/handler.go
package handler

import (
	"net/http"
	"strconv"

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
	Auth     *auth.Middleware
	Sessions sessions.Store
	Logger   zerolog.Logger
}

type Handler struct {
	repos    *repository.Repositories
	auth     *auth.Middleware
	sessions sessions.Store
	logger   zerolog.Logger
	RouteCtx dto.RouteContext
}

func New(cfg Config) *Handler {
	return &Handler{
		repos:    cfg.Repos,
		auth:     cfg.Auth,
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
	users.POST("/:user_id", h.HandleUpdateUser)
	users.DELETE("/:user_id", h.WithNav(h.HandleDeleteUser))
	users.GET("/:user_id/update", h.GetUpdateUserForm)
	users.GET("/:user_id/password", h.GetUpdatePasswordForm)
	users.POST("/:user_id/password", h.HandleUpdatePassword)

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

// Updated handler function

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
