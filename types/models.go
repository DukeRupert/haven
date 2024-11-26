package types

import (
	"strings"
	"time"
)

type UserRole string

const (
	UserRoleSuper UserRole = "super"
	UserRoleAdmin UserRole = "admin"
	UserRoleUser  UserRole = "user"
)

// String returns a formatted display string for the user role
func (r UserRole) String() string {
	return strings.ToLower(string(r))
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

type User struct {
	ID                    int       `db:"id" json:"id"`
	CreatedAt             time.Time `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time `db:"updated_at" json:"updated_at"`
	FirstName             string    `db:"first_name" json:"first_name"`
	LastName              string    `db:"last_name" json:"last_name"`
	Initials              string    `db:"initials" json:"initials"`
	Email                 string    `db:"email" json:"email"`
	Password              string    `db:"password" json:"-"` // Hashed password
	FacilityID            int       `db:"facility_id" json:"facility_id"`
	Role                  UserRole  `db:"role" json:"role" validate:"required,oneof=super admin user"`
	RegistrationCompleted bool      `db:"registration_completed" json:"registration_completed"`
}

// Facility represents a facility in the database
type Facility struct {
	ID        int       `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
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

func (s Schedule) IsZero() bool {
	return s.ID == 0 &&
		s.CreatedAt.IsZero() &&
		s.UpdatedAt.IsZero() &&
		s.UserID == 0 &&
		s.StartDate.IsZero()
}

type ProtectedDate struct {
	ID         int       `db:"id" json:"id"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
	ScheduleID int       `db:"schedule_id" json:"schedule_id" validate:"required"`
	Date       time.Time `db:"date" json:"date" validate:"required"`
	Available  bool      `db:"available" json:"available"`
	UserID     int       `db:"user_id" json:"user_id" validate:"required"`
	FacilityID int       `db:"facility_id" json:"facility_id" validate:"required"`
}

type UserDetails struct {
	User     User
	Facility Facility
	Schedule Schedule
}
