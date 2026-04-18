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
)

func TestTableRepo_Basic(t *testing.T) {
	stringPtr := func(s string) *string { return &s }

	t.Run("FindAllByStore", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer func() { _ = db.Close() }()
		repo := NewTableRepo(sqlx.NewDb(db, "postgres"))

		rows := sqlmock.NewRows([]string{"id", "store_id", "table_number", "capacity", "status", "notes", "is_active", "created_at", "updated_at", "deleted_at"}).
			AddRow("t1", "s1", "1A", 4, "available", "Near window", true, time.Now(), time.Now(), nil)
		mock.ExpectQuery(`(?is)SELECT .* FROM restaurant_tables WHERE store_id = \$1`).WithArgs("s1").WillReturnRows(rows)
		res, err := repo.FindAllByStore(context.Background(), "s1")
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("FindByID", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer func() { _ = db.Close() }()
		repo := NewTableRepo(sqlx.NewDb(db, "postgres"))

		rows := sqlmock.NewRows([]string{"id", "store_id", "table_number", "capacity", "status", "notes", "is_active", "created_at", "updated_at", "deleted_at"}).
			AddRow("t1", "s1", "1A", 4, "available", "Notes", true, time.Now(), time.Now(), nil)
		mock.ExpectQuery(`(?is)SELECT .* FROM restaurant_tables WHERE id = \$1`).WithArgs("t1").WillReturnRows(rows)
		res, err := repo.FindByID(context.Background(), "t1")
		assert.NoError(t, err)
		assert.NotNil(t, res)

		// Not Found
		mock.ExpectQuery(`(?is)SELECT .* FROM restaurant_tables WHERE id = \$1`).WithArgs("unknown").WillReturnError(sql.ErrNoRows)
		res, err = repo.FindByID(context.Background(), "unknown")
		assert.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run("Create", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer func() { _ = db.Close() }()
		repo := NewTableRepo(sqlx.NewDb(db, "postgres"))

		rows := sqlmock.NewRows([]string{"id", "store_id", "table_number", "capacity", "status", "notes", "is_active", "created_at", "updated_at", "deleted_at"}).
			AddRow("t1", "s1", "1A", 4, "available", "notes", true, time.Now(), time.Now(), nil)
		mock.ExpectQuery(`(?is)INSERT INTO restaurant_tables`).
			WithArgs("s1", "1A", 4, "available", "notes").
			WillReturnRows(rows)

		res, err := repo.Create(context.Background(), &domain.RestaurantTable{
			StoreID:     "s1",
			TableNumber: "1A",
			Capacity:    4,
			Status:      "available",
			Notes:       stringPtr("notes"),
		})
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})
}

func TestTableRepo_Manage(t *testing.T) {
	stringPtr := func(s string) *string { return &s }

	t.Run("Update", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer func() { _ = db.Close() }()
		repo := NewTableRepo(sqlx.NewDb(db, "postgres"))

		rows := sqlmock.NewRows([]string{"id", "store_id", "table_number", "capacity", "status", "notes", "is_active", "created_at", "updated_at", "deleted_at"}).
			AddRow("t1", "s1", "1B", 6, "available", "New", true, time.Now(), time.Now(), nil)
		mock.ExpectQuery(`(?is)UPDATE restaurant_tables SET table_number=\$1`).
			WithArgs("1B", 6, "New", true, "t1").
			WillReturnRows(rows)

		res, err := repo.Update(context.Background(), &domain.RestaurantTable{
			ID:          "t1",
			TableNumber: "1B",
			Capacity:    6,
			Notes:       stringPtr("New"),
			IsActive:    true,
		})
		assert.NoError(t, err)
		assert.Equal(t, "1B", res.TableNumber)

		// Not Found
		mock.ExpectQuery(`(?is)UPDATE restaurant_tables SET table_number=\$1`).
			WillReturnError(sql.ErrNoRows)
		res, err = repo.Update(context.Background(), &domain.RestaurantTable{ID: "unknown"})
		assert.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run("UpdateStatus", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer func() { _ = db.Close() }()
		repo := NewTableRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectExec(`(?is)UPDATE restaurant_tables SET status=\$1`).WithArgs("occupied", "t1").WillReturnResult(sqlmock.NewResult(1, 1))
		err := repo.UpdateStatus(context.Background(), "t1", "occupied")
		assert.NoError(t, err)

		// Not Found
		mock.ExpectExec(`(?is)UPDATE restaurant_tables SET status=\$1`).WillReturnResult(sqlmock.NewResult(0, 0))
		err = repo.UpdateStatus(context.Background(), "unknown", "occupied")
		assert.Error(t, err)
	})

	t.Run("SoftDelete", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer func() { _ = db.Close() }()
		repo := NewTableRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectExec(`(?is)UPDATE restaurant_tables SET deleted_at=NOW()`).WithArgs("t1").WillReturnResult(sqlmock.NewResult(1, 1))
		err := repo.SoftDelete(context.Background(), "t1")
		assert.NoError(t, err)

		// Not Found
		mock.ExpectExec(`(?is)UPDATE restaurant_tables SET deleted_at=NOW()`).WillReturnResult(sqlmock.NewResult(0, 0))
		err = repo.SoftDelete(context.Background(), "unknown")
		assert.Error(t, err)
	})
}

