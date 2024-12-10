// validation/facility_test.go
package validation

import (
	"strings"
	"testing"
	"errors"
)

func TestValidateFacilityName(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expected      FacilityName
		expectedError error
	}{
		{
			name:          "empty name",
			input:         "",
			expected:      "",
			expectedError: ErrEmptyName,
		},
		{
			name:          "only spaces",
			input:         "   ",
			expected:      "",
			expectedError: ErrEmptyName,
		},
		{
			name:          "name too long",
			input:         strings.Repeat("A", MaxNameLength+1),
			expected:      "",
			expectedError: ErrNameTooLong,
		},
		{
			name:          "invalid characters with spaces",
			input:         "Facility 123",
			expected:      "",
			expectedError: ErrNameNotAlphanumeric,
		},
		{
			name:          "valid name mixed case",
			input:         "TestFacility123",
			expected:      "testfacility123",
			expectedError: nil,
		},
		{
			name:          "valid name with mixed case",
			input:         "NorthWestFacility42",
			expected:      "northwestfacility42",
			expectedError: nil,
		},
		{
			name:          "already lowercase",
			input:         "southfacility99",
			expected:      "southfacility99",
			expectedError: nil,
		},
		{
			name:          "maximum length mixed case",
			input:         strings.Repeat("Ab", MaxNameLength/2),
			expected:      FacilityName(strings.ToLower(strings.Repeat("Ab", MaxNameLength/2))),
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateFacilityName(tt.input)

			// Check error
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

			// Check result
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestValidateFacilityCode(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expected      FacilityCode
		expectedError error
	}{
		{
			name:          "empty code",
			input:         "",
			expected:      "",
			expectedError: ErrEmptyCode,
		},
		{
			name:          "code too short",
			input:         "AB",
			expected:      "",
			expectedError: ErrInvalidLength,
		},
		{
			name:          "code too long",
			input:         "ABCD5",
			expected:      "",
			expectedError: ErrInvalidLength,
		},
		{
			name:          "invalid characters",
			input:         "AB#$",
			expected:      "",
			expectedError: ErrNotAlphanumeric,
		},
		{
			name:          "valid 3-char lowercase",
			input:         "abc",
			expected:      "kabc",
			expectedError: nil,
		},
		{
			name:          "valid 3-char mixed case",
			input:         "aBc",
			expected:      "kabc",
			expectedError: nil,
		},
		{
			name:          "valid 3-char uppercase",
			input:         "ABC",
			expected:      "kabc",
			expectedError: nil,
		},
		{
			name:          "valid 3-char with numbers",
			input:         "a12",
			expected:      "ka12",
			expectedError: nil,
		},
		{
			name:          "valid 4-char lowercase",
			input:         "abcd",
			expected:      "abcd",
			expectedError: nil,
		},
		{
			name:          "valid 4-char uppercase",
			input:         "ABCD",
			expected:      "abcd",
			expectedError: nil,
		},
		{
			name:          "valid 4-char mixed case with numbers",
			input:         "Ab2D",
			expected:      "ab2d",
			expectedError: nil,
		},
		{
			name:          "valid 4-char starting with k",
			input:         "KABC",
			expected:      "kabc",
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateFacilityCode(tt.input)

			// Check error
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

			// Check result
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestIsAlphanumeric(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"empty string", "", false},
		{"lowercase letters", "abc", true},
		{"uppercase letters", "ABC", true},
		{"numbers", "123", true},
		{"mixed alphanumeric lowercase", "abc123", true},
		{"mixed alphanumeric uppercase", "ABC123", true},
		{"special characters", "abc!", false},
		{"spaces", "abc 123", false},
		{"unicode letters", "abcé", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAlphanumeric(tt.input)
			if result != tt.expected {
				t.Errorf("expected %v for input %q, got %v", tt.expected, tt.input, result)
			}
		})
	}
}
