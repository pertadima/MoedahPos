package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
)

func TestSupplierRepo_Basic(t *testing.T) {
	ctx := context.Background()

	t.Run("Create", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()
		repo := NewSupplierRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectQuery(`(?is)INSERT INTO suppliers`).
			WithArgs("Supplier 1", "Contact 1", "phone1", "email1", "address1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("s1", "Supplier 1"))

		res, err := repo.Create(ctx, &domain.Supplier{
			Name: "Supplier 1", ContactName: "Contact 1", Phone: "phone1", Email: "email1", Address: "address1",
		})
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "s1", res.ID)
	})

	t.Run("FindAll", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()
		repo := NewSupplierRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectQuery(`(?is)SELECT COUNT\(\*\) FROM suppliers`).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		mock.ExpectQuery(`(?is)SELECT .* FROM suppliers`).
			WithArgs(20, 0).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("s1", "Supplier 1"))

		res, total, err := repo.FindAll(ctx, dto.SupplierListFilter{PaginationQuery: dto.PaginationQuery{PerPage: 20, Page: 1}})
		assert.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, res, 1)
	})

	t.Run("FindByID", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()
		repo := NewSupplierRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectQuery(`(?is)SELECT .* FROM suppliers WHERE id = \$1`).
			WithArgs("s1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("s1", "Supplier 1"))

		res, err := repo.FindByID(ctx, "s1")
		assert.NoError(t, err)
		assert.NotNil(t, res)

		// Not Found
		mock.ExpectQuery(`(?is)SELECT .* FROM suppliers WHERE id = \$1`).WithArgs("unknown").WillReturnError(sql.ErrNoRows)
		res, err = repo.FindByID(ctx, "unknown")
		assert.NoError(t, err)
		assert.Nil(t, res)
	})
}

