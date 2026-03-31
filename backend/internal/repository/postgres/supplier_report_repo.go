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
		SELECT
			TO_CHAR(created_at AT TIME ZONE 'Asia/Jakarta', 'YYYY-MM-DD') AS date,
			COUNT(*)                 AS transaction_count,
			ROUND(SUM(total)::numeric, 2)        AS total_sales,
			ROUND(SUM(tax_amt)::numeric, 2)      AS total_tax,
			ROUND(SUM(discount_amt)::numeric, 2) AS total_discount,
			ROUND((SUM(total) - SUM(discount_amt))::numeric, 2) AS total_net
		FROM transactions
		WHERE store_id = $1 AND status = 'completed'
		  AND created_at >= $2 AND created_at < $3
		GROUP BY 1
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
			ROUND(SUM(ti.quantity)::numeric, 3)  AS total_quantity,
			ROUND(SUM(ti.subtotal)::numeric, 2)  AS total_revenue,
			ROUND(SUM(ti.subtotal * ti.tax_rate / 100)::numeric, 2) AS total_tax
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
