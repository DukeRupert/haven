package types

import (
	"time"
)

type CreateUserParams struct {
	FirstName    string   `form:"first_name" validate:"required"`
	LastName     string   `form:"last_name" validate:"required"`
	Initials     string   `form:"initials" validate:"required"`
	Email        string   `form:"email" validate:"required,email"`
	Password     string   `form:"password" validate:"required,min=8"`
	Role         UserRole `form:"role" validate:"required,oneof=super admin user"`
	FacilityCode string   `form:"facility_code" validate:"required"` // Used to look up FacilityID
	FacilityID   int      // Set after facility lookup, not from form
}

type CreateScheduleByCodeParams struct {
	FacilityCode  string       `form:"facility_code"`
	UserInitials  string       `form:"user_initials"`
	FirstWeekday  time.Weekday `form:"first_weekday"`
	SecondWeekday time.Weekday `form:"second_weekday"`
	StartDate     time.Time    `form:"start_date"`
}

type UpdateScheduleParams struct {
	FirstWeekday  time.Weekday `form:"first_weekday" validate:"required,min=0,max=6"`
	SecondWeekday time.Weekday `form:"second_weekday" validate:"required,min=0,max=6"`
	StartDate     time.Time    `form:"start_date" validate:"required"`
}

// CreateFacilityParams holds the parameters needed to create a new facility
type CreateFacilityParams struct {
	Name string `json:"name" form:"name"`
	Code string `json:"code" form:"code"`
}

// UpdateFacilityParams holds the parameters needed to update a facility
type UpdateFacilityParams struct {
	Name string `json:"name" form:"name"`
	Code string `json:"code" form:"code"`
}

type CreateScheduleRequest struct {
	FirstWeekday  int    `form:"first_weekday" validate:"required,min=0,max=6"`
	SecondWeekday int    `form:"second_weekday" validate:"required,min=0,max=6"`
	StartDate     string `form:"start_date" validate:"required"`
}