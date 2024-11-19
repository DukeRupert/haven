package types

import (
	"fmt"

	"github.com/DukeRupert/haven/db"

	"github.com/rs/zerolog"

)

type Role = db.UserRole // alias for your existing UserRole type

type RouteContext struct {
    BasePath     string
    UserRole     Role
    UserInitials string
    FacilityID   int
    FacilityCode string
    User         *db.User     // Optional: full user object if needed
    Facility     *db.Facility // Optional: full facility object if needed
}

type NavItem struct {
    Path     string    // Full path including facility code if applicable
    Name     string    // Display name for the navigation item
    Icon     string    // Icon identifier (for CSS/SVG icons)
    Active   bool      // Whether this is the current active route
    Visible  bool      // Whether this item should be shown to the user
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

// func (rc RouteContext) GetBreadcrumbs() []Breadcrumb {
// 	parts := strings.Split(strings.Trim(rc.BuildURL(), "/"), "/")
// 	breadcrumbs := make([]Breadcrumb, len(parts))
// 	caser := cases.Title(language.English)

// 	for i, part := range parts {
// 		breadcrumbs[i] = Breadcrumb{
// 			Label: caser.String(strings.ReplaceAll(part, "_", " ")),
// 			URL:   "/" + strings.Join(parts[:i+1], "/"),
// 		}
// 	}

// 	return breadcrumbs
// }
