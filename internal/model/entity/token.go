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

// VerificationToken represents an email verification token
type VerificationToken struct {
    UserID    int       `db:"user_id"`
    Token     string    `db:"token"`
    Email     string    `db:"email"`
    CreatedAt time.Time `db:"created_at"`
    ExpiresAt time.Time `db:"expires_at"`
    Used      bool      `db:"used"`
}
