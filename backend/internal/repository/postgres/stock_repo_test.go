package postgres

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
)

func TestStockRepo_FindLevelsByStore(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := NewStockRepository(db)

	storeID := customerTestStoreID
	now := time.Now()

	t.Run("success_all", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "product_id", "store_id", "quantity", "min_quantity", "updated_at", "product_name", "product_sku", "unit", "cost_price"}).
			AddRow("1", "p1", storeID, 10.5, 5.0, now, "Product 1", "SKU1", "pcs", 1000.0).
			AddRow("2", "p2", storeID, 2.0, 5.0, now, "Product 2", "SKU2", "kg", 2000.0)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT sl.id, sl.product_id, sl.store_id, sl.quantity, sl.min_quantity, sl.updated_at, p.name AS product_name, p.sku AS product_sku, p.unit, p.cost_price FROM stock_levels sl JOIN products p ON p.id = sl.product_id AND p.deleted_at IS NULL AND p.is_active = true WHERE sl.store_id = $1 ORDER BY p.name ASC")).
			WithArgs(storeID).
			WillReturnRows(rows)

		levels, err := repo.FindLevelsByStore(context.Background(), storeID, false)
		assert.NoError(t, err)
		assert.Len(t, levels, 2)
		assert.Equal(t, "Product 1", levels[0].ProductName)
		assert.Equal(t, 10.5, levels[0].Quantity)
	})

	t.Run("success_low_stock", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "product_id", "store_id", "quantity", "min_quantity", "updated_at", "product_name", "product_sku", "unit", "cost_price"}).
			AddRow("2", "p2", storeID, 2.0, 5.0, now, "Product 2", "SKU2", "kg", 2000.0)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT sl.id, sl.product_id, sl.store_id, sl.quantity, sl.min_quantity, sl.updated_at, p.name AS product_name, p.sku AS product_sku, p.unit, p.cost_price FROM stock_levels sl JOIN products p ON p.id = sl.product_id AND p.deleted_at IS NULL AND p.is_active = true WHERE sl.store_id = $1 AND sl.quantity <= sl.min_quantity ORDER BY p.name ASC")).
			WithArgs(storeID).
			WillReturnRows(rows)

		levels, err := repo.FindLevelsByStore(context.Background(), storeID, true)
		assert.NoError(t, err)
		assert.Len(t, levels, 1)
		assert.Equal(t, "Product 2", levels[0].ProductName)
	})
}

func TestStockRepo_FindLevelByProduct(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := NewStockRepository(db)

	productID := "p1"
	storeID := customerTestStoreID
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "product_id", "store_id", "quantity", "min_quantity", "updated_at", "product_name", "product_sku", "unit"}).
			AddRow("1", productID, storeID, 10.5, 5.0, now, "Product 1", "SKU1", "pcs")

		mock.ExpectQuery(regexp.QuoteMeta("SELECT sl.id, sl.product_id, sl.store_id, sl.quantity, sl.min_quantity, sl.updated_at, p.name AS product_name, p.sku AS product_sku, p.unit FROM stock_levels sl JOIN products p ON p.id = sl.product_id WHERE sl.product_id = $1 AND sl.store_id = $2")).
			WithArgs(productID, storeID).
			WillReturnRows(rows)

		level, err := repo.FindLevelByProduct(context.Background(), productID, storeID)
		assert.NoError(t, err)
		assert.NotNil(t, level)
		assert.Equal(t, productID, level.ProductID)
		assert.Equal(t, 10.5, level.Quantity)
	})

	t.Run("not_found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT sl.id, sl.product_id, sl.store_id, sl.quantity, sl.min_quantity, sl.updated_at, p.name AS product_name, p.sku AS product_sku, p.unit FROM stock_levels sl JOIN products p ON p.id = sl.product_id WHERE sl.product_id = $1 AND sl.store_id = $2")).
			WithArgs(productID, storeID).
			WillReturnError(sql.ErrNoRows)

		level, err := repo.FindLevelByProduct(context.Background(), productID, storeID)
		assert.NoError(t, err)
		assert.Nil(t, level)
	})
}

func TestStockRepo_SetMinQuantity(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := NewStockRepository(db)

	productID := "p1"
	storeID := customerTestStoreID
	minQty := 5.0

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO stock_levels (product_id, store_id, min_quantity, updated_at) VALUES ($1, $2, $3, NOW()) ON CONFLICT (product_id, store_id) DO UPDATE SET min_quantity = $3, updated_at = NOW()")).
			WithArgs(productID, storeID, minQty).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.SetMinQuantity(context.Background(), productID, storeID, minQty)
		assert.NoError(t, err)
	})
}

