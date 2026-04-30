package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
	"github.com/moedahpos/backend/internal/repository"
)

// PriceHistoryRepo is the PostgreSQL implementation of PriceHistoryRepository.
type PriceHistoryRepo struct{ db *sqlx.DB }

// NewPriceHistoryRepository creates a new PriceHistoryRepository.
func NewPriceHistoryRepository(db *sql.DB) repository.PriceHistoryRepository {
	return &PriceHistoryRepo{db: sqlx.NewDb(db, "postgres")}
}

// Record inserts one price-change audit row.
func (r *PriceHistoryRepo) Record(ctx context.Context, h domain.PriceHistory) error {
	const q = `
		INSERT INTO price_history
		  (product_id, store_id, changed_by, old_cost, new_cost, old_sell, new_sell, source, ref_id, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	_, err := r.db.ExecContext(ctx, q,
		h.ProductID, h.StoreID, h.ChangedBy,
		h.OldCost, h.NewCost, h.OldSell, h.NewSell,
		h.Source, h.RefID, h.Notes,
	)
	if err != nil {
		return fmt.Errorf("PriceHistoryRepo.Record: %w", err)
	}
	return nil
}

// FindByProduct returns price-change history for one product, newest first.
func (r *PriceHistoryRepo) FindByProduct(ctx context.Context, productID string, f dto.PriceHistoryFilter) ([]*domain.PriceHistory, int, error) {
	const countQ = `SELECT COUNT(*) FROM price_history WHERE product_id = $1`
	var total int
	if err := r.db.QueryRowxContext(ctx, countQ, productID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("PriceHistoryRepo.FindByProduct count: %w", err)
	}

	const q = `
		SELECT
			ph.id, ph.product_id, p.name AS product_name,
			ph.store_id, ph.changed_by,
			COALESCE(u.name, '') AS changed_by_name,
			ph.old_cost, ph.new_cost, ph.old_sell, ph.new_sell,
			ph.source, ph.ref_id, ph.notes, ph.changed_at
		FROM price_history ph
		JOIN products p ON p.id = ph.product_id
		JOIN users    u ON u.id = ph.changed_by
		WHERE ph.product_id = $1
		ORDER BY ph.changed_at DESC
		LIMIT $2 OFFSET $3`

	var rows []*domain.PriceHistory
	if err := r.db.SelectContext(ctx, &rows, q, productID, f.PerPage, f.Offset()); err != nil {
		return nil, 0, fmt.Errorf("PriceHistoryRepo.FindByProduct: %w", err)
	}
	return rows, total, nil
}

// FindByStore returns price-change history across all products in a store, newest first.
func (r *PriceHistoryRepo) FindByStore(ctx context.Context, storeID string, f dto.PriceHistoryFilter) ([]*domain.PriceHistory, int, error) {
	args := []interface{}{storeID}
	conds := "ph.store_id = $1"
	i := 2

	if f.ProductID != "" {
		conds += fmt.Sprintf(" AND ph.product_id = $%d", i)
		args = append(args, f.ProductID)
		i++
	}
	if f.Source != "" {
		conds += fmt.Sprintf(" AND ph.source = $%d", i)
		args = append(args, f.Source)
		i++
	}

	var total int
	if err := r.db.QueryRowxContext(ctx,
		fmt.Sprintf("SELECT COUNT(*) FROM price_history ph WHERE %s", conds), args...,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("PriceHistoryRepo.FindByStore count: %w", err)
	}

	args = append(args, f.PerPage, f.Offset())
	q := fmt.Sprintf(`
		SELECT
			ph.id, ph.product_id, p.name AS product_name,
			ph.store_id, ph.changed_by,
			COALESCE(u.name, '') AS changed_by_name,
			ph.old_cost, ph.new_cost, ph.old_sell, ph.new_sell,
			ph.source, ph.ref_id, ph.notes, ph.changed_at
		FROM price_history ph
		JOIN products p ON p.id = ph.product_id
		JOIN users    u ON u.id = ph.changed_by
		WHERE %s
		ORDER BY ph.changed_at DESC
		LIMIT $%d OFFSET $%d`, conds, i, i+1)

	var rows []*domain.PriceHistory
	if err := r.db.SelectContext(ctx, &rows, q, args...); err != nil {
		return nil, 0, fmt.Errorf("PriceHistoryRepo.FindByStore: %w", err)
	}
	return rows, total, nil
}
