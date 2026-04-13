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
	CategoryID    *string           `json:"category_id"`
	Name          string            `json:"name"        validate:"required,max=200"`
	Description   string            `json:"description"`
	SellPrice     float64           `json:"sell_price"  validate:"min=0"`
	UseGlobalTax  *bool             `json:"use_global_tax"`
	TaxPercentage *float64          `json:"tax_percentage" validate:"omitempty,min=0,max=100"`
	PackagingCost float64           `json:"packaging_cost" validate:"min=0"`
	OverheadCost  float64           `json:"overhead_cost"  validate:"min=0"`
	LaborCost     float64           `json:"labor_cost"     validate:"min=0"`
	Ingredients   []IngredientInput `json:"ingredients"`
}

type UpdateMenuItemRequest struct {
	CategoryID    *string           `json:"category_id"`
	Name          string            `json:"name"        validate:"required,max=200"`
	Description   string            `json:"description"`
	SellPrice     float64           `json:"sell_price"  validate:"min=0"`
	UseGlobalTax  *bool             `json:"use_global_tax"`
	TaxPercentage *float64          `json:"tax_percentage" validate:"omitempty,min=0,max=100"`
	PackagingCost float64           `json:"packaging_cost" validate:"min=0"`
	OverheadCost  float64           `json:"overhead_cost"  validate:"min=0"`
	LaborCost     float64           `json:"labor_cost"     validate:"min=0"`
	IsActive      *bool             `json:"is_active"`
	Ingredients   []IngredientInput `json:"ingredients"`
}

type MenuItemResponse struct {
	ID             string               `json:"id"`
	StoreID        string               `json:"store_id"`
	CategoryID     *string              `json:"category_id,omitempty"`
	CategoryName   *string              `json:"category_name,omitempty"`
	Name           string               `json:"name"`
	Description    string               `json:"description"`
	SellPrice      float64              `json:"sell_price"`
	CostPrice      float64              `json:"cost_price"`
	IngredientCost float64              `json:"ingredient_cost"`
	PackagingCost  float64              `json:"packaging_cost"`
	OverheadCost   float64              `json:"overhead_cost"`
	LaborCost      float64              `json:"labor_cost"`
	UseGlobalTax   bool                 `json:"use_global_tax"`
	TaxPercentage  *float64             `json:"tax_percentage"`
	TaxRate        float64              `json:"tax_rate"` // computed effective rate for POS
	ImageURL       string               `json:"image_url,omitempty"`
	IsActive       bool                 `json:"is_active"`
	Ingredients    []IngredientResponse `json:"ingredients"`
	CreatedAt      string               `json:"created_at"`
	UpdatedAt      string               `json:"updated_at"`
}
