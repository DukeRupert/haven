package db

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

// String returns a formatted display string for the user role
func (r UserRole) String() string {
	switch r {
	case UserRoleSuper:
		return "Super Admin"
	case UserRoleAdmin:
		return "Admin"
	case UserRoleUser:
		return "User"
	default:
		return string(r) // fallback to the raw string value
	}
}

// Optionally, you might also want to add a method for getting CSS classes or styles:
func (r UserRole) BadgeClass() string {
	switch r {
	case UserRoleSuper:
		return "bg-purple-100 text-purple-800"
	case UserRoleAdmin:
		return "bg-blue-100 text-blue-800"
	case UserRoleUser:
		return "bg-green-100 text-green-800"
	default:
		return "bg-gray-100 text-gray-800"
	}
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
	ID            int          `db:"id" json:"id"`
	CreatedAt     time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time    `db:"updated_at" json:"updated_at"`
	UserID        int          `db:"user_id" json:"user_id" validate:"required"`
	FirstWeekday  time.Weekday `db:"first_weekday" json:"first_weekday" validate:"required,min=0,max=6"`
	SecondWeekday time.Weekday `db:"second_weekday" json:"second_weekday" validate:"required,min=0,max=6"`
	StartDate     time.Time    `db:"start_date" json:"start_date" validate:"required"`
}

type CreateScheduleByCodeParams struct {
	FacilityCode  string       `form:"facility_code"`
	UserInitials  string       `form:"user_initials"`
	FirstWeekday  time.Weekday `form:"first_weekday"`  // Changed from FirstDay
	SecondWeekday time.Weekday `form:"second_weekday"` // Changed from SecondDay
	StartDate     time.Time    `form:"start_date"`
}

type UpdateScheduleParams struct {
	FirstWeekday  time.Weekday `form:"first_weekday" validate:"required,min=0,max=6"`
	SecondWeekday time.Weekday `form:"second_weekday" validate:"required,min=0,max=6"`
	StartDate     time.Time    `form:"start_date" validate:"required"`
}

type ProtectedDate struct {
	ID         int       `db:"id" json:"id"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
	ScheduleID int       `db:"schedule_id" json:"schedule_id" validate:"required"`
	Date       time.Time `db:"date" json:"date" validate:"required"`
	Available  bool      `db:"available" json:"available"`
}

func (s Schedule) IsZero() bool {
	return s.ID == 0 &&
		s.CreatedAt.IsZero() &&
		s.UpdatedAt.IsZero() &&
		s.UserID == 0 &&
		s.StartDate.IsZero()
}

type AuthContext struct {
	UserID       int
	Role         UserRole
	Initials     string
	FacilityID   int
	FacilityCode string
}

