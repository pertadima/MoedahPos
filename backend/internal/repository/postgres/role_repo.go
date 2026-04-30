package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/repository"
)

// RoleRepo is the PostgreSQL implementation of repository.RoleRepository.
type RoleRepo struct {
	db *sqlx.DB
}

// NewRoleRepository creates a new RoleRepository.
func NewRoleRepository(db *sql.DB) repository.RoleRepository {
	return &RoleRepo{db: sqlx.NewDb(db, "postgres")}
}

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

// CreateRole creates a new role and assigns the given permissions.
func (r *RoleRepo) CreateRole(ctx context.Context, role *domain.Role, permissionIDs []string) (*domain.Role, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("RoleRepo.CreateRole begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	const qInsertRole = `
		INSERT INTO roles (id, name, description, created_at, updated_at)
		VALUES (uuid_generate_v4(), $1, $2, NOW(), NOW())
		RETURNING id, name, description, created_at, updated_at`

	if err := tx.QueryRowxContext(ctx, qInsertRole, role.Name, role.Description).StructScan(role); err != nil {
		return nil, fmt.Errorf("RoleRepo.CreateRole insert role: %w", err)
	}

	if len(permissionIDs) > 0 {
		const qInsertPerm = `INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2)`
		for _, pid := range permissionIDs {
			if _, err := tx.ExecContext(ctx, qInsertPerm, role.ID, pid); err != nil {
				return nil, fmt.Errorf("RoleRepo.CreateRole insert perm: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("RoleRepo.CreateRole commit: %w", err)
	}

	return role, nil
}

// UpdateRole updates an existing role's details and permissions.
func (r *RoleRepo) UpdateRole(ctx context.Context, role *domain.Role, permissionIDs []string) (*domain.Role, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("RoleRepo.UpdateRole begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	const qUpdateRole = `
		UPDATE roles 
		SET name = $1, description = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING id, name, description, created_at, updated_at`

	if err := tx.QueryRowxContext(ctx, qUpdateRole, role.Name, role.Description, role.ID).StructScan(role); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("role not found")
		}
		return nil, fmt.Errorf("RoleRepo.UpdateRole update role: %w", err)
	}

	const qDeletePerms = `DELETE FROM role_permissions WHERE role_id = $1`
	if _, err := tx.ExecContext(ctx, qDeletePerms, role.ID); err != nil {
		return nil, fmt.Errorf("RoleRepo.UpdateRole delete perms: %w", err)
	}

	if len(permissionIDs) > 0 {
		const qInsertPerm = `INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2)`
		for _, pid := range permissionIDs {
			if _, err := tx.ExecContext(ctx, qInsertPerm, role.ID, pid); err != nil {
				return nil, fmt.Errorf("RoleRepo.UpdateRole insert perm: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("RoleRepo.UpdateRole commit: %w", err)
	}

	return role, nil
}

// DeleteRole deletes a role and its associated permissions.
func (r *RoleRepo) DeleteRole(ctx context.Context, id string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("RoleRepo.DeleteRole begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	const qDeletePerms = `DELETE FROM role_permissions WHERE role_id = $1`
	if _, err := tx.ExecContext(ctx, qDeletePerms, id); err != nil {
		return fmt.Errorf("RoleRepo.DeleteRole delete perms: %w", err)
	}

	const qDeleteRole = `DELETE FROM roles WHERE id = $1`
	res, err := tx.ExecContext(ctx, qDeleteRole, id)
	if err != nil {
		return fmt.Errorf("RoleRepo.DeleteRole delete role: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("RoleRepo.DeleteRole affected rows: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("role not found")
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("RoleRepo.DeleteRole commit: %w", err)
	}

	return nil
}

// ListPermissions returns all available permissions.
func (r *RoleRepo) ListPermissions(ctx context.Context) ([]*domain.Permission, error) {
	const q = `SELECT id, name, description FROM permissions ORDER BY name`
	var perms []*domain.Permission
	if err := r.db.SelectContext(ctx, &perms, q); err != nil {
		return nil, fmt.Errorf("RoleRepo.ListPermissions: %w", err)
	}
	return perms, nil
}
