package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
)

func TestProductRepo_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer func() { _ = db.Close() }()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewProductRepo(sqlxDB)

	ctx := context.Background()
	desc := "Description 1"
	catID := "c1"
	barcode := "12345"
	tax := 10.0
	p := &domain.Product{
		StoreID:       "s1",
		CategoryID:    &catID,
		SKU:           "SKU1",
		Name:          "Product 1",
		Description:   &desc,
		Barcode:       &barcode,
		Unit:          "pcs",
		CostPrice:     50,
		SellPrice:     100,
		UseGlobalTax:  true,
		TaxPercentage: &tax,
	}

	rows := sqlmock.NewRows([]string{"id", "store_id", "category_id", "sku", "name", "description", "barcode", "unit", "cost_price", "sell_price", "use_global_tax", "tax_percentage", "image_url", "is_active", "created_at", "updated_at", "deleted_at"}).
		AddRow("p1", "s1", "c1", "SKU1", "Product 1", "Description 1", "12345", "pcs", 50, 100, true, 10, "", true, time.Now(), time.Now(), nil)

	mock.ExpectQuery(`INSERT INTO products`).
		WithArgs(p.StoreID, p.CategoryID, p.SKU, p.Name, p.Description, p.Barcode, p.Unit, p.CostPrice, p.SellPrice, p.UseGlobalTax, p.TaxPercentage, p.ImageURL).
		WillReturnRows(rows)

	res, err := repo.Create(ctx, p)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "p1", res.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCategoryRepo_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer func() { _ = db.Close() }()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewCategoryRepo(sqlxDB)

	ctx := context.Background()
	c := &domain.Category{
		StoreID: "s1",
		Name:    "Category 1",
	}

	rows := sqlmock.NewRows([]string{"id", "store_id", "name", "parent_id", "created_at", "updated_at", "deleted_at"}).
		AddRow("c1", "s1", "Category 1", nil, time.Now(), time.Now(), nil)

	mock.ExpectQuery(`INSERT INTO categories`).
		WithArgs(c.StoreID, c.Name, c.ParentID).
		WillReturnRows(rows)

	res, err := repo.Create(ctx, c)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "c1", res.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepo_FindByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer func() { _ = db.Close() }()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewProductRepo(sqlxDB)

	ctx := context.Background()
	rows := sqlmock.NewRows([]string{"id", "store_id", "category_id", "sku", "name", "description", "barcode", "unit", "cost_price", "sell_price", "use_global_tax", "tax_percentage", "image_url", "is_active", "created_at", "updated_at", "deleted_at", "category_name", "store_default_tax"}).
		AddRow("p1", "s1", "c1", "SKU1", "Product 1", "Desc", "123", "pcs", 50, 100, true, 10, "", true, time.Now(), time.Now(), nil, "Cat 1", 10.0)

	mock.ExpectQuery(`SELECT .* FROM products p`).WithArgs("p1").WillReturnRows(rows)

	res, err := repo.FindByID(ctx, "p1")
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "p1", res.ID)
}

func TestProductRepo_ExistsBySKU(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer func() { _ = db.Close() }()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewProductRepo(sqlxDB)

	ctx := context.Background()

	mock.ExpectQuery(`SELECT EXISTS`).WithArgs("s1", "SKU1").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	exists, err := repo.ExistsBySKU(ctx, "s1", "SKU1", "")
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestProductRepo_FindAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	repo := NewProductRepo(sqlx.NewDb(db, "postgres"))
	ctx := context.Background()

	f := dto.ProductListFilter{StoreID: "s1", WithStock: true}
	f.Defaults()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM products p`).WithArgs("s1").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT .* FROM products p`).WithArgs("s1", 20, 0).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("p1", "P1"))

	res, total, err := repo.FindAll(ctx, f)
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, res, 1)
}

func TestProductRepo_FindByBarcode(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	repo := NewProductRepo(sqlx.NewDb(db, "postgres"))
	ctx := context.Background()

	barcode := "123"
	mock.ExpectQuery(`SELECT .* FROM products p`).WithArgs("s1", "123").
		WillReturnRows(sqlmock.NewRows([]string{"id", "barcode"}).AddRow("p1", "123"))

	res, err := repo.FindByBarcode(ctx, "s1", "123")
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, barcode, *res.Barcode)
}
