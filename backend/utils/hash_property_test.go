package utils

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: k-admin-system
// Property 2: Password Encryption Round-Trip
// For any password string, encrypting with bcrypt and then validating against
// the hash SHALL return true for the original password and false for any different password
// Validates: Requirements 2.2
func TestProperty2_PasswordEncryptionRoundTrip(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("hashed password validates correctly for original password", prop.ForAll(
		func(password string) bool {
			// Skip empty passwords or passwords longer than 72 bytes (bcrypt limit)
			if password == "" || len(password) > 72 {
				return true
			}

			// Hash the password
			hashedPassword, err := HashPassword(password)
			if err != nil {
				t.Logf("Failed to hash password: %v", err)
				return false
			}

			// Verify hash is not empty
			if hashedPassword == "" {
				t.Logf("Hashed password is empty")
				return false
			}

			// Verify hash is different from original password
			if hashedPassword == password {
				t.Logf("Hashed password is same as original password")
				return false
			}

			// Verify original password validates against hash
			if !CheckPassword(hashedPassword, password) {
				t.Logf("Original password failed validation against hash")
				return false
			}

			return true
		},
		gen.AlphaString(),
	))

	properties.Property("hashed password rejects different passwords", prop.ForAll(
		func(password1 string, password2 string) bool {
			// Skip empty passwords or passwords longer than 72 bytes (bcrypt limit)
			if password1 == "" || password2 == "" || len(password1) > 72 || len(password2) > 72 {
				return true
			}

			// Skip if passwords are the same
			if password1 == password2 {
				return true
			}

			// Hash the first password
			hashedPassword, err := HashPassword(password1)
			if err != nil {
				t.Logf("Failed to hash password: %v", err)
				return false
			}

			// Verify second password does not validate against first password's hash
			if CheckPassword(hashedPassword, password2) {
				t.Logf("Different password incorrectly validated against hash")
				return false
			}

			return true
		},
		gen.AlphaString(),
		gen.AlphaString(),
	))

	properties.Property("same password produces different hashes", prop.ForAll(
		func(password string) bool {
			// Skip empty passwords or passwords longer than 72 bytes (bcrypt limit)
			if password == "" || len(password) > 72 {
				return true
			}

			// Hash the password twice
			hash1, err := HashPassword(password)
			if err != nil {
				t.Logf("Failed to hash password (first): %v", err)
				return false
			}

			hash2, err := HashPassword(password)
			if err != nil {
				t.Logf("Failed to hash password (second): %v", err)
				return false
			}

			// Verify hashes are different (bcrypt uses salt)
			if hash1 == hash2 {
				t.Logf("Same password produced identical hashes (salt not working)")
				return false
			}

			// Verify both hashes validate the original password
			if !CheckPassword(hash1, password) {
				t.Logf("First hash failed to validate password")
				return false
			}

			if !CheckPassword(hash2, password) {
				t.Logf("Second hash failed to validate password")
				return false
			}

			return true
		},
		gen.AlphaString(),
	))

	properties.Property("hash length is consistent", prop.ForAll(
		func(password string) bool {
			// Skip empty passwords or passwords longer than 72 bytes (bcrypt limit)
			if password == "" || len(password) > 72 {
				return true
			}

			// Hash the password
			hashedPassword, err := HashPassword(password)
			if err != nil {
				return false
			}

			// Bcrypt hashes are always 60 characters long
			expectedLength := 60
			if len(hashedPassword) != expectedLength {
				t.Logf("Hash length incorrect: expected %d, got %d", expectedLength, len(hashedPassword))
				return false
			}

			return true
		},
		gen.AlphaString(),
	))

	properties.Property("hash handles special characters", prop.ForAll(
		func(password string) bool {
			// Skip if base password is too long
			if len(password) > 60 {
				return true
			}

			// Test with various special characters (keeping within 72 byte limit)
			specialPasswords := []string{
				password + "!@#$",
				password + "中文",
				password + "🔒",
			}

			for _, pwd := range specialPasswords {
				if pwd == "" || len(pwd) > 72 {
					continue
				}

				// Hash the password
				hashedPassword, err := HashPassword(pwd)
				if err != nil {
					t.Logf("Failed to hash password with special characters: %v", err)
					return false
				}

				// Verify password validates
				if !CheckPassword(hashedPassword, pwd) {
					t.Logf("Password with special characters failed validation")
					return false
				}
			}

			return true
		},
		gen.AlphaString(),
	))

	properties.Property("hash handles long passwords", prop.ForAll(
		func(length int) bool {
			// Test passwords of various lengths (1-72 characters, bcrypt limit)
			if length < 1 || length > 72 {
				return true
			}

			// Generate password of specified length
			password := ""
			for i := 0; i < length; i++ {
				password += "a"
			}

			// Hash the password
			hashedPassword, err := HashPassword(password)
			if err != nil {
				t.Logf("Failed to hash long password (length %d): %v", length, err)
				return false
			}

			// Verify password validates
			if !CheckPassword(hashedPassword, password) {
				t.Logf("Long password (length %d) failed validation", length)
				return false
			}

			return true
		},
		gen.IntRange(1, 72),
	))

	properties.Property("empty hash always fails validation", prop.ForAll(
		func(password string) bool {
			// Skip empty passwords
			if password == "" {
				return true
			}

			// Verify empty hash always fails
			if CheckPassword("", password) {
				t.Logf("Empty hash incorrectly validated password")
				return false
			}

			return true
		},
		gen.AlphaString(),
	))

	properties.TestingRun(t)
}
