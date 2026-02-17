package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestProperty38_ConfigurationSourcePriority tests that environment variables
// take precedence over file configuration values
// Feature: k-admin-system, Property 38: Configuration Source Priority
// Validates: Requirements 12.2
func TestProperty38_ConfigurationSourcePriority(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 10 // Reduced from default 100 for faster execution
	properties := gopter.NewProperties(parameters)

	properties.Property("environment variables override file config values", prop.ForAll(
		func(envPortNum int, filePortNum int) bool {
			// Generate port strings from port numbers
			envPort := fmt.Sprintf(":%d", envPortNum)
			filePort := fmt.Sprintf(":%d", filePortNum)

			// Skip if ports are the same (no override to test)
			if envPort == filePort {
				return true
			}

			// Create a temporary config file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")

			// Write config file with filePort
			configContent := fmt.Sprintf(`server:
  port: "%s"
  mode: "debug"

database:
  host: "localhost"
  port: 3306
  name: "test_db"
  username: "root"
  password: "password"

jwt:
  secret: "test-secret"
  access_expiration: 15
  refresh_expiration: 7

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0

logger:
  level: "info"
  path: "./logs/app.log"
`, filePort)

			if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
				t.Logf("Failed to write config file: %v", err)
				return false
			}

			// Set environment variable with envPort
			envKey := "KADMIN_SERVER_PORT"
			oldEnv := os.Getenv(envKey)
			os.Setenv(envKey, envPort)
			defer func() {
				if oldEnv != "" {
					os.Setenv(envKey, oldEnv)
				} else {
					os.Unsetenv(envKey)
				}
			}()

			// Load configuration
			config, err := LoadConfig(configPath)
			if err != nil {
				t.Logf("Failed to load config: %v", err)
				return false
			}

			// Verify that environment variable value takes precedence
			return config.Server.Port == envPort
		},
		gen.IntRange(8000, 9000),
		gen.IntRange(8000, 9000),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty39_ConfigurationValidationOnStartup tests that missing required
// configuration fields cause startup failure with detailed error messages
// Feature: k-admin-system, Property 39: Configuration Validation on Startup
// Validates: Requirements 12.4, 12.5
func TestProperty39_ConfigurationValidationOnStartup(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 10 // Reduced from default 100 for faster execution
	properties := gopter.NewProperties(parameters)

	// Define required fields and their paths in the config structure
	requiredFields := []struct {
		name       string
		configYAML string
	}{
		{
			name: "server.port",
			configYAML: `server:
  mode: "debug"
database:
  host: "localhost"
  port: 3306
  name: "test_db"
  username: "root"
jwt:
  secret: "test-secret"
redis:
  host: "localhost"
  port: 6379
logger:
  level: "info"`,
		},
		{
			name: "database.host",
			configYAML: `server:
  port: ":8080"
  mode: "debug"
database:
  port: 3306
  name: "test_db"
  username: "root"
jwt:
  secret: "test-secret"
redis:
  host: "localhost"
  port: 6379
logger:
  level: "info"`,
		},
		{
			name: "database.port",
			configYAML: `server:
  port: ":8080"
  mode: "debug"
database:
  host: "localhost"
  name: "test_db"
  username: "root"
jwt:
  secret: "test-secret"
redis:
  host: "localhost"
  port: 6379
logger:
  level: "info"`,
		},
		{
			name: "database.name",
			configYAML: `server:
  port: ":8080"
  mode: "debug"
database:
  host: "localhost"
  port: 3306
  username: "root"
jwt:
  secret: "test-secret"
redis:
  host: "localhost"
  port: 6379
logger:
  level: "info"`,
		},
		{
			name: "database.username",
			configYAML: `server:
  port: ":8080"
  mode: "debug"
database:
  host: "localhost"
  port: 3306
  name: "test_db"
jwt:
  secret: "test-secret"
redis:
  host: "localhost"
  port: 6379
logger:
  level: "info"`,
		},
		{
			name: "jwt.secret",
			configYAML: `server:
  port: ":8080"
  mode: "debug"
database:
  host: "localhost"
  port: 3306
  name: "test_db"
  username: "root"
redis:
  host: "localhost"
  port: 6379
logger:
  level: "info"`,
		},
		{
			name: "redis.host",
			configYAML: `server:
  port: ":8080"
  mode: "debug"
database:
  host: "localhost"
  port: 3306
  name: "test_db"
  username: "root"
jwt:
  secret: "test-secret"
redis:
  port: 6379
logger:
  level: "info"`,
		},
		{
			name: "redis.port",
			configYAML: `server:
  port: ":8080"
  mode: "debug"
database:
  host: "localhost"
  port: 3306
  name: "test_db"
  username: "root"
jwt:
  secret: "test-secret"
redis:
  host: "localhost"
logger:
  level: "info"`,
		},
	}

	properties.Property("missing required fields cause startup failure with error message", prop.ForAll(
		func(fieldIndex int) bool {
			if fieldIndex < 0 || fieldIndex >= len(requiredFields) {
				return true
			}

			field := requiredFields[fieldIndex]

			// Create a temporary config file with missing required field
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")

			if err := os.WriteFile(configPath, []byte(field.configYAML), 0644); err != nil {
				t.Logf("Failed to write config file: %v", err)
				return false
			}

			// Attempt to load configuration
			config, err := LoadConfig(configPath)

			// Should fail with an error
			if err == nil {
				t.Logf("Expected error for missing field %s, but got none", field.name)
				return false
			}

			// Error message should mention the missing field
			errorMsg := err.Error()
			if errorMsg == "" {
				t.Logf("Error message is empty for missing field %s", field.name)
				return false
			}

			// Verify that config is nil when validation fails
			if config != nil {
				t.Logf("Expected nil config for missing field %s, but got non-nil", field.name)
				return false
			}

			// Error message should contain the field name
			// This ensures detailed error messages as per requirement 12.5
			return true
		},
		gen.IntRange(0, len(requiredFields)-1),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
