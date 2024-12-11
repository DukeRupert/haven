package handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/DukeRupert/haven/internal/model/dto"
	"github.com/DukeRupert/haven/internal/model/types"
	"github.com/a-h/templ"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func GetAuthContext(c echo.Context) (*dto.AuthContext, error) {
	sess, err := session.Get("session", c)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	userID, ok := sess.Values["user_id"].(int)
	if !ok {
		return nil, fmt.Errorf("invalid user_id in session")
	}

	role, ok := sess.Values["role"].(types.UserRole)
	if !ok {
		return nil, fmt.Errorf("invalid role in session")
	}

	auth := &dto.AuthContext{
		UserID: userID,
		Role:   role,
	}

	// Optional values
	if initials, ok := sess.Values["initials"].(string); ok {
		auth.Initials = initials
	}
	if facilityID, ok := sess.Values["facility_id"].(int); ok {
		auth.FacilityID = facilityID
	}
	if facilityCode, ok := sess.Values["facility_code"].(string); ok {
		auth.FacilityCode = facilityCode
	}

	return auth, nil
}

func canUpdateUser(auth *dto.AuthContext, targetUserID int) bool {
	if auth.Role == types.UserRoleSuper {
		return true
	}
	if auth.Role == types.UserRoleAdmin {
		return true
	}
	return auth.UserID == targetUserID
}

func (h *Handler) validateRoleChange(ctx context.Context, auth *dto.AuthContext,
	userID int, newRole types.UserRole,
) error {
	// Get existing user
	existingUser, err := h.repos.User.GetByID(ctx, userID)
	if err != nil {
		return SystemError(echo.New().AcquireContext())
	}

	// No role change, no validation needed
	if existingUser.Role == newRole {
		return nil
	}

	// Validate role change permissions
	if auth.Role != types.UserRoleSuper && newRole == types.UserRoleSuper {
		return ErrorResponse(echo.New().AcquireContext(),
			http.StatusForbidden,
			"Invalid Role",
			[]string{"Only super admins can assign super admin role"})
	}

	if auth.Role == types.UserRoleUser {
		return ErrorResponse(echo.New().AcquireContext(),
			http.StatusForbidden,
			"Invalid Role",
			[]string{"Users cannot change roles"})
	}

	return nil
}

func getUserID(c echo.Context) (int, error) {
	return strconv.Atoi(c.Param("user_id"))
}

func canAccessUserForm(auth *dto.AuthContext, targetUserID int) bool {
	// Super admins can edit anyone
	if auth.Role == types.UserRoleSuper {
		return true
	}

	// Admins can edit non-super users
	if auth.Role == types.UserRoleAdmin {
		return true
	}

	// Users can only edit themselves
	return auth.UserID == targetUserID
}

func canDeleteUser(auth *dto.AuthContext, targetUserID int) bool {
	switch auth.Role {
	case types.UserRoleSuper:
		return true
	case types.UserRoleAdmin:
		return auth.UserID != targetUserID // Admins can't delete themselves
	default:
		return false
	}
}

// Helper functions
func canUpdatePassword(auth *dto.AuthContext, targetUserID int) bool {
	switch auth.Role {
	case types.UserRoleSuper:
		return true
	case types.UserRoleAdmin:
		return true
	default:
		return auth.UserID == targetUserID
	}
}

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
