package rbac

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// RoleStore is an in-memory, pre-loaded map of role → set of permissions.
// It is populated once at application startup, eliminating per-request DB queries.
type RoleStore struct {
	// permissions maps roleName → { permissionName: true }
	permissions map[string]map[string]bool
}

type rolePermRow struct {
	RoleName       string `db:"role_name"`
	PermissionName string `db:"permission_name"`
}

// New loads all role-permission assignments from the database and builds
// the in-memory map. Call this once during application startup.
func New(db *sqlx.DB) (*RoleStore, error) {
	const q = `
		SELECT r.name AS role_name, p.name AS permission_name
		FROM role_permissions rp
		JOIN roles       r ON r.id = rp.role_id
		JOIN permissions p ON p.id = rp.permission_id`

	var rows []rolePermRow
	if err := db.SelectContext(context.Background(), &rows, q); err != nil {
		return nil, fmt.Errorf("rbac: loading permissions: %w", err)
	}

	store := &RoleStore{permissions: make(map[string]map[string]bool)}
	for _, row := range rows {
		if store.permissions[row.RoleName] == nil {
			store.permissions[row.RoleName] = make(map[string]bool)
		}
		store.permissions[row.RoleName][row.PermissionName] = true
	}

	return store, nil
}

// Has returns true if the given role has the specified permission.
func (rs *RoleStore) Has(roleName, permission string) bool {
	if perms, ok := rs.permissions[roleName]; ok {
		return perms[permission]
	}
	return false
}

// IsSuperAdmin returns true if the role name is "superadmin".
func IsSuperAdmin(roleName string) bool {
	return roleName == "superadmin"
}
