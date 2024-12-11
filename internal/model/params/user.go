// internal/model/params/user.go
package params

import (
	"github.com/DukeRupert/haven/internal/model/types"
)

type CreateUserParams struct {
	FirstName    string         `form:"first_name" validate:"required"`
	LastName     string         `form:"last_name" validate:"required"`
	Initials     string         `form:"initials" validate:"required"`
	Email        string         `form:"email" validate:"required,email"`
	Password     string         `form:"password" validate:"required,min=8"`
	Role         types.UserRole `form:"role" validate:"required,oneof=super admin user"`
	FacilityCode string         `form:"facility_code" validate:"required"` // Used to look up FacilityID
	FacilityID   int            // Set after facility lookup, not from form
}

type UpdateUserParams struct {
	FirstName  string         `form:"first_name"`
	LastName   string         `form:"last_name"`
	Initials   string         `form:"initials"`
	Email      string         `form:"email"`
	FacilityID int            `form:"facility_id"`
	Role       types.UserRole `form:"role"`
}

type UpdatePasswordParams struct {
	Password string `form:"password" validate:"required,min=8"`
	Confirm  string `form:"confirm" validate:"required,eqfield=Password"`
}
