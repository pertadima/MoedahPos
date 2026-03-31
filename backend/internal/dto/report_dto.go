package dto

// ─── Report Filter ────────────────────────────────────────────────────────────

// ReportFilter is the common query param set for all report endpoints.
// DateFrom and DateTo are YYYY-MM-DD strings (parsed in service layer).
type ReportFilter struct {
	StoreID  string
	DateFrom string
	DateTo   string
}

// ─── Sales Summary ────────────────────────────────────────────────────────────

// SalesSummaryRow is one day's aggregated sales totals (with profit).
type SalesSummaryRow struct {
	Date             string  `json:"date"              db:"date"`
	TransactionCount int     `json:"transaction_count" db:"transaction_count"`
	TotalSales       float64 `json:"total_sales"       db:"total_sales"`
	TotalTax         float64 `json:"total_tax"         db:"total_tax"`
	TotalDiscount    float64 `json:"total_discount"    db:"total_discount"`
	TotalNet         float64 `json:"total_net"         db:"total_net"` // sales - discount
	TotalCost        float64 `json:"total_cost"        db:"total_cost"`
	GrossProfit      float64 `json:"gross_profit"      db:"gross_profit"`
}

// SalesSummaryResponse wraps rows + period totals.
type SalesSummaryResponse struct {
	Rows              []SalesSummaryRow `json:"rows"`
	TotalSales        float64           `json:"total_sales"`
	TotalTransactions int               `json:"total_transactions"`
	TotalCost         float64           `json:"total_cost"`
	GrossProfit       float64           `json:"gross_profit"`
	ProfitMargin      float64           `json:"profit_margin"` // %
}

// ─── Sales by Product ─────────────────────────────────────────────────────────

// SalesByProductRow is aggregated sales for one product.
type SalesByProductRow struct {
	ProductID     string  `json:"product_id"     db:"product_id"`
	ProductName   string  `json:"product_name"   db:"product_name"`
	SKU           string  `json:"sku"            db:"sku"`
	TotalQuantity float64 `json:"total_quantity" db:"total_quantity"`
	TotalRevenue  float64 `json:"total_revenue"  db:"total_revenue"`
	TotalCost     float64 `json:"total_cost"     db:"total_cost"`
	GrossProfit   float64 `json:"gross_profit"   db:"gross_profit"`
	ProfitMargin  float64 `json:"profit_margin"  db:"profit_margin"` // %
	TotalTax      float64 `json:"total_tax"      db:"total_tax"`
}

// ─── Sales by Cashier ─────────────────────────────────────────────────────────

// SalesByCashierRow is aggregated sales for one cashier.
type SalesByCashierRow struct {
	CashierID        string  `json:"cashier_id"        db:"cashier_id"`
	CashierName      string  `json:"cashier_name"      db:"cashier_name"`
	TransactionCount int     `json:"transaction_count" db:"transaction_count"`
	TotalSales       float64 `json:"total_sales"       db:"total_sales"`
}

// ─── Stock Valuation ──────────────────────────────────────────────────────────

// StockValuationRow is the valuation for one product.
type StockValuationRow struct {
	ProductID   string  `json:"product_id"   db:"product_id"`
	ProductName string  `json:"product_name" db:"product_name"`
	SKU         string  `json:"sku"          db:"sku"`
	Unit        string  `json:"unit"         db:"unit"`
	CostPrice   float64 `json:"cost_price"   db:"cost_price"`
	Quantity    float64 `json:"quantity"     db:"quantity"`
	TotalValue  float64 `json:"total_value"  db:"total_value"`
}

// StockValuationResponse wraps rows + grand total.
type StockValuationResponse struct {
	Rows       []StockValuationRow `json:"rows"`
	GrandTotal float64             `json:"grand_total"`
}

// ─── Profit Summary (period) ─────────────────────────────────────────────────

// ProfitPeriodRow is aggregated profit for one day (or week/month grouping).
type ProfitPeriodRow struct {
	Period      string  `json:"period"       db:"period"`
	TotalSales  float64 `json:"total_sales"  db:"total_sales"`
	TotalCost   float64 `json:"total_cost"   db:"total_cost"`
	GrossProfit float64 `json:"gross_profit" db:"gross_profit"`
	ProfitMargin float64 `json:"profit_margin" db:"profit_margin"`
}

// ProfitSummaryResponse wraps rows + period totals.
type ProfitSummaryResponse struct {
	Rows         []ProfitPeriodRow `json:"rows"`
	TotalSales   float64           `json:"total_sales"`
	TotalCost    float64           `json:"total_cost"`
	GrossProfit  float64           `json:"gross_profit"`
	ProfitMargin float64           `json:"profit_margin"` // %
}
