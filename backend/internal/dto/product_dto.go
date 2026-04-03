package dto

// ─── Categories ───────────────────────────────────────────────────────────────

// CreateCategoryRequest is the input for POST /stores/:storeId/categories.
type CreateCategoryRequest struct {
	Name     string  `json:"name"      validate:"required,min=1,max=100"`
	ParentID *string `json:"parent_id" validate:"omitempty,uuid"`
}

// UpdateCategoryRequest is the input for PUT /stores/:storeId/categories/:id.
type UpdateCategoryRequest struct {
	Name     string  `json:"name"      validate:"required,min=1,max=100"`
	ParentID *string `json:"parent_id" validate:"omitempty,uuid"`
}

// CategoryResponse is a single category in responses.
type CategoryResponse struct {
	ID         string  `json:"id"`
	StoreID    string  `json:"store_id"`
	Name       string  `json:"name"`
	ParentID   *string `json:"parent_id,omitempty"`
	ParentName *string `json:"parent_name,omitempty"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
	DeletedAt  *string `json:"deleted_at,omitempty"`
}

// ─── Products ─────────────────────────────────────────────────────────────────

// CreateProductRequest is the input for POST /stores/:storeId/products.
type CreateProductRequest struct {
	CategoryID  *string `json:"category_id"  validate:"omitempty,uuid"`
	SKU         string  `json:"sku"          validate:"required,min=1,max=100"`
	Name        string  `json:"name"         validate:"required,min=1,max=200"`
	Description string  `json:"description"`
	Barcode     *string `json:"barcode"      validate:"omitempty,max=100"`
	Unit        string  `json:"unit"         validate:"required,max=20"`
	CostPrice   float64 `json:"cost_price"   validate:"min=0"`
	SellPrice   float64 `json:"sell_price"   validate:"min=0"`
	TaxRate     float64 `json:"tax_rate"     validate:"min=0,max=100"`
	ImageURL    *string `json:"image_url"`
	InitialQty  float64 `json:"initial_qty"` // sets stock level on creation
}

// UpdateProductRequest is the input for PUT /stores/:storeId/products/:id.
type UpdateProductRequest struct {
	CategoryID  *string `json:"category_id"  validate:"omitempty,uuid"`
	Name        string  `json:"name"         validate:"required,min=1,max=200"`
	Description string  `json:"description"`
	Barcode     *string `json:"barcode"      validate:"omitempty,max=100"`
	Unit        string  `json:"unit"         validate:"required,max=20"`
	CostPrice   float64 `json:"cost_price"   validate:"min=0"`
	SellPrice   float64 `json:"sell_price"   validate:"min=0"`
	TaxRate     float64 `json:"tax_rate"     validate:"min=0,max=100"`
	ImageURL    *string `json:"image_url"`
	IsActive    *bool   `json:"is_active"`
}

// ProductResponse is a single product in responses.
type ProductResponse struct {
	ID           string   `json:"id"`
	StoreID      string   `json:"store_id"`
	CategoryID   *string  `json:"category_id,omitempty"`
	CategoryName *string  `json:"category_name,omitempty"`
	SKU          string   `json:"sku"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Barcode      *string  `json:"barcode,omitempty"`
	Unit         string   `json:"unit"`
	CostPrice    float64  `json:"cost_price"`
	SellPrice    float64  `json:"sell_price"`
	TaxRate      float64  `json:"tax_rate"`
	ImageURL     *string  `json:"image_url,omitempty"`
	IsActive     bool     `json:"is_active"`
	StockQty     *float64 `json:"stock_qty,omitempty"` // joined from stock_levels
	CreatedAt    string   `json:"created_at"`
	UpdatedAt    string   `json:"updated_at"`
	DeletedAt    *string  `json:"deleted_at,omitempty"`
}

// ProductListFilter holds query params for the product list endpoint.
type ProductListFilter struct {
	PaginationQuery
	StoreID    string
	Search     string // searches name and SKU
	CategoryID string
	IsActive   *bool // nil = active only by default
	WithStock  bool  // join stock_levels
}
