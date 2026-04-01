package domain

import "time"

// User represents an application user (system-wide, not store-specific).
type User struct {
	ID           string     `db:"id"`
	Name         string     `db:"name"`
	Email        string     `db:"email"`
	PasswordHash string     `db:"password_hash"`
	IsActive     bool       `db:"is_active"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
	DeletedAt    *time.Time `db:"deleted_at"`
}

// RefreshToken represents a stored refresh token record.
type RefreshToken struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	TokenHash string    `db:"token_hash"`
	ExpiresAt time.Time `db:"expires_at"`
	Revoked   bool      `db:"revoked"`
	CreatedAt time.Time `db:"created_at"`
}

// UserStore represents a user's membership in a store with a specific role.
type UserStore struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	StoreID   string    `db:"store_id"`
	RoleID    string    `db:"role_id"`
	IsActive  bool      `db:"is_active"`
	CreatedAt time.Time `db:"created_at"`

	// Joined fields (not stored in users_stores, populated by queries)
	StoreName string `db:"store_name"`
	StoreType string `db:"store_type"`
	RoleName  string `db:"role_name"`
}

// StoreMember is the enriched projection for listing members of a store.
// It joins user, user_stores, and roles tables.
type StoreMember struct {
	UserID    string    `db:"user_id"`
	UserName  string    `db:"user_name"`
	UserEmail string    `db:"user_email"`
	StoreID   string    `db:"store_id"`
	RoleID    string    `db:"role_id"`
	RoleName  string    `db:"role_name"`
	IsActive  bool      `db:"is_active"`
	JoinedAt  time.Time `db:"joined_at"`
}

// StoreAssignment carries a store + role pair for bulk user-store assignment.
type StoreAssignment struct {
	StoreID string
	RoleID  string
}
