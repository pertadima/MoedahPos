package middleware

import (
	"net/http"

	"github.com/moedahpos/backend/pkg/rbac"
	"github.com/moedahpos/backend/pkg/response"
)

// RequirePermission returns middleware that checks the current user's store-role
// has the specified permission. Must be used AFTER Authenticate and StoreContext.
func RequirePermission(roleStore *rbac.RoleStore, permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			roleName := StoreRoleFromContext(r.Context())

			// superadmin bypasses all permission checks.
			if rbac.IsSuperAdmin(roleName) {
				next.ServeHTTP(w, r)
				return
			}

			if !roleStore.Has(roleName, permission) {
				response.Forbidden(w)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
