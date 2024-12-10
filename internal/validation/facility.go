// validation/facility.go
package validation

import (
	"fmt"
	"regexp"
	"strings"
)

const MaxNameLength = 250

// FacilityCode represents a validated facility code
type FacilityCode string

// FacilityName represents a validated facility name
type FacilityName string

// ValidateFacilityCode validates and normalizes a facility code
// Returns the normalized code and any validation error
func ValidateFacilityCode(code string) (FacilityCode, error) {
	if code == "" {
		return "", ErrEmptyCode
	}

	// Normalize to lowercase
	normalized := strings.ToLower(code)

	// Validate length
	if len(normalized) != 3 && len(normalized) != 4 {
		return "", ErrInvalidLength
	}

	// Validate alphanumeric
	if !isAlphanumeric(normalized) {
		return "", ErrNotAlphanumeric
	}

	// Prepend K to 3-character codes
	if len(normalized) == 3 {
		normalized = "k" + normalized
	}

	return FacilityCode(normalized), nil
}

// ValidateFacilityName validates a facility name
// Returns the normalized name and any validation error
func ValidateFacilityName(name string) (FacilityName, error) {
    // Trim spaces and check if empty
    normalized := strings.TrimSpace(name)
    if normalized == "" {
        return "", ErrEmptyName
    }

    // Check length before other validation
    if len(normalized) > MaxNameLength {
        return "", fmt.Errorf("%w: maximum length is %d characters", ErrNameTooLong, MaxNameLength)
    }

    // Check alphanumeric
    if !isAlphanumeric(normalized) {
        return "", ErrNameNotAlphanumeric
    }

    // Convert to lowercase
    normalized = strings.ToLower(normalized)

    return FacilityName(normalized), nil
}

// isAlphanumeric checks if a string contains only alphanumeric characters
func isAlphanumeric(s string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(s)
}