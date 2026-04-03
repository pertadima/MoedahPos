package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/moedahpos/backend/internal/domain"
)

// RoleRepo is the PostgreSQL implementation of repository.RoleRepository.
type RoleRepo struct {
	db *sqlx.DB
}

func NewRoleRepo(db *sqlx.DB) *RoleRepo { return &RoleRepo{db: db} }

// ListRoles returns all roles with their associated permission names.
func (r *RoleRepo) ListRoles(ctx context.Context) ([]*domain.Role, error) {
	const qRoles = `SELECT id, name, description, created_at, updated_at FROM roles ORDER BY name`
	var roles []*domain.Role
	if err := r.db.SelectContext(ctx, &roles, qRoles); err != nil {
		return nil, fmt.Errorf("RoleRepo.ListRoles: %w", err)
	}

	// Load permissions per role
	const qPerms = `
		SELECT r.id AS role_id, p.name AS perm_name
		FROM role_permissions rp
		JOIN roles r ON r.id = rp.role_id
		JOIN permissions p ON p.id = rp.permission_id`

	rows, err := r.db.QueryxContext(ctx, qPerms)
	if err != nil {
		return nil, fmt.Errorf("RoleRepo.ListRoles perms: %w", err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			_ = fmt.Errorf("RoleRepo.ListRoles rows.Close: %w", cerr)
		}
	}()

	permMap := make(map[string][]string)
	for rows.Next() {
		var roleID, permName string
		if err := rows.Scan(&roleID, &permName); err != nil {
			return nil, fmt.Errorf("RoleRepo.ListRoles scan: %w", err)
		}
		permMap[roleID] = append(permMap[roleID], permName)
	}

	for _, role := range roles {
		role.Permissions = permMap[role.ID]
	}
	return roles, nil
}
