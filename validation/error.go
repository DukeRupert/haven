// validation/errors.go
package validation

import "errors"

var (
	// Code-related errors
	ErrEmptyCode       = errors.New("facility code is required")
	ErrInvalidLength   = errors.New("facility code must be 3 or 4 characters")
	ErrNotAlphanumeric = errors.New("facility code must contain only letters and numbers")

	// Name-related errors
	ErrEmptyName           = errors.New("facility name is required")
	ErrNameTooLong         = errors.New("facility name too long")
	ErrNameNotAlphanumeric = errors.New("facility name must contain only letters and numbers")
)
