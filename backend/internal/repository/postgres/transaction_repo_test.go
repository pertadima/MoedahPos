package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
)

func TestTransactionRepo_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer func() { _ = db.Close() }()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewTransactionRepo(sqlxDB)

	ctx := context.Background()
	pid := "p1"
	input := domain.CreateTransactionInput{
		StoreID:   "s1",
		CashierID: "u1",
		Subtotal:  100,
		Total:     110,
		Status:    "completed",
		Items: []domain.CreateTransactionItemInput{
			{
				ProductID: &pid,
				Quantity:  1,
				UnitPrice: 100,
				Subtotal:  110,
			},
		},
	}

	mock.ExpectBegin()

	// 1. Transaction Header
	rows := sqlmock.NewRows([]string{"id", "store_id", "cashier_id", "table_id", "customer_name", "customer_phone", "subtotal", "discount_amt", "tax_amt", "total", "payment_method", "payment_amount", "change_amount", "status", "notes", "cart_discount_type", "cart_discount_value", "created_at", "updated_at"}).
		AddRow("t1", "s1", "u1", nil, "", "", 100, 0, 10, 110, "cash", 110, 0, "completed", "", "", 0, time.Now(), time.Now())

	mock.ExpectQuery(`INSERT INTO transactions`).
		WithArgs(input.StoreID, input.CashierID, input.TableID, input.CustomerName, input.CustomerPhone, input.Subtotal, input.DiscountAmt, input.TaxAmt, input.Total, input.PaymentMethod, input.PaymentAmount, input.ChangeAmount, "completed", input.Notes, input.CartDiscountType, input.CartDiscountValue).
		WillReturnRows(rows)

	// 2. Transaction Items
	itemRows := sqlmock.NewRows([]string{"id", "transaction_id", "product_id", "menu_item_id", "product_name", "sku", "quantity", "original_price", "unit_price", "cost_price", "discount_pct", "discount_type", "discount_value", "cart_discount_allocated", "tax_rate", "subtotal", "status", "completed_at"}).
		AddRow("ti1", "t1", "p1", nil, "Product 1", "SKU1", 1, 100, 100, 50, 0, "", 0, 0, 10, 110, "pending", nil)

	mock.ExpectQuery(`INSERT INTO transaction_items`).
		WillReturnRows(itemRows)

	// 3. Stock Movements & Levels
	mock.ExpectExec(`INSERT INTO stock_movements`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO stock_levels`).WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	// 4. Cashier Name
	mock.ExpectQuery(`SELECT name FROM users`).
		WithArgs("u1").
		WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("Cashier 1"))

	res, err := repo.Create(ctx, input)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "t1", res.ID)
	assert.Equal(t, "Cashier 1", res.CashierName)
	assert.Len(t, res.Items, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTransactionRepo_FindByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer func() { _ = db.Close() }()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewTransactionRepo(sqlxDB)

	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "store_id", "cashier_id", "table_id", "customer_name", "customer_phone", "subtotal", "discount_amt", "tax_amt", "total", "payment_method", "payment_amount", "change_amount", "status", "notes", "created_at", "updated_at", "cashier_name"}).
		AddRow("t1", "s1", "u1", nil, "", "", 100, 0, 10, 110, "cash", 110, 0, "completed", "", time.Now(), time.Now(), "User 1")

	mock.ExpectQuery(`SELECT .* FROM transactions t.* WHERE t.id = \$1`).
		WithArgs("t1").
		WillReturnRows(rows)

	itemRows := sqlmock.NewRows([]string{"id", "transaction_id", "product_id", "menu_item_id", "product_name", "sku", "quantity", "original_price", "unit_price", "cost_price", "discount_pct", "discount_type", "discount_value", "cart_discount_allocated", "tax_rate", "subtotal", "status", "completed_at"}).
		AddRow("ti1", "t1", "p1", nil, "Product 1", "SKU1", 1, 100, 100, 50, 0, "", 0, 0, 10, 110, "completed", nil)

	mock.ExpectQuery(`SELECT .* FROM transaction_items WHERE transaction_id = \$1`).
		WithArgs("t1").
		WillReturnRows(itemRows)

	res, err := repo.FindByID(ctx, "t1")

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "t1", res.ID)
	assert.Len(t, res.Items, 1)
}

func TestTransactionRepo_FindAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer func() { _ = db.Close() }()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewTransactionRepo(sqlxDB)

	ctx := context.Background()
	filter := dto.TransactionListFilter{
		StoreID:         "s1",
		Status:          "completed",
		PaginationQuery: dto.PaginationQuery{Page: 1, PerPage: 10},
	}

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM transactions t`).
		WithArgs("s1", "completed").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	rows := sqlmock.NewRows([]string{"id", "store_id", "cashier_id", "table_id", "customer_name", "customer_phone", "subtotal", "discount_amt", "tax_amt", "total", "payment_method", "payment_amount", "change_amount", "status", "notes", "created_at", "updated_at", "cashier_name"}).
		AddRow("t1", "s1", "u1", nil, "", "", 100, 0, 10, 110, "cash", 110, 0, "completed", "", time.Now(), time.Now(), "User 1")

	mock.ExpectQuery(`SELECT .* FROM transactions t.* JOIN users u`).
		WithArgs("s1", "completed", 10, 0).
		WillReturnRows(rows)

	res, total, err := repo.FindAll(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, res, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTransactionRepo_Void(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer func() { _ = db.Close() }()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewTransactionRepo(sqlxDB)

	ctx := context.Background()
	tid := "t1"
	uid := "u1"

	mock.ExpectQuery(`SELECT product_id, quantity FROM transaction_items WHERE transaction_id = \$1`).
		WithArgs(tid).
		WillReturnRows(sqlmock.NewRows([]string{"product_id", "quantity"}).AddRow("p1", 1))

	mock.ExpectQuery(`SELECT store_id FROM transactions WHERE id = \$1`).
		WithArgs(tid).
		WillReturnRows(sqlmock.NewRows([]string{"store_id"}).AddRow("s1"))

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE transactions SET status='voided'`).WithArgs(tid).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO stock_movements`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO stock_levels`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = repo.Void(ctx, tid, uid)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTransactionRepo_UpdateKDSItemStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer func() { _ = db.Close() }()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewTransactionRepo(sqlxDB)

	ctx := context.Background()

	mock.ExpectExec(`UPDATE transaction_items SET status = \$1, completed_at = NOW\(\)`).
		WithArgs("completed", "item1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.UpdateKDSItemStatus(ctx, "item1", "completed")
	assert.NoError(t, err)

	mock.ExpectExec(`UPDATE transaction_items SET status = \$1, completed_at = NULL`).
		WithArgs("pending", "item1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.UpdateKDSItemStatus(ctx, "item1", "pending")
	assert.NoError(t, err)
}

func TestTransactionRepo_GetKDSTickets(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer func() { _ = db.Close() }()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewTransactionRepo(sqlxDB)

	ctx := context.Background()

	mock.ExpectQuery(`SELECT .* FROM transactions t.* JOIN users u.* LEFT JOIN restaurant_tables`).
		WithArgs("s1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "store_id", "cashier_id", "table_id", "table_number", "customer_name", "customer_phone", "subtotal", "discount_amt", "tax_amt", "total", "payment_method", "payment_amount", "change_amount", "status", "notes", "created_at", "updated_at", "cashier_name"}).
			AddRow("t1", "s1", "u1", nil, nil, "", "", 100, 0, 10, 110, "cash", 110, 0, "draft", "", time.Now(), time.Now(), "User 1"))

	mock.ExpectQuery(`SELECT .* FROM transaction_items WHERE transaction_id IN`).
		WithArgs("t1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "transaction_id", "product_id", "menu_item_id", "product_name", "sku", "quantity", "original_price", "unit_price", "cost_price", "discount_pct", "discount_type", "discount_value", "cart_discount_allocated", "tax_rate", "subtotal", "status", "completed_at"}).
			AddRow("ti1", "t1", "p1", nil, "P1", "SKU1", 1, 100, 100, 50, 0, "", 0, 0, 10, 110, "pending", nil))

	res, err := repo.GetKDSTickets(ctx, "s1")
	assert.NoError(t, err)
	assert.Len(t, res, 1)
	assert.Len(t, res[0].Items, 1)
}
