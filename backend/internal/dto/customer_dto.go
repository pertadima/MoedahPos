package dto

// CreateCustomerRequest is the payload for creating a customer.
type CreateCustomerRequest struct {
	Name    string `json:"name"    validate:"required,max=150"`
	Phone   string `json:"phone"   validate:"max=30"`
	Email   string `json:"email"   validate:"max=150"`
	Address string `json:"address"`
	Notes   string `json:"notes"`
}

// UpdateCustomerRequest is the payload for updating a customer.
type UpdateCustomerRequest struct {
	Name    string `json:"name"    validate:"required,max=150"`
	Phone   string `json:"phone"   validate:"max=30"`
	Email   string `json:"email"   validate:"max=150"`
	Address string `json:"address"`
	Notes   string `json:"notes"`
}

// CustomerResponse is the API response shape for a customer.
type CustomerResponse struct {
	ID        string  `json:"id"`
	StoreID   string  `json:"store_id"`
	Name      string  `json:"name"`
	Phone     *string `json:"phone,omitempty"`
	Email     *string `json:"email,omitempty"`
	Address   *string `json:"address,omitempty"`
	Notes     *string `json:"notes,omitempty"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// CustomerListFilter holds query params for listing customers.
type CustomerListFilter struct {
	PaginationQuery
	StoreID string
	Search  string
}
