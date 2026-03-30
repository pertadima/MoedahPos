package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// Password hashes a plain-text password using bcrypt with the given cost.
func Password(plain string, cost int) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), cost)
	if err != nil {
		return "", fmt.Errorf("hashing password: %w", err)
	}
	return string(hashed), nil
}

// CheckPassword compares a plain-text password with a bcrypt hash.
// Returns nil on match, an error otherwise.
func CheckPassword(plain, hashed string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain))
}

// SHA256 returns a SHA-256 hex digest of the given string.
// Used for hashing refresh tokens before storing them in the database.
func SHA256(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
