// internal/model/dto/user.go
package dto

import (
	"fmt"
	"time"

	"github.com/DukeRupert/haven/internal/model/entity"
	"github.com/DukeRupert/haven/internal/model/types"

	"github.com/rs/zerolog"
)

type AuthContext struct {
	UserID       int
	Role         types.UserRole
	Initials     string
	FacilityID   int
	FacilityCode string
}

type RouteContext struct {
	BasePath     string
	UserRole     types.UserRole
	UserInitials string
	FacilityID   int
	FacilityCode string
	User         entity.User     // Optional: full user object if needed
	Facility     entity.Facility // Optional: full facility object if needed
	PathFacility string
	PathInitials string
}

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

// MarshalZerologObject implements zerolog.LogObjectMarshaler
func (rc RouteContext) MarshalZerologObject(e *zerolog.Event) {
	e.Str("facility_code", rc.FacilityCode).
		Str("user_initials", rc.UserInitials).
		Str("base_path", rc.BasePath)
}

func (r *RouteContext) BuildURL(path string) string {
	if r.FacilityCode == "" {
		return path
	}
	return fmt.Sprintf("/%s/%s", r.FacilityCode, path)
}

type Breadcrumb struct {
	Label string
	URL   string
}

type CalendarPageProps struct {
	Route       RouteContext
	NavItems    []NavItem
	Auth        AuthContext
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
