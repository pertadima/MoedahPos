package dto

type CreateExpenseRequest struct {
	CategoryID  string  `json:"category_id" validate:"required,uuid"`
	Amount      float64 `json:"amount" validate:"required,gt=0"`
	ExpenseDate string  `json:"expense_date" validate:"required"` // YYYY-MM-DD
	Notes       string  `json:"notes"`
}

type UpdateExpenseRequest struct {
	CategoryID  string  `json:"category_id" validate:"required,uuid"`
	Amount      float64 `json:"amount" validate:"required,gt=0"`
	ExpenseDate string  `json:"expense_date" validate:"required"` // YYYY-MM-DD
	Notes       string  `json:"notes"`
}

type CreateExpenseCategoryRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}

type ExpenseListFilter struct {
	PaginationQuery
	StoreID    string `json:"-"` // injected by middleware
	CategoryID string `query:"category_id"`
	DateFrom   string `query:"date_from"` // YYYY-MM-DD
	DateTo     string `query:"date_to"`   // YYYY-MM-DD
}

type ExpenseResponse struct {
	ID           string  `json:"id"`
	StoreID      string  `json:"store_id"`
	CategoryID   string  `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Amount       float64 `json:"amount"`
	ExpenseDate  string  `json:"expense_date"`
	Notes        string  `json:"notes"`
	CreatedBy    *string `json:"created_by"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

type ExpenseCategoryResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
}
