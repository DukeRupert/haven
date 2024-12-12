package auth

import (
    "fmt"
    "net/http"
    "time"

    "github.com/DukeRupert/haven/internal/store"
    "github.com/DukeRupert/haven/web/view/alert"

    "github.com/labstack/echo-contrib/session"
    "github.com/labstack/echo/v4"
    "github.com/rs/zerolog"
)

// Handler configuration
type HandlerConfig struct {
    Service *Service
    Logger  zerolog.Logger
}

// Handler struct
type Handler struct {
    service *Service
    logger  zerolog.Logger
}

// NewHandler creates a new auth handler
func NewHandler(cfg HandlerConfig) *Handler {
    return &Handler{
        service: cfg.Service,
        logger:  cfg.Logger.With().Str("component", "auth_handler").Logger(),
    }
}

// LoginParams represents the expected login request body
type LoginParams struct {
	Email    string `json:"email" form:"email" validate:"required,email"`
	Password string `json:"password" form:"password" validate:"required"`
}

// LoginResponse handles both the alert and potential redirect for login attempts
func (h *Handler) LoginResponse(c echo.Context, status int, heading string, messages []string, redirectURL string) error {
	c.Response().Status = status

	// For successful login with redirect
	if status == http.StatusOK && redirectURL != "" {
		c.Response().Header().Set("HX-Redirect", redirectURL)
		return c.String(http.StatusOK, "")
	}

	// For errors, return the alert component
	return alert.Error(heading, messages).Render(c.Request().Context(), c.Response().Writer)
}

func (h *Handler) LoginHandler() echo.HandlerFunc {
    return func(c echo.Context) error {
        logger := h.logger.With().
            Str("handler", "LoginHandler").
            Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
            Logger()

        // Get session first to ensure it's working
        sess, err := session.Get(store.DefaultSessionName, c)
        if err != nil {
            logger.Error().Err(err).Msg("failed to get session")
            return h.LoginResponse(c,
                http.StatusInternalServerError,
                "System Error",
                []string{"Unable to process login request"},
                "")
        }

        // Bind form data
        params := new(LoginParams)
        if err := c.Bind(params); err != nil {
            return h.LoginResponse(c, http.StatusBadRequest, "Invalid Request", 
                []string{"Please check your input"}, "")
        }

        // Authenticate using the service (don't verify password again)
        user, err := h.service.Authenticate(c.Request().Context(), params.Email, params.Password)
        if err != nil {
            logger.Debug().Err(err).Str("email", params.Email).Msg("authentication failed")
            return h.LoginResponse(c, http.StatusUnauthorized, "Login Failed",
                []string{"Invalid email or password"}, "")
        }

        // Get facility
        facility, err := h.service.repos.Facility.GetByID(c.Request().Context(), user.FacilityID)
        if err != nil {
            logger.Error().Err(err).Msg("failed to get facility")
            return h.LoginResponse(c, http.StatusInternalServerError, "System Error",
                []string{"Unable to complete login"}, "")
        }

        // Set session values
        sess.Values["user_id"] = user.ID
        sess.Values["user_role"] = user.Role
        sess.Values["user_initials"] = user.Initials
        sess.Values["facility_code"] = facility.Code
        sess.Values["facility_id"] = facility.ID
        sess.Values["last_access"] = time.Now()

        h.logger.Debug().
            Interface("session_values", sess.Values).
            Msg("session values before save")

        if err := sess.Save(c.Request(), c.Response()); err != nil {
            logger.Error().Err(err).Msg("failed to save session")
            return h.LoginResponse(c,
                http.StatusInternalServerError,
                "System Error",
                []string{"Unable to complete login process"},
                "")
        }

        redirectURL := fmt.Sprintf("/%s/calendar", facility.Code)
        logger.Info().
            Str("email", params.Email).
            Str("redirect_url", redirectURL).
            Msg("login successful")

        return h.LoginResponse(c, http.StatusOK, "", nil, redirectURL)
    }
}

func (h *Handler) LogoutHandler() echo.HandlerFunc {
    return func(c echo.Context) error {
        logger := h.logger.With().Str("handler", "LogoutHandler").Logger()

        // Get the session
        sess, err := session.Get(store.DefaultSessionName, c)
        if err != nil {
            logger.Error().Err(err).Msg("failed to get session")
        }

        logger.Debug().
            Bool("is_new", sess.IsNew).
            Str("session_id", sess.ID).
            Interface("values", convertSessionValues(sess.Values)).
            Msg("starting logout process")

        // Clear session
        sess.Values = make(map[interface{}]interface{})
        sess.Options.MaxAge = -1

        if err = sess.Save(c.Request(), c.Response()); err != nil {
            logger.Error().Err(err).Msg("failed to save cleared session")
        }

        // Prevent caching
        c.Response().Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0, pre-check=0")

        logger.Info().Msg("logout completed")
        return c.Redirect(http.StatusSeeOther, "/login")
    }
}

// Helper function to convert session values for logging
func convertSessionValues(values map[interface{}]interface{}) map[string]interface{} {
	converted := make(map[string]interface{})
	for k, v := range values {
		if key, ok := k.(string); ok {
			converted[key] = v
		}
	}
	return converted
}

func redirectToLogin(c echo.Context) error {
	// If it's an API request, return 401
	if isAPIRequest(c) {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}
	// For regular requests, redirect to login
	return c.Redirect(http.StatusTemporaryRedirect, "/login")
}

func isAPIRequest(c echo.Context) bool {
	// Check Accept header
	if c.Request().Header.Get("Accept") == "application/json" {
		return true
	}
	// Check if it's an XHR request
	if c.Request().Header.Get("X-Requested-With") == "XMLHttpRequest" {
		return true
	}
	return false
}