func TestMenuItemRepo_Basic(t *testing.T) {
	t.Run("FindAllByStore", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer func() { _ = db.Close() }()
		repo := NewMenuItemRepo(sqlx.NewDb(db, "postgres"))

		rows := sqlmock.NewRows([]string{"id", "store_id", "category_id", "name", "description", "sell_price", "use_global_tax", "tax_percentage", "image_url", "is_active", "packaging_cost", "overhead_cost", "labor_cost", "created_at", "updated_at", "deleted_at", "category_name", "store_default_tax"}).
			AddRow("m1", "s1", "c1", "Coffee", "Hot", 20000.0, true, 10.0, "", true, 1000.0, 500.0, 2000.0, time.Now(), time.Now(), nil, "Drinks", 10.0)
		mock.ExpectQuery(`(?is)SELECT .* FROM menu_items`).WithArgs("s1").WillReturnRows(rows)

		ingRows := sqlmock.NewRows([]string{"id", "menu_item_id", "product_id", "quantity", "product_name", "product_sku", "unit", "cost_price"}).
			AddRow("msi1", "m1", "p1", 10.0, "Beans", "SKU1", "gram", 100.0)
		mock.ExpectQuery(`(?is)SELECT .* FROM menu_item_ingredients`).WithArgs("m1").WillReturnRows(ingRows)

		res, err := repo.FindAllByStore(context.Background(), "s1")
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("FindByID", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer func() { _ = db.Close() }()
		repo := NewMenuItemRepo(sqlx.NewDb(db, "postgres"))

		cols := []string{"id", "store_id", "category_id", "name", "description", "sell_price", "use_global_tax", "tax_percentage", "image_url", "is_active", "packaging_cost", "overhead_cost", "labor_cost", "created_at", "updated_at", "deleted_at", "category_name", "store_default_tax"}
		rows := sqlmock.NewRows(cols).
			AddRow("m1", "s1", nil, "Coffee", nil, 20000.0, true, nil, "img.png", true, 1000.0, 500.0, 2000.0, time.Now(), time.Now(), nil, "None", 10.0)

		mock.ExpectQuery(`(?is)SELECT .* FROM menu_items mi .* INNER JOIN stores s .* WHERE mi.id = \$1`).
			WithArgs("m1").
			WillReturnRows(rows)

		ingRows := sqlmock.NewRows([]string{"id", "menu_item_id", "product_id", "quantity", "product_name", "product_sku", "unit", "cost_price"}).
			AddRow("msi1", "m1", "p1", 10.0, "Beans", "SKU1", "gram", 100.0)
		mock.ExpectQuery(`(?is)SELECT .* FROM menu_item_ingredients`).WithArgs(sqlmock.AnyArg()).WillReturnRows(ingRows)

		res, err := repo.FindByID(context.Background(), "m1")
		require.NoError(t, err)
		require.NotNil(t, res)

		// Not Found
		mock.ExpectQuery(`(?is)SELECT .* FROM menu_items mi .*`).WithArgs("unknown").WillReturnError(sql.ErrNoRows)
		res, err = repo.FindByID(context.Background(), "unknown")
		assert.NoError(t, err)
		assert.Nil(t, res)
	})
}

