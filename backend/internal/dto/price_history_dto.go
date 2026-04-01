package dto

// PriceHistoryFilter is the query filter for price history list endpoints.
type PriceHistoryFilter struct {
	StoreID   string
	ProductID string
	Source    string // optional: "manual" | "purchase_order"
	PaginationQuery
}

// PriceHistoryRow is the response shape for one price-change audit entry.
type PriceHistoryRow struct {
	ID            string  `json:"id"`
	ProductID     string  `json:"product_id"`
	ProductName   string  `json:"product_name"`
	StoreID       string  `json:"store_id"`
	ChangedBy     string  `json:"changed_by"`
	ChangedByName string  `json:"changed_by_name"`
	OldCost       float64 `json:"old_cost"`
	NewCost       float64 `json:"new_cost"`
	OldSell       float64 `json:"old_sell"`
	NewSell       float64 `json:"new_sell"`
	Source        string  `json:"source"`
	RefID         *string `json:"ref_id,omitempty"`
	Notes         *string `json:"notes,omitempty"`
	ChangedAt     string  `json:"changed_at"`
}
