package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"github.com/moedahpos/backend/internal/domain"
)

func TestTransactionRepo_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer db.Close()

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
	defer db.Close()

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