func TestStockRepo_Adjust(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := NewStockRepository(db)

	input := domain.AdjustInput{
		ProductID: "p1",
		StoreID:   customerTestStoreID,
		Delta:     10.0,
		RefType:   "adjustment",
		RefID:     nil,
		Notes:     "",
		CreatedBy: "user-1",
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO stock_movements (product_id, store_id, ref_type, ref_id, quantity_delta, notes, created_by) VALUES ($1, $2, $3, $4, $5, $6, $7)")).
			WithArgs(input.ProductID, input.StoreID, input.RefType, input.RefID, input.Delta, input.Notes, input.CreatedBy).
			WillReturnResult(sqlmock.NewResult(1, 1))

		slRows := sqlmock.NewRows([]string{"id", "product_id", "store_id", "quantity", "min_quantity", "updated_at"}).
			AddRow("1", input.ProductID, input.StoreID, 20.0, 5.0, time.Now())

		mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO stock_levels (product_id, store_id, quantity, updated_at) VALUES ($1, $2, $3, NOW()) ON CONFLICT (product_id, store_id) DO UPDATE SET quantity = stock_levels.quantity + $3, updated_at = NOW() RETURNING id, product_id, store_id, quantity, min_quantity, updated_at")).
			WithArgs(input.ProductID, input.StoreID, input.Delta).
			WillReturnRows(slRows)

		mock.ExpectCommit()

		pRows := sqlmock.NewRows([]string{"product_name", "product_sku", "unit"}).
			AddRow("Product 1", "SKU1", "pcs")
		mock.ExpectQuery(regexp.QuoteMeta("SELECT name AS product_name, sku AS product_sku, unit FROM products WHERE id = $1")).
			WithArgs(input.ProductID).
			WillReturnRows(pRows)

		level, err := repo.Adjust(context.Background(), input)
		assert.NoError(t, err)
		assert.NotNil(t, level)
		assert.Equal(t, 20.0, level.Quantity)
		assert.Equal(t, "Product 1", level.ProductName)
	})

	t.Run("rollback_on_error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO stock_movements (product_id, store_id, ref_type, ref_id, quantity_delta, notes, created_by) VALUES ($1, $2, $3, $4, $5, $6, $7)")).
			WithArgs(input.ProductID, input.StoreID, input.RefType, input.RefID, input.Delta, input.Notes, input.CreatedBy).
			WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		level, err := repo.Adjust(context.Background(), input)
		assert.Error(t, err)
		assert.Nil(t, level)
	})
}

func TestStockRepo_FindMovements(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := NewStockRepository(db)

	filter := dto.StockMovementFilter{
		StoreID: "store-1",
		PaginationQuery: dto.PaginationQuery{
			Page:    1,
			PerPage: 10,
		},
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM stock_movements sm WHERE sm.store_id = $1")).
			WithArgs(filter.StoreID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		rows := sqlmock.NewRows([]string{"id", "product_id", "store_id", "ref_type", "ref_id", "quantity_delta", "notes", "created_by", "created_at", "product_name", "created_by_name"}).
			AddRow("1", "p1", filter.StoreID, "adjustment", nil, 10.0, "Notes", "u1", time.Now(), "Product 1", "User 1")

		mock.ExpectQuery(regexp.QuoteMeta("SELECT sm.id, sm.product_id, sm.store_id, sm.ref_type, sm.ref_id, sm.quantity_delta, sm.notes, sm.created_by, sm.created_at, p.name AS product_name, u.name AS created_by_name FROM stock_movements sm JOIN products p ON p.id = sm.product_id JOIN users u ON u.id = sm.created_by WHERE sm.store_id = $1 ORDER BY sm.created_at DESC LIMIT $2 OFFSET $3")).
			WithArgs(filter.StoreID, 10, 0).
			WillReturnRows(rows)

		movements, total, err := repo.FindMovements(context.Background(), filter)
		assert.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, movements, 1)
		assert.Equal(t, "Product 1", movements[0].ProductName)
	})
}

func TestStockRepo_DeductStock(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := NewStockRepository(db)

	productID := "p1"
	storeID := "store-1"
	qty := 2.0
	refID := "txn-1"
	cashierID := "user-1"

	t.Run("success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO stock_movements (product_id, store_id, ref_type, ref_id, quantity_delta, notes, created_by) VALUES ($1, $2, 'sale', $3, $4, 'Penjualan menu restoran', $5)")).
			WithArgs(productID, storeID, refID, -qty, cashierID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(regexp.QuoteMeta("UPDATE stock_levels SET quantity = GREATEST(0, quantity - $1), updated_at = NOW() WHERE product_id = $2 AND store_id = $3")).
			WithArgs(qty, productID, storeID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		err := repo.DeductStock(context.Background(), productID, storeID, qty, refID, cashierID)
		assert.NoError(t, err)
	})
}
