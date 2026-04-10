package dto

// ─── Income Categories ────────────────────────────────────────────────────────

type CreateIncomeCategoryRequest struct {
	Name        string `json:"name"        validate:"required,max=255"`
	Description string `json:"description"`
}

type IncomeCategoryResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	CreatedAt   string  `json:"created_at"`
}

// ─── Incomes ──────────────────────────────────────────────────────────────────

type CreateIncomeRequest struct {
	CategoryID    string  `json:"category_id"    validate:"required,uuid"`
	Amount        float64 `json:"amount"         validate:"required,gt=0"`
	IncomeDate    string  `json:"income_date"    validate:"required"`
	PaymentMethod string  `json:"payment_method" validate:"required,oneof=cash transfer qris other"`
	Reference     string  `json:"reference"`
	Notes         string  `json:"notes"`
}

type UpdateIncomeRequest struct {
	CategoryID    string  `json:"category_id"    validate:"required,uuid"`
	Amount        float64 `json:"amount"         validate:"required,gt=0"`
	IncomeDate    string  `json:"income_date"    validate:"required"`
	PaymentMethod string  `json:"payment_method" validate:"required,oneof=cash transfer qris other"`
	Reference     string  `json:"reference"`
	Notes         string  `json:"notes"`
}

type IncomeResponse struct {
	ID            string  `json:"id"`
	StoreID       string  `json:"store_id"`
	CategoryID    string  `json:"category_id"`
	CategoryName  string  `json:"category_name"`
	Amount        float64 `json:"amount"`
	IncomeDate    string  `json:"income_date"`
	PaymentMethod string  `json:"payment_method"`
	Reference     *string `json:"reference,omitempty"`
	Notes         *string `json:"notes,omitempty"`
	CreatedBy     *string `json:"created_by,omitempty"`
	CreatedByName *string `json:"created_by_name,omitempty"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

type IncomeListFilter struct {
	StoreID    string
	CategoryID string
	DateFrom   string
	DateTo     string
	PaginationQuery
}
