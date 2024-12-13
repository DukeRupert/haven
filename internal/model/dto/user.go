// internal/model/dto/user.go
package dto

import (
	"time"

	"github.com/DukeRupert/haven/internal/model/entity"
	"github.com/DukeRupert/haven/internal/model/types"
)

// PageContext combines both for templates that need both
type PageContext struct {
	Route *RouteContext
	Auth  *AuthContext
	Nav   []NavItem
}

// AuthContext handles all user and authorization data
type AuthContext struct {
	UserID       int            `json:"user_id"`
	Role         types.UserRole `json:"role"`
	Initials     string         `json:"initials"`
	FacilityID   int            `json:"facility_id"`
	FacilityCode string         `json:"facility_code"`
}

// RouteContext focuses only on routing/path information
type RouteContext struct {
	BasePath    string // Base path including facility prefix (e.g., "/facility/KHLN")
	CurrentPath string // Current route pattern with parameters (e.g., "/controllers/:id")
	FullPath    string // Actual full URL path (e.g., "/facility/KHLN/controllers/123")
}

// NavItem remains focused on navigation structure
type NavItem struct {
	Path    string // Full path including facility code if applicable
	Name    string // Display name for the navigation item
	Icon    string // Icon identifier (for CSS/SVG icons)
	Active  bool   // Whether this is the current active route
	Visible bool   // Whether this item should be shown to the user
}

type Route struct {
	Path          string
	Name          string
	Icon          string
	MinRole       types.UserRole // Minimum role required
	NeedsFacility bool           // Whether route requires facility context
}

type UserPageProps struct {
	PageCtx		PageContext
    Title       string
    Description string
    Details     *UserDetails
}

type CalendarPageProps struct {
	PageCtx		PageContext
	Title       string
	Description string
	Calendar    CalendarProps
}

type CalendarProps struct {
	CurrentMonth   time.Time
	FacilityCode   string
	ProtectedDates []entity.ProtectedDate
	UserRole       types.UserRole
	CurrentUserID  int
}

type CalendarDayProps struct {
	Date           time.Time
	CurrentMonth   time.Time
	ProtectedDates []entity.ProtectedDate
	UserRole       types.UserRole
	CurrentUserID  int
	FacilityCode   string
}

type ProtectedDateGroup struct {
	Date  time.Time
	Dates []entity.ProtectedDate
}

type UserDetails struct {
	User     entity.User
	Facility entity.Facility
	Schedule entity.Schedule
}

type UserPageData struct {
	Title       string
	Description string
	Auth        AuthContext
	User        UserDetails
}
