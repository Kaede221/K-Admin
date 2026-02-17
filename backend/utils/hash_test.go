package utils

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "testPassword123"

	// Test hashing
	hashed, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hashed == "" {
		t.Error("Hashed password should not be empty")
	}

	if hashed == password {
		t.Error("Hashed password should not equal plain password")
	}
}

func TestCheckPassword(t *testing.T) {
	password := "testPassword123"
	wrongPassword := "wrongPassword"

	// Hash the password
	hashed, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Test correct password
	if !CheckPassword(hashed, password) {
		t.Error("CheckPassword should return true for correct password")
	}

	// Test wrong password
	if CheckPassword(hashed, wrongPassword) {
		t.Error("CheckPassword should return false for wrong password")
	}
}

func TestHashPassword_EmptyString(t *testing.T) {
	hashed, err := HashPassword("")
	if err != nil {
		t.Fatalf("HashPassword should handle empty string: %v", err)
	}

	if hashed == "" {
		t.Error("Hashed password should not be empty even for empty input")
	}
}
