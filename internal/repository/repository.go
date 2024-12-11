// internal/repository/repository.go
package repository

import (
	"github.com/DukeRupert/haven/internal/repository/facility"
	"github.com/DukeRupert/haven/internal/repository/schedule"
	"github.com/DukeRupert/haven/internal/repository/session"
	"github.com/DukeRupert/haven/internal/repository/token"
	"github.com/DukeRupert/haven/internal/repository/user"
)

type Repositories struct {
	Facility *facility.Repository
	User     *user.Repository
	Schedule *schedule.Repository
	Token    *token.Repository
	Session  *session.Repository
}

func NewRepositories(db *DB) *Repositories {
	// Initialize repositories in order of dependencies
	facilityRepo := facility.New(db.pool)
	scheduleRepo := schedule.New(db.pool)
	tokenRepo := token.New(db.pool)
	sessionRepo := session.New(db.pool)

	// User repository depends on facility and schedule
	userRepo := user.New(
		db.pool,
		facilityRepo,
		scheduleRepo,
	)

	return &Repositories{
		Facility: facilityRepo,
		User:     userRepo,
		Schedule: scheduleRepo,
		Token:    tokenRepo,
		Session:  sessionRepo,
	}
}
