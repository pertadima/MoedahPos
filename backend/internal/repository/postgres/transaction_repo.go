package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
)

// TransactionRepo is the PostgreSQL implementation of repository.TransactionRepository.
type TransactionRepo struct{ db *sqlx.DB }

func NewTransactionRepo(db *sqlx.DB) *TransactionRepo { return &TransactionRepo{db: db} }

// Create persists a full transaction with items and stock movements in a single DB transaction.
func (r *TransactionRepo) Create(ctx context.Context, input domain.CreateTransactionInput) (*domain.Transaction, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("TransactionRepo.Create begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// 1. INSERT transaction header
	const txnQ = `
		INSERT INTO transactions
		  (store_id, cashier_id, customer_name, customer_phone,
		   subtotal, discount_amt, tax_amt, total,
		   payment_method, payment_amount, change_amount, status, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,'completed',$12)
		RETURNING id, store_id, cashier_id, customer_name, customer_phone,
		          subtotal, discount_amt, tax_amt, total,
		          payment_method, payment_amount, change_amount, status, notes,
		          created_at, updated_at`

	txn := &domain.Transaction{}
	err = tx.QueryRowxContext(ctx, txnQ,
		input.StoreID, input.CashierID, input.CustomerName, input.CustomerPhone,
		input.Subtotal, input.DiscountAmt, input.TaxAmt, input.Total,
		input.PaymentMethod, input.PaymentAmount, input.ChangeAmount, input.Notes,
	).StructScan(txn)
	if err != nil {
		return nil, fmt.Errorf("TransactionRepo.Create insert txn: %w", err)
	}

	// 2. INSERT items + stock movements per item
	const itemQ = `
		INSERT INTO transaction_items
		  (transaction_id, product_id, product_name, sku, quantity, unit_price, discount_pct, tax_rate, subtotal)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id, transaction_id, product_id, product_name, sku, quantity, unit_price, discount_pct, tax_rate, subtotal`

	const mvQ = `
		INSERT INTO stock_movements (product_id, store_id, ref_type, ref_id, quantity_delta, notes, created_by)
		VALUES ($1,$2,'sale',$3,$4,'Sale transaction',$5)`

	const slQ = `
		INSERT INTO stock_levels (product_id, store_id, quantity, updated_at)
		VALUES ($1,$2,$3,NOW())
		ON CONFLICT (product_id, store_id)
		DO UPDATE SET quantity = stock_levels.quantity + $3, updated_at = NOW()`

	for _, item := range input.Items {
		ti := &domain.TransactionItem{}
		if err := tx.QueryRowxContext(ctx, itemQ,
			txn.ID, item.ProductID, item.ProductName, item.SKU,
			item.Quantity, item.UnitPrice, item.DiscountPct, item.TaxRate, item.Subtotal,
		).StructScan(ti); err != nil {
			return nil, fmt.Errorf("TransactionRepo.Create insert item: %w", err)
		}
		txn.Items = append(txn.Items, *ti)

		if item.ProductID != nil {
			// Stock movement (negative delta = stock out)
			if _, err := tx.ExecContext(ctx, mvQ,
				*item.ProductID, input.StoreID, txn.ID, -item.Quantity, input.CashierID,
			); err != nil {
				return nil, fmt.Errorf("TransactionRepo.Create stock movement: %w", err)
			}
			// Upsert stock level
			if _, err := tx.ExecContext(ctx, slQ, *item.ProductID, input.StoreID, -item.Quantity); err != nil {
				return nil, fmt.Errorf("TransactionRepo.Create stock level: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("TransactionRepo.Create commit: %w", err)
	}

	// Load cashier name
	_ = r.db.QueryRowxContext(ctx, `SELECT name FROM users WHERE id = $1`, input.CashierID).Scan(&txn.CashierName)

	return txn, nil
}

// FindAll returns a paginated list of transactions.
func (r *TransactionRepo) FindAll(ctx context.Context, f dto.TransactionListFilter) ([]*domain.Transaction, int, error) {
	args := []interface{}{f.StoreID}
	conds := []string{"t.store_id = $1"}
	i := 2

	if f.Status != "" {
		conds = append(conds, fmt.Sprintf("t.status = $%d", i))
		args = append(args, f.Status)
		i++
	}
	if f.CashierID != "" {
		conds = append(conds, fmt.Sprintf("t.cashier_id = $%d", i))
		args = append(args, f.CashierID)
		i++
	}
	if f.DateFrom != "" {
		conds = append(conds, fmt.Sprintf("t.created_at >= $%d", i))
		args = append(args, f.DateFrom)
		i++
	}
	if f.DateTo != "" {
		conds = append(conds, fmt.Sprintf("t.created_at < $%d", i))
		args = append(args, f.DateTo)
		i++
	}

	where := "WHERE " + strings.Join(conds, " AND ")

	var total int
	if err := r.db.QueryRowxContext(ctx,
		fmt.Sprintf("SELECT COUNT(*) FROM transactions t %s", where), args...,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("TransactionRepo.FindAll count: %w", err)
	}

	args = append(args, f.PerPage, f.Offset())
	dataQ := fmt.Sprintf(`
		SELECT t.id, t.store_id, t.cashier_id, t.customer_name, t.customer_phone,
		       t.subtotal, t.discount_amt, t.tax_amt, t.total,
		       t.payment_method, t.payment_amount, t.change_amount, t.status, t.notes,
		       t.created_at, t.updated_at, u.name AS cashier_name
		FROM transactions t
		JOIN users u ON u.id = t.cashier_id
		%s
		ORDER BY t.created_at DESC
		LIMIT $%d OFFSET $%d`, where, i, i+1)

	var txns []*domain.Transaction
	if err := r.db.SelectContext(ctx, &txns, dataQ, args...); err != nil {
		return nil, 0, fmt.Errorf("TransactionRepo.FindAll: %w", err)
	}
	return txns, total, nil
}

// FindByID returns a transaction with all its items.
func (r *TransactionRepo) FindByID(ctx context.Context, id string) (*domain.Transaction, error) {
	const txnQ = `
		SELECT t.id, t.store_id, t.cashier_id, t.customer_name, t.customer_phone,
		       t.subtotal, t.discount_amt, t.tax_amt, t.total,
		       t.payment_method, t.payment_amount, t.change_amount, t.status, t.notes,
		       t.created_at, t.updated_at, u.name AS cashier_name
		FROM transactions t
		JOIN users u ON u.id = t.cashier_id
		WHERE t.id = $1`

	txn := &domain.Transaction{}
	if err := r.db.QueryRowxContext(ctx, txnQ, id).StructScan(txn); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("TransactionRepo.FindByID: %w", err)
	}

	const itemQ = `
		SELECT id, transaction_id, product_id, product_name, sku,
		       quantity, unit_price, discount_pct, tax_rate, subtotal
		FROM transaction_items WHERE transaction_id = $1`

	if err := r.db.SelectContext(ctx, &txn.Items, itemQ, id); err != nil {
		return nil, fmt.Errorf("TransactionRepo.FindByID items: %w", err)
	}
	return txn, nil
}

// Void marks a transaction voided and restores stock atomically.
func (r *TransactionRepo) Void(ctx context.Context, txnID, userID string) error {
	// Load items first (outside tx — read-only)
	const itemQ = `SELECT product_id, quantity FROM transaction_items WHERE transaction_id = $1`

	type itemRow struct {
		ProductID *string `db:"product_id"`
		Quantity  float64 `db:"quantity"`
	}
	var items []itemRow
	if err := r.db.SelectContext(ctx, &items, itemQ, txnID); err != nil {
		return fmt.Errorf("TransactionRepo.Void load items: %w", err)
	}

	// Load store_id
	var storeID string
	if err := r.db.QueryRowxContext(ctx, `SELECT store_id FROM transactions WHERE id = $1`, txnID).Scan(&storeID); err != nil {
		return fmt.Errorf("TransactionRepo.Void load store: %w", err)
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("TransactionRepo.Void begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Update transaction status
	if _, err := tx.ExecContext(ctx,
		`UPDATE transactions SET status='voided', updated_at=NOW() WHERE id=$1`, txnID,
	); err != nil {
		return fmt.Errorf("TransactionRepo.Void status: %w", err)
	}

	// Reverse stock for each item
	const mvQ = `
		INSERT INTO stock_movements (product_id, store_id, ref_type, ref_id, quantity_delta, notes, created_by)
		VALUES ($1,$2,'void',$3,$4,'Transaction voided - stock restored',$5)`
	const slQ = `
		INSERT INTO stock_levels (product_id, store_id, quantity, updated_at)
		VALUES ($1,$2,$3,NOW())
		ON CONFLICT (product_id, store_id)
		DO UPDATE SET quantity = stock_levels.quantity + $3, updated_at = NOW()`

	for _, item := range items {
		if item.ProductID == nil {
			continue
		}
		if _, err := tx.ExecContext(ctx, mvQ, *item.ProductID, storeID, txnID, item.Quantity, userID); err != nil {
			return fmt.Errorf("TransactionRepo.Void movement: %w", err)
		}
		if _, err := tx.ExecContext(ctx, slQ, *item.ProductID, storeID, item.Quantity); err != nil {
			return fmt.Errorf("TransactionRepo.Void stock: %w", err)
		}
	}

	return tx.Commit()
}
