package models

import "time"

// UserRole is a custom type for the user_role enum
type UserRole string

const (
	UserRoleSuper UserRole = "super"
	UserRoleAdmin UserRole = "admin"
	UserRoleUser  UserRole = "user"
)

type User struct {
	ID         int       `db:"id" json:"id"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
	FirstName  string    `db:"first_name" json:"first_name"`
	LastName   string    `db:"last_name" json:"last_name"`
	Initials   string    `db:"initials" json:"initials"`
	Email      string    `db:"email" json:"email"`
	Password   string    `db:"password" json:"-"` // Hashed password
	FacilityID int       `db:"facility_id" json:"facility_id"`
	Role       UserRole  `db:"role" json:"role" validate:"required,oneof=super admin user"`
}

// CreateUserParams represents the parameters needed to create a new user
type CreateUserParams struct {
	FirstName  string   `json:"first_name" validate:"required"`
	LastName   string   `json:"last_name" validate:"required"`
	Initials   string   `json:"initials" validate:"required,max=10"`
	Email      string   `json:"email" validate:"required,email"`
	Password   string   `json:"password" validate:"required,min=8"`
	FacilityID int      `json:"facility_id" validate:"required,min=1"`
	Role       UserRole `json:"role" validate:"required,oneof=super admin user"`
}

type Schedule struct {
	ID           int       `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	RDOs         []int     `json:"rdos"`
	Anchor       time.Time `json:"anchor"`
	ControllerID int       `json:"controller_id"`
}

type CreateScheduleParams struct {
	RDOs         []int     `json:"rdos"`
	Anchor       time.Time `json:"anchor"`
	ControllerID int       `json:"controller_id"`
}

type UpdateScheduleParams struct {
	RDOs   []int     `json:"rdos"`
	Anchor time.Time `json:"anchor"`
}

type AuthContext struct {
	UserID       int
	Role         UserRole
	Initials     string
	FacilityID   int
	FacilityCode string
}
