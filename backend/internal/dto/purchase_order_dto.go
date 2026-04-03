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
}
