package dto

// ─── User Admin DTOs ──────────────────────────────────────────────────────────

type CreateUserRequest struct {
	Name     string           `json:"name"     validate:"required,min=2,max=100"`
	Email    string           `json:"email"    validate:"required,email"`
	Password string           `json:"password" validate:"required,min=6"`
	Stores   []StoreAssignDTO `json:"stores"`
}

type UpdateUserRequest struct {
	Name  string `json:"name"  validate:"required,min=2,max=100"`
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Password string `json:"password" validate:"required,min=6"`
}

type SetUserStoresRequest struct {
	Stores []StoreAssignDTO `json:"stores" validate:"required"`
}

type StoreAssignDTO struct {
	StoreID string `json:"store_id" validate:"required,uuid4"`
	RoleID  string `json:"role_id"  validate:"required,uuid4"`
}

// ─── User Admin Response ──────────────────────────────────────────────────────

type UserResponse struct {
	ID         string              `json:"id"`
	Name       string              `json:"name"`
	Email      string              `json:"email"`
	IsActive   bool                `json:"is_active"`
	CreatedAt  string              `json:"created_at"`
	UpdatedAt  string              `json:"updated_at"`
	DeletedAt  *string             `json:"deleted_at,omitempty"`
	StoreCount int                 `json:"store_count"`
	Stores     []UserStoreResponse `json:"stores,omitempty"`
}

type UserStoreResponse struct {
	StoreID   string `json:"store_id"`
	StoreName string `json:"store_name"`
	StoreType string `json:"store_type"`
	RoleID    string `json:"role_id"`
	RoleName  string `json:"role_name"`
	IsActive  bool   `json:"is_active"`
}

type UserListResponse struct {
	Data []UserResponse `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

// ─── Role Response ────────────────────────────────────────────────────────────

type RoleResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}
