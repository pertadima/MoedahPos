package dto

// ─── Stock Adjustment ─────────────────────────────────────────────────────────

// AdjustStockRequest is the input for POST /stores/:storeId/stock/adjust.
type AdjustStockRequest struct {
	ProductID string  `json:"product_id" validate:"required,uuid"`
	Delta     float64 `json:"delta"      validate:"required"` // positive = in, negative = out
	Notes     string  `json:"notes"      validate:"max=500"`
}

// ─── Stock Level Response ─────────────────────────────────────────────────────

// StockLevelResponse represents a single product's stock at a store.
type StockLevelResponse struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	ProductSKU  string  `json:"product_sku"`
	Unit        string  `json:"unit"`
	StoreID     string  `json:"store_id"`
	Quantity    float64 `json:"quantity"`
	MinQuantity float64 `json:"min_quantity"`
	IsLowStock  bool    `json:"is_low_stock"` // quantity <= min_quantity
	UpdatedAt   string  `json:"updated_at"`
}

// SetMinStockRequest updates the minimum quantity threshold for a product.
type SetMinStockRequest struct {
	ProductID   string  `json:"product_id"   validate:"required,uuid"`
	MinQuantity float64 `json:"min_quantity" validate:"min=0"`
}

// ─── Stock Movement Response ──────────────────────────────────────────────────

// StockMovementResponse represents one movement record.
type StockMovementResponse struct {
	ID            string  `json:"id"`
	ProductID     string  `json:"product_id"`
	ProductName   string  `json:"product_name"`
	StoreID       string  `json:"store_id"`
	RefType       string  `json:"ref_type"`
	RefID         *string `json:"ref_id,omitempty"`
	QuantityDelta float64 `json:"quantity_delta"`
	Notes         string  `json:"notes"`
	CreatedBy     string  `json:"created_by"`
	CreatedByName string  `json:"created_by_name"`
	CreatedAt     string  `json:"created_at"`
}

// StockMovementFilter holds query params for the movement history endpoint.
type StockMovementFilter struct {
	PaginationQuery
	StoreID   string
	ProductID string
	RefType   string
}
