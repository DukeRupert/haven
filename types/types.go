package types

import (
	"path"
	"strings"

	"github.com/rs/zerolog"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type RouteContext struct {
	FacilityCode string
	UserInitials string
	BasePath     string
}

// MarshalZerologObject implements zerolog.LogObjectMarshaler
func (rc RouteContext) MarshalZerologObject(e *zerolog.Event) {
	e.Str("facility_code", rc.FacilityCode).
		Str("user_initials", rc.UserInitials).
		Str("base_path", rc.BasePath)
}

func (rc RouteContext) BuildURL(parts ...string) string {
	urlParts := []string{rc.BasePath}

	if rc.FacilityCode != "" {
		urlParts = append(urlParts, rc.FacilityCode)
		if rc.UserInitials != "" {
			urlParts = append(urlParts, rc.UserInitials)
		}
	}

	if len(parts) > 0 {
		urlParts = append(urlParts, parts...)
	}

	// Join parts with forward slashes and ensure single leading slash
	return "/" + strings.TrimPrefix(path.Join(urlParts...), "/")
}

type Breadcrumb struct {
	Label string
	URL   string
}

func (rc RouteContext) GetBreadcrumbs() []Breadcrumb {
	parts := strings.Split(strings.Trim(rc.BuildURL(), "/"), "/")
	breadcrumbs := make([]Breadcrumb, len(parts))
	caser := cases.Title(language.English)

	for i, part := range parts {
		breadcrumbs[i] = Breadcrumb{
			Label: caser.String(strings.ReplaceAll(part, "_", " ")),
			URL:   "/" + strings.Join(parts[:i+1], "/"),
		}
	}

	return breadcrumbs
}
