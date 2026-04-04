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

// PORepo is the PostgreSQL implementation of repository.PurchaseOrderRepository.
type PORepo struct{ db *sqlx.DB }

func NewPORepo(db *sqlx.DB) *PORepo { return &PORepo{db: db} }

func (r *PORepo) Create(ctx context.Context, po *domain.PurchaseOrder, items []domain.POItem) (*domain.PurchaseOrder, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("PORepo.Create begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	const poQ = `
		INSERT INTO purchase_orders (store_id, supplier_id, po_number, status, total_amount, ordered_by, notes)
		VALUES ($1,$2,$3,'draft',$4,$5,$6)
		RETURNING id, store_id, supplier_id, po_number, status, total_amount,
		          ordered_by, received_by, ordered_at, received_at, notes, created_at, updated_at`
	result := &domain.PurchaseOrder{}
	if err := tx.QueryRowxContext(ctx, poQ,
		po.StoreID, po.SupplierID, po.PONumber, po.TotalAmount, po.OrderedBy, po.Notes,
	).StructScan(result); err != nil {
		return nil, fmt.Errorf("PORepo.Create insert po: %w", err)
	}

	result.Items, err = r.insertItems(ctx, tx, result.ID, items)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("PORepo.Create commit: %w", err)
	}
	r.populateJoins(ctx, result)
	return result, nil
}

func (r *PORepo) insertItems(ctx context.Context, tx *sqlx.Tx, poID string, items []domain.POItem) ([]domain.POItem, error) {
	const q = `
		INSERT INTO purchase_order_items (po_id, product_id, quantity, unit_cost, subtotal)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id, po_id, product_id, quantity, unit_cost, received_qty, subtotal`
	var result []domain.POItem
	for _, item := range items {
		row := domain.POItem{}
		if err := tx.QueryRowxContext(ctx, q, poID, item.ProductID, item.Quantity, item.UnitCost, item.Subtotal).StructScan(&row); err != nil {
			return nil, fmt.Errorf("PORepo.insertItems: %w", err)
		}
		result = append(result, row)
	}
	return result, nil
}

func (r *PORepo) FindAll(ctx context.Context, f dto.POListFilter) ([]*domain.PurchaseOrder, int, error) {
	args := []interface{}{f.StoreID}
	conds := []string{"po.store_id = $1"}
	i := 2
	if f.Status != "" {
		conds = append(conds, fmt.Sprintf("po.status = $%d", i))
		args = append(args, f.Status)
		i++
	}
	if f.DateFrom != "" {
		conds = append(conds, fmt.Sprintf("po.created_at >= $%d", i))
		args = append(args, f.DateFrom)
		i++
	}
	if f.DateTo != "" {
		conds = append(conds, fmt.Sprintf("po.created_at < $%d", i))
		args = append(args, f.DateTo)
		i++
	}
	where := "WHERE " + strings.Join(conds, " AND ")

	var total int
	if err := r.db.QueryRowxContext(ctx,
		fmt.Sprintf("SELECT COUNT(*) FROM purchase_orders po %s", where), args...,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("PORepo.FindAll count: %w", err)
	}

	args = append(args, f.PerPage, f.Offset())
	dataQ := fmt.Sprintf(`
		WITH termin_agg AS (
			SELECT t.po_id,
			       min(t.due_date) FILTER (WHERE t.status != 'paid') AS next_deadline,
			       SUM(t.amount - COALESCE(pr.amount_paid, 0)) AS amount_due,
			       SUM(COALESCE(pr.amount_paid, 0)) AS amount_paid
			FROM purchase_order_termins t
			LEFT JOIN (
				SELECT termin_id, SUM(amount_paid) AS amount_paid
				FROM payment_records
				GROUP BY termin_id
			) pr ON pr.termin_id = t.id
			GROUP BY t.po_id
		)
		SELECT po.id, po.store_id, po.supplier_id, po.po_number, po.status, po.total_amount,
		       po.ordered_by, po.received_by, po.ordered_at, po.received_at, po.notes,
		       po.created_at, po.updated_at,
		       s.name AS supplier_name,
		       u.name AS ordered_by_name,
		       ru.name AS received_by_name,
		       ta.next_deadline AS next_deadline,
		       COALESCE(ta.amount_due, po.total_amount) AS amount_due,
		       COALESCE(ta.amount_paid, 0) AS amount_paid,
		       (SELECT COUNT(*) FROM purchase_order_items WHERE po_id=po.id) AS total_items
		FROM purchase_orders po
		JOIN users u ON u.id = po.ordered_by
		LEFT JOIN suppliers s  ON s.id  = po.supplier_id
		LEFT JOIN users     ru ON ru.id = po.received_by
		LEFT JOIN termin_agg ta ON ta.po_id = po.id
		%s ORDER BY ta.next_deadline ASC NULLS LAST, po.created_at DESC LIMIT $%d OFFSET $%d`, where, i, i+1)

	var pos []*domain.PurchaseOrder
	if err := r.db.SelectContext(ctx, &pos, dataQ, args...); err != nil {
		return nil, 0, fmt.Errorf("PORepo.FindAll: %w", err)
	}
	return pos, total, nil
}

func (r *PORepo) FindByID(ctx context.Context, id string) (*domain.PurchaseOrder, error) {
	const poQ = `
		SELECT po.id, po.store_id, po.supplier_id, po.po_number, po.status, po.total_amount,
		       po.ordered_by, po.received_by, po.ordered_at, po.received_at, po.notes,
		       po.created_at, po.updated_at,
		       s.name AS supplier_name, u.name AS ordered_by_name, ru.name AS received_by_name,
		       (SELECT COUNT(*) FROM purchase_order_items WHERE po_id=po.id) AS total_items
		FROM purchase_orders po
		JOIN users u ON u.id = po.ordered_by
		LEFT JOIN suppliers s  ON s.id  = po.supplier_id
		LEFT JOIN users     ru ON ru.id = po.received_by
		WHERE po.id = $1`

	po := &domain.PurchaseOrder{}
	if err := r.db.QueryRowxContext(ctx, poQ, id).StructScan(po); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("PORepo.FindByID: %w", err)
	}

	const itemQ = `
		SELECT poi.id, poi.po_id, poi.product_id, poi.quantity, poi.unit_cost, poi.received_qty, poi.subtotal,
		       p.name AS product_name, p.sku AS product_sku, p.unit
		FROM purchase_order_items poi
		JOIN products p ON p.id = poi.product_id
		WHERE poi.po_id = $1`
	if err := r.db.SelectContext(ctx, &po.Items, itemQ, id); err != nil {
		return nil, fmt.Errorf("PORepo.FindByID items: %w", err)
	}
	return po, nil
}

func (r *PORepo) Update(ctx context.Context, po *domain.PurchaseOrder, items []domain.POItem) (*domain.PurchaseOrder, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("PORepo.Update begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	const q = `
		UPDATE purchase_orders SET supplier_id=$1, notes=$2, total_amount=$3, updated_at=NOW()
		WHERE id=$4 AND status='draft'
		RETURNING id, store_id, supplier_id, po_number, status, total_amount,
		          ordered_by, received_by, ordered_at, received_at, notes, created_at, updated_at`
	result := &domain.PurchaseOrder{}
	if err := tx.QueryRowxContext(ctx, q, po.SupplierID, po.Notes, po.TotalAmount, po.ID).StructScan(result); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("PORepo.Update: %w", err)
	}

	// Replace items
	if _, err := tx.ExecContext(ctx, `DELETE FROM purchase_order_items WHERE po_id = $1`, po.ID); err != nil {
		return nil, fmt.Errorf("PORepo.Update delete items: %w", err)
	}
	result.Items, err = r.insertItems(ctx, tx, po.ID, items)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("PORepo.Update commit: %w", err)
	}
	r.populateJoins(ctx, result)
	return result, nil
}

func (r *PORepo) Submit(ctx context.Context, poID, _ string) error {
	const q = `
		UPDATE purchase_orders
		SET status='ordered', ordered_at=NOW(), updated_at=NOW()
		WHERE id=$1 AND status='draft'`
	res, err := r.db.ExecContext(ctx, q, poID)
	if err != nil {
		return fmt.Errorf("PORepo.Submit: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("PO not found or not in draft status")
	}
	return nil
}

// Receive marks a PO as received, updates stock levels and product cost prices atomically.
func (r *PORepo) Receive(ctx context.Context, poID, userID string) error { //nolint:cyclop // PO receiving is inherently complex
	// Load items (now inc. unit_cost)
	type itemRow struct {
		ProductID string  `db:"product_id"`
		Quantity  float64 `db:"quantity"`
		UnitCost  float64 `db:"unit_cost"`
	}
	var items []itemRow
	if err := r.db.SelectContext(ctx, &items,
		`SELECT product_id, quantity, unit_cost FROM purchase_order_items WHERE po_id = $1`, poID,
	); err != nil {
		return fmt.Errorf("PORepo.Receive load items: %w", err)
	}

	var storeID string
	if err := r.db.QueryRowxContext(ctx, `SELECT store_id FROM purchase_orders WHERE id = $1`, poID).Scan(&storeID); err != nil {
		return fmt.Errorf("PORepo.Receive load store: %w", err)
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("PORepo.Receive begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Update PO status
	res, err := tx.ExecContext(ctx, `
		UPDATE purchase_orders
		SET status='received', received_by=$1, received_at=NOW(), updated_at=NOW()
		WHERE id=$2 AND status='ordered'`, userID, poID)
	if err != nil {
		return fmt.Errorf("PORepo.Receive update po: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("PO not found or not in ordered status")
	}

	// Update received_qty on items
	if _, err := tx.ExecContext(ctx,
		`UPDATE purchase_order_items SET received_qty = quantity WHERE po_id = $1`, poID,
	); err != nil {
		return fmt.Errorf("PORepo.Receive update items: %w", err)
	}

	const mvQ = `
		INSERT INTO stock_movements (product_id, store_id, ref_type, ref_id, quantity_delta, notes, created_by)
		VALUES ($1,$2,'purchase',$3,$4,'Purchase order received',$5)`
	const slQ = `
		INSERT INTO stock_levels (product_id, store_id, quantity, updated_at)
		VALUES ($1,$2,$3,NOW())
		ON CONFLICT (product_id, store_id)
		DO UPDATE SET quantity = stock_levels.quantity + $3, updated_at = NOW()`
	// Update product cost_price to the received unit_cost (Last Cost method)
	const cpQ = `
		UPDATE products SET cost_price = $1, updated_at = NOW()
		WHERE id = $2`
	// FIFO batch insert: each PO item creates one batch record with received_at=NOW().
	// received_at determines FIFO position — items received first are deducted first.
	const batchQ = `
		INSERT INTO stock_batches
			(product_id, store_id, po_id, quantity_remaining, purchase_price, received_at)
		VALUES ($1, $2, $3, $4, $5, NOW())`

	for _, item := range items {
		if _, err := tx.ExecContext(ctx, mvQ, item.ProductID, storeID, poID, item.Quantity, userID); err != nil {
			return fmt.Errorf("PORepo.Receive movement: %w", err)
		}
		if _, err := tx.ExecContext(ctx, slQ, item.ProductID, storeID, item.Quantity); err != nil {
			return fmt.Errorf("PORepo.Receive stock level: %w", err)
		}
		// Always update cost_price to the actual purchase cost
		if _, err := tx.ExecContext(ctx, cpQ, item.UnitCost, item.ProductID); err != nil {
			return fmt.Errorf("PORepo.Receive cost price: %w", err)
		}
		// Create FIFO batch — committed atomically with the rest of the receive operation.
		if _, err := tx.ExecContext(ctx, batchQ, item.ProductID, storeID, poID, item.Quantity, item.UnitCost); err != nil {
			return fmt.Errorf("PORepo.Receive create batch: %w", err)
		}
	}

	return tx.Commit()
}

func (r *PORepo) Cancel(ctx context.Context, poID string) error {
	// dbCancelled matches the DB enum value (British spelling kept for DB compat).
	const dbCancelled = "cancelled" //nolint:misspell // DB enum value
	q := "UPDATE purchase_orders SET status='" + dbCancelled + "', updated_at=NOW()" +
		" WHERE id=$1 AND status IN ('draft','ordered')"
	res, err := r.db.ExecContext(ctx, q, poID)
	if err != nil {
		return fmt.Errorf("PORepo.Cancel: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("PO not found or cannot be canceled")
	}
	return nil
}

func (r *PORepo) populateJoins(ctx context.Context, po *domain.PurchaseOrder) {
	type joinRow struct {
		OrderedByName  string  `db:"ordered_by_name"`
		SupplierName   *string `db:"supplier_name"`
		ReceivedByName *string `db:"received_by_name"`
	}
	var row joinRow
	_ = r.db.QueryRowxContext(ctx, `
		SELECT u.name AS ordered_by_name, s.name AS supplier_name, ru.name AS received_by_name
		FROM purchase_orders po
		JOIN users u ON u.id = po.ordered_by
		LEFT JOIN suppliers s  ON s.id  = po.supplier_id
		LEFT JOIN users     ru ON ru.id = po.received_by
		WHERE po.id = $1`, po.ID,
	).StructScan(&row)
	po.OrderedByName = row.OrderedByName
	po.SupplierName = row.SupplierName
	po.ReceivedByName = row.ReceivedByName
}
