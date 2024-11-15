// validation/user_test.go
package validation

import (
	"errors"
	"strings"
	"testing"

	"github.com/DukeRupert/haven/db"
)

func TestValidateUserName(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		fieldName     string
		expected      UserName
		expectedError error
	}{
		{
			name:          "empty name",
			input:         "",
			fieldName:     "First name",
			expected:      "",
			expectedError: ErrEmptyField,
		},
		{
			name:          "only spaces",
			input:         "   ",
			fieldName:     "Last name",
			expected:      "",
			expectedError: ErrEmptyField,
		},
		{
			name:          "valid simple name",
			input:         "John",
			fieldName:     "First name",
			expected:      "John",
			expectedError: nil,
		},
		{
			name:          "valid compound name",
			input:         "Mary Jane",
			fieldName:     "First name",
			expected:      "Mary Jane",
			expectedError: nil,
		},
		{
			name:          "valid hyphenated name",
			input:         "Smith-Jones",
			fieldName:     "Last name",
			expected:      "Smith-Jones",
			expectedError: nil,
		},
		{
			name:          "valid name with apostrophe",
			input:         "O'Connor",
			fieldName:     "Last name",
			expected:      "O'Connor",
			expectedError: nil,
		},
		{
			name:          "invalid name with numbers",
			input:         "John2",
			fieldName:     "First name",
			expected:      "",
			expectedError: ErrInvalidName,
		},
		{
			name:          "invalid name with special characters",
			input:         "John!",
			fieldName:     "First name",
			expected:      "",
			expectedError: ErrInvalidName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateUserName(tt.input, tt.fieldName)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error containing %v, got nil", tt.expectedError)
					return
				}
				if !strings.Contains(err.Error(), tt.expectedError.Error()) {
					t.Errorf("expected error containing %v, got %v", tt.expectedError, err)
					return
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestValidateUserInitials(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expected      UserInitials
		expectedError error
	}{
		{
			name:          "empty initials",
			input:         "",
			expected:      "",
			expectedError: ErrEmptyInitials,
		},
		{
			name:          "only spaces",
			input:         "   ",
			expected:      "",
			expectedError: ErrEmptyInitials,
		},
		{
			name:          "too long",
			input:         "ABCDEFGHIJK",
			expected:      "",
			expectedError: ErrInitialsTooLong,
		},
		{
			name:          "valid simple initials",
			input:         "JD",
			expected:      "JD",
			expectedError: nil,
		},
		{
			name:          "valid initials with periods",
			input:         "J.D.",
			expected:      "JD",
			expectedError: nil,
		},
		{
			name:          "lowercase initials",
			input:         "jd",
			expected:      "JD",
			expectedError: nil,
		},
		{
			name:          "invalid initials with numbers",
			input:         "JD2",
			expected:      "",
			expectedError: ErrInvalidInitials,
		},
		{
			name:          "invalid initials with special characters",
			input:         "J&D",
			expected:      "",
			expectedError: ErrInvalidInitials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateUserInitials(tt.input)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
					return
				}
				if !errors.Is(err, tt.expectedError) {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
					return
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestValidateUserEmail(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expected      UserEmail
		expectedError error
	}{
		{
			name:          "empty email",
			input:         "",
			expected:      "",
			expectedError: ErrEmptyEmail,
		},
		{
			name:          "only spaces",
			input:         "   ",
			expected:      "",
			expectedError: ErrEmptyEmail,
		},
		{
			name:          "valid email",
			input:         "test@example.com",
			expected:      "test@example.com",
			expectedError: nil,
		},
		{
			name:          "valid email with uppercase",
			input:         "Test@Example.com",
			expected:      "test@example.com",
			expectedError: nil,
		},
		{
			name:          "valid email with plus",
			input:         "test+label@example.com",
			expected:      "test+label@example.com",
			expectedError: nil,
		},
		{
			name:          "invalid email no @",
			input:         "testexample.com",
			expected:      "",
			expectedError: ErrInvalidEmail,
		},
		{
			name:          "invalid email no domain",
			input:         "test@",
			expected:      "",
			expectedError: ErrInvalidEmail,
		},
		{
			name:          "invalid email no local part",
			input:         "@example.com",
			expected:      "",
			expectedError: ErrInvalidEmail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateUserEmail(tt.input)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
					return
				}
				if err != tt.expectedError {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
					return
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestValidateUserPassword(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expected      UserPassword
		expectedError error
	}{
		{
			name:          "empty password",
			input:         "",
			expected:      "",
			expectedError: ErrEmptyPassword,
		},
		{
			name:          "only spaces",
			input:         "   ",
			expected:      "",
			expectedError: ErrEmptyPassword,
		},
		{
			name:          "too short",
			input:         "Abc123!",
			expected:      "",
			expectedError: ErrPasswordTooShort,
		},
		{
			name:          "valid password with all types",
			input:         "Abcd123!@",
			expected:      "Abcd123!@",
			expectedError: nil,
		},
		{
			name:          "valid password no special chars",
			input:         "Abcd1234",
			expected:      "Abcd1234",
			expectedError: nil,
		},
		{
			name:          "valid password no numbers",
			input:         "Abcd!@#$",
			expected:      "Abcd!@#$",
			expectedError: nil,
		},
		{
			name:          "weak password only lowercase",
			input:         "abcdefghijk",
			expected:      "",
			expectedError: ErrWeakPassword,
		},
		{
			name:          "weak password only numbers",
			input:         "12345678",
			expected:      "",
			expectedError: ErrWeakPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateUserPassword(tt.input)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
					return
				}
				if !errors.Is(err, tt.expectedError) {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
					return
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestValidateUserRole(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expected      db.UserRole
		expectedError error
	}{
		{
			name:          "empty role",
			input:         "",
			expected:      "",
			expectedError: ErrEmptyRole,
		},
		{
			name:          "only spaces",
			input:         "   ",
			expected:      "",
			expectedError: ErrEmptyRole,
		},
		{
			name:          "valid role super",
			input:         "super",
			expected:      "super",
			expectedError: nil,
		},
		{
			name:          "valid role admin",
			input:         "admin",
			expected:      "admin",
			expectedError: nil,
		},
		{
			name:          "valid role user",
			input:         "user",
			expected:      "user",
			expectedError: nil,
		},
		{
			name:          "valid role with spaces",
			input:         " admin ",
			expected:      "admin",
			expectedError: nil,
		},
		{
			name:          "valid role uppercase",
			input:         "ADMIN",
			expected:      "admin",
			expectedError: nil,
		},
		{
			name:          "invalid role",
			input:         "guest",
			expected:      "",
			expectedError: ErrInvalidRole,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateUserRole(tt.input)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
					return
				}
				if err != tt.expectedError {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
					return
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
