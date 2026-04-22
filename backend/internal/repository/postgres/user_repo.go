package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/moedahpos/backend/internal/domain"
)

// UserRepo is the PostgreSQL implementation of repository.UserRepository.
type UserRepo struct {
	db *sqlx.DB
}

// NewUserRepo creates a new UserRepo.
func NewUserRepo(db *sqlx.DB) *UserRepo {
	return &UserRepo{db: db}
}

// Create inserts a new user record and returns the fully hydrated entity.
func (r *UserRepo) Create(ctx context.Context, u *domain.User) (*domain.User, error) {
	const q = `
		INSERT INTO users (name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, name, email, password_hash, is_active, created_at, updated_at, deleted_at`

	row := &domain.User{}
	if err := r.db.QueryRowxContext(ctx, q, u.Name, u.Email, u.PasswordHash).StructScan(row); err != nil {
		return nil, fmt.Errorf("UserRepo.Create: %w", err)
	}
	return row, nil
}

// FindByID retrieves a user by UUID, excluding soft-deleted records.
func (r *UserRepo) FindByID(ctx context.Context, id string) (*domain.User, error) {
	const q = `
		SELECT id, name, email, password_hash, is_active, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL`

	user := &domain.User{}
	if err := r.db.QueryRowxContext(ctx, q, id).StructScan(user); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("UserRepo.FindByID: %w", err)
	}
	return user, nil
}

// FindByEmail retrieves a user by email, excluding soft-deleted records.
func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	const q = `
		SELECT id, name, email, password_hash, is_active, created_at, updated_at, deleted_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL`

	user := &domain.User{}
	if err := r.db.QueryRowxContext(ctx, q, email).StructScan(user); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("UserRepo.FindByEmail: %w", err)
	}
	return user, nil
}

// ExistsByEmail checks whether the given email is already in use.
func (r *UserRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	const q = `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL)`
	var exists bool
	if err := r.db.QueryRowxContext(ctx, q, email).Scan(&exists); err != nil {
		return false, fmt.Errorf("UserRepo.ExistsByEmail: %w", err)
	}
	return exists, nil
}

// FindStoresByUserID returns all stores a user belongs to, along with their role name.
func (r *UserRepo) FindStoresByUserID(ctx context.Context, userID string) ([]domain.UserStore, error) {
	const q = `
		SELECT
			us.id,
			us.user_id,
			us.store_id,
			us.role_id,
			us.is_active,
			us.created_at,
			s.name       AS store_name,
			s.store_type AS store_type,
			ro.name      AS role_name,
			s.loyalty_points_per_rupiah
		FROM user_stores us
		JOIN stores s  ON s.id  = us.store_id AND s.deleted_at IS NULL
		JOIN roles  ro ON ro.id = us.role_id
		WHERE us.user_id = $1 AND us.is_active = true`

	var stores []domain.UserStore
	if err := r.db.SelectContext(ctx, &stores, q, userID); err != nil {
		return nil, fmt.Errorf("UserRepo.FindStoresByUserID: %w", err)
	}
	return stores, nil
}

// ─── Admin Operations ─────────────────────────────────────────────────────────