func TestSupplierRepo_Manage(t *testing.T) {
	ctx := context.Background()

	t.Run("Update", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()
		repo := NewSupplierRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectQuery(`(?is)UPDATE suppliers`).
			WithArgs("Updated", "Contact", "phone", "email", "addr", true, "s1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("s1", "Updated"))

		res, err := repo.Update(ctx, &domain.Supplier{ID: "s1", Name: "Updated", ContactName: "Contact", Phone: "phone", Email: "email", Address: "addr", IsActive: true})
		assert.NoError(t, err)
		assert.NotNil(t, res)

		// Not Found
		mock.ExpectQuery(`(?is)UPDATE suppliers`).WillReturnError(sql.ErrNoRows)
		res, err = repo.Update(ctx, &domain.Supplier{ID: "unknown"})
		assert.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run("SoftDelete", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()
		repo := NewSupplierRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectExec(`(?is)UPDATE suppliers SET deleted_at`).
			WithArgs("s1").
			WillReturnResult(sqlmock.NewResult(1, 1))

		err = repo.SoftDelete(ctx, "s1")
		assert.NoError(t, err)

		// Not Found
		mock.ExpectExec(`(?is)UPDATE suppliers SET deleted_at`).WillReturnResult(sqlmock.NewResult(0, 0))
		err = repo.SoftDelete(ctx, "unknown")
		assert.Error(t, err)
	})
}

func TestReportRepo_Sales(t *testing.T) {
	ctx := context.Background()
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	t.Run("SalesSummary", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()
		repo := NewReportRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectQuery(`(?is)WITH sales AS .* SELECT .* FROM sales s FULL OUTER JOIN exp_agg e`).
			WithArgs("st1", from, to).
			WillReturnRows(sqlmock.NewRows([]string{"date", "total_sales", "net_profit"}).AddRow("2024-01-01", 1000.0, 500.0))

		res, err := repo.SalesSummary(ctx, "st1", from, to)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("SalesByProduct", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()
		repo := NewReportRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectQuery(`(?is)SELECT .* FROM transaction_items ti .* GROUP BY ti.product_id`).
			WithArgs("st1", from, to).
			WillReturnRows(sqlmock.NewRows([]string{"product_id", "product_name", "total_revenue"}).AddRow("p1", "Prod 1", 1000.0))

		res, err := repo.SalesByProduct(ctx, "st1", from, to)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("SalesByCashier", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()
		repo := NewReportRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectQuery(`(?is)SELECT .* FROM transactions t JOIN users u .* GROUP BY u.id`).
			WithArgs("st1", from, to).
			WillReturnRows(sqlmock.NewRows([]string{"cashier_id", "cashier_name", "total_sales"}).AddRow("u1", "User 1", 1000.0))

		res, err := repo.SalesByCashier(ctx, "st1", from, to)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})
}

func TestReportRepo_Financials(t *testing.T) {
	ctx := context.Background()
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	t.Run("StockValuation", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()
		repo := NewReportRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectQuery(`(?is)SELECT .* FROM products p LEFT JOIN stock_levels sl`).
			WithArgs("st1").
			WillReturnRows(sqlmock.NewRows([]string{"product_id", "product_name", "total_value"}).AddRow("p1", "Prod 1", 5000.0))

		res, err := repo.StockValuation(ctx, "st1")
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("ProfitSummary", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()
		repo := NewReportRepo(sqlx.NewDb(db, "postgres"))

		// Test for "day"
		mock.ExpectQuery(`(?is)WITH sales AS .* SELECT .* FROM sales s FULL OUTER JOIN exp_agg e`).
			WithArgs("st1", from, to).
			WillReturnRows(sqlmock.NewRows([]string{"period", "total_sales", "net_profit"}).AddRow("2024-01-01", 1000.0, 500.0))

		res, err := repo.ProfitSummary(ctx, "st1", from, to, "day")
		assert.NoError(t, err)
		assert.Len(t, res, 1)

		// Test for "week"
		mock.ExpectQuery(`(?is)WITH sales AS`).WillReturnRows(sqlmock.NewRows([]string{"period"}).AddRow("2024-01-01"))
		_, _ = repo.ProfitSummary(ctx, "st1", from, to, "week")

		// Test for "month"
		mock.ExpectQuery(`(?is)WITH sales AS`).WillReturnRows(sqlmock.NewRows([]string{"period"}).AddRow("2024-01"))
		_, _ = repo.ProfitSummary(ctx, "st1", from, to, "month")
	})

	t.Run("CashFlowSummary", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()
		repo := NewReportRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectQuery(`(?is)SELECT .* AS date, SUM\(t.total\) AS cash_in`).
			WithArgs("st1", from, to).
			WillReturnRows(sqlmock.NewRows([]string{"date", "cash_in", "payment_method"}).AddRow("2024-01-01", 1000.0, "cash"))

		mock.ExpectQuery(`(?is)SELECT date, SUM\(cash_out\) AS cash_out FROM`).
			WithArgs("st1", from, to).
			WillReturnRows(sqlmock.NewRows([]string{"date", "cash_out"}).AddRow("2024-01-01", 200.0))

		mock.ExpectQuery(`(?is)SELECT .* AS date, SUM\(amount\) AS other_in FROM incomes`).
			WithArgs("st1", from, to).
			WillReturnRows(sqlmock.NewRows([]string{"date", "other_in"}).AddRow("2024-01-01", 100.0))

		res, err := repo.CashFlowSummary(ctx, "st1", from, to)
		assert.NoError(t, err)
		require.NotEmpty(t, res)
		assert.Equal(t, 900.0, res[0].NetCash) // 1000 + 100 - 200
	})

	t.Run("CashFlowDetail", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()
		repo := NewReportRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectQuery(`(?is)SELECT 'SALE' AS type`).
			WithArgs("st1", from, to).
			WillReturnRows(sqlmock.NewRows([]string{"type", "label", "amount", "payment_method", "timestamp"}).
				AddRow("SALE", "Sale #1", 100.0, "cash", "2024-01-01 10:00:00"))

		res, err := repo.CashFlowDetail(ctx, "st1", from, to)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})
}
