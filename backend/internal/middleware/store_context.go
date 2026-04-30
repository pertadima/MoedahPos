package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"

	"github.com/moedahpos/backend/pkg/response"
)

type storeContextKey string

const (
	storeIDContextKey          storeContextKey = "store_id"
	storeRolesContextKey       storeContextKey = "store_roles"
	storePermissionsContextKey storeContextKey = "store_permissions"
)

// StoreContext is middleware that:
//  1. Reads :storeId from the URL.
//  2. Validates the authenticated user has an active membership in that store.
//  3. Injects store_id, the user's role names, and permissions into the request context.
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

			roles, permissions, err := getUserRolesAndPermissionsInStore(r.Context(), db, userID, storeID)
			if err != nil {
				response.Forbidden(w)
				return
			}

			ctx := context.WithValue(r.Context(), storeIDContextKey, storeID)
			ctx = context.WithValue(ctx, storeRolesContextKey, roles)
			ctx = context.WithValue(ctx, storePermissionsContextKey, permissions)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// StoreIDFromContext extracts the current store ID from the request context.
func StoreIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(storeIDContextKey).(string)
	return v
}

// StoreRolesFromContext extracts the user's role names in the current store from context.
func StoreRolesFromContext(ctx context.Context) []string {
	v, _ := ctx.Value(storeRolesContextKey).([]string)
	return v
}

// StorePermissionsFromContext extracts the user's permissions in the current store from context.
func StorePermissionsFromContext(ctx context.Context) []string {
	v, _ := ctx.Value(storePermissionsContextKey).([]string)
	return v
}

// getUserRolesAndPermissionsInStore queries the user_stores + roles + role_permissions to get the user's roles and permissions.
func getUserRolesAndPermissionsInStore(ctx context.Context, db *sqlx.DB, userID, storeID string) ([]string, []string, error) {
	// 1. Get roles
	const roleQuery = `
		SELECT r.name
		FROM user_stores us
		JOIN roles r ON r.id = us.role_id
		WHERE us.user_id = $1 AND us.store_id = $2 AND us.is_active = true`

	var roles []string
	if err := db.SelectContext(ctx, &roles, roleQuery, userID, storeID); err != nil {
		return nil, nil, err
	}
	if len(roles) == 0 {
		return nil, nil, fmt.Errorf("user has no active access to store")
	}

	// 2. Get distinct permissions for those roles
	const permQuery = `
		SELECT DISTINCT p.name
		FROM user_stores us
		JOIN role_permissions rp ON rp.role_id = us.role_id
		JOIN permissions p ON p.id = rp.permission_id
		WHERE us.user_id = $1 AND us.store_id = $2 AND us.is_active = true`

	var permissions []string
	if err := db.SelectContext(ctx, &permissions, permQuery, userID, storeID); err != nil {
		return nil, nil, err
	}

	return roles, permissions, nil
}
