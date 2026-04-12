package domain

import (
	"time"
)

// ExpenseCategory represents a global category for expenses.
type ExpenseCategory struct {
	ID          string     `db:"id"          json:"id"`
	Name        string     `db:"name"        json:"name"`
	Description string     `db:"description" json:"description"`
	IsActive    bool       `db:"is_active"   json:"is_active"`
	CreatedAt   time.Time  `db:"created_at"  json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"  json:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at"  json:"-"`
}

// Expense represents an operational cost for a specific store.
type Expense struct {
	ID            string    `db:"id"           json:"id"`
	StoreID       string    `db:"store_id"     json:"store_id"`
	CategoryID    string    `db:"category_id"  json:"category_id"`
	CategoryName  string    `db:"category_name" json:"category_name"` // Joined field
	Amount        float64   `db:"amount"       json:"amount"`
	ExpenseDate   time.Time `db:"expense_date" json:"expense_date"`
	Notes         string    `db:"notes"        json:"notes"`
	PaymentStatus string    `db:"payment_status" json:"payment_status"` // paid, unpaid, canceled
	CreatedBy     *string   `db:"created_by"   json:"created_by"`
	CreatedAt     time.Time `db:"created_at"   json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"   json:"updated_at"`
}

// RecurringExpense represents an expense that is generated automatically on a schedule.
type RecurringExpense struct {
	ID              string     `db:"id"               json:"id"`
	StoreID         string     `db:"store_id"         json:"store_id"`
	CategoryID      string     `db:"category_id"      json:"category_id"`
	CategoryName    string     `db:"category_name"    json:"category_name"` // Joined field
	Name            string     `db:"name"             json:"name"`
	Amount          float64    `db:"amount"           json:"amount"`
	Interval        string     `db:"interval"         json:"interval"` // daily, weekly, monthly, yearly
	IntervalValue   int        `db:"interval_value"   json:"interval_value"`
	StartDate       time.Time  `db:"start_date"       json:"start_date"`
	EndDate         *time.Time `db:"end_date"         json:"end_date"`
	NextRunDate     time.Time  `db:"next_run_date"    json:"next_run_date"`
	Notes           string     `db:"notes"            json:"notes"`
	IsActive        bool       `db:"is_active"        json:"is_active"`
	CreatedBy       *string    `db:"created_by"       json:"created_by"`
	CreatedAt       time.Time  `db:"created_at"       json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"       json:"updated_at"`
	LastGeneratedAt *time.Time `db:"last_generated_at" json:"last_generated_at"`
}
