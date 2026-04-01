package domain

import "time"

// Role represents a named set of permissions (e.g., admin, cashier).
type Role struct {
	ID          string    `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
	Permissions []string  `db:"-"` // populated by service layer
}

// Permission represents a granular action (e.g., "products.create").
type Permission struct {
	ID          string `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
}
