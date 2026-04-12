package domain

import "time"

// IncomeCategory is a classification for non-POS cash inflows.
type IncomeCategory struct {
	ID          string     `db:"id"`
	Name        string     `db:"name"`
	Description *string    `db:"description"`
	IsActive    bool       `db:"is_active"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at"`
}

// Income records a single non-POS cash inflow for a store.
type Income struct {
	ID            string    `db:"id"`
	StoreID       string    `db:"store_id"`
	CategoryID    string    `db:"category_id"`
	Amount        float64   `db:"amount"`
	IncomeDate    time.Time `db:"income_date"`
	PaymentMethod string    `db:"payment_method"`
	Reference     *string   `db:"reference"`
	Notes         *string   `db:"notes"`
	CreatedBy     *string   `db:"created_by"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`

	// Populated via JOIN
	CategoryName  string  `db:"category_name"`
	CreatedByName *string `db:"created_by_name"`
}
