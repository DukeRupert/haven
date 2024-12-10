package entity

import (
	"time"

	"github.com/DukeRupert/haven/internal/model/types"
)

type User struct {
	ID                    int            `db:"id" json:"id"`
	CreatedAt             time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time      `db:"updated_at" json:"updated_at"`
	FirstName             string         `db:"first_name" json:"first_name"`
	LastName              string         `db:"last_name" json:"last_name"`
	Initials              string         `db:"initials" json:"initials"`
	Email                 string         `db:"email" json:"email"`
	Password              string         `db:"password" json:"-"` // Hashed password
	FacilityID            int            `db:"facility_id" json:"facility_id"`
	Role                  types.UserRole `db:"role" json:"role" validate:"required,oneof=super admin user"`
	RegistrationCompleted bool           `db:"registration_completed" json:"registration_completed"`
}
