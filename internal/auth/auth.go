package auth

import (
    "context"
    "errors"

    "github.com/DukeRupert/haven/internal/model/types"
    "github.com/DukeRupert/haven/internal/model/entity"
    "github.com/DukeRupert/haven/internal/repository"
    "github.com/rs/zerolog"
    "golang.org/x/crypto/bcrypt"
)

// Common errors
var (
    ErrInvalidCredentials = errors.New("invalid credentials")
)

// SessionKeys defines constants for session value keys
const (
    SessionKeyUserID       = "user_id"
    SessionKeyRole        = "role"         
    SessionKeyInitials    = "initials"     
    SessionKeyFacilityID  = "facility_id"
    SessionKeyFacilityCode = "facility_code"
    SessionKeyLastAccess  = "last_access"
)


// Service handles authentication-related operations
type Service struct {
    repos  *repository.Repositories
    logger zerolog.Logger
}

// Config holds service configuration
type Config struct {
    Repos  *repository.Repositories
    Logger zerolog.Logger
}

func NewService(cfg Config) *Service {
    return &Service{
        repos:  cfg.Repos,
        logger: cfg.Logger.With().Str("component", "auth").Logger(),
    }
}

// Authenticate verifies user credentials
func (s *Service) Authenticate(ctx context.Context, email, password string) (*entity.User, error) {
	log := s.logger.With().Str("method", "Authenticate").Logger()

	user, err := s.repos.User.GetByEmail(ctx, email)
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

// RoleLevel represents the hierarchy level of a role
type RoleLevel int

const (
    UserLevel  RoleLevel = 1
    AdminLevel RoleLevel = 2
    SuperLevel RoleLevel = 3
)

// HasMinimumRole checks if a role meets or exceeds the minimum required role
func HasMinimumRole(current, minimum types.UserRole) bool {
    roleValues := map[types.UserRole]RoleLevel{
        types.UserRoleUser:  UserLevel,
        types.UserRoleAdmin: AdminLevel,
        types.UserRoleSuper: SuperLevel,
    }

    currentLevel, ok := roleValues[current]
    if !ok {
        return false
    }

    requiredLevel, ok := roleValues[minimum]
    if !ok {
        return false
    }

    return currentLevel >= requiredLevel
}
