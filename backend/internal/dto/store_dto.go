package dto

// ─── Stores ───────────────────────────────────────────────────────────────────

// CreateStoreRequest is the input for POST /stores.
type CreateStoreRequest struct {
	Name      string `json:"name"       validate:"required,min=2,max=100"`
	Address   string `json:"address"    validate:"max=500"`
	Phone     string `json:"phone"      validate:"max=20"`
	TaxNumber string `json:"tax_number" validate:"max=50"`
	Currency  string `json:"currency"   validate:"len=3"`
	StoreType string `json:"store_type" validate:"oneof=retail restaurant"`
}

// UpdateStoreRequest is the input for PUT /stores/:id.
type UpdateStoreRequest struct {
	Name      string `json:"name"       validate:"required,min=2,max=100"`
	Address   string `json:"address"    validate:"max=500"`
	Phone     string `json:"phone"      validate:"max=20"`
	TaxNumber string `json:"tax_number" validate:"max=50"`
	Currency  string `json:"currency"   validate:"len=3"`
	StoreType string `json:"store_type" validate:"oneof=retail restaurant"`
	IsActive  *bool  `json:"is_active"`
}

// StoreResponse is the shape returned for a single store.
type StoreResponse struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Address   string  `json:"address"`
	Phone     string  `json:"phone"`
	TaxNumber string  `json:"tax_number"`
	Currency  string  `json:"currency"`
	StoreType string  `json:"store_type"`
	IsActive  bool    `json:"is_active"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
	DeletedAt *string `json:"deleted_at,omitempty"`
}

// StoreListFilter holds query params for the store list endpoint.
type StoreListFilter struct {
	PaginationQuery
	Search   string
	IsActive *bool // nil = all, true/false = filter
}

// ─── Store Members ────────────────────────────────────────────────────────────

// AddMemberRequest is the input for POST /stores/:id/members.
type AddMemberRequest struct {
	UserID string `json:"user_id" validate:"required,uuid"`
	RoleID string `json:"role_id" validate:"required,uuid"`
}

// UpdateMemberRoleRequest is the input for PUT /stores/:id/members/:userId.
type UpdateMemberRoleRequest struct {
	RoleID string `json:"role_id" validate:"required,uuid"`
}

// MemberResponse is a store member entry in list responses.
type MemberResponse struct {
	UserID    string `json:"user_id"`
	UserName  string `json:"user_name"`
	UserEmail string `json:"user_email"`
	RoleID    string `json:"role_id"`
	RoleName  string `json:"role_name"`
	IsActive  bool   `json:"is_active"`
	JoinedAt  string `json:"joined_at"`
}
