package domain

import "time"

// StockAdjustment tracking for discrepancies like lost or damaged goods.
type StockAdjustment struct {
	ID        string    `json:"id" db:"id"`
	ProductID string    `json:"product_id" db:"product_id"`
	StoreID   string    `json:"store_id" db:"store_id"`
	Type      string    `json:"type" db:"type"`     // IN, OUT
	Reason    string    `json:"reason" db:"reason"` // DAMAGED, LOST, MANUAL_CORRECTION
	Quantity  float64   `json:"quantity" db:"quantity"`
	Notes     string    `json:"notes" db:"notes"`
	CreatedBy string    `json:"created_by" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// Populated via JOINs
	ProductName   string `json:"product_name" db:"product_name"`
	ProductSKU    string `json:"product_sku" db:"product_sku"`
	Unit          string `json:"unit" db:"unit"`
	CreatedByName string `json:"created_by_name" db:"created_by_name"`
}

type StockAdjustmentBatch struct {
	ID           string  `json:"id" db:"id"`
	AdjustmentID string  `json:"adjustment_id" db:"adjustment_id"`
	BatchID      string  `json:"batch_id" db:"batch_id"`
	DeductedQty  float64 `json:"deducted_qty" db:"deducted_qty"`
}

// AdjustInput encapsulates a stock adjustment request from the service layer.
type CreateAdjustmentInput struct {
	ProductID string  `json:"product_id" validate:"required"`
	Type      string  `json:"type" validate:"required,oneof=IN OUT"`
	Reason    string  `json:"reason" validate:"required,oneof=DAMAGED LOST MANUAL_CORRECTION"`
	Quantity  float64 `json:"quantity" validate:"required,gt=0"`
	Notes     string  `json:"notes"`
}
