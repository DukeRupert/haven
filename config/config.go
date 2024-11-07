package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port         string
	DatabaseURL  string
	SessionKey   string
	Environment  string
}

// LoadConfig loads configuration from environment variables
// and returns a Config struct or an error if required values are missing
func Load() (*Config, error) {
	// Load .env file if it exists
	err := godotenv.Load()
	if err != nil {
		// Only log a warning since .env file is optional
		fmt.Printf("Warning: .env file not found: %v\n", err)
	}

	config := &Config{}

	// Load and validate each configuration value
	var missingVars []string

	// Port configuration with default
	config.Port = getEnvWithDefault("PORT", "8080")
	if !isValidPort(config.Port) {
		return nil, fmt.Errorf("invalid PORT value: %s", config.Port)
	}

	// Database URL is required
	config.DatabaseURL = os.Getenv("DATABASE_URL")
	if config.DatabaseURL == "" {
		missingVars = append(missingVars, "DATABASE_URL")
	}

	// Session key is required and must be at least 32 characters
	config.SessionKey = os.Getenv("SESSION_KEY")
	if config.SessionKey == "" {
		missingVars = append(missingVars, "SESSION_KEY")
	} else if len(config.SessionKey) < 32 {
		return nil, errors.New("SESSION_KEY must be at least 32 characters long")
	}

	// Environment with default and validation
	config.Environment = strings.ToLower(getEnvWithDefault("ENVIRONMENT", "development"))
	if !isValidEnvironment(config.Environment) {
		return nil, fmt.Errorf("invalid ENVIRONMENT value: %s", config.Environment)
	}

	// Check for any missing required variables
	if len(missingVars) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %s", strings.Join(missingVars, ", "))
	}

	return config, nil
}

// getEnvWithDefault returns the environment variable value or the default if not set
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// isValidPort checks if the port value is valid
func isValidPort(port string) bool {
	// Add any port validation logic you need
	// This is a simple example
	return port != ""
}

// isValidEnvironment checks if the environment value is valid
func isValidEnvironment(env string) bool {
	validEnvs := map[string]bool{
		"development": true,
		"staging":     true,
		"production":  true,
	}
	return validEnvs[env]
}

// String returns a string representation of the config for logging
// Sensitive fields are redacted
func (c *Config) String() string {
	return fmt.Sprintf(
		"Config{Port: %s, DatabaseURL: [REDACTED], Environment: %s}",
		c.Port,
		c.Environment,
	)
}