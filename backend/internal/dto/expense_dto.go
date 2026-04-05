package dto

type CreateExpenseRequest struct {
	CategoryID  string  `json:"category_id" validate:"required,uuid"`
	Amount      float64 `json:"amount" validate:"required,gt=0"`
	ExpenseDate   string  `json:"expense_date" validate:"required"` // YYYY-MM-DD
	Notes         string  `json:"notes"`
	PaymentStatus string  `json:"payment_status"` // paid, unpaid (defaults to paid if empty)
}

type UpdateExpenseRequest struct {
	CategoryID  string  `json:"category_id" validate:"required,uuid"`
	Amount      float64 `json:"amount" validate:"required,gt=0"`
	ExpenseDate   string  `json:"expense_date" validate:"required"` // YYYY-MM-DD
	Notes         string  `json:"notes"`
	PaymentStatus string  `json:"payment_status"`
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
	Notes           string  `json:"notes"`
	PaymentStatus   string  `json:"payment_status"`
	CreatedBy       *string `json:"created_by"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

type ExpenseCategoryResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
}

type UpdateExpenseStatusRequest struct {
	PaymentStatus string `json:"payment_status" validate:"required,oneof=paid unpaid cancelled"`
}

type CreateRecurringExpenseRequest struct {
	CategoryID    string  `json:"category_id" validate:"required,uuid"`
	Name          string  `json:"name" validate:"required"`
	Amount        float64 `json:"amount" validate:"required,gt=0"`
	Interval      string  `json:"interval" validate:"required,oneof=daily weekly monthly yearly"`
	IntervalValue int     `json:"interval_value" validate:"required,gt=0"`
	StartDate     string  `json:"start_date" validate:"required"` // YYYY-MM-DD
	EndDate       *string `json:"end_date"`                       // YYYY-MM-DD
	Notes         string  `json:"notes"`
}

type UpdateRecurringExpenseRequest struct {
	CategoryID    string  `json:"category_id" validate:"required,uuid"`
	Name          string  `json:"name" validate:"required"`
	Amount        float64 `json:"amount" validate:"required,gt=0"`
	Interval      string  `json:"interval" validate:"required,oneof=daily weekly monthly yearly"`
	IntervalValue int     `json:"interval_value" validate:"required,gt=0"`
	StartDate     string  `json:"start_date" validate:"required"` // YYYY-MM-DD
	EndDate       *string `json:"end_date"`                       // YYYY-MM-DD
	IsActive      bool    `json:"is_active"`
	Notes         string  `json:"notes"`
}

type RecurringExpenseResponse struct {
	ID              string  `json:"id"`
	StoreID         string  `json:"store_id"`
	CategoryID      string  `json:"category_id"`
	CategoryName    string  `json:"category_name"`
	Name            string  `json:"name"`
	Amount          float64 `json:"amount"`
	Interval        string  `json:"interval"`
	IntervalValue   int     `json:"interval_value"`
	StartDate       string  `json:"start_date"`
	EndDate         *string `json:"end_date"`
	NextRunDate     string  `json:"next_run_date"`
	Notes           string  `json:"notes"`
	IsActive        bool    `json:"is_active"`
	CreatedBy       *string `json:"created_by"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
	LastGeneratedAt *string `json:"last_generated_at"`
}
