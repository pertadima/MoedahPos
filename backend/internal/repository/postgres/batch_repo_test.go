package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
)

func TestBatchRepo_Lifecycle(t *testing.T) {
	ctx := context.Background()

	t.Run("CreateBatch", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()
		repo := NewBatchRepository(db)

		batch := &domain.StockBatch{
			ProductID:         "p1",
			StoreID:           "s1",
			QuantityRemaining: 10,
			PurchasePrice:     50,
			ReceivedAt:        time.Now(),
		}
		mock.ExpectExec(`INSERT INTO stock_batches`).
			WithArgs(batch.ProductID, batch.StoreID, batch.POID, batch.QuantityRemaining, batch.PurchasePrice, batch.ReceivedAt).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err = repo.CreateBatch(ctx, batch)
		assert.NoError(t, err)
	})

	t.Run("GetBatchesByProduct", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()
		repo := NewBatchRepository(db)

		rows := sqlmock.NewRows([]string{"id", "product_id", "store_id", "po_id", "quantity_remaining", "purchase_price", "received_at", "created_at", "product_name", "product_sku", "unit"}).
			AddRow("b1", "p1", "s1", nil, 10.0, 50.0, time.Now(), time.Now(), "P1", "SKU1", "pcs")

		mock.ExpectQuery(`SELECT .* FROM stock_batches`).
			WithArgs("p1", "s1").
			WillReturnRows(rows)

		res, err := repo.GetBatchesByProduct(ctx, "p1", "s1")
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("GetBatchesByStore", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()
		repo := NewBatchRepository(db)

		filter := dto.BatchListFilter{StoreID: "s1", ProductID: "p1"}
		rows := sqlmock.NewRows([]string{"id", "product_id", "store_id", "po_id", "quantity_remaining", "purchase_price", "received_at", "created_at", "product_name", "product_sku", "unit"}).
			AddRow("b1", "p1", "s1", nil, 10.0, 50.0, time.Now(), time.Now(), "P1", "SKU1", "pcs")

		mock.ExpectQuery(`SELECT .* FROM stock_batches .* WHERE sb.store_id = \$1 .* AND sb.product_id = \$2`).
			WithArgs("s1", "p1").
			WillReturnRows(rows)

		res, err := repo.GetBatchesByStore(ctx, filter)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("GetStockSummary", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()
		repo := NewBatchRepository(db)

		rows := sqlmock.NewRows([]string{"product_id", "product_name", "product_sku", "unit", "total_qty", "batch_count", "avg_cost_price"}).
			AddRow("p1", "P1", "SKU1", "pcs", 20.0, 2, 55.0)

		mock.ExpectQuery(`SELECT .* FROM stock_batches .* WHERE sb.store_id = \$1`).
			WithArgs("s1").
			WillReturnRows(rows)

		res, err := repo.GetStockSummary(ctx, "s1")
		assert.NoError(t, err)
		require.Len(t, res, 1)
		assert.Equal(t, 20.0, res[0].TotalQty)
	})
}

func TestBatchRepo_DeductFIFO(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := NewBatchRepository(db)
	ctx := context.Background()

	t.Run("Success Multiple Batches", func(t *testing.T) {
		mock.ExpectBegin()

		// Step 1: Lock batches
		rows := sqlmock.NewRows([]string{"id", "quantity_remaining"}).
			AddRow("b1", 10.0). // Oldest
			AddRow("b2", 10.0)  // Newer

		mock.ExpectQuery(`SELECT id, quantity_remaining FROM stock_batches .* FOR UPDATE`).
			WithArgs("p1", "s1").
			WillReturnRows(rows)

		// Step 3: Updates
		// Deduct 15: covers b1 fully (10) and b2 partially (5)
		mock.ExpectExec(`UPDATE stock_batches SET quantity_remaining = \$1 WHERE id = \$2`).
			WithArgs(0.0, "b1").
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectExec(`UPDATE stock_batches SET quantity_remaining = \$1 WHERE id = \$2`).
			WithArgs(5.0, "b2").
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectCommit()

		err := repo.DeductFIFO(ctx, "p1", "s1", 15.0)
		assert.NoError(t, err)
	})

	t.Run("Insufficient Stock", func(t *testing.T) {
		mock.ExpectBegin()
		rows := sqlmock.NewRows([]string{"id", "quantity_remaining"}).
			AddRow("b1", 5.0)

		mock.ExpectQuery(`SELECT id, quantity_remaining FROM stock_batches`).
			WithArgs("p1", "s1").
			WillReturnRows(rows)

		mock.ExpectRollback()

		err := repo.DeductFIFO(ctx, "p1", "s1", 10.0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient batch stock")
	})
}