// ListAll returns a paginated list of users; optionally includes inactive/deleted records.
func (r *UserRepo) ListAll(ctx context.Context, search string, includeInactive bool, page, perPage int) ([]*domain.User, int, error) {
	offset := (page - 1) * perPage
	where := "WHERE 1=1"
	args := []any{}
	argIdx := 1

	if !includeInactive {
		where += " AND deleted_at IS NULL AND is_active = true"
	}
	if search != "" {
		where += fmt.Sprintf(" AND (name ILIKE $%d OR email ILIKE $%d)", argIdx, argIdx+1)
		like := "%" + search + "%"
		args = append(args, like, like)
		argIdx += 2
	}

	// total
	var total int
	cq := fmt.Sprintf("SELECT COUNT(*) FROM users %s", where)
	if err := r.db.QueryRowxContext(ctx, cq, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("UserRepo.ListAll count: %w", err)
	}

	// rows
	args = append(args, perPage, offset)
	q := fmt.Sprintf(`
		SELECT id, name, email, password_hash, is_active, created_at, updated_at, deleted_at
		FROM users %s
		ORDER BY name ASC
		LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)

	var users []*domain.User
	if err := r.db.SelectContext(ctx, &users, q, args...); err != nil {
		return nil, 0, fmt.Errorf("UserRepo.ListAll: %w", err)
	}
	return users, total, nil
}

// Update changes user name and/or email.
func (r *UserRepo) Update(ctx context.Context, id, name, email string) (*domain.User, error) {
	const q = `
		UPDATE users SET name=$2, email=$3, updated_at=NOW()
		WHERE id=$1 AND deleted_at IS NULL
		RETURNING id, name, email, password_hash, is_active, created_at, updated_at, deleted_at`
	u := &domain.User{}
	if err := r.db.QueryRowxContext(ctx, q, id, name, email).StructScan(u); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("UserRepo.Update: %w", err)
	}
	return u, nil
}

// SoftDelete marks a user as deleted and deactivated.
func (r *UserRepo) SoftDelete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE users SET deleted_at=NOW(), is_active=false, updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("UserRepo.SoftDelete: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// Deactivate sets is_active=false without setting deleted_at.
func (r *UserRepo) Deactivate(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE users SET is_active=false, updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("UserRepo.Deactivate: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// ResetPassword updates the password hash for a user.
func (r *UserRepo) ResetPassword(ctx context.Context, id, hash string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE users SET password_hash=$2, updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id, hash)
	if err != nil {
		return fmt.Errorf("UserRepo.ResetPassword: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// SetStores atomically replaces all store memberships for a user.
// It deactivates existing memberships then upserts the new ones.
func (r *UserRepo) SetStores(ctx context.Context, userID string, assignments []domain.StoreAssignment) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("UserRepo.SetStores begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }() // intentional: no-op after commit

	// Deactivate all existing
	if _, err = tx.ExecContext(ctx, `UPDATE user_stores SET is_active=false WHERE user_id=$1`, userID); err != nil {
		return fmt.Errorf("UserRepo.SetStores deactivate: %w", err)
	}

	// Upsert new
	for _, a := range assignments {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO user_stores (user_id, store_id, role_id, is_active)
			VALUES ($1, $2, $3, true)
			ON CONFLICT (user_id, store_id)
			DO UPDATE SET role_id=$3, is_active=true`,
			userID, a.StoreID, a.RoleID)
		if err != nil {
			return fmt.Errorf("UserRepo.SetStores upsert: %w", err)
		}
	}
	return tx.Commit()
}

// ─── RefreshTokenRepo ─────────────────────────────────────────────────────────

// RefreshTokenRepo is the PostgreSQL implementation of repository.RefreshTokenRepository.
type RefreshTokenRepo struct {
	db *sqlx.DB
}

// NewRefreshTokenRepo creates a new RefreshTokenRepo.
func NewRefreshTokenRepo(db *sqlx.DB) *RefreshTokenRepo {
	return &RefreshTokenRepo{db: db}
}

// Create inserts a new refresh token record.
func (r *RefreshTokenRepo) Create(ctx context.Context, t *domain.RefreshToken) error {
	const q = `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)`
	if _, err := r.db.ExecContext(ctx, q, t.UserID, t.TokenHash, t.ExpiresAt); err != nil {
		return fmt.Errorf("RefreshTokenRepo.Create: %w", err)
	}
	return nil
}

// FindByHash finds a valid (non-revoked, non-expired) token by its hash.
func (r *RefreshTokenRepo) FindByHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	const q = `
		SELECT id, user_id, token_hash, expires_at, revoked, created_at
		FROM refresh_tokens
		WHERE token_hash = $1 AND revoked = false AND expires_at > NOW()`

	t := &domain.RefreshToken{}
	if err := r.db.QueryRowxContext(ctx, q, tokenHash).StructScan(t); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("RefreshTokenRepo.FindByHash: %w", err)
	}
	return t, nil
}

// RevokeByID marks a single token as revoked.
func (r *RefreshTokenRepo) RevokeByID(ctx context.Context, id string) error {
	const q = `UPDATE refresh_tokens SET revoked = true WHERE id = $1`
	if _, err := r.db.ExecContext(ctx, q, id); err != nil {
		return fmt.Errorf("RefreshTokenRepo.RevokeByID: %w", err)
	}
	return nil
}

// RevokeAllByUserID revokes every token belonging to a user (logout-all).
func (r *RefreshTokenRepo) RevokeAllByUserID(ctx context.Context, userID string) error {
	const q = `UPDATE refresh_tokens SET revoked = true WHERE user_id = $1`
	if _, err := r.db.ExecContext(ctx, q, userID); err != nil {
		return fmt.Errorf("RefreshTokenRepo.RevokeAllByUserID: %w", err)
	}
	return nil
}

// DeleteExpired removes stale token rows older than `before`.
func (r *RefreshTokenRepo) DeleteExpired(ctx context.Context, before time.Time) error {
	const q = `DELETE FROM refresh_tokens WHERE expires_at < $1`
	if _, err := r.db.ExecContext(ctx, q, before); err != nil {
		return fmt.Errorf("RefreshTokenRepo.DeleteExpired: %w", err)
	}
	return nil
}
