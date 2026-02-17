package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"k-admin-system/config"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestProperty40_LogLevelFiltering tests that only log messages at the configured
// level or higher severity are output
// Feature: k-admin-system, Property 40: Log Level Filtering
// Validates: Requirements 13.3
func TestProperty40_LogLevelFiltering(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 5 // Reduced for faster execution
	properties := gopter.NewProperties(parameters)

	// Define log levels in order of severity (lowest to highest)
	logLevels := []string{"debug", "info", "warn", "error"}

	properties.Property("only logs at configured level or higher are output", prop.ForAll(
		func(configuredLevelIndex int) bool {
			if configuredLevelIndex < 0 || configuredLevelIndex >= len(logLevels) {
				return true
			}

			configuredLevel := logLevels[configuredLevelIndex]

			// Create a temporary log file
			tmpDir := t.TempDir()
			logPath := filepath.Join(tmpDir, "test.log")

			cfg := &config.Config{
				Server: config.ServerConfig{
					Mode: "release", // File only to make verification easier
					Port: ":8080",
				},
				Logger: config.LoggerConfig{
					Level:      configuredLevel,
					Path:       logPath,
					MaxSize:    10,
					MaxAge:     7,
					MaxBackups: 3,
					Compress:   false,
				},
			}

			logger, err := InitLogger(cfg)
			if err != nil {
				t.Logf("Failed to initialize logger: %v", err)
				return false
			}
			defer func() {
				_ = logger.Sync()
			}()

			// Log messages at all levels with unique identifiers
			logger.Debug("DEBUG_MESSAGE")
			logger.Info("INFO_MESSAGE")
			logger.Warn("WARN_MESSAGE")
			logger.Error("ERROR_MESSAGE")

			// Sync to ensure all logs are written
			_ = logger.Sync()

			// Read the log file
			content, err := os.ReadFile(logPath)
			if err != nil {
				t.Logf("Failed to read log file: %v", err)
				return false
			}

			logContent := string(content)

			// Verify that only messages at configured level or higher are present
			for i, level := range logLevels {
				messageMarker := strings.ToUpper(level) + "_MESSAGE"

				if i < configuredLevelIndex {
					// Lower severity - should NOT be in logs
					if strings.Contains(logContent, messageMarker) {
						t.Logf("Found %s in logs when configured level is %s (should be filtered)", level, configuredLevel)
						return false
					}
				} else {
					// Same or higher severity - SHOULD be in logs
					if !strings.Contains(logContent, messageMarker) {
						t.Logf("Did not find %s in logs when configured level is %s (should be present)", level, configuredLevel)
						return false
					}
				}
			}

			return true
		},
		gen.IntRange(0, len(logLevels)-1),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty43_EnvironmentSpecificLogOutput tests that log files are created
// in both development and production modes
// Feature: k-admin-system, Property 43: Environment-Specific Log Output
// Validates: Requirements 13.7, 13.8
func TestProperty43_EnvironmentSpecificLogOutput(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 5 // Reduced for faster execution
	properties := gopter.NewProperties(parameters)

	properties.Property("log files are created in all modes", prop.ForAll(
		func(mode string) bool {
			// Only test valid modes
			validModes := map[string]bool{"debug": true, "test": true, "release": true}
			if !validModes[mode] {
				return true
			}

			tmpDir := t.TempDir()
			logPath := filepath.Join(tmpDir, "test.log")

			cfg := &config.Config{
				Server: config.ServerConfig{
					Mode: mode,
					Port: ":8080",
				},
				Logger: config.LoggerConfig{
					Level:      "info",
					Path:       logPath,
					MaxSize:    10,
					MaxAge:     7,
					MaxBackups: 3,
					Compress:   false,
				},
			}

			logger, err := InitLogger(cfg)
			if err != nil {
				t.Logf("Failed to initialize logger: %v", err)
				return false
			}
			defer func() {
				_ = logger.Sync()
			}()

			// Log a test message
			testMessage := "test_mode_verification_" + mode
			logger.Info(testMessage)
			_ = logger.Sync()

			// Read the log file
			fileContent, err := os.ReadFile(logPath)
			if err != nil {
				t.Logf("Failed to read log file: %v", err)
				return false
			}

			// File should always contain the message
			if !strings.Contains(string(fileContent), testMessage) {
				t.Logf("Message not found in file for mode %s", mode)
				return false
			}

			return true
		},
		gen.OneConstOf("debug", "test", "release"),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
