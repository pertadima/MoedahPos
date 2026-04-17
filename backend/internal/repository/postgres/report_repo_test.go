package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestReportRepo_StockValuation(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewReportRepo(sqlxDB)

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
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewReportRepo(sqlxDB)

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
