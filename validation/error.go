// validation/errors.go
package validation

import "errors"

var (
	// Facility Code-related errors
	ErrEmptyCode       = errors.New("facility code is required")
	ErrInvalidLength   = errors.New("facility code must be 3 or 4 characters")
	ErrNotAlphanumeric = errors.New("facility code must contain only letters and numbers")

	// Facility Name-related errors
	ErrEmptyName           = errors.New("facility name is required")
	ErrNameTooLong         = errors.New("facility name too long")
	ErrNameNotAlphanumeric = errors.New("facility name must contain only letters and numbers")

	// User related errors
	ErrEmptyField = errors.New("field is empty")
	ErrInvalidName = errors.New("invalid name")
	
	ErrEmptyInitials = errors.New("initials are required")
	ErrInitialsTooLong = errors.New("initials too long")
	ErrInvalidInitials = errors.New("initials must contain only letters")

	ErrEmptyEmail = errors.New("email is required")
	ErrInvalidEmail = errors.New("invalid email format")

	ErrEmptyPassword = errors.New("password is required")
	ErrPasswordTooShort = errors.New("password too short")
	ErrWeakPassword = errors.New("password must contain at least 3 of: uppercase letters, lowercase letters, numbers, special characters")

	ErrEmptyRole = errors.New("role is required")
	ErrInvalidRole = errors.New("invalid role specified")
)
