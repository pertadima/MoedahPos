package jwt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJWTManager(t *testing.T) {
	secret := "secret-key"
	accessTTL := 15 * time.Minute
	refreshTTL := 24 * time.Hour
	m := New(secret, accessTTL, refreshTTL)

	t.Run("Generate and Parse Access Token", func(t *testing.T) {
		userID := "u1"
		email := "test@example.com"
		token, err := m.GenerateAccessToken(userID, email)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		claims, err := m.ParseAccessToken(token)
		assert.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, email, claims.Email)
	})

	t.Run("Invalid Token Parsing", func(t *testing.T) {
		_, err := m.ParseAccessToken("invalid.token.here")
		assert.Error(t, err)
	})

	t.Run("Generate Refresh Token", func(t *testing.T) {
		token, expiresAt, err := m.GenerateRefreshToken()
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.WithinDuration(t, time.Now().Add(refreshTTL), expiresAt, 1*time.Second)
	})

	t.Run("AccessTTLSeconds", func(t *testing.T) {
		assert.Equal(t, int64(900), m.AccessTTLSeconds())
	})
}
