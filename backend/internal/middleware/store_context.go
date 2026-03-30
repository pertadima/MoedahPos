package middleware

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"github.com/moedahpos/backend/pkg/response"
)

type storeContextKey string

const (
	storeIDContextKey   storeContextKey = "store_id"
	storeRoleContextKey storeContextKey = "store_role"
)

// StoreContext is middleware that:
//  1. Reads :storeId from the URL.
//  2. Validates the authenticated user has an active membership in that store.
//  3. Injects store_id and the user's role name into the request context.
//
// It must be used AFTER the Authenticate middleware.
func StoreContext(db *sqlx.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			storeID := chi.URLParam(r, "storeId")
			if storeID == "" {
				response.Error(w, http.StatusBadRequest, "store_id is required")
				return
			}

			userID := UserIDFromContext(r.Context())
			if userID == "" {
				response.Unauthorized(w, "")
				return
			}

			roleName, err := getUserRoleInStore(r.Context(), db, userID, storeID)
			if err != nil {
				response.Forbidden(w)
				return
			}

			ctx := context.WithValue(r.Context(), storeIDContextKey, storeID)
			ctx = context.WithValue(ctx, storeRoleContextKey, roleName)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// StoreIDFromContext extracts the current store ID from the request context.
func StoreIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(storeIDContextKey).(string)
	return v
}

// StoreRoleFromContext extracts the user's role name in the current store from context.
func StoreRoleFromContext(ctx context.Context) string {
	v, _ := ctx.Value(storeRoleContextKey).(string)
	return v
}

// getUserRoleInStore queries the user_stores + roles tables to get the user's role name.
func getUserRoleInStore(ctx context.Context, db *sqlx.DB, userID, storeID string) (string, error) {
	const q = `
		SELECT r.name
		FROM user_stores us
		JOIN roles r ON r.id = us.role_id
		WHERE us.user_id = $1 AND us.store_id = $2 AND us.is_active = true`
	var roleName string
	if err := db.QueryRowxContext(ctx, q, userID, storeID).Scan(&roleName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("user has no access to store")
		}
		return "", err
	}
	return roleName, nil
}
