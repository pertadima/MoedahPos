package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/moedahpos/backend/pkg/jwt"
	"github.com/moedahpos/backend/pkg/response"
)

// contextKey is an unexported type for context keys to prevent collisions.
type contextKey string

const (
	UserIDKey contextKey = "user_id"
	EmailKey  contextKey = "email"
)

// Authenticate is a middleware that validates the Authorization: Bearer <token> header.
// On success, it injects user_id and email into the request context.
func Authenticate(jwtMgr *jwt.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" || !strings.HasPrefix(header, "Bearer ") {
				response.Unauthorized(w, "Missing or invalid Authorization header")
				return
			}

			tokenStr := strings.TrimPrefix(header, "Bearer ")

			claims, err := jwtMgr.ParseAccessToken(tokenStr)
			if err != nil {
				response.Unauthorized(w, "Invalid or expired access token")
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, EmailKey, claims.Email)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext extracts the authenticated user's ID from the context.
func UserIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(UserIDKey).(string)
	return id
}

// EmailFromContext extracts the authenticated user's email from the context.
func EmailFromContext(ctx context.Context) string {
	email, _ := ctx.Value(EmailKey).(string)
	return email
}