func TestMenuItemRepo_Manage(t *testing.T) {
	stringPtr := func(s string) *string { return &s }

	t.Run("Create", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer func() { _ = db.Close() }()
		repo := NewMenuItemRepo(sqlx.NewDb(db, "postgres"))

		rows := sqlmock.NewRows([]string{"id", "store_id", "category_id", "name", "description", "sell_price", "use_global_tax", "tax_percentage", "packaging_cost", "overhead_cost", "labor_cost", "image_url", "is_active", "created_at", "updated_at", "deleted_at"}).
			AddRow("m1", "s1", "c1", "Coffee", nil, 20000.0, true, nil, 1000.0, 500.0, 2000.0, nil, true, time.Now(), time.Now(), nil)
		mock.ExpectQuery(`(?is)INSERT INTO menu_items`).
			WithArgs("s1", "c1", "Coffee", nil, 20000.0, true, nil, 1000.0, 500.0, 2000.0, nil).
			WillReturnRows(rows)

		res, err := repo.Create(context.Background(), &domain.MenuItem{
			StoreID:       "s1",
			CategoryID:    stringPtr("c1"),
			Name:          "Coffee",
			SellPrice:     20000.0,
			UseGlobalTax:  true,
			PackagingCost: 1000.0,
			OverheadCost:  500.0,
			LaborCost:     2000.0,
		})
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("Update", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer func() { _ = db.Close() }()
		repo := NewMenuItemRepo(sqlx.NewDb(db, "postgres"))

		rows := sqlmock.NewRows([]string{"id", "store_id", "category_id", "name", "description", "sell_price", "use_global_tax", "tax_percentage", "packaging_cost", "overhead_cost", "labor_cost", "image_url", "is_active", "created_at", "updated_at", "deleted_at"}).
			AddRow("m1", "s1", "c1", "New Coffee", "Desc", 25000.0, true, nil, 1000.0, 500.0, 2000.0, nil, true, time.Now(), time.Now(), nil)
		mock.ExpectQuery(`(?is)UPDATE menu_items SET`).
			WithArgs("c1", "New Coffee", "Desc", 25000.0, true, nil, 1000.0, 500.0, 2000.0, nil, true, "m1").
			WillReturnRows(rows)

		res, err := repo.Update(context.Background(), &domain.MenuItem{
			ID:            "m1",
			CategoryID:    stringPtr("c1"),
			Name:          "New Coffee",
			Description:   stringPtr("Desc"),
			SellPrice:     25000.0,
			UseGlobalTax:  true,
			IsActive:      true,
			PackagingCost: 1000.0,
			OverheadCost:  500.0,
			LaborCost:     2000.0,
		})
		require.NoError(t, err)
		require.NotNil(t, res)
	})
}

func TestMenuItemRepo_Manage_Ingredients(t *testing.T) {
	t.Run("ReplaceIngredients", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer func() { _ = db.Close() }()
		repo := NewMenuItemRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectBegin()
		mock.ExpectExec(`(?is)DELETE FROM menu_item_ingredients WHERE menu_item_id = \$1`).WithArgs("m1").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`(?is)INSERT INTO menu_item_ingredients`).WithArgs("m1", "p1", 10.0).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.ReplaceIngredients(context.Background(), "m1", []domain.MenuItemIngredient{{ProductID: "p1", Quantity: 10.0}})
		assert.NoError(t, err)
	})

	t.Run("SoftDelete", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer func() { _ = db.Close() }()
		repo := NewMenuItemRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectExec(`(?is)UPDATE menu_items SET deleted_at=NOW()`).WithArgs("m1").WillReturnResult(sqlmock.NewResult(1, 1))
		err := repo.SoftDelete(context.Background(), "m1")
		assert.NoError(t, err)

		// Not Found
		mock.ExpectExec(`(?is)UPDATE menu_items SET deleted_at=NOW()`).WillReturnResult(sqlmock.NewResult(0, 0))
		err = repo.SoftDelete(context.Background(), "unknown")
		assert.Error(t, err)
	})
}
