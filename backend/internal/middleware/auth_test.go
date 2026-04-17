package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/moedahpos/backend/pkg/jwt"
)

func TestAuthenticate(t *testing.T) {
	jwtMgr := jwt.New("secret", time.Hour, time.Hour)
	authMiddleware := Authenticate(jwtMgr)

	t.Run("Success", func(t *testing.T) {
		token, _ := jwtMgr.GenerateAccessToken("u123", "test@example.com")

		handler := authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := UserIDFromContext(r.Context())
			email := EmailFromContext(r.Context())
			assert.Equal(t, "u123", userID)
			assert.Equal(t, "test@example.com", email)
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Missing Header", func(t *testing.T) {
		handler := authMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Invalid Token", func(t *testing.T) {
		handler := authMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
