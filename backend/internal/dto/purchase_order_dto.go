package dto

// ─── Purchase Orders ──────────────────────────────────────────────────────────

// CreatePORequest is the input for POST /stores/:storeId/purchase-orders.
type CreatePORequest struct {
	SupplierID *string       `json:"supplier_id" validate:"omitempty,uuid"`
	Notes      string        `json:"notes"       validate:"max=500"`
	Items      []POItemInput `json:"items"       validate:"required,min=1,dive"`
}

// UpdatePORequest replaces all items of a draft PO.
type UpdatePORequest struct {
	SupplierID *string       `json:"supplier_id" validate:"omitempty,uuid"`
	Notes      string        `json:"notes"       validate:"max=500"`
	Items      []POItemInput `json:"items"       validate:"required,min=1,dive"`
}

// POItemInput is a single line in a PO.
type POItemInput struct {
	ProductID string  `json:"product_id" validate:"required,uuid"`
	Quantity  float64 `json:"quantity"   validate:"required,gt=0"`
	UnitCost  float64 `json:"unit_cost"  validate:"min=0"`
}

// ─── Responses ────────────────────────────────────────────────────────────────

// POItemResponse is a single line in a PO response.
type POItemResponse struct {
	ID          string  `json:"id"`
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	ProductSKU  string  `json:"product_sku"`
	Unit        string  `json:"unit"`
	Quantity    float64 `json:"quantity"`
	UnitCost    float64 `json:"unit_cost"`
	ReceivedQty float64 `json:"received_qty"`
	Subtotal    float64 `json:"subtotal"`
}

// POResponse is the full purchase order response.
type POResponse struct {
	ID             string           `json:"id"`
	StoreID        string           `json:"store_id"`
	SupplierID     *string          `json:"supplier_id,omitempty"`
	SupplierName   *string          `json:"supplier_name,omitempty"`
	PONumber       string           `json:"po_number"`
	Status         string           `json:"status"`
	TotalItems     int              `json:"total_items"`
	TotalAmount    float64          `json:"total_amount"`
	AmountPaid     float64          `json:"amount_paid"`
	AmountDue      float64          `json:"amount_due"`
	PaymentStatus  string           `json:"payment_status"` // unpaid | partial | paid
	OrderedByName  string           `json:"ordered_by_name"`
	ReceivedByName *string          `json:"received_by_name,omitempty"`
	OrderedAt      *string          `json:"ordered_at,omitempty"`
	ReceivedAt     *string          `json:"received_at,omitempty"`
	Notes          string           `json:"notes,omitempty"`
	Items          []POItemResponse `json:"items"`
	CreatedAt      string           `json:"created_at"`
	UpdatedAt      string           `json:"updated_at"`
}

// POListFilter holds list query params.
type POListFilter struct {
	PaginationQuery
	StoreID  string
	Status   string
	DateFrom string
	DateTo   string
}

// POPaymentRequest is the input for recording a payment.
type POPaymentRequest struct {
	Amount float64 `json:"amount" validate:"required,gt=0"`
	Note   string  `json:"note"`
}

// POPaymentResponse is one payment record.
type POPaymentResponse struct {
	ID         string  `json:"id"`
	POID       string  `json:"po_id"`
	Amount     float64 `json:"amount"`
	Note       *string `json:"note,omitempty"`
	PaidBy     string  `json:"paid_by"`
	PaidByName string  `json:"paid_by_name"`
	PaidAt     string  `json:"paid_at"`
}

// PayableSummary aggregates outstanding debt per store.
type PayableSummary struct {
	TotalDebt        float64 `json:"total_debt"`
	TotalPaid        float64 `json:"total_paid"`
	TotalOutstanding float64 `json:"total_outstanding"`
	UnpaidCount      int     `json:"unpaid_count"`
	PartialCount     int     `json:"partial_count"`

	// Termin-based debt aging
	OverdueDebt float64 `json:"overdue_debt"`
	DueSoonDebt float64 `json:"due_soon_debt"`
	FutureDebt  float64 `json:"future_debt"`
}

// ─── Termin (Installment) DTOs ────────────────────────────────────────────────

// TerminInput is one installment in a termin schedule creation request.
type TerminInput struct {
	TerminNumber int     `json:"termin_number" validate:"required,min=1"`
	Amount       float64 `json:"amount"        validate:"required,gt=0"`
	DueDate      string  `json:"due_date"      validate:"required"` // YYYY-MM-DD
	Notes        string  `json:"notes"         validate:"max=500"`
}

// CreateTerminScheduleRequest creates or replaces the termin schedule for a PO.
type CreateTerminScheduleRequest struct {
	Termins []TerminInput `json:"termins" validate:"required,min=1,dive"`
}

// PaymentRecordResponse represents one recorded payment.
type PaymentRecordResponse struct {
	ID             string  `json:"id"`
	TerminID       string  `json:"termin_id"`
	AmountPaid     float64 `json:"amount_paid"`
	PaymentDate    string  `json:"payment_date"`
	PaymentMethod  string  `json:"payment_method"`
	Notes          string  `json:"notes,omitempty"`
	RecordedByName string  `json:"recorded_by_name,omitempty"`
	CreatedAt      string  `json:"created_at"`
}

// TerminResponse represents one termin with its payment aggregation.
type TerminResponse struct {
	ID           string                  `json:"id"`
	POID         string                  `json:"po_id"`
	TerminNumber int                     `json:"termin_number"`
	Amount       float64                 `json:"amount"`
	DueDate      string                  `json:"due_date"`
	Status       string                  `json:"status"` // unpaid | partial | paid | overdue
	Notes        string                  `json:"notes,omitempty"`
	AmountPaid   float64                 `json:"amount_paid"`
	AmountDue    float64                 `json:"amount_due"`
	IsOverdue    bool                    `json:"is_overdue"`
	Payments     []PaymentRecordResponse `json:"payments"`
	CreatedAt    string                  `json:"created_at"`
}

// RecordPaymentRequest is the input for POST /{terminId}/payments.
type RecordPaymentRequest struct {
	AmountPaid    float64 `json:"amount_paid"    validate:"required,gt=0"`
	PaymentDate   string  `json:"payment_date"   validate:"required"` // YYYY-MM-DD
	PaymentMethod string  `json:"payment_method" validate:"required,oneof=cash transfer check other"`
	Notes         string  `json:"notes"          validate:"max=500"`
}

// PODebtSummaryResponse is the aggregated debt view for a single PO.
type PODebtSummaryResponse struct {
	POID          string  `json:"po_id"`
	PONumber      string  `json:"po_number"`
	TotalAmount   float64 `json:"total_amount"`
	TotalTermin   float64 `json:"total_termin"`
	TotalPaid     float64 `json:"total_paid"`
	RemainingDebt float64 `json:"remaining_debt"`
	Status        string  `json:"status"` // unpaid | partial | paid
	TerminCount   int     `json:"termin_count"`
	OverdueCount  int     `json:"overdue_count"`
}

// ─── Document Generation ──────────────────────────────────────────────────────

// PODocumentData is the full payload returned by GET /{poId}/document.
// The frontend renders this into a printable HTML page.
type PODocumentData struct {
	DocType      string                `json:"doc_type"` // invoice | receipt | termin_agreement
	GeneratedAt  string                `json:"generated_at"`
	PO           POResponse            `json:"po"`
	DebtSummary  PODebtSummaryResponse `json:"debt_summary"`
	Termins      []TerminResponse      `json:"termins"`
	SupplierName string                `json:"supplier_name"`
	StoreName    string                `json:"store_name,omitempty"`
}
