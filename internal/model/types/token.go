// internal/model/types/token.go
package types

// Token types
type TokenType string

const (
    TokenTypeRegistration TokenType = "registration"
    TokenTypeVerification TokenType = "verification"
)