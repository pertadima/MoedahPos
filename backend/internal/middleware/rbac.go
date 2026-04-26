package middleware

import (
	"net/http"

	"github.com/moedahpos/backend/pkg/rbac"
	"github.com/moedahpos/backend/pkg/response"
)

// RequirePermission returns middleware that checks if the current user's store permissions
// include the specified permission. Must be used AFTER Authenticate and StoreContext.
func RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			roles := StoreRolesFromContext(r.Context())

			// superadmin bypasses all permission checks.
			for _, roleName := range roles {
				if rbac.IsSuperAdmin(roleName) {
					next.ServeHTTP(w, r)
					return
				}
			}

			permissions := StorePermissionsFromContext(r.Context())
			hasPerm := false
			for _, p := range permissions {
				if p == permission {
					hasPerm = true
					break
				}
			}

			if !hasPerm {
				response.Forbidden(w)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
