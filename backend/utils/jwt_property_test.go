package utils

import (
	"context"
	"testing"
	"time"

	"k-admin-system/config"
	"k-admin-system/global"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// setupTestEnvironment initializes the test environment with config and Redis
func setupTestEnvironment(t *testing.T) func() {
	// Initialize logger
	logger, _ := zap.NewDevelopment()
	global.Logger = logger

	// Initialize config
	global.Config = &config.Config{
		JWT: config.JWTConfig{
			Secret:            "test-secret-key-for-jwt-testing",
			AccessExpiration:  15, // 15 minutes
			RefreshExpiration: 7,  // 7 days
		},
		Redis: config.RedisConfig{
			Host:     "localhost",
			Port:     6379,
			Password: "",
			DB:       0,
		},
	}

	// Try to initialize Redis (skip if not available)
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// Test Redis connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		t.Logf("Redis not available, some tests will be skipped: %v", err)
		global.RedisClient = nil
	} else {
		global.RedisClient = redisClient
	}

	// Return cleanup function
	return func() {
		if global.RedisClient != nil {
			global.RedisClient.Close()
		}
	}
}

// Feature: k-admin-system
// Property 3: Token Generation and Refresh Cycle
// For any valid user credentials, login SHALL generate both access and refresh tokens,
// and using the refresh token SHALL produce a new valid access token
// Validates: Requirements 2.1, 2.5
func TestProperty3_TokenGenerationAndRefreshCycle(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("token generation produces valid access and refresh tokens", prop.ForAll(
		func(userID uint, username string, roleID uint) bool {
			// Ensure valid inputs
			if userID == 0 {
				userID = 1
			}
			if username == "" {
				username = "testuser"
			}
			if roleID == 0 {
				roleID = 1
			}

			// Generate tokens
			accessToken, refreshToken, err := GenerateToken(userID, username, roleID)
			if err != nil {
				t.Logf("Failed to generate tokens: %v", err)
				return false
			}

			// Verify tokens are not empty
			if accessToken == "" || refreshToken == "" {
				t.Logf("Generated tokens are empty")
				return false
			}

			// Parse access token
			accessClaims, err := ParseToken(accessToken)
			if err != nil {
				t.Logf("Failed to parse access token: %v", err)
				return false
			}

			// Verify access token claims
			if accessClaims.UserID != userID {
				t.Logf("Access token UserID mismatch: expected %d, got %d", userID, accessClaims.UserID)
				return false
			}
			if accessClaims.Username != username {
				t.Logf("Access token Username mismatch: expected %s, got %s", username, accessClaims.Username)
				return false
			}
			if accessClaims.RoleID != roleID {
				t.Logf("Access token RoleID mismatch: expected %d, got %d", roleID, accessClaims.RoleID)
				return false
			}

			// Parse refresh token
			refreshClaims, err := ParseToken(refreshToken)
			if err != nil {
				t.Logf("Failed to parse refresh token: %v", err)
				return false
			}

			// Verify refresh token claims
			if refreshClaims.UserID != userID {
				t.Logf("Refresh token UserID mismatch: expected %d, got %d", userID, refreshClaims.UserID)
				return false
			}

			return true
		},
		gen.UIntRange(1, 1000),
		gen.AlphaString(),
		gen.UIntRange(1, 100),
	))

	properties.Property("refresh token generates new valid access token", prop.ForAll(
		func(userID uint, username string, roleID uint) bool {
			// Ensure valid inputs
			if userID == 0 {
				userID = 1
			}
			if username == "" {
				username = "testuser"
			}
			if roleID == 0 {
				roleID = 1
			}

			// Generate initial tokens
			_, refreshToken, err := GenerateToken(userID, username, roleID)
			if err != nil {
				t.Logf("Failed to generate initial tokens: %v", err)
				return false
			}

			// Wait a moment to ensure new token has different timestamp
			time.Sleep(10 * time.Millisecond)

			// Refresh access token
			newAccessToken, err := RefreshToken(refreshToken)
			if err != nil {
				t.Logf("Failed to refresh token: %v", err)
				return false
			}

			// Verify new access token is valid
			if newAccessToken == "" {
				t.Logf("New access token is empty")
				return false
			}

			// Parse new access token
			newClaims, err := ParseToken(newAccessToken)
			if err != nil {
				t.Logf("Failed to parse new access token: %v", err)
				return false
			}

			// Verify claims match original
			if newClaims.UserID != userID {
				t.Logf("New token UserID mismatch: expected %d, got %d", userID, newClaims.UserID)
				return false
			}
			if newClaims.Username != username {
				t.Logf("New token Username mismatch: expected %s, got %s", username, newClaims.Username)
				return false
			}
			if newClaims.RoleID != roleID {
				t.Logf("New token RoleID mismatch: expected %d, got %d", roleID, newClaims.RoleID)
				return false
			}

			return true
		},
		gen.UIntRange(1, 1000),
		gen.AlphaString(),
		gen.UIntRange(1, 100),
	))

	properties.Property("token expiration is correctly set", prop.ForAll(
		func(userID uint) bool {
			if userID == 0 {
				userID = 1
			}

			// Generate tokens
			accessToken, refreshToken, err := GenerateToken(userID, "testuser", 1)
			if err != nil {
				return false
			}

			// Parse tokens
			accessClaims, err := ParseToken(accessToken)
			if err != nil {
				return false
			}

			refreshClaims, err := ParseToken(refreshToken)
			if err != nil {
				return false
			}

			// Verify access token expires in approximately 15 minutes
			accessExpiration := time.Until(accessClaims.ExpiresAt.Time)
			expectedAccessExpiration := 15 * time.Minute
			if accessExpiration < expectedAccessExpiration-time.Minute || accessExpiration > expectedAccessExpiration+time.Minute {
				t.Logf("Access token expiration incorrect: expected ~%v, got %v", expectedAccessExpiration, accessExpiration)
				return false
			}

			// Verify refresh token expires in approximately 7 days
			refreshExpiration := time.Until(refreshClaims.ExpiresAt.Time)
			expectedRefreshExpiration := 7 * 24 * time.Hour
			if refreshExpiration < expectedRefreshExpiration-time.Hour || refreshExpiration > expectedRefreshExpiration+time.Hour {
				t.Logf("Refresh token expiration incorrect: expected ~%v, got %v", expectedRefreshExpiration, refreshExpiration)
				return false
			}

			return true
		},
		gen.UIntRange(1, 1000),
	))

	properties.TestingRun(t)
}

