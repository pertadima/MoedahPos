package domain

import "time"

// Supplier is a goods provider linked to purchase orders.
type Supplier struct {
	ID          string     `db:"id"`
	Name        string     `db:"name"`
	ContactName string     `db:"contact_name"`
	Phone       string     `db:"phone"`
	Email       string     `db:"email"`
	Address     string     `db:"address"`
	IsActive    bool       `db:"is_active"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at"`
}
