package domain

import "time"

// PurchaseOrder tracks goods ordered from a supplier.
type PurchaseOrder struct {
	ID                string     `db:"id"`
	StoreID           string     `db:"store_id"`
	SupplierID        *string    `db:"supplier_id"`
	PONumber          string     `db:"po_number"`
	Status            string     `db:"status"` // draft | ordered | received | canceled
	TotalAmount       float64    `db:"total_amount"`
	OrderedBy         string     `db:"ordered_by"`
	ReceivedBy        *string    `db:"received_by"`
	OrderedAt         *time.Time `db:"ordered_at"`
	ReceivedAt        *time.Time `db:"received_at"`
	Notes             *string    `db:"notes"`
	CreatedAt         time.Time  `db:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at"`
	SupplierSignature *string    `db:"supplier_signature"`
	BuyerSignature    *string    `db:"buyer_signature"`
	SupplierSignedAt  *time.Time `db:"supplier_signed_at"`
	BuyerSignedAt     *time.Time `db:"buyer_signed_at"`

	// Populated via JOIN
	SupplierName   *string  `db:"supplier_name"`
	OrderedByName  string   `db:"ordered_by_name"`
	ReceivedByName *string  `db:"received_by_name"`
	TotalItems     int      `db:"total_items"`
	Items          []POItem `db:"-"`

	// Payment aggregation (loaded separately)
	AmountPaid    float64    `db:"amount_paid"`
	AmountDue     float64    `db:"amount_due"`
	NextDeadline  *time.Time `db:"next_deadline"`
	PaymentStatus string     `db:"payment_status"` // unpaid | partial | paid
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

// POPayment records a simple payment against a purchase order (legacy).
type POPayment struct {
	ID         string    `db:"id"`
	POID       string    `db:"po_id"`
	StoreID    string    `db:"store_id"`
	Amount     float64   `db:"amount"`
	Note       *string   `db:"note"`
	PaidBy     string    `db:"paid_by"`
	PaidByName string    `db:"paid_by_name"`
	PaidAt     time.Time `db:"paid_at"`
}

// ─── Termin (Installment) System ───────────────────────────────────────────────────

// POTermin is one installment in a purchase order payment schedule.
// status transitions: unpaid → partial → paid (overdue is a derived view).
type POTermin struct {
	ID           string    `db:"id"`
	POID         string    `db:"po_id"`
	TerminNumber int       `db:"termin_number"`
	Amount       float64   `db:"amount"`
	DueDate      time.Time `db:"due_date"`
	Status       string    `db:"status"` // unpaid | partial | paid | overdue
	Notes        string    `db:"notes"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`

	// Populated via aggregation
	AmountPaid float64         `db:"amount_paid"`
	AmountDue  float64         `db:"amount_due"`
	Payments   []PaymentRecord `db:"-"`
}

// PaymentRecord is a single payment transaction against one POTermin.
type PaymentRecord struct {
	ID            string    `db:"id"`
	TerminID      string    `db:"termin_id"`
	AmountPaid    float64   `db:"amount_paid"`
	PaymentDate   time.Time `db:"payment_date"`
	PaymentMethod string    `db:"payment_method"` // cash | transfer | check | other
	Notes         string    `db:"notes"`
	RecordedBy    *string   `db:"recorded_by"`
	CreatedAt     time.Time `db:"created_at"`

	// Populated via JOIN
	RecordedByName string `db:"recorded_by_name"`
}

// PODebtSummary aggregates payment status across all termins for a PO.
type PODebtSummary struct {
	POID          string  `db:"po_id"`
	TotalAmount   float64 `db:"total_amount"`
	TotalTermin   float64 `db:"total_termin"` // sum of termin amounts
	TotalPaid     float64 `db:"total_paid"`
	RemainingDebt float64 `db:"remaining_debt"`
	Status        string  // unpaid | partial | paid
	TerminCount   int     `db:"termin_count"`
	OverdueCount  int     `db:"overdue_count"`
}
