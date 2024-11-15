// validation/user.go
package validation

import (
	"fmt"
	"net/mail"
	"strings"
	"unicode"

	"github.com/DukeRupert/haven/db"
)

const (
	MaxInitialsLength = 10
	MinPasswordLength = 8
)

type UserRole string

const (
	RoleSuper UserRole = "super"
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
)

// ValidUserRoles contains all valid user roles
var ValidUserRoles = map[UserRole]bool{
	RoleSuper: true,
	RoleAdmin: true,
	RoleUser:  true,
}

type UserName string
type UserInitials string
type UserEmail string
type UserPassword string
type HashedPassword string

// ValidateUserName validates both first and last names
func ValidateUserName(name string, fieldName string) (UserName, error) {
	if name = strings.TrimSpace(name); name == "" {
		return "", fmt.Errorf("%w: %s is required", ErrEmptyField, fieldName)
	}

	if !isValidName(name) {
		return "", fmt.Errorf("%w: %s contains invalid characters", ErrInvalidName, fieldName)
	}

	return UserName(name), nil
}

// ValidateUserInitials validates user initials
func ValidateUserInitials(initials string) (UserInitials, error) {
	if initials = strings.TrimSpace(initials); initials == "" {
		return "", ErrEmptyInitials
	}

	if len(initials) > MaxInitialsLength {
		return "", fmt.Errorf("%w: maximum length is %d characters", ErrInitialsTooLong, MaxInitialsLength)
	}

	// Convert to uppercase and remove any periods
	normalized := strings.ToUpper(strings.ReplaceAll(initials, ".", ""))

	if !isValidInitials(normalized) {
		return "", ErrInvalidInitials
	}

	return UserInitials(normalized), nil
}

// ValidateUserEmail validates email format
func ValidateUserEmail(email string) (UserEmail, error) {
	if email = strings.TrimSpace(email); email == "" {
		return "", ErrEmptyEmail
	}

	// Parse email address
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return "", ErrInvalidEmail
	}

	// Convert to lowercase
	return UserEmail(strings.ToLower(addr.Address)), nil
}

// ValidateUserPassword validates password strength
func ValidateUserPassword(password string) (UserPassword, error) {
	if password = strings.TrimSpace(password); password == "" {
		return "", ErrEmptyPassword
	}

	if len(password) < MinPasswordLength {
		return "", fmt.Errorf("%w: minimum length is %d characters", ErrPasswordTooShort, MinPasswordLength)
	}

	if !isStrongPassword(password) {
		return "", ErrWeakPassword
	}

	return UserPassword(password), nil
}

// Update the return type to match your db package
func ValidateUserRole(role string) (db.UserRole, error) {
    normalizedRole := strings.ToLower(strings.TrimSpace(role))
    
    if normalizedRole == "" {
        return "", ErrEmptyRole
    }

    switch normalizedRole {
    case "super", "admin", "user":
        return db.UserRole(normalizedRole), nil
    default:
        return "", ErrInvalidRole
    }
}

// Helper functions

func isValidName(name string) bool {
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsSpace(r) && r != '-' && r != '\'' {
			return false
		}
	}
	return true
}

func isValidInitials(initials string) bool {
	for _, r := range initials {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func isStrongPassword(password string) bool {
	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	// Require at least 3 of the 4 character types
	requirements := 0
	if hasUpper {
		requirements++
	}
	if hasLower {
		requirements++
	}
	if hasNumber {
		requirements++
	}
	if hasSpecial {
		requirements++
	}

	return requirements >= 3
}