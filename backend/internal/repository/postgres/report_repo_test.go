package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestReportRepo_StockValuation(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer func() { _ = db.Close() }()

	repo := NewReportRepository(db)

	ctx := context.Background()
	storeID := "s1"

	rows := sqlmock.NewRows([]string{"product_id", "product_name", "sku", "unit", "cost_price", "quantity", "total_value"}).
		AddRow("p1", "Product 1", "SKU1", "pcs", 50, 10, 500.0).
		AddRow("p2", "Product 2", "SKU2", "kg", 100, 5, 500.0)

	mock.ExpectQuery(`SELECT (.+) FROM products p (.+) WHERE p.store_id = \$1`).
		WithArgs(storeID).
		WillReturnRows(rows)

	res, err := repo.StockValuation(ctx, storeID)

	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.Equal(t, 500.0, res[0].TotalValue)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepo_SalesByProduct(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer func() { _ = db.Close() }()

	repo := NewReportRepository(db)

	ctx := context.Background()
	storeID := "s1"
	from := time.Now().Add(-24 * time.Hour)
	to := time.Now()

	rows := sqlmock.NewRows([]string{"product_id", "product_name", "sku", "total_quantity", "total_revenue", "total_cost", "gross_profit", "profit_margin", "total_tax"}).
		AddRow("p1", "Product 1", "SKU1", 10, 1000.0, 500.0, 500.0, 50.0, 100.0)

	mock.ExpectQuery(`SELECT (.+) FROM transaction_items ti (.+) WHERE t.store_id = \$1`).
		WithArgs(storeID, from, to).
		WillReturnRows(rows)

	res, err := repo.SalesByProduct(ctx, storeID, from, to)

	assert.NoError(t, err)
	assert.Len(t, res, 1)
	assert.Equal(t, 1000.0, res[0].TotalRevenue)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepo_CashFlow(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer func() { _ = db.Close() }()

	repo := NewReportRepository(db)

	ctx := context.Background()
	storeID := "s1"
	from := time.Now().Add(-24 * time.Hour)
	to := time.Now()

	t.Run("CashFlowSummary", func(t *testing.T) {
		mock.ExpectQuery(`SELECT .* FROM transactions t`).WillReturnRows(sqlmock.NewRows([]string{"date", "cash_in", "payment_method"}).AddRow("2024-01-01", 1000.0, "cash"))
		mock.ExpectQuery(`SELECT date, SUM\(cash_out\) .* FROM \(.*SELECT .* FROM po_payments`).WillReturnRows(sqlmock.NewRows([]string{"date", "cash_out"}).AddRow("2024-01-01", 500.0))
		mock.ExpectQuery(`SELECT .* FROM incomes`).WillReturnRows(sqlmock.NewRows([]string{"date", "other_in"}).AddRow("2024-01-01", 200.0))

		res, err := repo.CashFlowSummary(ctx, storeID, from, to)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
		assert.Equal(t, 1200.0, res[0].CashIn) // 1000 (sales) + 200 (other)
		assert.Equal(t, 500.0, res[0].CashOut)
		assert.Equal(t, 700.0, res[0].NetCash)
	})

	t.Run("CashFlowDetail", func(t *testing.T) {
		mock.ExpectQuery(`SELECT 'SALE' AS type`).WillReturnRows(sqlmock.NewRows([]string{"type", "label", "amount", "payment_method", "category", "notes", "timestamp"}).
			AddRow("SALE", "Penjualan #abc", 100.0, "cash", nil, nil, "2024-01-01 10:00:00"))

		res, err := repo.CashFlowDetail(ctx, storeID, from, to)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
		assert.Equal(t, "SALE", res[0].Type)
	})
}
