package dto

// ─── Tables ───────────────────────────────────────────────────────────────────

type CreateTableRequest struct {
	TableNumber string  `json:"table_number" validate:"required,max=20"`
	Capacity    int     `json:"capacity"     validate:"required,min=1,max=100"`
	Notes       *string `json:"notes"`
}

type UpdateTableRequest struct {
	TableNumber string  `json:"table_number" validate:"required,max=20"`
	Capacity    int     `json:"capacity"     validate:"required,min=1,max=100"`
	Notes       *string `json:"notes"`
	IsActive    *bool   `json:"is_active"`
}

type UpdateTableStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=available occupied reserved"`
}

type TableResponse struct {
	ID          string `json:"id"`
	StoreID     string `json:"store_id"`
	TableNumber string `json:"table_number"`
	Capacity    int    `json:"capacity"`
	Status      string `json:"status"`
	Notes       string `json:"notes,omitempty"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// ─── Menu Items ───────────────────────────────────────────────────────────────

type IngredientInput struct {
	ProductID string  `json:"product_id" validate:"required,uuid"`
	Quantity  float64 `json:"quantity"   validate:"required,gt=0"`
}

type IngredientResponse struct {
	ID          string  `json:"id"`
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	ProductSKU  string  `json:"product_sku"`
	Unit        string  `json:"unit"`
	Quantity    float64 `json:"quantity"`
	CostPrice   float64 `json:"cost_price"`
}

type CreateMenuItemRequest struct {
	CategoryID  *string           `json:"category_id"`
	Name        string            `json:"name"        validate:"required,max=200"`
	Description string            `json:"description"`
	SellPrice   float64           `json:"sell_price"  validate:"min=0"`
	TaxRate     float64           `json:"tax_rate"    validate:"min=0,max=100"`
	Ingredients []IngredientInput `json:"ingredients"`
}

type UpdateMenuItemRequest struct {
	CategoryID  *string           `json:"category_id"`
	Name        string            `json:"name"        validate:"required,max=200"`
	Description string            `json:"description"`
	SellPrice   float64           `json:"sell_price"  validate:"min=0"`
	TaxRate     float64           `json:"tax_rate"    validate:"min=0,max=100"`
	IsActive    *bool             `json:"is_active"`
	Ingredients []IngredientInput `json:"ingredients"`
}

type MenuItemResponse struct {
	ID           string               `json:"id"`
	StoreID      string               `json:"store_id"`
	CategoryID   *string              `json:"category_id,omitempty"`
	CategoryName *string              `json:"category_name,omitempty"`
	Name         string               `json:"name"`
	Description  string               `json:"description"`
	SellPrice    float64              `json:"sell_price"`
	CostPrice    float64              `json:"cost_price"`
	TaxRate      float64              `json:"tax_rate"`
	ImageURL     string               `json:"image_url,omitempty"`
	IsActive     bool                 `json:"is_active"`
	Ingredients  []IngredientResponse `json:"ingredients"`
	CreatedAt    string               `json:"created_at"`
	UpdatedAt    string               `json:"updated_at"`
}
