package models

import "time"

// Facility represents a facility in the database
type Facility struct {
	ID        int       `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
}

// CreateFacilityParams holds the parameters needed to create a new facility
type CreateFacilityParams struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

// CreateFacilityParams holds the parameters needed to create a new facility
type UpdateFacilityParams struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// Controller represents a controller in the database
type Controller struct {
	ID         int       `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Initials   string    `json:"initials"`
	Email      string    `json:"email"`
	Password   string    `json:"-"` // Hashed password
	FacilityID int       `json:"facility_id"`
	Role       string    `json:"role" validate:"required,oneof=super admin user"`
}

// CreateControllerParams holds the parameters needed to create a new controller
type CreateControllerParams struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Initials   string `json:"initials"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	FacilityID int    `json:"facility_id"`
	Role       string `json:"role" validate:"required,oneof=super admin user"`
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