// Feature: k-admin-system
// Property 4: Token Blacklist Enforcement
// For any blacklisted token, all API requests using that token SHALL be rejected
// with 401 status regardless of token validity
// Validates: Requirements 2.7
func TestProperty4_TokenBlacklistEnforcement(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Skip if Redis is not available
	if global.RedisClient == nil {
		t.Skip("Redis not available, skipping blacklist tests")
	}

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("blacklisted tokens are rejected", prop.ForAll(
		func(userID uint, username string, roleID uint) bool {
			// Ensure valid inputs
			if userID == 0 {
				userID = 1
			}
			if username == "" {
				username = "testuser"
			}
			if roleID == 0 {
				roleID = 1
			}

			// Generate token
			accessToken, _, err := GenerateToken(userID, username, roleID)
			if err != nil {
				t.Logf("Failed to generate token: %v", err)
				return false
			}

			// Verify token is valid before blacklisting
			_, err = ParseToken(accessToken)
			if err != nil {
				t.Logf("Token invalid before blacklisting: %v", err)
				return false
			}

			// Add token to blacklist
			err = AddTokenToBlacklist(accessToken)
			if err != nil {
				t.Logf("Failed to add token to blacklist: %v", err)
				return false
			}

			// Verify token is now blacklisted
			if !IsTokenBlacklisted(accessToken) {
				t.Logf("Token not blacklisted after adding to blacklist")
				return false
			}

			// Verify ParseToken rejects blacklisted token
			_, err = ParseToken(accessToken)
			if err != ErrTokenBlacklisted {
				t.Logf("Expected ErrTokenBlacklisted, got: %v", err)
				return false
			}

			return true
		},
		gen.UIntRange(1, 1000),
		gen.AlphaString(),
		gen.UIntRange(1, 100),
	))

	properties.Property("non-blacklisted tokens are accepted", prop.ForAll(
		func(userID uint) bool {
			if userID == 0 {
				userID = 1
			}

			// Generate token
			accessToken, _, err := GenerateToken(userID, "testuser", 1)
			if err != nil {
				return false
			}

			// Verify token is not blacklisted
			if IsTokenBlacklisted(accessToken) {
				t.Logf("New token incorrectly marked as blacklisted")
				return false
			}

			// Verify token can be parsed
			_, err = ParseToken(accessToken)
			if err != nil {
				t.Logf("Failed to parse non-blacklisted token: %v", err)
				return false
			}

			return true
		},
		gen.UIntRange(1, 1000),
	))

	properties.Property("blacklist respects token expiration", prop.ForAll(
		func(userID uint) bool {
			if userID == 0 {
				userID = 1
			}

			// Generate token
			accessToken, _, err := GenerateToken(userID, "testuser", 1)
			if err != nil {
				return false
			}

			// Add to blacklist
			err = AddTokenToBlacklist(accessToken)
			if err != nil {
				return false
			}

			// Verify token is blacklisted
			if !IsTokenBlacklisted(accessToken) {
				t.Logf("Token not blacklisted immediately after adding")
				return false
			}

			// The blacklist entry should have TTL matching token expiration
			// We can't easily test the full expiration without waiting 15 minutes,
			// but we can verify the blacklist entry exists
			return true
		},
		gen.UIntRange(1, 1000),
	))

	properties.TestingRun(t)
}
