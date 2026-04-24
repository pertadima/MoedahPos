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

// BatchRepo is the PostgreSQL implementation of repository.BatchRepository.
type BatchRepo struct{ db *sqlx.DB }

// NewBatchRepository creates a new BatchRepository.
func NewBatchRepository(db *sql.DB) repository.BatchRepository {
	return &BatchRepo{db: sqlx.NewDb(db, "postgres")}
}

// CreateBatch inserts one stock batch record generated when a PO item is received.
func (r *BatchRepo) CreateBatch(ctx context.Context, batch *domain.StockBatch) error {
	const q = `
		INSERT INTO stock_batches
			(product_id, store_id, po_id, quantity_remaining, purchase_price, received_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.ExecContext(ctx, q,
		batch.ProductID, batch.StoreID, batch.POID,
		batch.QuantityRemaining, batch.PurchasePrice, batch.ReceivedAt,
	)
	if err != nil {
		return fmt.Errorf("BatchRepo.CreateBatch: %w", err)
	}
	return nil
}

// GetBatchesByProduct returns all non-empty batches for a product at a store,
// ordered oldest-first (FIFO order). Used for audit and debugging purposes.
func (r *BatchRepo) GetBatchesByProduct(ctx context.Context, productID, storeID string) ([]*domain.StockBatch, error) {
	const q = `
		SELECT sb.id, sb.product_id, sb.store_id, sb.po_id,
		       sb.quantity_remaining, sb.purchase_price, sb.received_at, sb.created_at,
		       p.name AS product_name, p.sku AS product_sku, p.unit
		FROM stock_batches sb
		JOIN products p ON p.id = sb.product_id
		WHERE sb.product_id = $1 AND sb.store_id = $2 AND sb.quantity_remaining > 0
		ORDER BY sb.received_at ASC` // oldest first for clear FIFO display
	var batches []*domain.StockBatch
	if err := r.db.SelectContext(ctx, &batches, q, productID, storeID); err != nil {
		return nil, fmt.Errorf("BatchRepo.GetBatchesByProduct: %w", err)
	}
	return batches, nil
}

// GetBatchesByStore returns non-empty batches for a store, grouped by product.
// If filter.ProductID is set, only that product's batches are returned.
func (r *BatchRepo) GetBatchesByStore(ctx context.Context, f dto.BatchListFilter) ([]*domain.StockBatch, error) {
	q := `
		SELECT sb.id, sb.product_id, sb.store_id, sb.po_id,
		       sb.quantity_remaining, sb.purchase_price, sb.received_at, sb.created_at,
		       p.name AS product_name, p.sku AS product_sku, p.unit
		FROM stock_batches sb
		JOIN products p ON p.id = sb.product_id
		WHERE sb.store_id = $1 AND sb.quantity_remaining > 0`
	args := []interface{}{f.StoreID}
	if f.ProductID != "" {
		// Narrow to a single product when product_id query param is provided.
		q += " AND sb.product_id = $2"
		args = append(args, f.ProductID)
	}
	q += " ORDER BY p.name, sb.received_at ASC"
	var batches []*domain.StockBatch
	if err := r.db.SelectContext(ctx, &batches, q, args...); err != nil {
		return nil, fmt.Errorf("BatchRepo.GetBatchesByStore: %w", err)
	}
	return batches, nil
}

// GetStockSummary returns total stock per product aggregated across all batches.
// This gives a per-product view of available batch-tracked inventory.
func (r *BatchRepo) GetStockSummary(ctx context.Context, storeID string) ([]*domain.BatchStockSummary, error) {
	const q = `
		SELECT sb.product_id,
		       p.name  AS product_name,
		       p.sku   AS product_sku,
		       p.unit,
		       COALESCE(SUM(sb.quantity_remaining),  0) AS total_qty,
		       COUNT(*)                                  AS batch_count,
		       COALESCE(AVG(sb.purchase_price),      0) AS avg_cost_price
		FROM stock_batches sb
		JOIN products p ON p.id = sb.product_id
		WHERE sb.store_id = $1 AND sb.quantity_remaining > 0
		GROUP BY sb.product_id, p.name, p.sku, p.unit
		ORDER BY p.name`
	var rows []*domain.BatchStockSummary
	if err := r.db.SelectContext(ctx, &rows, q, storeID); err != nil {
		return nil, fmt.Errorf("BatchRepo.GetStockSummary: %w", err)
	}
	return rows, nil
}

// DeductFIFO deducts qty from the oldest available batches within one DB transaction.
//
// FIFO algorithm:
//  1. Lock all non-empty batches for the product ordered by received_at ASC (FOR UPDATE).
//  2. Verify total available ≥ requested qty; return error immediately if not.
//  3. Iterate batches oldest-first, subtracting from each until qty is fully covered.
//  4. Commit — any concurrent caller will block on step 1 until we release the locks.
func (r *BatchRepo) DeductFIFO(ctx context.Context, productID, storeID string, qty float64) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("BatchRepo.DeductFIFO begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Step 1 – lock batches in FIFO order.
	// FOR UPDATE prevents concurrent transactions from reading stale quantities.
	const selectQ = `
		SELECT id, quantity_remaining
		FROM stock_batches
		WHERE product_id = $1 AND store_id = $2 AND quantity_remaining > 0
		ORDER BY received_at ASC
		FOR UPDATE`

	type batchRow struct {
		ID                string  `db:"id"`
		QuantityRemaining float64 `db:"quantity_remaining"`
	}

	var batches []batchRow
	if err := tx.SelectContext(ctx, &batches, selectQ, productID, storeID); err != nil {
		return fmt.Errorf("BatchRepo.DeductFIFO lock batches: %w", err)
	}

	// Step 2 – verify sufficient batch stock.
	var totalAvail float64
	for _, b := range batches {
		totalAvail += b.QuantityRemaining
	}
	if totalAvail < qty {
		return fmt.Errorf("insufficient batch stock for product %s (available=%.3f, needed=%.3f)",
			productID, totalAvail, qty)
	}

	// Step 3 – FIFO deduction: consume oldest batches first.
	remaining := qty
	for _, b := range batches {
		if remaining <= 0 {
			break // fully covered — stop iterating
		}
		// Deduct as much as this batch can provide.
		deduct := b.QuantityRemaining
		if deduct > remaining {
			deduct = remaining
		}
		newQty := b.QuantityRemaining - deduct
		if _, err := tx.ExecContext(ctx,
			`UPDATE stock_batches SET quantity_remaining = $1 WHERE id = $2`,
			newQty, b.ID,
		); err != nil {
			return fmt.Errorf("BatchRepo.DeductFIFO update batch %s: %w", b.ID, err)
		}
		remaining -= deduct
	}

	// Step 4 – commit releases all row locks.
	return tx.Commit()
}
