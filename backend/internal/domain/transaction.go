package domain

import "time"

// Transaction is a completed sale record.
type Transaction struct {
	ID                string    `db:"id"`
	StoreID           string    `db:"store_id"`
	CashierID         string    `db:"cashier_id"`
	TableID           *string   `db:"table_id"` // nil for retail; set for restaurant draft orders
	CustomerName      string    `db:"customer_name"`
	CustomerPhone     string    `db:"customer_phone"`
	Subtotal          float64   `db:"subtotal"`
	DiscountAmt       float64   `db:"discount_amt"`
	TaxAmt            float64   `db:"tax_amt"`
	Total             float64   `db:"total"`
	PaymentMethod     string    `db:"payment_method"` // cash | card | qris | transfer
	PaymentAmount     float64   `db:"payment_amount"`
	ChangeAmount      float64   `db:"change_amount"`
	Status            string    `db:"status"` // draft | completed | voided
	Notes             string    `db:"notes"`
	CartDiscountType  string    `db:"cart_discount_type"`  // PERCENTAGE | FIXED
	CartDiscountValue float64   `db:"cart_discount_value"` // the cart-level discount amount
	CreatedAt         time.Time `db:"created_at"`
	UpdatedAt         time.Time `db:"updated_at"`
	ServerUpdatedAt   time.Time `db:"server_updated_at"`
	SyncVersion       int       `db:"sync_version"`

	// Populated via JOIN
	CashierName string            `db:"cashier_name"`
	TableNumber *string           `db:"table_number"`
	Items       []TransactionItem `db:"-"`
}

// TransactionItem is a single line in a transaction (price snapshot at sale time).
type TransactionItem struct {
	ID                    string     `db:"id"`
	TransactionID         string     `db:"transaction_id"`
	ProductID             *string    `db:"product_id"`
	MenuItemID            *string    `db:"menu_item_id"`
	ProductName           string     `db:"product_name"` // snapshot
	SKU                   string     `db:"sku"`          // snapshot
	Quantity              float64    `db:"quantity"`
	OriginalPrice         float64    `db:"original_price"`          // product's sell_price at time of sale
	UnitPrice             float64    `db:"unit_price"`              // final price after ALL discounts
	CostPrice             float64    `db:"cost_price"`              // snapshot of cost at time of sale
	DiscountPct           float64    `db:"discount_pct"`            // kept for legacy
	DiscountType          string     `db:"discount_type"`           // PERCENTAGE | FIXED | OVERRIDE
	DiscountValue         float64    `db:"discount_value"`          // item discount amount
	CartDiscountAllocated float64    `db:"cart_discount_allocated"` // cart discount allocated per unit
	TaxRate               float64    `db:"tax_rate"`
	Subtotal              float64    `db:"subtotal"`
	Status                string     `db:"status"`       // KDS status: pending | completed
	CompletedAt           *time.Time `db:"completed_at"` // KDS completion time
}

// CreateTransactionInput carries service-calculated values to the repository.
// The service resolves products and calculates all amounts before calling the repo.
type CreateTransactionInput struct {
	StoreID           string
	CashierID         string
	TableID           *string // nil = retail
	Status            string  // "draft" or "completed"
	CustomerName      string
	CustomerPhone     string
	PaymentMethod     string
	PaymentAmount     float64
	ChangeAmount      float64
	Notes             string
	Subtotal          float64
	DiscountAmt       float64
	TaxAmt            float64
	Total             float64
	CartDiscountType  string  // PERCENTAGE | FIXED
	CartDiscountValue float64 // audit: raw value entered by cashier
	Items             []CreateTransactionItemInput
}

// PayDraftInput is what the service passes to the repo when finalizing a held order.
type PayDraftInput struct {
	TransactionID string
	PaymentMethod string
	PaymentAmount float64
	ChangeAmount  float64
	CustomerName  string
	CustomerPhone string
}

// CreateTransactionItemInput is a pre-calculated item ready for DB insertion.
type CreateTransactionItemInput struct {
	ProductID             *string
	MenuItemID            *string // set for restaurant menu items
	ProductName           string
	SKU                   string
	Quantity              float64
	OriginalPrice         float64 // product's sell_price at time of sale
	UnitPrice             float64 // final price after ALL discounts (item + cart)
	CostPrice             float64 // snapshot of cost at time of sale
	DiscountPct           float64 // kept for legacy; equals DiscountValue when type=PERCENTAGE
	DiscountType          string  // PERCENTAGE | FIXED | OVERRIDE
	DiscountValue         float64 // item discount amount
	CartDiscountAllocated float64 // portion of cart discount allocated to this item (per unit)
	TaxRate               float64
	Subtotal              float64
	Status                string
}
