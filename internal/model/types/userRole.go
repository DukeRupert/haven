// internal/model/types/userRole.go
package types

import "strings"

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