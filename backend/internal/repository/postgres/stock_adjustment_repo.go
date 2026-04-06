package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/moedahpos/backend/internal/domain"
)

type StockAdjustmentRepo struct {
	db *sqlx.DB
}

func NewStockAdjustmentRepo(db *sqlx.DB) *StockAdjustmentRepo {
	return &StockAdjustmentRepo{db: db}
}

//nolint:funlen,cyclop,gocognit // The accounting implementation demands strict integrity rules in a single tx
func (r *StockAdjustmentRepo) CreateAdjustment(ctx context.Context, storeID, userID string, input domain.CreateAdjustmentInput) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("StockAdjustmentRepo.CreateAdjustment begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	adjID := uuid.NewString()

	// 1. Insert header
	const insertAdjQ = `
		INSERT INTO stock_adjustments (id, product_id, store_id, type, reason, quantity, notes, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	if _, err := tx.ExecContext(ctx, insertAdjQ, adjID, input.ProductID, storeID, input.Type, input.Reason, input.Quantity, input.Notes, userID); err != nil {
		return fmt.Errorf("failed to insert stock_adjustments: %w", err)
	}

	if input.Type == "OUT" {
		// Log OUT: we must deduct from existing batches in FIFO order
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
		if err := tx.SelectContext(ctx, &batches, selectQ, input.ProductID, storeID); err != nil {
			return fmt.Errorf("lock batches for adjustment: %w", err)
		}

		// Verify sufficient stock
		var totalAvail float64
		for _, b := range batches {
			totalAvail += b.QuantityRemaining
		}
		if totalAvail < input.Quantity {
			return fmt.Errorf("insufficient stock for adjustment (available=%.3f, needed=%.3f)", totalAvail, input.Quantity)
		}

		remaining := input.Quantity
		for _, b := range batches {
			if remaining <= 0 {
				break
			}
			deduct := b.QuantityRemaining
			if deduct > remaining {
				deduct = remaining
			}
			newQty := b.QuantityRemaining - deduct

			// Update batch
			if _, err := tx.ExecContext(ctx, `UPDATE stock_batches SET quantity_remaining = $1 WHERE id = $2`, newQty, b.ID); err != nil {
				return fmt.Errorf("update batch %s: %w", b.ID, err)
			}

			// Insert junction tracking record
			batchAssocID := uuid.NewString()
			if _, err := tx.ExecContext(ctx, `INSERT INTO stock_adjustment_batches (id, adjustment_id, batch_id, deducted_qty) VALUES ($1, $2, $3, $4)`, batchAssocID, adjID, b.ID, deduct); err != nil {
				return fmt.Errorf("insert stock_adjustment_batches: %w", err)
			}

			remaining -= deduct
		}

		// Update total stock level
		if _, err := tx.ExecContext(ctx, `UPDATE stock_levels SET quantity = quantity - $1 WHERE product_id = $2 AND store_id = $3`, input.Quantity, input.ProductID, storeID); err != nil {
			return fmt.Errorf("update stock_levels OUT: %w", err)
		}

		// Write to overarching stock_movements ledger
		movementID := uuid.NewString()
		if _, err := tx.ExecContext(ctx, `INSERT INTO stock_movements (id, product_id, store_id, ref_type, ref_id, quantity_delta, notes, created_by) VALUES ($1, $2, $3, 'adjustment', $4, $5, $6, $7)`,
			movementID, input.ProductID, storeID, adjID, -input.Quantity, input.Notes, userID); err != nil {
			return fmt.Errorf("insert stock_movements OUT: %w", err)
		}

	} else if input.Type == "IN" {
		// Log IN: we create a new batch. We need the purchase price.
		// Try to find the most recent purchase price
		var recentPrice float64
		_ = tx.GetContext(ctx, &recentPrice, `SELECT purchase_price FROM stock_batches WHERE product_id = $1 AND store_id = $2 ORDER BY received_at DESC LIMIT 1`, input.ProductID, storeID)

		if recentPrice == 0 {
			// fallback to cost_price from products
			_ = tx.GetContext(ctx, &recentPrice, `SELECT cost_price FROM products WHERE id = $1`, input.ProductID)
		}

		newBatchID := uuid.NewString()
		if _, err := tx.ExecContext(ctx, `INSERT INTO stock_batches (id, product_id, store_id, quantity_remaining, purchase_price, received_at) VALUES ($1, $2, $3, $4, $5, NOW())`,
			newBatchID, input.ProductID, storeID, input.Quantity, recentPrice); err != nil {
			return fmt.Errorf("insert stock_batches for IN adjustment: %w", err)
		}

		// Insert junction
		assocID := uuid.NewString()
		if _, err := tx.ExecContext(ctx, `INSERT INTO stock_adjustment_batches (id, adjustment_id, batch_id, deducted_qty) VALUES ($1, $2, $3, $4)`, assocID, adjID, newBatchID, input.Quantity); err != nil {
			return fmt.Errorf("insert stock_adjustment_batches IN: %w", err)
		}

		// Update total stock level
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO stock_levels (id, product_id, store_id, quantity, min_quantity) 
			VALUES ($1, $2, $3, $4, 0)
			ON CONFLICT (product_id, store_id) DO UPDATE SET quantity = stock_levels.quantity + $4`,
			uuid.NewString(), input.ProductID, storeID, input.Quantity); err != nil {
			return fmt.Errorf("update stock_levels IN: %w", err)
		}

		// Write to overarching movements ledger
		movementID := uuid.NewString()
		if _, err := tx.ExecContext(ctx, `INSERT INTO stock_movements (id, product_id, store_id, ref_type, ref_id, quantity_delta, notes, created_by) VALUES ($1, $2, $3, 'adjustment', $4, $5, $6, $7)`,
			movementID, input.ProductID, storeID, adjID, input.Quantity, input.Notes, userID); err != nil {
			return fmt.Errorf("insert stock_movements IN: %w", err)
		}
	}

	return tx.Commit()
}

func (r *StockAdjustmentRepo) GetStockAdjustmentHistory(ctx context.Context, storeID string, productID *string) ([]*domain.StockAdjustment, error) {
	q := `
		SELECT 
			a.id, a.product_id, a.store_id, a.type, a.reason, a.quantity, a.notes, a.created_by, a.created_at, a.updated_at,
			p.name AS product_name, p.sku AS product_sku, p.unit,
			u.name AS created_by_name
		FROM stock_adjustments a
		JOIN products p ON p.id = a.product_id
		JOIN users u ON u.id = a.created_by
		WHERE a.store_id = $1
	`
	args := []interface{}{storeID}

	if productID != nil && *productID != "" {
		q += ` AND a.product_id = $2`
		args = append(args, *productID)
	}

	q += ` ORDER BY a.created_at DESC`

	var adjustments []*domain.StockAdjustment
	if err := r.db.SelectContext(ctx, &adjustments, q, args...); err != nil {
		return nil, fmt.Errorf("GetStockAdjustmentHistory: %w", err)
	}

	return adjustments, nil
}
