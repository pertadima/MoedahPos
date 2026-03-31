package dto

// ─── Register ─────────────────────────────────────────────────────────────────

// RegisterRequest is the input for POST /auth/register.
type RegisterRequest struct {
	Name     string `json:"name"     validate:"required,min=2,max=100"`
	Email    string `json:"email"    validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=128"`
}

// RegisterResponse is returned on successful registration.
type RegisterResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

// ─── Login ────────────────────────────────────────────────────────────────────

// LoginRequest is the input for POST /auth/login.
type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse is returned on successful authentication.
type LoginResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	TokenType    string   `json:"token_type"`
	ExpiresIn    int64    `json:"expires_in"` // seconds
	User         UserInfo `json:"user"`
}

// UserInfo is a lightweight user representation embedded in responses.
type UserInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// ─── Refresh ──────────────────────────────────────────────────────────────────

// RefreshRequest is the input for POST /auth/refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshResponse is returned on a successful token refresh.
type RefreshResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

// ─── Me ───────────────────────────────────────────────────────────────────────

// MeResponse is returned for GET /auth/me.
type MeResponse struct {
	ID     string          `json:"id"`
	Name   string          `json:"name"`
	Email  string          `json:"email"`
	Stores []StoreRoleInfo `json:"stores"`
}

// StoreRoleInfo is a store membership entry in /me.
type StoreRoleInfo struct {
	StoreID   string `json:"store_id"`
	StoreName string `json:"store_name"`
	StoreType string `json:"store_type"`
	Role      string `json:"role"`
}

// ─── Validation Error ─────────────────────────────────────────────────────────

// FieldError represents a single field-level validation error.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
