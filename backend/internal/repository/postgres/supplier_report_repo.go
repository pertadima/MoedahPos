package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
)

// ── Supplier Repo ─────────────────────────────────────────────────────────────

type SupplierRepo struct{ db *sqlx.DB }

func NewSupplierRepo(db *sqlx.DB) *SupplierRepo { return &SupplierRepo{db: db} }

func (r *SupplierRepo) Create(ctx context.Context, s *domain.Supplier) (*domain.Supplier, error) {
	const q = `
		INSERT INTO suppliers (name, contact_name, phone, email, address)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id, name, contact_name, phone, email, address, is_active, created_at, updated_at, deleted_at`
	row := &domain.Supplier{}
	if err := r.db.QueryRowxContext(ctx, q, s.Name, s.ContactName, s.Phone, s.Email, s.Address).StructScan(row); err != nil {
		return nil, fmt.Errorf("SupplierRepo.Create: %w", err)
	}
	return row, nil
}

func (r *SupplierRepo) FindAll(ctx context.Context, f dto.SupplierListFilter) ([]*domain.Supplier, int, error) {
	args := []interface{}{}
	conds := []string{"deleted_at IS NULL"}
	i := 1

	if f.Search != "" {
		conds = append(conds, fmt.Sprintf("name ILIKE $%d", i))
		args = append(args, "%"+f.Search+"%")
		i++
	}
	if f.IsActive != nil {
		conds = append(conds, fmt.Sprintf("is_active = $%d", i))
		args = append(args, *f.IsActive)
		i++
	}
	where := "WHERE " + strings.Join(conds, " AND ")

	var total int
	if err := r.db.QueryRowxContext(ctx,
		fmt.Sprintf("SELECT COUNT(*) FROM suppliers %s", where), args...,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("SupplierRepo.FindAll count: %w", err)
	}

	args = append(args, f.PerPage, f.Offset())
	dataQ := fmt.Sprintf(`
		SELECT id, name, contact_name, phone, email, address, is_active, created_at, updated_at, deleted_at
		FROM suppliers %s ORDER BY name ASC LIMIT $%d OFFSET $%d`,
		where, i, i+1)

	var suppliers []*domain.Supplier
	if err := r.db.SelectContext(ctx, &suppliers, dataQ, args...); err != nil {
		return nil, 0, fmt.Errorf("SupplierRepo.FindAll: %w", err)
	}
	return suppliers, total, nil
}

func (r *SupplierRepo) FindByID(ctx context.Context, id string) (*domain.Supplier, error) {
	const q = `
		SELECT id, name, contact_name, phone, email, address, is_active, created_at, updated_at, deleted_at
		FROM suppliers WHERE id = $1 AND deleted_at IS NULL`
	s := &domain.Supplier{}
	if err := r.db.QueryRowxContext(ctx, q, id).StructScan(s); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("SupplierRepo.FindByID: %w", err)
	}
	return s, nil
}

func (r *SupplierRepo) Update(ctx context.Context, s *domain.Supplier) (*domain.Supplier, error) {
	const q = `
		UPDATE suppliers SET name=$1, contact_name=$2, phone=$3, email=$4, address=$5, is_active=$6, updated_at=NOW()
		WHERE id=$7 AND deleted_at IS NULL
		RETURNING id, name, contact_name, phone, email, address, is_active, created_at, updated_at, deleted_at`
	row := &domain.Supplier{}
	if err := r.db.QueryRowxContext(ctx, q, s.Name, s.ContactName, s.Phone, s.Email, s.Address, s.IsActive, s.ID).StructScan(row); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("SupplierRepo.Update: %w", err)
	}
	return row, nil
}

