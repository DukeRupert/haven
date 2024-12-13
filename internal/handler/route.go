// internal/handler/routes.go
package handler

import (
	"github.com/DukeRupert/haven/internal/auth"
	"github.com/DukeRupert/haven/internal/context"
	"github.com/DukeRupert/haven/internal/model/types"

	"github.com/labstack/echo/v4"
)

/*
User Self-service: 	/api/user/*
Facility admin: 	/api/facility/:facility_id/*
Super admin: 		/api/admin/facilities/*
*/

func SetupRoutes(e *echo.Echo, h *Handler, auth *auth.Middleware, authHandler *auth.Handler, ctx *context.RouteContextMiddleware) {
	// Global middleware
	e.Use(auth.Authenticate())
	e.Use(ctx.WithRouteContext())

	// Public routes (no auth required)
	public := e.Group("")
	public.Use(auth.EnsurePublic())
	{
		public.GET("/", h.GetHome)
		public.GET("/login", h.GetLogin, auth.RedirectAuthenticated())
		public.POST("/login", authHandler.LoginHandler())
		public.POST("/logout", authHandler.LogoutHandler())
		public.GET("/register", h.GetRegister)
		public.POST("/register", h.HandleRegistration)
		public.GET("/set-password", h.GetSetPassword)
		public.POST("/set-password", h.HandleSetPassword)
	}

	// API routes
	api := e.Group("/api")
	{
		// User self-service endpoints
		self := api.Group("/user", auth.RequireRole(types.UserRoleUser))
		{
			self.GET("", h.WithNav(h.HandleGetUser))
			self.PUT("", h.HandleUpdateUser)
			self.GET("/edit", h.GetUpdateUserForm)
			self.PUT("/password", h.HandleUpdatePassword)
			self.GET("/password", h.GetUpdatePasswordForm)
			self.POST("/availability", h.HandleAvailabilityToggle)
		}

		// Facility management (super admin only)
		admin := api.Group("/admin", auth.RequireRole(types.UserRoleSuper))
		{
			admin.GET("/facilities", h.WithNav(h.HandleGetFacilities))
			admin.POST("/facilities", h.HandleCreateFacility)
			admin.PUT("/facilities/:facility_id", h.HandleUpdateFacility)
			admin.DELETE("/facilities/:facility_id", h.HandleDeleteFacility)
		}

		// Facility-specific routes
		facility := api.Group("/facility/:facility_id", auth.ValidateFacility())
		{
			// Calendar & availability
			facility.GET("/calendar", h.WithNav(h.HandleCalendar))

			// User management (admin only)
			users := facility.Group("/users", auth.RequireRole(types.UserRoleAdmin))
			{
				users.GET("", h.WithNav(h.HandleUsers))
				users.POST("", h.HandleCreateUser)
				users.GET("/new", h.GetCreateUserForm)
				users.GET("/:user_id", h.WithNav(h.HandleGetUser))
				users.PUT("/:user_id", h.HandleAdminUpdateUser)
				users.DELETE("/:user_id", h.WithNav(h.HandleDeleteUser))
				users.GET("/:user_id/edit", h.GetAdminUpdateUserForm)
				users.GET("/:user_id/password", h.GetUpdatePasswordForm)
				users.PUT("/:user_id/availability", h.HandleAdminAvailabilityToggle)
			}

			// Schedule management
			schedules := facility.Group("/schedules")
			{
				schedules.POST("", h.HandleCreateSchedule)
				schedules.GET("/new/:user_id", h.GetCreateScheduleForm)
				schedules.GET("/:schedule_id", h.HandleGetSchedule)
				schedules.PUT("/:schedule_id", h.HandleUpdateSchedule)
				schedules.GET("/:schedule_id/edit", h.GetUpdateScheduleForm)
			}
		}
	}
}
