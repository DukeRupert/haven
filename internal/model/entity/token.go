// internal/model/entity/token.go
package entity

import (
	"time"
)

// RegistrationToken represents a token used for user registration
type RegistrationToken struct {
	UserID    int       `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}
