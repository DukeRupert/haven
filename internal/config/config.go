package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server Configuration
	Port        string
	Environment string
	SessionKey  string
	BaseURL		string

	// Database Configuration
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DatabaseURL string

	// Email config
	PostmarkServerToken string
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

	// Server Configuration
	config.Port = getEnvWithDefault("PORT", "8080")
	if !isValidPort(config.Port) {
		return nil, fmt.Errorf("invalid PORT value: %s", config.Port)
	}

	config.Environment = strings.ToLower(getEnvWithDefault("ENVIRONMENT", "development"))
	if !isValidEnvironment(config.Environment) {
		return nil, fmt.Errorf("invalid ENVIRONMENT value: %s", config.Environment)
	}

	config.BaseURL = getEnvWithDefault("BASE_URL", "http://localhost")
	config.PostmarkServerToken = getEnvWithDefault("POSTMARK_SERVER_TOKEN", "")

	// Session key is required and must be at least 32 characters
	config.SessionKey = os.Getenv("SESSION_KEY")
	if config.SessionKey == "" {
		missingVars = append(missingVars, "SESSION_KEY")
	} else if len(config.SessionKey) < 32 {
		return nil, errors.New("SESSION_KEY must be at least 32 characters long")
	}

	// Database Configuration
	config.DBHost = getEnvWithDefault("DB_HOST", "localhost")
	config.DBPort = getEnvWithDefault("DB_PORT", "5432")
	config.DBUser = os.Getenv("DB_USER")
	config.DBPassword = os.Getenv("DB_PASSWORD")
	config.DBName = os.Getenv("DB_NAME")

	// Check required database variables
	for _, v := range []struct{ key, value string }{
		{"DB_USER", config.DBUser},
		{"DB_PASSWORD", config.DBPassword},
		{"DB_NAME", config.DBName},
	} {
		if v.value == "" {
			missingVars = append(missingVars, v.key)
		}
	}

	// Construct Database URL
	config.DatabaseURL = fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.DBUser,
		config.DBPassword,
		config.DBHost,
		config.DBPort,
		config.DBName,
	)

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
		"Config{"+
			"Port: %s, "+
			"Environment: %s, "+
			"DBHost: %s, "+
			"DBPort: %s, "+
			"DBUser: %s, "+
			"DBName: %s, "+
			"DatabaseURL: [REDACTED]"+
			"}",
		c.Port,
		c.Environment,
		c.DBHost,
		c.DBPort,
		c.DBUser,
		c.DBName,
	)
}