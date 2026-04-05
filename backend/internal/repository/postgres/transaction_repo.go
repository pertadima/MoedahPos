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

// Create persists a full transaction with items and (for completed orders) stock movements.
// Set input.Status = "draft" to hold an order without deducting stock.
func (r *TransactionRepo) Create(ctx context.Context, input domain.CreateTransactionInput) (*domain.Transaction, error) { //nolint:cyclop,funlen // transaction creation is inherently complex
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("TransactionRepo.Create begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	status := input.Status
	if status == "" {
		status = "completed"
	}

	// 1. INSERT transaction header (table_id may be nil for retail)
	const txnQ = `
		INSERT INTO transactions
		  (store_id, cashier_id, table_id, customer_name, customer_phone,
		   subtotal, discount_amt, tax_amt, total,
		   payment_method, payment_amount, change_amount, status, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		RETURNING id, store_id, cashier_id, table_id,
		          COALESCE(customer_name,'')  AS customer_name,
		          COALESCE(customer_phone,'') AS customer_phone,
		          subtotal, discount_amt, tax_amt, total,
		          payment_method, payment_amount, change_amount, status,
		          COALESCE(notes,'') AS notes,
		          created_at, updated_at`

	txn := &domain.Transaction{}
	err = tx.QueryRowxContext(ctx, txnQ,
		input.StoreID, input.CashierID, input.TableID,
		input.CustomerName, input.CustomerPhone,
		input.Subtotal, input.DiscountAmt, input.TaxAmt, input.Total,
		input.PaymentMethod, input.PaymentAmount, input.ChangeAmount, status, input.Notes,
	).StructScan(txn)
	if err != nil {
		return nil, fmt.Errorf("TransactionRepo.Create insert txn: %w", err)
	}

	// 2. INSERT items and (if completed) stock movements
	const itemQ = `
		INSERT INTO transaction_items
		  (transaction_id, product_id, menu_item_id, product_name, sku, quantity, unit_price, cost_price, discount_pct, tax_rate, subtotal, status)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,'pending')
		RETURNING id, transaction_id, product_id, menu_item_id, product_name, sku, quantity, unit_price, cost_price, discount_pct, tax_rate, subtotal, status, completed_at`

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
			txn.ID, item.ProductID, item.MenuItemID, item.ProductName, item.SKU,
			item.Quantity, item.UnitPrice, item.CostPrice, item.DiscountPct, item.TaxRate, item.Subtotal,
		).StructScan(ti); err != nil {
			return nil, fmt.Errorf("TransactionRepo.Create insert item: %w", err)
		}
		txn.Items = append(txn.Items, *ti)

		// Only deduct stock for completed (paid) orders
		if status == "completed" && item.ProductID != nil {
			if _, err := tx.ExecContext(ctx, mvQ,
				*item.ProductID, input.StoreID, txn.ID, -item.Quantity, input.CashierID,
			); err != nil {
				return nil, fmt.Errorf("TransactionRepo.Create stock movement: %w", err)
			}
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

// GetDraftByTable returns the open (draft) transaction for a given table, or nil if none.
func (r *TransactionRepo) GetDraftByTable(ctx context.Context, storeID, tableID string) (*domain.Transaction, error) {
	const q = `
		SELECT t.id, t.store_id, t.cashier_id, t.table_id,
		       COALESCE(t.customer_name,'')  AS customer_name,
		       COALESCE(t.customer_phone,'') AS customer_phone,
		       t.subtotal, t.discount_amt, t.tax_amt, t.total,
		       t.payment_method, t.payment_amount, t.change_amount, t.status,
		       COALESCE(t.notes,'') AS notes,
		       t.created_at, t.updated_at, u.name AS cashier_name
		FROM transactions t
		JOIN users u ON u.id = t.cashier_id
		WHERE t.store_id = $1 AND t.table_id = $2 AND t.status = 'draft'
		ORDER BY t.created_at DESC
		LIMIT 1`

	txn := &domain.Transaction{}
	if err := r.db.QueryRowxContext(ctx, q, storeID, tableID).StructScan(txn); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("TransactionRepo.GetDraftByTable: %w", err)
	}

	const itemQ = `
		SELECT id, transaction_id, product_id, menu_item_id, product_name, sku,
		       quantity, unit_price, cost_price, discount_pct, tax_rate, subtotal, status, completed_at
		FROM transaction_items WHERE transaction_id = $1`

	if err := r.db.SelectContext(ctx, &txn.Items, itemQ, txn.ID); err != nil {
		return nil, fmt.Errorf("TransactionRepo.GetDraftByTable items: %w", err)
	}
	return txn, nil
}

// UpdateDraftItems replaces all items on a draft transaction and recalculates totals.
func (r *TransactionRepo) UpdateDraftItems(ctx context.Context, txnID string, items []domain.CreateTransactionItemInput,
	subtotal, discountAmt, taxAmt, total float64, customerName, notes string) (*domain.Transaction, error) {

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("TransactionRepo.UpdateDraftItems begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Fetch existing items for this draft to calculate diffs
	var existingItems []domain.TransactionItem
	if err := tx.SelectContext(ctx, &existingItems, `SELECT id, product_id, menu_item_id, quantity, status, unit_price, discount_pct, tax_rate FROM transaction_items WHERE transaction_id = $1`, txnID); err != nil {
		return nil, fmt.Errorf("TransactionRepo.UpdateDraftItems fetch existing: %w", err)
	}

	type ItemKey struct {
		IsMenu bool
		ID     string
	}

	getKey := func(productID, menuID *string) ItemKey {
		if menuID != nil {
			return ItemKey{IsMenu: true, ID: *menuID}
		}
		if productID != nil {
			return ItemKey{IsMenu: false, ID: *productID}
		}
		return ItemKey{} // should not happen
	}

	// incoming quantities map
	incomingQty := make(map[ItemKey]domain.CreateTransactionItemInput)
	for _, it := range items {
		k := getKey(it.ProductID, it.MenuItemID)
		if existing, ok := incomingQty[k]; ok {
			existing.Quantity += it.Quantity
			incomingQty[k] = existing
		} else {
			incomingQty[k] = it
		}
	}

	// existing items grouped by key
	existingGroups := make(map[ItemKey][]domain.TransactionItem)
	for _, it := range existingItems {
		k := getKey(it.ProductID, it.MenuItemID)
		existingGroups[k] = append(existingGroups[k], it)
	}

	// Diffing map incoming against existing
	for k, inItem := range incomingQty {
		exItems := existingGroups[k]
		var exTotal float64
		for _, e := range exItems {
			exTotal += e.Quantity
		}

		diff := inItem.Quantity - exTotal
		if diff > 0 {
			const itemQ = `
				INSERT INTO transaction_items
				  (transaction_id, product_id, menu_item_id, product_name, sku, quantity, unit_price, cost_price, discount_pct, tax_rate, subtotal, status)
				VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,'pending')`
			
			unitNet := inItem.UnitPrice * (1 - inItem.DiscountPct/100)
			diffTax := (unitNet * diff) * inItem.TaxRate / 100
			diffSubtotal := (unitNet * diff) + diffTax

			if _, err := tx.ExecContext(ctx, itemQ,
				txnID, inItem.ProductID, inItem.MenuItemID, inItem.ProductName, inItem.SKU,
				diff, inItem.UnitPrice, inItem.CostPrice, inItem.DiscountPct, inItem.TaxRate, diffSubtotal,
			); err != nil {
				return nil, fmt.Errorf("TransactionRepo.UpdateDraftItems insert diff: %w", err)
			}
		} else if diff < 0 {
			shortfall := -diff
			// Pass 0: pending items, Pass 1: completed items
			for pass := 0; pass < 2; pass++ {
				for _, e := range exItems {
					if shortfall <= 0 {
						break
					}
					isPending := e.Status == "pending"
					if (pass == 0 && !isPending) || (pass == 1 && isPending) {
						continue
					}
					
					if e.Quantity <= shortfall {
						if _, err := tx.ExecContext(ctx, `DELETE FROM transaction_items WHERE id = $1`, e.ID); err != nil {
							return nil, fmt.Errorf("TransactionRepo.UpdateDraftItems delete item: %w", err)
						}
						shortfall -= e.Quantity
					} else {
						newQty := e.Quantity - shortfall
						unitNet := e.UnitPrice * (1 - e.DiscountPct/100)
						newSubtotal := (unitNet * newQty) * (1 + e.TaxRate/100)
						if _, err := tx.ExecContext(ctx, `UPDATE transaction_items SET quantity = $1, subtotal = $2 WHERE id = $3`, newQty, newSubtotal, e.ID); err != nil {
							return nil, fmt.Errorf("TransactionRepo.UpdateDraftItems update item qty: %w", err)
						}
						shortfall = 0
					}
				}
			}
		}
	}

	// Identify completely removed items
	for k, exItems := range existingGroups {
		if _, ok := incomingQty[k]; !ok {
			for _, e := range exItems {
				if _, err := tx.ExecContext(ctx, `DELETE FROM transaction_items WHERE id = $1`, e.ID); err != nil {
					return nil, fmt.Errorf("TransactionRepo.UpdateDraftItems delete unlisted item: %w", err)
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("TransactionRepo.UpdateDraftItems commit: %w", err)
	}

	return r.FindByID(ctx, txnID)
}

// PayDraft finalizes a held order: sets payment details, completes it, and deducts stock.
func (r *TransactionRepo) PayDraft(ctx context.Context, input domain.PayDraftInput, storeID, cashierID string) (*domain.Transaction, error) {
	// Load items first (to deduct stock)
	const itemQ = `SELECT product_id, quantity FROM transaction_items WHERE transaction_id = $1`
	type itemRow struct {
		ProductID *string `db:"product_id"`
		Quantity  float64 `db:"quantity"`
	}
	var existing []itemRow
	if err := r.db.SelectContext(ctx, &existing, itemQ, input.TransactionID); err != nil {
		return nil, fmt.Errorf("TransactionRepo.PayDraft load items: %w", err)
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("TransactionRepo.PayDraft begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Update header: set payment info + mark completed
	const updQ = `
		UPDATE transactions
		SET payment_method=$2, payment_amount=$3, change_amount=$4,
		    customer_name=COALESCE(NULLIF($5,''), customer_name),
		    customer_phone=COALESCE(NULLIF($6,''), customer_phone),
		    status='completed', table_id=NULL, updated_at=NOW()
		WHERE id=$1 AND status='draft'`
	res, err := tx.ExecContext(ctx, updQ,
		input.TransactionID, input.PaymentMethod, input.PaymentAmount, input.ChangeAmount,
		input.CustomerName, input.CustomerPhone,
	)
	if err != nil {
		return nil, fmt.Errorf("TransactionRepo.PayDraft update: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, fmt.Errorf("draft transaction not found or already completed")
	}

	// Deduct stock for each product item
	const mvQ = `
		INSERT INTO stock_movements (product_id, store_id, ref_type, ref_id, quantity_delta, notes, created_by)
		VALUES ($1,$2,'sale',$3,$4,'Sale transaction',$5)`
	const slQ = `
		INSERT INTO stock_levels (product_id, store_id, quantity, updated_at)
		VALUES ($1,$2,$3,NOW())
		ON CONFLICT (product_id, store_id)
		DO UPDATE SET quantity = stock_levels.quantity + $3, updated_at = NOW()`

	for _, item := range existing {
		if item.ProductID == nil {
			continue
		}
		if _, err := tx.ExecContext(ctx, mvQ, *item.ProductID, storeID, input.TransactionID, -item.Quantity, cashierID); err != nil {
			return nil, fmt.Errorf("TransactionRepo.PayDraft movement: %w", err)
		}
		if _, err := tx.ExecContext(ctx, slQ, *item.ProductID, storeID, -item.Quantity); err != nil {
			return nil, fmt.Errorf("TransactionRepo.PayDraft stock level: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("TransactionRepo.PayDraft commit: %w", err)
	}

	return r.FindByID(ctx, input.TransactionID)
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
	if f.TableID != "" {
		conds = append(conds, fmt.Sprintf("t.table_id = $%d", i))
		args = append(args, f.TableID)
		i++
	}
	if f.DateFrom != "" {
		conds = append(conds, fmt.Sprintf("t.created_at >= $%d::date", i))
		args = append(args, f.DateFrom)
		i++
	}
	if f.DateTo != "" {
		conds = append(conds, fmt.Sprintf("t.created_at < ($%d::date + INTERVAL '1 day')", i))
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
		SELECT t.id, t.store_id, t.cashier_id, t.table_id,
		       COALESCE(t.customer_name,'')  AS customer_name,
		       COALESCE(t.customer_phone,'') AS customer_phone,
		       t.subtotal, t.discount_amt, t.tax_amt, t.total,
		       t.payment_method, t.payment_amount, t.change_amount, t.status,
		       COALESCE(t.notes,'') AS notes,
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
		SELECT t.id, t.store_id, t.cashier_id, t.table_id,
		       COALESCE(t.customer_name,'')  AS customer_name,
		       COALESCE(t.customer_phone,'') AS customer_phone,
		       t.subtotal, t.discount_amt, t.tax_amt, t.total,
		       t.payment_method, t.payment_amount, t.change_amount, t.status,
		       COALESCE(t.notes,'') AS notes,
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
		SELECT id, transaction_id, product_id, menu_item_id, product_name, sku,
		       quantity, unit_price, cost_price, discount_pct, tax_rate, subtotal, status, completed_at
		FROM transaction_items WHERE transaction_id = $1`

	if err := r.db.SelectContext(ctx, &txn.Items, itemQ, id); err != nil {
		return nil, fmt.Errorf("TransactionRepo.FindByID items: %w", err)
	}
	return txn, nil
}

// Void marks a transaction voided and restores stock atomically.
func (r *TransactionRepo) Void(ctx context.Context, txnID, userID string) error {
	const itemQ = `SELECT product_id, quantity FROM transaction_items WHERE transaction_id = $1`

	type itemRow struct {
		ProductID *string `db:"product_id"`
		Quantity  float64 `db:"quantity"`
	}
	var items []itemRow
	if err := r.db.SelectContext(ctx, &items, itemQ, txnID); err != nil {
		return fmt.Errorf("TransactionRepo.Void load items: %w", err)
	}

	var storeID string
	if err := r.db.QueryRowxContext(ctx, `SELECT store_id FROM transactions WHERE id = $1`, txnID).Scan(&storeID); err != nil {
		return fmt.Errorf("TransactionRepo.Void load store: %w", err)
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("TransactionRepo.Void begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx,
		`UPDATE transactions SET status='voided', updated_at=NOW() WHERE id=$1`, txnID,
	); err != nil {
		return fmt.Errorf("TransactionRepo.Void status: %w", err)
	}

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

// GetKDSTickets returns all transactions (draft or completed) for a restaurant that still have pending items.
func (r *TransactionRepo) GetKDSTickets(ctx context.Context, storeID string) ([]*domain.Transaction, error) {
	const q = `
		SELECT t.id, t.store_id, t.cashier_id, t.table_id,
		       rt.table_number,
		       COALESCE(t.customer_name,'')  AS customer_name,
		       COALESCE(t.customer_phone,'') AS customer_phone,
		       t.subtotal, t.discount_amt, t.tax_amt, t.total,
		       t.payment_method, t.payment_amount, t.change_amount, t.status,
		       COALESCE(t.notes,'') AS notes,
		       t.created_at, t.updated_at, u.name AS cashier_name
		FROM transactions t
		JOIN users u ON u.id = t.cashier_id
		LEFT JOIN restaurant_tables rt ON t.table_id = rt.id
		WHERE t.store_id = $1 AND (
			t.status = 'draft' OR EXISTS (
				SELECT 1 FROM transaction_items ti WHERE ti.transaction_id = t.id AND ti.status = 'pending'
			)
		)
		ORDER BY t.created_at ASC`

	var txns []*domain.Transaction
	if err := r.db.SelectContext(ctx, &txns, q, storeID); err != nil {
		return nil, fmt.Errorf("TransactionRepo.GetKDSTickets headers: %w", err)
	}

	if len(txns) == 0 {
		return txns, nil
	}

	var txnIDs []string
	for _, t := range txns {
		txnIDs = append(txnIDs, t.ID)
	}

	qItems, args, err := sqlx.In(`
		SELECT id, transaction_id, product_id, menu_item_id, product_name, sku,
		       quantity, unit_price, cost_price, discount_pct, tax_rate, subtotal, status, completed_at
		FROM transaction_items WHERE transaction_id IN (?)
		ORDER BY id ASC`, txnIDs)
	if err != nil {
		return nil, fmt.Errorf("TransactionRepo.GetKDSTickets in query: %w", err)
	}
	qItems = r.db.Rebind(qItems)

	var allItems []domain.TransactionItem
	if err := r.db.SelectContext(ctx, &allItems, qItems, args...); err != nil {
		return nil, fmt.Errorf("TransactionRepo.GetKDSTickets items: %w", err)
	}

	itemMap := make(map[string][]domain.TransactionItem)
	for _, it := range allItems {
		itemMap[it.TransactionID] = append(itemMap[it.TransactionID], it)
	}

	for _, t := range txns {
		t.Items = itemMap[t.ID]
	}

	return txns, nil
}

// UpdateKDSItemStatus updates the completion status of a specific KDS ticket item.
func (r *TransactionRepo) UpdateKDSItemStatus(ctx context.Context, itemID, status string) error {
	if status == "completed" {
		_, err := r.db.ExecContext(ctx, `UPDATE transaction_items SET status = $1, completed_at = NOW() WHERE id = $2`, status, itemID)
		return err
	}
	_, err := r.db.ExecContext(ctx, `UPDATE transaction_items SET status = $1, completed_at = NULL WHERE id = $2`, status, itemID)
	return err
}
