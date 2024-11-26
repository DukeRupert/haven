package handler

import (
	"fmt"
	"net/http"

	"github.com/DukeRupert/haven/types"
	"github.com/DukeRupert/haven/view/alert"
	"github.com/DukeRupert/haven/view/component"
	"github.com/DukeRupert/haven/view/page"
	
	"golang.org/x/crypto/bcrypt"
	"github.com/labstack/echo/v4"
)

// Helper function to validate role
func isValidRole(role types.UserRole) bool {
	validRoles := []types.UserRole{"super", "admin", "user"}
	for _, r := range validRoles {
		if role == r {
			return true
		}
	}
	return false
}

// Updated handler function
func (h *Handler) handleProfile(c echo.Context, routeCtx *types.RouteContext, navItems []types.NavItem) error {
	logger := h.logger.With().
		Str("handler", "handleProfile").
		Str("facility", c.Param("facility")).
		Logger()

	auth, err := GetAuthContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "auth context error")
	}

	// Check for initials query param
	initials := c.Param("initials")
	
	// If no initials provided, show the current user's profile
	if initials == "" {
		initials = auth.Initials // Assuming this exists in your auth context
		logger.Debug().
			Str("initials", initials).
			Msg("no initials provided, showing own profile")
	} else {
		// Log if we're viewing someone else's profile
		logger.Debug().
			Str("initials", initials).
			Str("viewer", auth.Initials).
			Str("route_initials", routeCtx.UserInitials).
			Msg("viewing other user's profile")

		// Optional: Check if user has permission to view other profiles
		if !canViewOtherProfiles(auth) {
			logger.Warn().
				Str("initials", initials).
				Str("viewer", auth.Initials).
				Msg("unauthorized attempt to view other profile")
			return echo.NewHTTPError(http.StatusForbidden, "Not authorized to view other profiles")
		}
	}

	details, err := h.db.GetUserDetails(
		c.Request().Context(),
		initials,
		auth.FacilityID,
	)
	if err != nil {
		logger.Error().
			Err(err).
			Str("initials", initials).
			Msg("failed to find user details")
		return c.String(http.StatusNotFound, "User not found")
	}

	title := "Profile"
	description := "Manage your account information and schedule"

	// Optionally modify title/description when viewing other profiles
	if initials != auth.Initials {
		title = fmt.Sprintf("%s's Profile", details.User.Initials) // Adjust based on your user details structure
		description = fmt.Sprintf("View %s's information and schedule", details.User.Initials)
	}

	component := page.Profile(
		*routeCtx,
		navItems,
		title,
		description,
		auth,
		details,
	)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

// Helper function to check if user can view other profiles
func canViewOtherProfiles(auth *types.AuthContext) bool {
	// Implement your permission logic here
	// For example:
	switch auth.Role {
	case types.UserRoleAdmin, types.UserRoleSuper:
		return true
	default:
		return false
	}
}

// Create handles POST /api/user
func (h *Handler) handleCreateUser(c echo.Context) error {
    logger := h.logger

    // Create a struct specifically for form binding
    var formData struct {
        FirstName    string `form:"first_name"`
        LastName     string `form:"last_name"`
        Initials     string `form:"initials"`
        Email        string `form:"email"`
        Password     string `form:"password"`
        Role         string `form:"role"`
        FacilityCode string `form:"facility_code"`
    }

    // Bind the form data first
    if err := c.Bind(&formData); err != nil {
        return h.ErrorResponse(c, http.StatusBadRequest, "Invalid Request", 
            []string{"The submitted form data was invalid"})
    }

    // Validate and collect all validation errors
    var errors []string
    var params types.CreateUserParams

    // Validate facility code and get facility
    if formData.FacilityCode == "" {
        errors = append(errors, "Facility code is required")
    } else {
        facility, err := h.db.GetFacilityByCode(c.Request().Context(), formData.FacilityCode)
        if err != nil {
            logger.Error().
                Err(err).
                Str("facility_code", formData.FacilityCode).
                Msg("failed to find facility by code")
            errors = append(errors, "The specified facility code was not found")
        } else {
            params.FacilityID = facility.ID
        }
    }

    // ... rest of your validation code ...

    // Return all validation errors if any exist
    if len(errors) > 0 {
        return h.ValidationError(c, errors)
    }

    // Hash the password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
    if err != nil {
        return h.SystemError(c, err, "failed to hash password")
    }
    params.Password = string(hashedPassword)

    // Create the user
    user, err := h.db.CreateUser(c.Request().Context(), params)
    if err != nil {
        return h.SystemError(c, err, "failed to create user in database")
    }

    // Log success
    logger.Info().
        Int("user_id", user.ID).
        Str("email", user.Email).
        Int("facility_id", user.FacilityID).
        Str("role", string(user.Role)).
        Msg("user created successfully")

    // For successful creation, you might want to show a success message
    if c.Request().Header.Get("HX-Request") == "true" {
        // Return both the success alert and the new user list item
        return render(c, ComponentGroup(
            alert.Success("User Created", fmt.Sprintf("Successfully created user %s %s", user.FirstName, user.LastName)),
            page.UserListItem(h.RouteCtx, *user),
        ))
    }

    return render(c, page.UserListItem(h.RouteCtx, *user))
}

func (h *Handler) createUserForm(c echo.Context) error {
	// Get facility code from route parameter
	code := c.Param("facility")
	if code == "" {
		h.logger.Error().Msg("facility code is missing from request")
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "facility code is required",
		})
	}
	return render(c, component.CreateUserForm(code))
}

func (h *Handler) handleUsers(c echo.Context, routeCtx *types.RouteContext, navItems []types.NavItem) error {
	// Get facility code from route parameter
	code := c.Param("facility")
	if code == "" {
		h.logger.Error().Msg("facility code is missing from request")
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "facility code is required",
		})
	}

	// Get users from database
	// // Track database query duration specifically
	users, err := h.db.GetUsersByFacilityCode(c.Request().Context(), code)
	if err != nil {
		h.logger.Error().
			Err(err).
			Msg("failed to retrieve users from database")

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to retrieve users",
		})
	}

	// If no users found, return empty array instead of null
	if users == nil {
		users = []types.User{}
	}

	title := "Controllers"
	description := "A list of all controllers assigned to the facility."
	auth, err := GetAuthContext(c)

	component := page.ShowUsers(
		*routeCtx,
		navItems,
		title,
		description,
		auth.Role,
		users,
	)
	return component.Render(c.Request().Context(), c.Response().Writer)
}


