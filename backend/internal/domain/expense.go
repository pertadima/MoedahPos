package domain

import (
	"time"
)

// ExpenseCategory represents a global category for expenses.
type ExpenseCategory struct {
	ID          string     `db:"id"          json:"id"`
	Name        string     `db:"name"        json:"name"`
	Description string     `db:"description" json:"description"`
	CreatedAt   time.Time  `db:"created_at"  json:"created_at"`
}

// Expense represents an operational cost for a specific store.
type Expense struct {
	ID           string    `db:"id"           json:"id"`
	StoreID      string    `db:"store_id"     json:"store_id"`
	CategoryID   string    `db:"category_id"  json:"category_id"`
	CategoryName string    `db:"category_name" json:"category_name"` // Joined field
	Amount       float64   `db:"amount"       json:"amount"`
	ExpenseDate  time.Time `db:"expense_date" json:"expense_date"`
	Notes        string    `db:"notes"        json:"notes"`
	CreatedBy    *string   `db:"created_by"   json:"created_by"`
	CreatedAt    time.Time `db:"created_at"   json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"   json:"updated_at"`
}
