// internal/handler/routes.go
package handler

import (
	"github.com/DukeRupert/haven/internal/auth"
	"github.com/DukeRupert/haven/internal/context"
	"github.com/DukeRupert/haven/internal/model/types"

	"github.com/labstack/echo/v4"
)

// SetupRoutes configures all application routes
func SetupRoutes(e *echo.Echo, h *Handler, auth *auth.Middleware, authHandler *auth.Handler, ctx *context.RouteContextMiddleware) {
	// Other middleware
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

	public.GET("/", h.GetHome)
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

	// Facility specific endpoints
	facilities := e.Group("/:facility")
	facilities.Use(auth.ValidateFacility())
	setupFacilityRoutes(facilities, h, auth)
}

// setupFacilityRoutes configures facility-specific endpoints
func setupFacilityRoutes(g *echo.Group, h *Handler, auth *auth.Middleware) {
	g.GET("/calendar", h.WithNav(h.HandleCalendar))
	g.GET("/profile", h.WithNav(h.HandleProfile))
	g.GET("/:initials", h.WithNav(h.HandleProfile))
	g.POST("/available/:id", h.HandleAvailabilityToggle)
	g.POST("/schedule/:initials", h.HandleCreateSchedule)
	g.GET("/schedule/:id", h.HandleGetSchedule)
	g.POST("/schedule/:id", h.HandleUpdateSchedule)
	g.GET("/controllers", h.WithNav(h.HandleUsers), auth.RequireRole(types.UserRoleAdmin))
}

// setupAPIRoutes configures API endpoints
func setupAPIRoutes(e *echo.Echo, h *Handler, auth *auth.Middleware) {
	api := e.Group("/api")

	u := e.Group("/user")
	u.DELETE(":user_id", h.WithNav(h.HandleDeleteUser))
	u.POST("/:user_id", h.HandleUpdateUser)
	u.GET("/:user_id/update", h.GetUpdateUserForm)
	u.GET("/:user_id/password", h.GetUpdatePasswordForm)
	u.POST("/:user_id/password", h.HandleUpdatePassword)

	// Facility specific routes
	f := api.Group("/:facility", auth.ValidateFacility(), auth.RequireRole(types.UserRoleAdmin))
	f.GET("/user", h.GetCreateUserForm)
	f.POST("/user", h.HandleCreateUser)
	f.GET("/schedule/:initials", h.GetCreateScheduleForm)
	f.GET("/schedule/update/:id", h.GetUpdateScheduleForm)
}
