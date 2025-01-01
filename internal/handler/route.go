// internal/handler/routes.go
package handler

import (
	"github.com/DukeRupert/haven/internal/middleware"
	"github.com/DukeRupert/haven/internal/model/types"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

/*
Public: 			/
Authenticated:		/app/
Super:				/app/facilities/
Facility:			/app/facility_code/
User:				/app/facility_code/user_initials/
*/

const (
	PathFacilityID   = "/:facility_id"
	PathFacilityCode = "/:facility_code"
	PathUserInitials = "/:user_initials"
	PathScheduleID   = "/:schedule_id"
)

func SetupRoutes(e *echo.Echo, h *Handler, m *middleware.Middleware) {
	setupGlobalMiddleware(e, h)
	setupPublicRoutes(e, h)
	setupAppRoutes(e, h, m)
}

func setupGlobalMiddleware(e *echo.Echo, h *Handler) {
	e.Pre(echoMiddleware.RemoveTrailingSlash())
	e.Use(
		echoMiddleware.RequestID(),
		middleware.RequestLogger(h.logger),
		middleware.ErrorLogger(h.logger),
		echoMiddleware.Recover(),
	)

	// CORS configuration
	e.Use(echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
		AllowOrigins: []string{h.config.BaseURL, "https://sturdy-train-vq455j4p4rwf666v-8080.app.github.dev"},
		AllowMethods: []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	// Error handling
	e.HTTPErrorHandler = CustomHTTPErrorHandler
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if err := next(c); err != nil {
				c.Echo().HTTPErrorHandler(err, c)
			}
			return nil
		}
	})
}

func setupPublicRoutes(e *echo.Echo, h *Handler) {
	e.GET("/", h.GetHome)
	e.GET("/login", h.GetLogin)
	e.POST("/login", h.LoginHandler)
	e.POST("/logout", h.LogoutHandler)
	e.GET("/register", h.GetRegistration)
	e.POST("/register", h.HandleRegistration)
	e.POST("/verify", h.InitiateEmailVerification)
	e.GET("/verify", h.GetVerificationPage)
	e.GET("/set-password", h.GetSetPassword)
	e.POST("/set-password", h.HandleSetPassword)
	e.GET("/resend-verification", h.HandleResendVerification)
}

func setupAppRoutes(e *echo.Echo, h *Handler, m *middleware.Middleware) {
   // Base app group with auth
   app := e.Group("/app", m.Auth(), m.RouteContext())
   // Complete path: /app/calendar
   app.GET("/calendar", h.HandleCalendar)
   // Complete path: /app/profile  
   app.GET("/profile", h.HandleGetUser)

   // Facility management (require super role)
   facilities := app.Group("/facilities", m.RequireRole(types.UserRoleSuper))
   {
       // Complete path: /app/facilities
       facilities.GET("", h.HandleGetFacilities)
       facilities.POST("", h.HandleCreateFacility)
       // Complete path: /app/facilities/create
       facilities.GET("/create", h.GetCreateFacilityForm)
       // Complete path: /app/facilities/edit
       facilities.GET("/edit", h.GetUpdateFacilityForm)
       // Complete path: /app/facilities/:facility_id
       facilities.PUT(PathFacilityID, h.HandleUpdateFacility)
       facilities.DELETE(PathFacilityID, h.HandleDeleteFacility)
   }

   // Facility routes (require facility access)
   facility := app.Group(PathFacilityCode, m.RequireFacilityAccess())
   {
       // Complete path: /app/:facility_code/calendar
       facility.GET("/calendar", h.HandleCalendar)
       // Complete path: /app/:facility_code/publish
       facility.PUT("/publish", h.HandleUpdatePublishedThrough)
   }

   // User management routes (requires admin role)
   users := facility.Group("/users", m.RequireRole(types.UserRoleAdmin))
   {
       // Complete path: /app/:facility_code/users
       users.GET("", h.HandleUsers)
       users.POST("", h.HandleCreateUser)
       // Complete path: /app/:facility_code/users/create
       users.GET("/create", h.GetCreateUserForm)
   }

   // User routes (requires profile access)
   user := facility.Group(PathUserInitials, m.RequireProfileAccess())
   {
       // Complete path: /app/:facility_code/:user_initials
       user.GET("", h.HandleGetUser)
       user.PUT("", h.HandleUpdateUser)
       user.DELETE("", h.HandleDeleteUser, m.RequireRole(types.UserRoleAdmin))
       // Complete path: /app/:facility_code/:user_initials/edit
       user.GET("/edit", h.GetUpdateUserForm)
       // Complete path: /app/:facility_code/:user_initials/password
       user.GET("/password", h.GetUpdatePasswordForm)
       user.PUT("/password", h.HandleUpdatePassword)
       // Complete path: /app/:facility_code/:user_initials/availability/:id
       user.POST("/availability/:id", h.HandleAvailabilityToggle)
   }

   // Schedule routes (require admin role)
   schedule := user.Group("/schedule",  m.RequireRole(types.UserRoleAdmin))
   {
       // Complete path: /app/:facility_code/:user_initials/schedule
       schedule.POST("", h.HandleCreateSchedule)
       // Complete path: /app/:facility_code/:user_initials/schedule/create
       schedule.GET("/create", h.GetCreateScheduleForm)
       // Complete path: /app/:facility_code/:user_initials/schedule/:schedule_id
       schedule.GET(PathScheduleID, h.HandleGetSchedule)
       schedule.PUT(PathScheduleID, h.HandleUpdateSchedule)
       // Complete path: /app/:facility_code/:user_initials/schedule/:schedule_id/edit
       schedule.GET("/:schedule_id/edit", h.GetUpdateScheduleForm)
   }
}
