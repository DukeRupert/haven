// internal/handler/routes.go
package handler

import (
	"github.com/DukeRupert/haven/internal/auth"
	"github.com/DukeRupert/haven/internal/context"
	"github.com/DukeRupert/haven/internal/model/types"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

/*
User Self-service: 	/api/user/*
Facility admin: 	/api/facility/:facility_id/*
Super admin: 		/api/admin/facilities/*
*/

func SetupRoutes(e *echo.Echo, h *Handler, auth *auth.Middleware, authHandler *auth.Handler, ctx *context.RouteContextMiddleware) {
	// Global middleware
	e.Pre(middleware.RemoveTrailingSlash())
	e.Static("/static", "web/assets")
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"https://sturdy-train-vq455j4p4rwf666v-8080.app.github.dev"},
		AllowMethods: []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(auth.Auth())
	routeCtxMiddleware := context.NewRouteContextMiddleware(h.logger)
	e.Use(routeCtxMiddleware.WithRouteContext())
	e.HTTPErrorHandler = CustomHTTPErrorHandler
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err != nil {
				c.Echo().HTTPErrorHandler(err, c)
			}
			return nil
		}
	})

	// Public routes - no group or additional middleware needed
	e.GET("/", h.GetHome)
	e.GET("/login", h.GetLogin, auth.RedirectAuthenticated())
	e.POST("/login", authHandler.LoginHandler())
	e.POST("/logout", authHandler.LogoutHandler())
	e.GET("/register", h.GetRegistration)
	e.POST("/register", h.HandleRegistration)
	e.POST("/verify", h.InitiateEmailVerification)
	e.GET("/verify", h.GetVerificationPage)
	e.GET("/set-password", h.GetSetPassword)
	e.POST("/set-password", h.HandleSetPassword)

	// Protected routes

	/* 	User self-service endpoints
	/profile
	*/
	self := e.Group("/profile", auth.RequireRole(types.UserRoleUser))
	{
		self.GET("", h.WithNav(h.HandleGetUser))
		self.PUT("/:user_id", h.HandleUpdateUser)
		self.GET("/:user_id/edit", h.GetUpdateUserForm)
		self.PUT("/:user_id/password", h.HandleUpdatePassword)
		self.GET("/:user_id/password", h.GetUpdatePasswordForm)
		self.POST("/availability/:id", h.HandleAvailabilityToggle)
	}

	/* 	Facility specific endpoints (admin only)
	/facility/:facility_id
	*/
	facility := e.Group("/facility/:facility_id", auth.ValidateFacility())
	facility.GET("/calendar", h.WithNav(h.HandleCalendar))

	/*
		/facility/:facility_id/users
	*/
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
	}

	// Facility Management (super only)
	e.GET("/facilities", h.WithNav(h.HandleGetFacilities), auth.RequireRole(types.UserRoleSuper))

	// API routes
	api := e.Group("/api")
	{

		/* User self-service endpoints
		/api/user/:user_id
		*/
		self := api.Group("/user/:user_id", auth.RequireRole(types.UserRoleUser))
		{
			self.GET("", h.WithNav(h.HandleGetUser))
			self.PUT("", h.HandleUpdateUser)
			self.DELETE("", h.WithNav(h.HandleDeleteUser))
			self.GET("/edit", h.GetUpdateUserForm)
			self.PUT("/password", h.HandleUpdatePassword)
			self.GET("/password", h.GetUpdatePasswordForm)
			self.POST("/availability/:id", h.HandleAvailabilityToggle)
		}

		/* Facility management (super admin only
		/api/admin/facilities
		*/
		admin := api.Group("/admin", auth.RequireRole(types.UserRoleSuper))
		{
			admin.POST("/facilities", h.HandleCreateFacility)
			admin.GET("/facilities/new", h.GetCreateFacilityForm)
			admin.GET("/facilities/edit", h.GetUpdateFacilityForm)
			admin.PUT("/facilities/:facility_id", h.HandleUpdateFacility)
			admin.DELETE("/facilities/:facility_id", h.HandleDeleteFacility)
		}

		/* Facility-specific routes
		/api/facility/:facility_id
		*/
		facility := api.Group("/facility/:facility_id", auth.ValidateFacility())
		{

			/* User management (admin only)
			/api/facility/:facility_id/users
			*/
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
			}

			/* Schedule management
			/api/facility/:facility_id/schedule
			*/
			schedule := facility.Group("/schedule")
			{
				schedule.POST("", h.HandleCreateSchedule)
				schedule.GET("/new/:user_id", h.GetCreateScheduleForm)
				schedule.GET("/:schedule_id", h.HandleGetSchedule)
				schedule.PUT("/:schedule_id", h.HandleUpdateSchedule)
				schedule.GET("/:schedule_id/edit", h.GetUpdateScheduleForm)
			}
		}
	}
}
