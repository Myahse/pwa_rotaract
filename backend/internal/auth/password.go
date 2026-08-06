package auth

import (
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const defaultCost = 12

// OAuthPasswordMarker is stored in password_hash for Google-only accounts.
// Password login must reject these markers.
const OAuthPasswordMarker = "oauth:google"

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), defaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

func CheckPassword(hash, password string) error {
	if hash == "" || strings.HasPrefix(hash, "oauth:") {
		return bcrypt.ErrMismatchedHashAndPassword
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func IsOAuthPassword(hash string) bool {
	return strings.HasPrefix(hash, "oauth:")
}
