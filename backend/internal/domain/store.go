package domain

import "time"

// Store represents a physical or virtual point-of-sale location.
type Store struct {
	ID        string     `db:"id"`
	Name      string     `db:"name"`
	Address   string     `db:"address"`
	Phone     string     `db:"phone"`
	TaxNumber string     `db:"tax_number"`
	Currency  string     `db:"currency"`
	IsActive  bool       `db:"is_active"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}
