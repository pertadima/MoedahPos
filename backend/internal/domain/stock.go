package domain

import "time"

// StockLevel tracks current quantity for a product at a specific store.
type StockLevel struct {
	ID          string    `db:"id"`
	ProductID   string    `db:"product_id"`
	StoreID     string    `db:"store_id"`
	Quantity    float64   `db:"quantity"`
	MinQuantity float64   `db:"min_quantity"`
	UpdatedAt   time.Time `db:"updated_at"`

	// Populated via JOIN
	ProductName string `db:"product_name"`
	ProductSKU  string `db:"product_sku"`
	Unit        string `db:"unit"`
}

// StockMovement is an immutable audit-trail entry for every stock change.
type StockMovement struct {
	ID            string    `db:"id"`
	ProductID     string    `db:"product_id"`
	StoreID       string    `db:"store_id"`
	RefType       string    `db:"ref_type"` // sale | purchase | adjustment | transfer
	RefID         *string   `db:"ref_id"`
	QuantityDelta float64   `db:"quantity_delta"`
	Notes         string    `db:"notes"`
	CreatedBy     string    `db:"created_by"`
	CreatedAt     time.Time `db:"created_at"`

	// Populated via JOIN
	ProductName   string `db:"product_name"`
	CreatedByName string `db:"created_by_name"`
}

// AdjustInput encapsulates a stock adjustment request for the repository.
type AdjustInput struct {
	ProductID string
	StoreID   string
	Delta     float64
	RefType   string
	RefID     *string
	Notes     string
	CreatedBy string
}

// ─── FIFO Batch Inventory ──────────────────────────────────────────────────────

// StockBatch represents a single inventory batch created when a PO is received.
// Batches are deducted in FIFO order (oldest received_at first) when a sale occurs.
type StockBatch struct {
	ID                string    `db:"id"`
	ProductID         string    `db:"product_id"`
	StoreID           string    `db:"store_id"`
	POID              *string   `db:"po_id"`
	QuantityRemaining float64   `db:"quantity_remaining"`
	PurchasePrice     float64   `db:"purchase_price"`
	ReceivedAt        time.Time `db:"received_at"`
	CreatedAt         time.Time `db:"created_at"`

	// Populated via JOIN with products table.
	ProductName string `db:"product_name"`
	ProductSKU  string `db:"product_sku"`
	Unit        string `db:"unit"`
}

// BatchStockSummary holds aggregated batch stock across all batches for one product.
type BatchStockSummary struct {
	ProductID    string  `db:"product_id"`
	ProductName  string  `db:"product_name"`
	ProductSKU   string  `db:"product_sku"`
	Unit         string  `db:"unit"`
	TotalQty     float64 `db:"total_qty"`
	BatchCount   int     `db:"batch_count"`
	AvgCostPrice float64 `db:"avg_cost_price"`
}
