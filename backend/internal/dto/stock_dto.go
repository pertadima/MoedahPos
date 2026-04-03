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

// ─── FIFO Batch DTOs ──────────────────────────────────────────────────────────

// StockBatchResponse represents one inventory batch record.
type StockBatchResponse struct {
	ID                string  `json:"id"`
	ProductID         string  `json:"product_id"`
	ProductName       string  `json:"product_name"`
	ProductSKU        string  `json:"product_sku"`
	Unit              string  `json:"unit"`
	StoreID           string  `json:"store_id"`
	POID              *string `json:"po_id,omitempty"`
	QuantityRemaining float64 `json:"quantity_remaining"`
	PurchasePrice     float64 `json:"purchase_price"`
	ReceivedAt        string  `json:"received_at"`
	CreatedAt         string  `json:"created_at"`
}

// BatchStockSummaryResponse aggregates batch stock totals per product.
type BatchStockSummaryResponse struct {
	ProductID    string  `json:"product_id"`
	ProductName  string  `json:"product_name"`
	ProductSKU   string  `json:"product_sku"`
	Unit         string  `json:"unit"`
	TotalQty     float64 `json:"total_qty"`
	BatchCount   int     `json:"batch_count"`
	AvgCostPrice float64 `json:"avg_cost_price"`
}

// BatchListFilter holds query parameters for the batch listing endpoint.
type BatchListFilter struct {
	StoreID   string
	ProductID string // optional; empty means all products
}

// POBatchItem carries the per-item data needed to create a stock batch
// when a purchase order is received.
type POBatchItem struct {
	ProductID string
	Quantity  float64
	UnitCost  float64
}
