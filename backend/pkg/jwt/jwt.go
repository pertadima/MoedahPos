package jwt

import (
	"fmt"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims are the custom JWT payload fields.
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	gojwt.RegisteredClaims
}

// Manager handles JWT generation and parsing.
type Manager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// New creates a JWT Manager.
func New(secret string, accessTTL, refreshTTL time.Duration) *Manager {
	return &Manager{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

// GenerateAccessToken creates a signed access token for the given user.
func (m *Manager) GenerateAccessToken(userID, email string) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: gojwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  gojwt.NewNumericDate(now),
			ExpiresAt: gojwt.NewNumericDate(now.Add(m.accessTTL)),
			ID:        uuid.NewString(),
		},
	}
	token := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// GenerateRefreshToken creates a random opaque refresh token string.
// The caller is responsible for hashing and storing it.
func (m *Manager) GenerateRefreshToken() (string, time.Time, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", time.Time{}, fmt.Errorf("generating refresh token: %w", err)
	}
	expiresAt := time.Now().Add(m.refreshTTL)
	return id.String(), expiresAt, nil
}

// ParseAccessToken validates and parses a JWT string, returning its claims.
func (m *Manager) ParseAccessToken(tokenStr string) (*Claims, error) {
	token, err := gojwt.ParseWithClaims(tokenStr, &Claims{}, func(t *gojwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*gojwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parsing token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

// AccessTTLSeconds returns the access token TTL in seconds (for API response).
func (m *Manager) AccessTTLSeconds() int64 {
	return int64(m.accessTTL.Seconds())
}
