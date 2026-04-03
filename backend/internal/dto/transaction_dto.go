package dto

// ─── Create Transaction (Retail Checkout) ──────────────────────────────────────

// CreateTransactionRequest is the input for POST /stores/:storeId/transactions.
type CreateTransactionRequest struct {
	CustomerName  string        `json:"customer_name"  validate:"max=100"`
	CustomerPhone string        `json:"customer_phone" validate:"max=20"`
	PaymentMethod string        `json:"payment_method" validate:"required,oneof=cash card qris transfer"`
	PaymentAmount float64       `json:"payment_amount" validate:"required,min=0"`
	Notes         string        `json:"notes"          validate:"max=500"`
	Items         []TxItemInput `json:"items"          validate:"required,min=1,dive"`
}

// TxItemInput is a single line in a sale request.
// For retail: set product_id. For restaurant menus: set menu_item_id.
type TxItemInput struct {
	ProductID   string  `json:"product_id"   validate:"omitempty,uuid"`
	MenuItemID  string  `json:"menu_item_id" validate:"omitempty,uuid"`
	Quantity    float64 `json:"quantity"     validate:"required,gt=0"`
	DiscountPct float64 `json:"discount_pct" validate:"min=0,max=100"`
}

// ─── Draft Order (Restaurant Table Orders) ────────────────────────────────────

// CreateDraftRequest holds an order for a table without payment yet.
type CreateDraftRequest struct {
	TableID      string        `json:"table_id"       validate:"required,uuid"`
	CustomerName string        `json:"customer_name"  validate:"max=100"`
	Notes        string        `json:"notes"          validate:"max=500"`
	Items        []TxItemInput `json:"items"          validate:"required,min=1,dive"`
}

// UpdateDraftRequest replaces the items of an existing draft (idempotent).
type UpdateDraftRequest struct {
	CustomerName string        `json:"customer_name" validate:"max=100"`
	Notes        string        `json:"notes"         validate:"max=500"`
	Items        []TxItemInput `json:"items"         validate:"required,min=1,dive"`
}

// PayDraftRequest finalizes a held order with payment details.
type PayDraftRequest struct {
	PaymentMethod string  `json:"payment_method" validate:"required,oneof=cash card qris transfer"`
	PaymentAmount float64 `json:"payment_amount" validate:"required,min=0"`
	CustomerName  string  `json:"customer_name"  validate:"max=100"`
	CustomerPhone string  `json:"customer_phone" validate:"max=20"`
}

// ─── Responses ────────────────────────────────────────────────────────────────

// TransactionItemResponse is a single line item in a receipt.
type TransactionItemResponse struct {
	ID          string  `json:"id"`
	ProductID   *string `json:"product_id,omitempty"`
	ProductName string  `json:"product_name"`
	SKU         string  `json:"sku"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	DiscountPct float64 `json:"discount_pct"`
	TaxRate     float64 `json:"tax_rate"`
	Subtotal    float64 `json:"subtotal"`
}

// TransactionResponse is the full receipt — returned on create and get.
type TransactionResponse struct {
	ID            string                    `json:"id"`
	StoreID       string                    `json:"store_id"`
	CashierID     string                    `json:"cashier_id"`
	CashierName   string                    `json:"cashier_name"`
	TableID       *string                   `json:"table_id,omitempty"`
	CustomerName  string                    `json:"customer_name,omitempty"`
	CustomerPhone string                    `json:"customer_phone,omitempty"`
	Subtotal      float64                   `json:"subtotal"`
	DiscountAmt   float64                   `json:"discount_amt"`
	TaxAmt        float64                   `json:"tax_amt"`
	Total         float64                   `json:"total"`
	PaymentMethod string                    `json:"payment_method"`
	PaymentAmount float64                   `json:"payment_amount"`
	ChangeAmount  float64                   `json:"change_amount"`
	Status        string                    `json:"status"`
	Notes         string                    `json:"notes,omitempty"`
	Items         []TransactionItemResponse `json:"items"`
	CreatedAt     string                    `json:"created_at"`
	UpdatedAt     string                    `json:"updated_at"`
}

// TransactionListFilter holds query params for list endpoints.
type TransactionListFilter struct {
	PaginationQuery
	StoreID   string
	TableID   string
	Status    string
	CashierID string
	DateFrom  string
	DateTo    string
}
