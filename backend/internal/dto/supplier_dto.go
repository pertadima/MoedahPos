package dto

// ─── Suppliers ────────────────────────────────────────────────────────────────

// CreateSupplierRequest is the input for POST /suppliers.
type CreateSupplierRequest struct {
	Name        string `json:"name"         validate:"required,min=2,max=100"`
	ContactName string `json:"contact_name" validate:"max=100"`
	Phone       string `json:"phone"        validate:"max=20"`
	Email       string `json:"email"        validate:"omitempty,email,max=255"`
	Address     string `json:"address"      validate:"max=500"`
}

// UpdateSupplierRequest is the input for PUT /suppliers/:id.
type UpdateSupplierRequest struct {
	Name        string `json:"name"         validate:"required,min=2,max=100"`
	ContactName string `json:"contact_name" validate:"max=100"`
	Phone       string `json:"phone"        validate:"max=20"`
	Email       string `json:"email"        validate:"omitempty,email,max=255"`
	Address     string `json:"address"      validate:"max=500"`
	IsActive    *bool  `json:"is_active"`
}

// SupplierResponse is a supplier in responses.
type SupplierResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	ContactName string  `json:"contact_name"`
	Phone       string  `json:"phone"`
	Email       string  `json:"email"`
	Address     string  `json:"address"`
	IsActive    bool    `json:"is_active"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
	DeletedAt   *string `json:"deleted_at,omitempty"`
}

// SupplierListFilter holds list query params.
type SupplierListFilter struct {
	PaginationQuery
	Search   string
	IsActive *bool
}
