package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/DukeRupert/haven/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler holds dependencies for auth handlers
type AuthHandler struct {
	db     *pgxpool.Pool
	logger zerolog.Logger
}

// NewAuthHandler creates a new AuthHandler instance
func NewAuthHandler(db *pgxpool.Pool, logger zerolog.Logger) *AuthHandler {
	return &AuthHandler{
		db:     db,
		logger: logger.With().Str("component", "auth").Logger(),
	}
}

// LoginParams represents the expected login request body
type LoginParams struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents the response sent back after successful login
type LoginResponse struct {
	User  *models.User `json:"user"`
	Token string       `json:"token,omitempty"`
}

// LoginHandler handles user login requests
func (h *AuthHandler) LoginHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		log := h.logger.With().Str("handler", "LoginHandler").Logger()

		// Parse and validate request body
		params := new(LoginParams)
		if err := c.Bind(params); err != nil {
			log.Error().Err(err).Msg("failed to bind request body")
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
		}

		// Normalize email
		params.Email = strings.ToLower(strings.TrimSpace(params.Email))

		// Attempt login
		user, err := authenticateUser(ctx, h.db, params.Email, params.Password)
		if err != nil {
			if errors.Is(err, ErrInvalidCredentials) {
				log.Debug().
					Str("email", params.Email).
					Msg("invalid login attempt")
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid email or password")
			}
			log.Error().Err(err).Msg("login error")
			return echo.NewHTTPError(http.StatusInternalServerError, "login failed")
		}

		// Create session
		sess, err := session.Get("session", c)
		if err != nil {
			log.Error().Err(err).Msg("failed to get session")
			return echo.NewHTTPError(http.StatusInternalServerError, "session error")
		}

		// Set session values
		sess.Values["user_id"] = user.ID
		sess.Values["role"] = user.Role
		sess.Values["last_access"] = time.Now()

		if err := sess.Save(c.Request(), c.Response()); err != nil {
			log.Error().Err(err).Msg("failed to save session")
			return echo.NewHTTPError(http.StatusInternalServerError, "session error")
		}

		// Prepare response (omit sensitive fields)
		user.Password = "" // Clear password hash from response

		return c.JSON(http.StatusOK, LoginResponse{
			User: user,
		})
	}
}

var ErrInvalidCredentials = errors.New("invalid credentials")

// authenticateUser verifies the email and password combination
func authenticateUser(ctx context.Context, db *pgxpool.Pool, email, password string) (*models.User, error) {
	log := zerolog.Ctx(ctx).With().Str("method", "authenticateUser").Logger()

	var user models.User
	err := db.QueryRow(ctx,
		`SELECT id, created_at, updated_at, first_name, last_name, 
                initials, email, password, facility_id, role 
         FROM users 
         WHERE email = $1`,
		email).Scan(
		&user.ID, &user.CreatedAt, &user.UpdatedAt, &user.FirstName,
		&user.LastName, &user.Initials, &user.Email, &user.Password,
		&user.FacilityID, &user.Role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Debug().Str("email", email).Msg("user not found")
			return nil, ErrInvalidCredentials
		}
		log.Error().Err(err).Msg("database error")
		return nil, err
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		log.Debug().Str("email", email).Msg("invalid password")
		return nil, ErrInvalidCredentials
	}

	return &user, nil
}
