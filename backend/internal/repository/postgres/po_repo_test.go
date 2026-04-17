package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"github.com/moedahpos/backend/internal/domain"
)

func TestPORepo_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewPORepo(sqlxDB)
	ctx := context.Background()

	po := &domain.PurchaseOrder{
		StoreID:     "s1",
		SupplierID:  ptrToStringPtr("supp1"),
		PONumber:    "PO-123",
		TotalAmount: 1000,
		OrderedBy:   "u1",
		Notes:       ptrToStringPtr("note"),
	}
	items := []domain.POItem{
		{ProductID: "p1", Quantity: 10, UnitCost: 100, Subtotal: 1000},
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO purchase_orders`).WithArgs(
		po.StoreID, po.SupplierID, po.PONumber, po.TotalAmount, po.OrderedBy, po.Notes,
	).WillReturnRows(sqlmock.NewRows([]string{"id", "store_id", "supplier_id", "po_number", "status", "total_amount", "ordered_by", "received_by", "ordered_at", "received_at", "notes", "created_at", "updated_at"}).
		AddRow("po1", "s1", "supp1", "PO-123", "draft", 1000, "u1", nil, time.Now(), nil, "note", time.Now(), time.Now()))

	mock.ExpectQuery(`INSERT INTO purchase_order_items`).WithArgs(
		"po1", "p1", float64(10), float64(100), float64(1000),
	).WillReturnRows(sqlmock.NewRows([]string{"id", "po_id", "product_id", "quantity", "unit_cost", "received_qty", "subtotal"}).
		AddRow("item1", "po1", "p1", 10, 100, 0, 1000))

	mock.ExpectCommit()

	mock.ExpectQuery(`SELECT u.name AS ordered_by_name`).WithArgs("po1").
		WillReturnRows(sqlmock.NewRows([]string{"ordered_by_name", "supplier_name", "received_by_name"}).
			AddRow("User 1", "Supplier 1", nil))

	result, err := repo.Create(ctx, po, items)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "po1", result.ID)
	assert.Equal(t, "User 1", result.OrderedByName)
}

func TestPORepo_FindByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewPORepo(sqlxDB)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT po.id`).WithArgs("po1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "store_id", "supplier_id", "po_number", "status", "total_amount", "ordered_by", "received_by", "ordered_at", "received_at", "notes", "created_at", "updated_at", "supplier_name", "ordered_by_name", "received_by_name", "total_items"}).
			AddRow("po1", "s1", "supp1", "PO-123", "draft", 1000, "u1", nil, time.Now(), nil, "note", time.Now(), time.Now(), "Supplier 1", "User 1", nil, 1))

	mock.ExpectQuery(`SELECT poi.id`).WithArgs("po1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "po_id", "product_id", "quantity", "unit_cost", "received_qty", "subtotal", "product_name", "product_sku", "unit"}).
			AddRow("item1", "po1", "p1", 10, 100, 0, 1000, "Product 1", "SKU1", "pcs"))

	result, err := repo.FindByID(ctx, "po1")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "po1", result.ID)
	assert.Len(t, result.Items, 1)

	// Not Found
	mock.ExpectQuery(`SELECT po.id`).WithArgs("unknown").WillReturnError(sql.ErrNoRows)
	result, err = repo.FindByID(ctx, "unknown")
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestPORepo_Submit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewPORepo(sqlxDB)
	ctx := context.Background()

	mock.ExpectExec(`UPDATE purchase_orders`).WithArgs("po1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Submit(ctx, "po1", "u1")
	assert.NoError(t, err)

	// Fail (no rows affected)
	mock.ExpectExec(`UPDATE purchase_orders`).WithArgs("po2").
		WillReturnResult(sqlmock.NewResult(0, 0))
	err = repo.Submit(ctx, "po2", "u1")
	assert.Error(t, err)
}

func TestPORepo_Receive(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewPORepo(sqlxDB)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT product_id, quantity, unit_cost`).WithArgs("po1").
		WillReturnRows(sqlmock.NewRows([]string{"product_id", "quantity", "unit_cost"}).
			AddRow("p1", 10.0, 100.0))

	mock.ExpectQuery(`SELECT store_id FROM purchase_orders`).WithArgs("po1").
		WillReturnRows(sqlmock.NewRows([]string{"store_id"}).AddRow("s1"))

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE purchase_orders`).WithArgs("u1", "po1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE purchase_order_items`).WithArgs("po1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(`INSERT INTO stock_movements`).WithArgs("p1", "s1", "po1", 10.0, "u1").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO stock_levels`).WithArgs("p1", "s1", 10.0).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`UPDATE products`).WithArgs(100.0, "p1").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO stock_batches`).WithArgs("p1", "s1", "po1", 10.0, 100.0).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err = repo.Receive(ctx, "po1", "u1")
	assert.NoError(t, err)
}

func ptrToStringPtr(s string) *string {
	return &s
}
