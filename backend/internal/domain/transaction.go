package domain

import "time"

// Transaction is a completed sale record.
type Transaction struct {
	ID            string    `db:"id"`
	StoreID       string    `db:"store_id"`
	CashierID     string    `db:"cashier_id"`
	CustomerName  string    `db:"customer_name"`
	CustomerPhone string    `db:"customer_phone"`
	Subtotal      float64   `db:"subtotal"`
	DiscountAmt   float64   `db:"discount_amt"`
	TaxAmt        float64   `db:"tax_amt"`
	Total         float64   `db:"total"`
	PaymentMethod string    `db:"payment_method"` // cash | card | qris | transfer
	PaymentAmount float64   `db:"payment_amount"`
	ChangeAmount  float64   `db:"change_amount"`
	Status        string    `db:"status"` // draft | completed | voided
	Notes         string    `db:"notes"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`

	// Populated via JOIN
	CashierName string            `db:"cashier_name"`
	Items       []TransactionItem `db:"-"`
}

// TransactionItem is a single line in a transaction (price snapshot at sale time).
type TransactionItem struct {
	ID            string  `db:"id"`
	TransactionID string  `db:"transaction_id"`
	ProductID     *string `db:"product_id"`
	ProductName   string  `db:"product_name"` // snapshot
	SKU           string  `db:"sku"`           // snapshot
	Quantity      float64 `db:"quantity"`
	UnitPrice     float64 `db:"unit_price"`
	DiscountPct   float64 `db:"discount_pct"`
	TaxRate       float64 `db:"tax_rate"`
	Subtotal      float64 `db:"subtotal"`
}

// CreateTransactionInput carries service-calculated values to the repository.
// The service resolves products and calculates all amounts before calling the repo.
type CreateTransactionInput struct {
	StoreID       string
	CashierID     string
	CustomerName  string
	CustomerPhone string
	PaymentMethod string
	PaymentAmount float64
	ChangeAmount  float64
	Notes         string
	Subtotal      float64
	DiscountAmt   float64
	TaxAmt        float64
	Total         float64
	Items         []CreateTransactionItemInput
}

// CreateTransactionItemInput is a pre-calculated item ready for DB insertion.
type CreateTransactionItemInput struct {
	ProductID   *string
	MenuItemID  *string // set for restaurant menu items
	ProductName string
	SKU         string
	Quantity    float64
	UnitPrice   float64
	DiscountPct float64
	TaxRate     float64
	Subtotal    float64
}
