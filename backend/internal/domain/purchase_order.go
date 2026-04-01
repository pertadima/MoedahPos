package domain

import "time"

// PurchaseOrder tracks goods ordered from a supplier.
type PurchaseOrder struct {
	ID          string     `db:"id"`
	StoreID     string     `db:"store_id"`
	SupplierID  *string    `db:"supplier_id"`
	PONumber    string     `db:"po_number"`
	Status      string     `db:"status"` // draft | ordered | received | cancelled
	TotalAmount float64    `db:"total_amount"`
	OrderedBy   string     `db:"ordered_by"`
	ReceivedBy  *string    `db:"received_by"`
	OrderedAt   *time.Time `db:"ordered_at"`
	ReceivedAt  *time.Time `db:"received_at"`
	Notes       string     `db:"notes"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`

	// Populated via JOIN
	SupplierName   *string  `db:"supplier_name"`
	OrderedByName  string   `db:"ordered_by_name"`
	ReceivedByName *string  `db:"received_by_name"`
	Items          []POItem `db:"-"`

	// Payment aggregation (loaded separately)
	AmountPaid    float64 `db:"amount_paid"`
	PaymentStatus string  `db:"payment_status"` // unpaid | partial | paid
}

// POItem is a single line in a purchase order.
type POItem struct {
	ID          string  `db:"id"`
	POID        string  `db:"po_id"`
	ProductID   string  `db:"product_id"`
	Quantity    float64 `db:"quantity"`
	UnitCost    float64 `db:"unit_cost"`
	ReceivedQty float64 `db:"received_qty"`
	Subtotal    float64 `db:"subtotal"`

	// Populated via JOIN
	ProductName string `db:"product_name"`
	ProductSKU  string `db:"product_sku"`
	Unit        string `db:"unit"`
}

// POPayment records a single payment against a purchase order.
type POPayment struct {
	ID        string    `db:"id"`
	POID      string    `db:"po_id"`
	StoreID   string    `db:"store_id"`
	Amount    float64   `db:"amount"`
	Note      *string   `db:"note"`
	PaidBy    string    `db:"paid_by"`
	PaidByName string   `db:"paid_by_name"`
	PaidAt    time.Time `db:"paid_at"`
}

