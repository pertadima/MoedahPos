package postgres

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/moedahpos/backend/internal/domain"
)

const (
	stockTestStoreID = "store-1"
	testUserID       = "user-1"
)

func TestStockAdjustmentRepo_CreateAdjustment_IN(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := NewStockAdjustmentRepository(db)

	storeID := stockTestStoreID
	userID := testUserID
	input := domain.CreateAdjustmentInput{
		ProductID: "p1",
		Type:      "IN",
		Reason:    "MANUAL_CORRECTION",
		Quantity:  10.0,
		Notes:     "Correction",
	}

	t.Run("success_IN", func(t *testing.T) {
		mock.ExpectBegin()

		// 1. Insert header
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO stock_adjustments (id, product_id, store_id, type, reason, quantity, notes, created_by) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)")).
			WithArgs(sqlmock.AnyArg(), input.ProductID, storeID, input.Type, input.Reason, input.Quantity, input.Notes, userID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// 2. Find recent price
		mock.ExpectQuery(regexp.QuoteMeta("SELECT purchase_price FROM stock_batches WHERE product_id = $1 AND store_id = $2 ORDER BY received_at DESC LIMIT 1")).
			WithArgs(input.ProductID, storeID).
			WillReturnRows(sqlmock.NewRows([]string{"purchase_price"}).AddRow(1000.0))

		// 3. Insert batch
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO stock_batches (id, product_id, store_id, quantity_remaining, purchase_price, received_at) VALUES ($1, $2, $3, $4, $5, NOW())")).
			WithArgs(sqlmock.AnyArg(), input.ProductID, storeID, input.Quantity, 1000.0).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// 4. Insert junction
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO stock_adjustment_batches (id, adjustment_id, batch_id, deducted_qty) VALUES ($1, $2, $3, $4)")).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), input.Quantity).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// 5. Update stock level
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO stock_levels (id, product_id, store_id, quantity, min_quantity) VALUES ($1, $2, $3, $4, 0) ON CONFLICT (product_id, store_id) DO UPDATE SET quantity = stock_levels.quantity + $4")).
			WithArgs(sqlmock.AnyArg(), input.ProductID, storeID, input.Quantity).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// 6. Insert movement
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO stock_movements (id, product_id, store_id, ref_type, ref_id, quantity_delta, notes, created_by) VALUES ($1, $2, $3, 'adjustment', $4, $5, $6, $7)")).
			WithArgs(sqlmock.AnyArg(), input.ProductID, storeID, sqlmock.AnyArg(), input.Quantity, input.Notes, userID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		err := repo.CreateAdjustment(context.Background(), storeID, userID, input)
		assert.NoError(t, err)
	})
}

func TestStockAdjustmentRepo_CreateAdjustment_OUT(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := NewStockAdjustmentRepository(db)

	storeID := stockTestStoreID
	userID := testUserID
	input := domain.CreateAdjustmentInput{
		ProductID: "p1",
		Type:      "OUT",
		Reason:    "DAMAGED",
		Quantity:  5.0,
		Notes:     "Damaged item",
	}

	t.Run("success_OUT", func(t *testing.T) {
		mock.ExpectBegin()

		// 1. Insert header
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO stock_adjustments (id, product_id, store_id, type, reason, quantity, notes, created_by) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)")).
			WithArgs(sqlmock.AnyArg(), input.ProductID, storeID, input.Type, input.Reason, input.Quantity, input.Notes, userID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// 2. Select batches for UPDATE
		rows := sqlmock.NewRows([]string{"id", "quantity_remaining"}).
			AddRow("b1", 10.0)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, quantity_remaining FROM stock_batches WHERE product_id = $1 AND store_id = $2 AND quantity_remaining > 0 ORDER BY received_at ASC FOR UPDATE")).
			WithArgs(input.ProductID, storeID).
			WillReturnRows(rows)

		// 3. Update batch
		mock.ExpectExec(regexp.QuoteMeta("UPDATE stock_batches SET quantity_remaining = $1 WHERE id = $2")).
			WithArgs(5.0, "b1").
			WillReturnResult(sqlmock.NewResult(1, 1))

		// 4. Insert junction tracking
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO stock_adjustment_batches (id, adjustment_id, batch_id, deducted_qty) VALUES ($1, $2, $3, $4)")).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "b1", 5.0).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// 5. Update stock level
		mock.ExpectExec(regexp.QuoteMeta("UPDATE stock_levels SET quantity = quantity - $1 WHERE product_id = $2 AND store_id = $3")).
			WithArgs(input.Quantity, input.ProductID, storeID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// 6. Insert movement
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO stock_movements (id, product_id, store_id, ref_type, ref_id, quantity_delta, notes, created_by) VALUES ($1, $2, $3, 'adjustment', $4, $5, $6, $7)")).
			WithArgs(sqlmock.AnyArg(), input.ProductID, storeID, sqlmock.AnyArg(), -input.Quantity, input.Notes, userID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		err := repo.CreateAdjustment(context.Background(), storeID, userID, input)
		assert.NoError(t, err)
	})

	t.Run("insufficient_stock", func(t *testing.T) {
		mock.ExpectBegin()

		// 1. Insert header
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO stock_adjustments (id, product_id, store_id, type, reason, quantity, notes, created_by) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)")).
			WithArgs(sqlmock.AnyArg(), input.ProductID, storeID, input.Type, input.Reason, input.Quantity, input.Notes, userID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// 2. Select batches for UPDATE (low stock)
		rows := sqlmock.NewRows([]string{"id", "quantity_remaining"}).
			AddRow("b1", 2.0)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, quantity_remaining FROM stock_batches WHERE product_id = $1 AND store_id = $2 AND quantity_remaining > 0 ORDER BY received_at ASC FOR UPDATE")).
			WithArgs(input.ProductID, storeID).
			WillReturnRows(rows)

		mock.ExpectRollback()

		err := repo.CreateAdjustment(context.Background(), storeID, userID, input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient stock")
	})
}

func TestStockAdjustmentRepo_History(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	repo := NewStockAdjustmentRepository(db)
	ctx := context.Background()

	t.Run("success_all", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "product_id", "product_name"}).
			AddRow("adj1", "p1", "Product 1")
		mock.ExpectQuery(`SELECT .* FROM stock_adjustments a`).WithArgs("st1").WillReturnRows(rows)

		res, err := repo.GetStockAdjustmentHistory(ctx, "st1", nil)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("success_with_product", func(t *testing.T) {
		productID := "p1"
		rows := sqlmock.NewRows([]string{"id", "product_id"}).AddRow("adj1", "p1")
		mock.ExpectQuery(`SELECT .* FROM stock_adjustments a .* WHERE a.store_id = \$1 AND a.product_id = \$2`).
			WithArgs("st1", productID).
			WillReturnRows(rows)

		res, err := repo.GetStockAdjustmentHistory(ctx, "st1", &productID)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})
}
