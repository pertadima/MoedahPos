package domain

import "time"

// Product is a sellable item belonging to a store.
type Product struct {
	ID            string     `db:"id"`
	StoreID       string     `db:"store_id"`
	CategoryID    *string    `db:"category_id"`
	SKU           string     `db:"sku"`
	Name          string     `db:"name"`
	Description   *string    `db:"description"`
	Barcode       *string    `db:"barcode"`
	Unit          string     `db:"unit"`
	CostPrice     float64    `db:"cost_price"`
	SellPrice     float64    `db:"sell_price"`
	UseGlobalTax  bool       `db:"use_global_tax"`
	TaxPercentage *float64   `db:"tax_percentage"`
	ImageURL      *string    `db:"image_url"`
	IsActive      bool       `db:"is_active"`
	CreatedAt     time.Time  `db:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at"`
	DeletedAt     *time.Time `db:"deleted_at"`
	ServerUpdatedAt time.Time  `db:"server_updated_at"`
	SyncVersion     int        `db:"sync_version"`

	// Populated via JOIN
	CategoryName    *string  `db:"category_name"`
	StockQty        *float64 `db:"stock_qty"` // from stock_levels join
	StoreDefaultTax float64  `db:"store_default_tax"`
}

func (p *Product) EffectiveTaxRate() float64 {
	if p.UseGlobalTax {
		return p.StoreDefaultTax
	}
	if p.TaxPercentage != nil {
		return *p.TaxPercentage
	}
	return 0
}

// Category groups products within a store (supports nested parent/child).
type Category struct {
	ID        string     `db:"id"`
	StoreID   string     `db:"store_id"`
	Name      string     `db:"name"`
	ParentID  *string    `db:"parent_id"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
	ServerUpdatedAt time.Time  `db:"server_updated_at"`
	SyncVersion     int        `db:"sync_version"`

	// Populated via JOIN
	ParentName *string `db:"parent_name"`
}

// PriceHistory is an audit log entry for cost / sell price changes on a product.
type PriceHistory struct {
	ID            string    `db:"id"`
	ProductID     string    `db:"product_id"`
	ProductName   string    `db:"product_name"` // denormalised via JOIN
	StoreID       string    `db:"store_id"`
	ChangedBy     string    `db:"changed_by"`
	ChangedByName string    `db:"changed_by_name"` // JOIN from users
	OldCost       float64   `db:"old_cost"`
	NewCost       float64   `db:"new_cost"`
	OldSell       float64   `db:"old_sell"`
	NewSell       float64   `db:"new_sell"`
	Source        string    `db:"source"` // manual | purchase_order
	RefID         *string   `db:"ref_id"`
	Notes         *string   `db:"notes"`
	ChangedAt     time.Time `db:"changed_at"`
}
