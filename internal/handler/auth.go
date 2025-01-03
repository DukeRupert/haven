package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/DukeRupert/haven/internal/model/entity"
	"github.com/DukeRupert/haven/internal/store"
	"github.com/DukeRupert/haven/web/view/alert"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// Common errors
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

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

func (h *Handler) LoginHandler(c echo.Context) error {
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
	user, err := h.Authenticate(c.Request().Context(), params.Email, params.Password)
	if err != nil {
		logger.Debug().Err(err).Str("email", params.Email).Msg("authentication failed")
		return h.LoginResponse(c, http.StatusUnauthorized, "Login Failed",
			[]string{"Invalid email or password"}, "")
	}

	// Get facility
	facility, err := h.repos.Facility.GetByID(c.Request().Context(), user.FacilityID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get facility")
		return h.LoginResponse(c, http.StatusInternalServerError, "System Error",
			[]string{"Unable to complete login"}, "")
	}

	// Set session values
	sess.Values[store.SessionKeyUserID] = user.ID
	sess.Values[store.SessionKeyRole] = user.Role
	sess.Values[store.SessionKeyInitials] = user.Initials
	sess.Values[store.SessionKeyFacilityCode] = facility.Code
	sess.Values[store.SessionKeyFacilityID] = facility.ID
	sess.Values[store.SessionKeyLastAccess] = time.Now()

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

	logger.Debug().
		Interface("session_values", map[string]interface{}{
			"user_id":       sess.Values[store.SessionKeyUserID],
			"role":          sess.Values[store.SessionKeyRole],
			"initials":      sess.Values[store.SessionKeyInitials],
			"facility_code": sess.Values[store.SessionKeyFacilityCode],
			"facility_id":   sess.Values[store.SessionKeyFacilityID],
		}).
		Msg("session values after save")

	redirectURL := fmt.Sprintf("/app/calendar")
	return h.LoginResponse(c, http.StatusOK, "", nil, redirectURL)
}

func (h *Handler) LogoutHandler(c echo.Context) error {
	logger := h.logger.With().
		Str("handler", "LogoutHandler").
		Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
		Logger()

	// Get the session
	sess, err := session.Get(store.DefaultSessionName, c)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get session")
		// Continue with logout even if we can't get the session
	}

	logger.Debug().
		Bool("is_new", sess.IsNew).
		Str("session_id", sess.ID).
		Interface("values", convertSessionValues(sess.Values)).
		Msg("starting logout process")

	// Clear all session values
	for k := range sess.Values {
		delete(sess.Values, k)
	}

	// Configure session for deletion
	sess.Options.MaxAge = -1
	sess.Options.Path = "/"      // Ensure cookie path matches
	sess.Options.HttpOnly = true // Security best practice
	sess.Options.Secure = true   // Assuming HTTPS

	// Save the session (this should delete it)
	if err = sess.Save(c.Request(), c.Response()); err != nil {
		logger.Error().Err(err).Msg("failed to save cleared session")
		// Continue with logout even if save fails
	}

	// Clear any additional cookies your app might use
	clearCookie(c, store.DefaultSessionName)

	// Prevent caching
	c.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Response().Header().Set("Pragma", "no-cache")
	c.Response().Header().Set("Expires", "0")

	logger.Info().Msg("logout completed")

	// Use 303 See Other to ensure a GET request to login page
	return c.Redirect(http.StatusSeeOther, "/login")
}

// Authenticate verifies user credentials
func (h *Handler) Authenticate(ctx context.Context, email, password string) (*entity.User, error) {
	log := h.logger.With().Str("method", "Authenticate").Logger()

	user, err := h.repos.User.GetByEmail(ctx, email)
	if err != nil {
		log.Debug().Str("email", email).Err(err).Msg("user lookup failed")
		return nil, ErrInvalidCredentials
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		log.Debug().Str("email", email).Msg("invalid password")
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

// Clear any other auth-related cookies if they exist
func clearCookie(c echo.Context, name string) {
	c.SetCookie(&http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
	})
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
