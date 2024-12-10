package handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"github.com/DukeRupert/haven/types"
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Reduce boilerplate for simple templ component renders
func render(c echo.Context, component templ.Component) error {
	c.Response().Header().Set("Content-Type", "text/html")
	return component.Render(c.Request().Context(), c.Response())
}

// ComponentGroup combines multiple templ components into a single component
func ComponentGroup(components ...templ.Component) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		for _, component := range components {
			if err := component.Render(ctx, w); err != nil {
				return err
			}
		}
		return nil
	})
}

// Helper function to build correct paths
func BuildNav(routeCtx *types.RouteContext, currentPath string) []types.NavItem {
	strippedPath := strings.TrimPrefix(currentPath, "/"+routeCtx.FacilityCode)

	navItems := []types.NavItem{}

	// Add nav items based on role access
	for path, config := range RouteConfigs {
		if IsAtLeastRole(string(routeCtx.UserRole), string(config.MinRole)) {
			navPath := path
			if config.RequiresFacility && routeCtx.FacilityCode != "" {
				navPath = fmt.Sprintf("/%s%s", routeCtx.FacilityCode, path)
			}

			title := cases.Title(language.English)

			navItems = append(navItems, types.NavItem{
				Path:    navPath,
				Name:    title.String(strings.TrimPrefix(path, "/")),
				Active:  strippedPath == path,
				Visible: true,
			})
		}
	}

	return navItems
}

// isAuthorizedToToggle checks if a user is authorized to toggle a protected date's availability
func isAuthorizedToToggle(userID int, role types.UserRole, protectedDate types.ProtectedDate) bool {
	// Allow access if user is admin or super
	if role == "admin" || role == "super" {
		return true
	}

	// Allow access if user owns the protected date
	if role == "user" && userID == protectedDate.UserID {
		return true
	}

	return false
}

// hashPassword creates a bcrypt hash of a password
func hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("error hashing password: %w", err)
	}
	return string(hashedBytes), nil
}

// generateSecureToken creates a cryptographically secure token for registration
func generateSecureToken() (string, error) {
	// Generate 32 bytes of random data
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("error generating secure token: %w", err)
	}

	// Encode to base64URL (URL-safe version of base64)
	// Use RawURLEncoding to avoid padding characters
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
