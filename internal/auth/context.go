// internal/auth/context.go
package auth

import (
	"fmt"

	"github.com/DukeRupert/haven/internal/model/dto"
	"github.com/DukeRupert/haven/internal/model/types"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

// AuthContext fetches authentication details from the session
func (m *Middleware) GetAuthContext(c echo.Context) (*dto.AuthContext, error) {
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
