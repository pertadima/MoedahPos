package middleware

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/moedahpos/backend/pkg/response"
)

// RequireAdminOrSuperAdmin checks that the authenticated user holds the
// 'superadmin' or 'admin' role in at least one store (or is a superadmin).
// This is used for global admin endpoints (/admin/*) that are not store-scoped.
func RequireAdminOrSuperAdmin(db *sqlx.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := UserIDFromContext(r.Context())
			if userID == "" {
				response.Forbidden(w)
				return
			}

			const q = `
				SELECT COUNT(*)
				FROM user_stores us
				JOIN roles r ON r.id = us.role_id
				WHERE us.user_id = $1
				  AND us.is_active = true
				  AND r.name IN ('superadmin', 'admin')`

			var count int
			if err := db.QueryRowContext(r.Context(), q, userID).Scan(&count); err != nil || count == 0 {
				response.Forbidden(w)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