func (r *SupplierRepo) SoftDelete(ctx context.Context, id string) error {
	const q = `UPDATE suppliers SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	res, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("SupplierRepo.SoftDelete: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("supplier not found")
	}
	return nil
}

// ── Report Repo ───────────────────────────────────────────────────────────────

type ReportRepo struct{ db *sqlx.DB }

func NewReportRepo(db *sqlx.DB) *ReportRepo { return &ReportRepo{db: db} }

func (r *ReportRepo) SalesSummary(ctx context.Context, storeID string, from, to time.Time) ([]dto.SalesSummaryRow, error) {
	const q = `
		WITH sales AS (
			SELECT
				TO_CHAR(t.created_at AT TIME ZONE 'Asia/Jakarta', 'YYYY-MM-DD') AS date,
				COUNT(t.id) AS transaction_count,
				SUM(t.total) AS total_sales,
				SUM(t.tax_amt) AS total_tax,
				SUM(t.discount_amt) AS total_discount,
				SUM(t.total) - SUM(t.discount_amt) AS total_net,
				COALESCE(SUM(ti_agg.item_cost), 0) AS total_cost,
				SUM(t.total) - COALESCE(SUM(ti_agg.item_cost), 0) AS gross_profit
			FROM transactions t
			LEFT JOIN (
				SELECT transaction_id, SUM(cost_price * quantity) AS item_cost
				FROM transaction_items
				GROUP BY transaction_id
			) ti_agg ON ti_agg.transaction_id = t.id
			WHERE t.store_id = $1 AND t.status = 'completed'
			  AND t.created_at >= $2 AND t.created_at < $3
			GROUP BY 1
		),
		exp AS (
			SELECT
				TO_CHAR(e.expense_date, 'YYYY-MM-DD') AS date,
				SUM(e.amount) AS total_expense
			FROM expenses e
			WHERE e.store_id = $1 AND e.expense_date >= $2 AND e.expense_date < $3
			GROUP BY 1
		)
		SELECT
			COALESCE(s.date, e.date) AS date,
			COALESCE(s.transaction_count, 0) AS transaction_count,
			ROUND(COALESCE(s.total_sales, 0)::numeric, 2) AS total_sales,
			ROUND(COALESCE(s.total_tax, 0)::numeric, 2) AS total_tax,
			ROUND(COALESCE(s.total_discount, 0)::numeric, 2) AS total_discount,
			ROUND(COALESCE(s.total_net, 0)::numeric, 2) AS total_net,
			ROUND(COALESCE(s.total_cost, 0)::numeric, 2) AS total_cost,
			ROUND(COALESCE(s.gross_profit, 0)::numeric, 2) AS gross_profit,
			ROUND(COALESCE(e.total_expense, 0)::numeric, 2) AS total_expense,
			ROUND((COALESCE(s.gross_profit, 0) - COALESCE(e.total_expense, 0))::numeric, 2) AS net_profit
		FROM sales s
		FULL OUTER JOIN exp e ON e.date = s.date
		ORDER BY 1 DESC`
	var rows []dto.SalesSummaryRow
	if err := r.db.SelectContext(ctx, &rows, q, storeID, from, to); err != nil {
		return nil, fmt.Errorf("ReportRepo.SalesSummary: %w", err)
	}
	return rows, nil
}

func (r *ReportRepo) SalesByProduct(ctx context.Context, storeID string, from, to time.Time) ([]dto.SalesByProductRow, error) {
	const q = `
		SELECT
			COALESCE(ti.product_id::text, '') AS product_id,
			ti.product_name,
			ti.sku,
			ROUND(SUM(ti.quantity)::numeric, 3)                                   AS total_quantity,
			ROUND(SUM(ti.subtotal)::numeric, 2)                                   AS total_revenue,
			ROUND(SUM(ti.cost_price * ti.quantity)::numeric, 2)                   AS total_cost,
			ROUND((SUM(ti.subtotal) - SUM(ti.cost_price * ti.quantity))::numeric, 2) AS gross_profit,
			CASE WHEN SUM(ti.subtotal) > 0
			     THEN ROUND(((SUM(ti.subtotal) - SUM(ti.cost_price * ti.quantity)) / SUM(ti.subtotal) * 100)::numeric, 1)
			     ELSE 0 END                                                        AS profit_margin,
			ROUND(SUM(ti.subtotal * ti.tax_rate / 100)::numeric, 2)               AS total_tax
		FROM transaction_items ti
		JOIN transactions t ON t.id = ti.transaction_id
		WHERE t.store_id = $1 AND t.status = 'completed'
		  AND t.created_at >= $2 AND t.created_at < $3
		GROUP BY ti.product_id, ti.product_name, ti.sku
		ORDER BY total_revenue DESC`
	var rows []dto.SalesByProductRow
	if err := r.db.SelectContext(ctx, &rows, q, storeID, from, to); err != nil {
		return nil, fmt.Errorf("ReportRepo.SalesByProduct: %w", err)
	}
	return rows, nil
}

func (r *ReportRepo) SalesByCashier(ctx context.Context, storeID string, from, to time.Time) ([]dto.SalesByCashierRow, error) {
	const q = `
		SELECT
			u.id   AS cashier_id,
			u.name AS cashier_name,
			COUNT(t.id)           AS transaction_count,
			ROUND(SUM(t.total)::numeric, 2) AS total_sales
		FROM transactions t
		JOIN users u ON u.id = t.cashier_id
		WHERE t.store_id = $1 AND t.status = 'completed'
		  AND t.created_at >= $2 AND t.created_at < $3
		GROUP BY u.id, u.name
		ORDER BY total_sales DESC`
	var rows []dto.SalesByCashierRow
	if err := r.db.SelectContext(ctx, &rows, q, storeID, from, to); err != nil {
		return nil, fmt.Errorf("ReportRepo.SalesByCashier: %w", err)
	}
	return rows, nil
}

func (r *ReportRepo) StockValuation(ctx context.Context, storeID string) ([]dto.StockValuationRow, error) {
	const q = `
		SELECT
			p.id   AS product_id,
			p.name AS product_name,
			p.sku,
			p.unit,
			p.cost_price,
			COALESCE(sl.quantity, 0) AS quantity,
			ROUND((COALESCE(sl.quantity, 0) * p.cost_price)::numeric, 2) AS total_value
		FROM products p
		LEFT JOIN stock_levels sl ON sl.product_id = p.id AND sl.store_id = $1
		WHERE p.store_id = $1 AND p.deleted_at IS NULL AND p.is_active = true
		ORDER BY total_value DESC`
	var rows []dto.StockValuationRow
	if err := r.db.SelectContext(ctx, &rows, q, storeID); err != nil {
		return nil, fmt.Errorf("ReportRepo.StockValuation: %w", err)
	}
	return rows, nil
}

// ProfitSummary returns profit grouped by the given period expression.
// groupBy must be a trusted pg expression (e.g. "day" | "week" | "month").
func (r *ReportRepo) ProfitSummary(ctx context.Context, storeID string, from, to time.Time, groupBy string) ([]dto.ProfitPeriodRow, error) {
	// Safe allowlist — never interpolated from user input directly
	periodExpr := map[string]string{
		"day":   "TO_CHAR(t.created_at AT TIME ZONE 'Asia/Jakarta', 'YYYY-MM-DD')",
		"week":  "TO_CHAR(DATE_TRUNC('week', t.created_at AT TIME ZONE 'Asia/Jakarta'), 'YYYY-MM-DD')",
		"month": "TO_CHAR(DATE_TRUNC('month', t.created_at AT TIME ZONE 'Asia/Jakarta'), 'YYYY-MM')",
	}
	expensePeriodExpr := map[string]string{
		"day":   "TO_CHAR(e.expense_date, 'YYYY-MM-DD')",
		"week":  "TO_CHAR(DATE_TRUNC('week', e.expense_date), 'YYYY-MM-DD')",
		"month": "TO_CHAR(DATE_TRUNC('month', e.expense_date), 'YYYY-MM')",
	}

	expr, ok := periodExpr[groupBy]
	if !ok {
		expr = periodExpr["day"]
	}
	expExpr, ok := expensePeriodExpr[groupBy]
	if !ok {
		expExpr = expensePeriodExpr["day"]
	}

	q := fmt.Sprintf(`
		WITH sales AS (
			SELECT
				%s AS period,
				SUM(t.total) AS total_sales,
				COALESCE(SUM(ti_agg.item_cost), 0) AS total_cost
			FROM transactions t
			LEFT JOIN (
				SELECT transaction_id, SUM(cost_price * quantity) AS item_cost
				FROM transaction_items
				GROUP BY transaction_id
			) ti_agg ON ti_agg.transaction_id = t.id
			WHERE t.store_id = $1 AND t.status = 'completed'
			  AND t.created_at >= $2 AND t.created_at < $3
			GROUP BY 1
		),
		exp AS (
			SELECT
				%s AS period,
				SUM(e.amount) AS total_expense
			FROM expenses e
			WHERE e.store_id = $1 AND e.expense_date >= $2 AND e.expense_date < $3
			GROUP BY 1
		)
		SELECT
			COALESCE(s.period, e.period) AS period,
			ROUND(COALESCE(s.total_sales, 0)::numeric, 2) AS total_sales,
			ROUND(COALESCE(s.total_cost, 0)::numeric, 2) AS total_cost,
			ROUND((COALESCE(s.total_sales, 0) - COALESCE(s.total_cost, 0))::numeric, 2) AS gross_profit,
			ROUND(COALESCE(e.total_expense, 0)::numeric, 2) AS total_expense,
			ROUND((COALESCE(s.total_sales, 0) - COALESCE(s.total_cost, 0) - COALESCE(e.total_expense, 0))::numeric, 2) AS net_profit,
			CASE WHEN COALESCE(s.total_sales, 0) > 0
			     THEN ROUND(((COALESCE(s.total_sales, 0) - COALESCE(s.total_cost, 0) - COALESCE(e.total_expense, 0)) / COALESCE(s.total_sales, 0) * 100)::numeric, 1)
			     ELSE 0 END AS profit_margin
		FROM sales s
		FULL OUTER JOIN exp e ON e.period = s.period
		ORDER BY 1`, expr, expExpr)

	var rows []dto.ProfitPeriodRow
	if err := r.db.SelectContext(ctx, &rows, q, storeID, from, to); err != nil {
		return nil, fmt.Errorf("ReportRepo.ProfitSummary: %w", err)
	}
	return rows, nil
}

// cashInRow is a scanned row from the cash_in CTE (sales transactions).
type cashInRow struct {
	Date          string  `db:"date"`
	CashIn        float64 `db:"cash_in"`
	PaymentMethod string  `db:"payment_method"`
}

// incomeInRow is a scanned row from the incomes table.
type incomeInRow struct {
	Date    string  `db:"date"`
	OtherIn float64 `db:"other_in"`
}

// cashOutRow is a scanned row from the cash_out aggregation.
type cashOutRow struct {
	Date    string  `db:"date"`
	CashOut float64 `db:"cash_out"`
}

// cashFlowDayData is the in-memory accumulator for a single day.
type cashFlowDayData struct {
	cashIn         float64
	cashOut        float64
	salesIn        float64
	otherIn        float64
	cashInByMethod map[string]float64
}

const cashInQ = `
	SELECT
		TO_CHAR(t.created_at AT TIME ZONE 'Asia/Jakarta', 'YYYY-MM-DD') AS date,
		SUM(t.total) AS cash_in,
		t.payment_method
	FROM transactions t
	WHERE t.store_id = $1
	  AND t.status = 'completed'
	  AND t.created_at >= $2 AND t.created_at < $3
	GROUP BY 1, t.payment_method
	ORDER BY 1`

const cashOutQ = `
	SELECT date, SUM(cash_out) AS cash_out FROM (
		SELECT
			TO_CHAR(pp.paid_at AT TIME ZONE 'Asia/Jakarta', 'YYYY-MM-DD') AS date,
			SUM(pp.amount) AS cash_out
		FROM po_payments pp
		WHERE pp.store_id = $1
		  AND pp.paid_at >= $2 AND pp.paid_at < $3
		GROUP BY 1
		UNION ALL
		SELECT
			TO_CHAR(pr.created_at AT TIME ZONE 'Asia/Jakarta', 'YYYY-MM-DD') AS date,
			SUM(pr.amount_paid) AS cash_out
		FROM payment_records pr
		JOIN purchase_order_termins pot ON pot.id = pr.termin_id
		JOIN purchase_orders po ON po.id = pot.po_id
		WHERE po.store_id = $1
		  AND pr.created_at >= $2 AND pr.created_at < $3
		GROUP BY 1
		UNION ALL
		SELECT
			TO_CHAR(e.expense_date, 'YYYY-MM-DD') AS date,
			SUM(e.amount) AS cash_out
		FROM expenses e
		WHERE e.store_id = $1
		  AND e.payment_status = 'paid'
		  AND e.expense_date >= $2::date AND e.expense_date < $3::date
		GROUP BY 1
	) sub
	GROUP BY date
	ORDER BY date`

// mergeCashFlowRows combines in/out/income rows into a sorted day-level slice.
func mergeCashFlowRows(inRows []cashInRow, outRows []cashOutRow, incomeRows []incomeInRow) []dto.CashFlowDayRow {
	dayMap := map[string]*cashFlowDayData{}
	allDates := map[string]struct{}{}

	for _, row := range inRows {
		allDates[row.Date] = struct{}{}
		d, ok := dayMap[row.Date]
		if !ok {
			d = &cashFlowDayData{cashInByMethod: map[string]float64{}}
			dayMap[row.Date] = d
		}
		d.cashIn += row.CashIn
		d.salesIn += row.CashIn
		d.cashInByMethod[row.PaymentMethod] += row.CashIn
	}
	for _, row := range incomeRows {
		allDates[row.Date] = struct{}{}
		d, ok := dayMap[row.Date]
		if !ok {
			d = &cashFlowDayData{cashInByMethod: map[string]float64{}}
			dayMap[row.Date] = d
		}
		d.cashIn += row.OtherIn
		d.otherIn += row.OtherIn
	}
	for _, row := range outRows {
		allDates[row.Date] = struct{}{}
		d, ok := dayMap[row.Date]
		if !ok {
			d = &cashFlowDayData{cashInByMethod: map[string]float64{}}
			dayMap[row.Date] = d
		}
		d.cashOut += row.CashOut
	}

	dates := make([]string, 0, len(allDates))
	for d := range allDates {
		dates = append(dates, d)
	}
	// YYYY-MM-DD strings sort lexicographically = chronologically
	for i := range dates {
		for j := i + 1; j < len(dates); j++ {
			if dates[i] > dates[j] {
				dates[i], dates[j] = dates[j], dates[i]
			}
		}
	}

	result := make([]dto.CashFlowDayRow, 0, len(dates))
	for _, d := range dates {
		dd := dayMap[d]
		result = append(result, dto.CashFlowDayRow{
			Date:           d,
			CashIn:         dd.cashIn,
			CashOut:        dd.cashOut,
			NetCash:        dd.cashIn - dd.cashOut,
			CashInByMethod: dd.cashInByMethod,
			SalesIn:        dd.salesIn,
			OtherIn:        dd.otherIn,
		})
	}
	return result
}

// incomeInQ returns per-day income totals from the incomes table.
const incomeInQ = `
	SELECT
		TO_CHAR(income_date, 'YYYY-MM-DD') AS date,
		SUM(amount) AS other_in
	FROM incomes
	WHERE store_id = $1
	  AND income_date >= $2::date AND income_date < $3::date
	GROUP BY 1
	ORDER BY 1`

// CashFlowSummary builds per-day cash in/out rows from actual paid transactions and incomes.
func (r *ReportRepo) CashFlowSummary(ctx context.Context, storeID string, from, to time.Time) ([]dto.CashFlowDayRow, error) {
	var inRows []cashInRow
	if err := r.db.SelectContext(ctx, &inRows, cashInQ, storeID, from, to); err != nil {
		return nil, fmt.Errorf("ReportRepo.CashFlowSummary cash_in: %w", err)
	}
	var outRows []cashOutRow
	if err := r.db.SelectContext(ctx, &outRows, cashOutQ, storeID, from, to); err != nil {
		return nil, fmt.Errorf("ReportRepo.CashFlowSummary cash_out: %w", err)
	}
	var incomeRows []incomeInRow
	if err := r.db.SelectContext(ctx, &incomeRows, incomeInQ, storeID, from, to); err != nil {
		return nil, fmt.Errorf("ReportRepo.CashFlowSummary income_in: %w", err)
	}
	return mergeCashFlowRows(inRows, outRows, incomeRows), nil
}
