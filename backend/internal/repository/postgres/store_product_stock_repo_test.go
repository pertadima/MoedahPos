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

func TestStoreRepo(t *testing.T) {
	ctx := context.Background()

	t.Run("Create", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()
		repo := NewStoreRepo(sqlx.NewDb(db, "postgres"))

		s := &domain.Store{Name: "Store 1", StoreType: "retail"}
		rows := sqlmock.NewRows([]string{"id", "name", "address", "phone", "tax_number", "currency", "store_type", "default_tax_percentage", "is_active", "created_at", "updated_at", "deleted_at"}).
			AddRow("st1", "Store 1", "", "", "", "IDR", "retail", 10.0, true, time.Now(), time.Now(), nil)

		mock.ExpectQuery(`INSERT INTO stores`).WithArgs(s.Name, s.Address, s.Phone, s.TaxNumber, s.Currency, s.StoreType, s.DefaultTaxPercentage).WillReturnRows(rows)

		res, err := repo.Create(ctx, s)
		assert.NoError(t, err)
		assert.Equal(t, "st1", res.ID)
	})

	t.Run("FindMember", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()
		repo := NewStoreRepo(sqlx.NewDb(db, "postgres"))

		rows := sqlmock.NewRows([]string{"id", "user_id", "store_id", "role_id", "is_active", "created_at"}).
			AddRow("m1", "u1", "st1", "admin", true, time.Now())

		mock.ExpectQuery(`SELECT .* FROM user_stores`).WithArgs("u1", "st1").WillReturnRows(rows)

		res, err := repo.FindMember(ctx, "u1", "st1")
		assert.NoError(t, err)
		assert.Equal(t, "admin", res.RoleID)

		// Not Found
		mock.ExpectQuery(`SELECT .* FROM user_stores`).WithArgs("unknown", "st1").WillReturnError(sql.ErrNoRows)
		res, err = repo.FindMember(ctx, "unknown", "st1")
		assert.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run("FindAll", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()
		repo := NewStoreRepo(sqlx.NewDb(db, "postgres"))

		f := dto.StoreListFilter{Search: "S1"}
		f.Defaults()

		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM stores s`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectQuery(`SELECT .* FROM stores s`).WithArgs("%S1%", 20, 0).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("st1", "S1"))

		res, total, err := repo.FindAll(ctx, f)
		assert.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, res, 1)
	})

	t.Run("Update", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()
		repo := NewStoreRepo(sqlx.NewDb(db, "postgres"))

		s := &domain.Store{ID: "st1", Name: "S1 Updated"}
		mock.ExpectQuery(`UPDATE stores`).WithArgs(s.Name, s.Address, s.Phone, s.TaxNumber, s.Currency, s.StoreType, s.DefaultTaxPercentage, s.IsActive, s.ID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("st1", "S1 Updated"))

		res, err := repo.Update(ctx, s)
		assert.NoError(t, err)
		assert.Equal(t, "S1 Updated", res.Name)

		// Not Found
		mock.ExpectQuery(`UPDATE stores`).WillReturnError(sql.ErrNoRows)
		res, err = repo.Update(ctx, &domain.Store{ID: "unknown"})
		assert.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run("MemberManagement", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()
		repo := NewStoreRepo(sqlx.NewDb(db, "postgres"))

		// ListMembers
		mock.ExpectQuery(`SELECT .* FROM user_stores us`).WithArgs("st1").WillReturnRows(sqlmock.NewRows([]string{"user_id", "user_name"}).AddRow("u1", "U1"))
		members, err := repo.ListMembers(ctx, "st1")
		assert.NoError(t, err)
		assert.Len(t, members, 1)

		// AddMember
		mock.ExpectExec(`INSERT INTO user_stores`).WithArgs("u1", "st1", "admin").WillReturnResult(sqlmock.NewResult(1, 1))
		err = repo.AddMember(ctx, &domain.UserStore{UserID: "u1", StoreID: "st1", RoleID: "admin"})
		assert.NoError(t, err)

		// UpdateMemberRole
		mock.ExpectExec(`UPDATE user_stores SET role_id`).WithArgs("staff", "u1", "st1").WillReturnResult(sqlmock.NewResult(1, 1))
		err = repo.UpdateMemberRole(ctx, "u1", "st1", "staff")
		assert.NoError(t, err)

		// DeactivateMember
		mock.ExpectExec(`UPDATE user_stores SET is_active = false`).WithArgs("u1", "st1").WillReturnResult(sqlmock.NewResult(1, 1))
		err = repo.DeactivateMember(ctx, "u1", "st1")
		assert.NoError(t, err)

		// Not Found
		mock.ExpectExec(`UPDATE user_stores SET is_active = false`).WillReturnResult(sqlmock.NewResult(0, 0))
		err = repo.DeactivateMember(ctx, "unknown", "st1")
		assert.Error(t, err)
	})

	t.Run("FindByID", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()
		repo := NewStoreRepo(sqlx.NewDb(db, "postgres"))

		rows := sqlmock.NewRows([]string{"id", "name"}).AddRow("st1", "Store 1")
		mock.ExpectQuery(`SELECT .* FROM stores WHERE id = \$1`).WithArgs("st1").WillReturnRows(rows)

		res, err := repo.FindByID(ctx, "st1")
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "st1", res.ID)

		// Not Found
		mock.ExpectQuery(`SELECT .* FROM stores WHERE id = \$1`).WithArgs("unknown").WillReturnError(sql.ErrNoRows)
		res, err = repo.FindByID(ctx, "unknown")
		assert.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run("SoftDelete", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()
		repo := NewStoreRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectExec(`UPDATE stores SET deleted_at = NOW\(\)`).WithArgs("st1").WillReturnResult(sqlmock.NewResult(1, 1))

		err = repo.SoftDelete(ctx, "st1")
		assert.NoError(t, err)

		// Not Found
		mock.ExpectExec(`UPDATE stores SET deleted_at = NOW\(\)`).WillReturnResult(sqlmock.NewResult(0, 0))
		err = repo.SoftDelete(ctx, "unknown")
		assert.Error(t, err)
	})
}

func TestCategoryRepo(t *testing.T) {
	ctx := context.Background()

	t.Run("CRUD", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()
		repo := NewCategoryRepo(sqlx.NewDb(db, "postgres"))

		// Create
		c := &domain.Category{StoreID: "st1", Name: "Cat 1"}
		mock.ExpectQuery(`INSERT INTO categories`).WithArgs(c.StoreID, c.Name, c.ParentID).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("c1"))
		res, _ := repo.Create(ctx, c)
		assert.Equal(t, "c1", res.ID)

		// FindAllByStore
		mock.ExpectQuery(`SELECT .* FROM categories`).WithArgs("st1").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("c1", "Cat 1"))
		cats, _ := repo.FindAllByStore(ctx, "st1")
		assert.Len(t, cats, 1)

		// Update
		c.ID = "c1"
		c.Name = "Cat 1 Updated"
		mock.ExpectQuery(`UPDATE categories`).WithArgs(c.Name, c.ParentID, c.ID).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("c1", "Cat 1 Updated"))
		res, _ = repo.Update(ctx, c)
		assert.Equal(t, "Cat 1 Updated", res.Name)

		// Not Found
		mock.ExpectQuery(`UPDATE categories`).WillReturnError(sql.ErrNoRows)
		res, err = repo.Update(ctx, &domain.Category{ID: "unknown"})
		assert.NoError(t, err)
		assert.Nil(t, res)

		// SoftDelete
		mock.ExpectExec(`UPDATE categories SET deleted_at`).WithArgs("c1").WillReturnResult(sqlmock.NewResult(1, 1))
		err = repo.SoftDelete(ctx, "c1")
		assert.NoError(t, err)

		// Not Found
		mock.ExpectExec(`UPDATE categories SET deleted_at`).WillReturnResult(sqlmock.NewResult(0, 0))
		err = repo.SoftDelete(ctx, "unknown")
		assert.Error(t, err)
	})

	t.Run("FindByID", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()
		repo := NewCategoryRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectQuery(`SELECT .* FROM categories WHERE id = \$1`).WithArgs("c1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("c1", "Cat 1"))

		res, err := repo.FindByID(ctx, "c1")
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "c1", res.ID)

		// Not Found
		mock.ExpectQuery(`SELECT .* FROM categories WHERE id = \$1`).WillReturnError(sql.ErrNoRows)
		res, err = repo.FindByID(ctx, "unknown")
		assert.NoError(t, err)
		assert.Nil(t, res)
	})
}

func TestProductRepo(t *testing.T) {
	ctx := context.Background()

	t.Run("CRUD", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()
		repo := NewProductRepo(sqlx.NewDb(db, "postgres"))

		// FindByID
		cols := []string{"id", "store_id", "category_id", "sku", "name", "description", "barcode", "unit", "cost_price", "sell_price", "use_global_tax", "tax_percentage", "image_url", "is_active", "created_at", "updated_at", "deleted_at", "category_name", "store_default_tax"}
		mock.ExpectQuery(`SELECT .* FROM products`).WithArgs("p1").
			WillReturnRows(sqlmock.NewRows(cols).AddRow("p1", "st1", nil, "SKU1", "P1", nil, nil, "pcs", 0.0, 0.0, true, nil, nil, true, time.Now(), time.Now(), nil, nil, 10.0))
		p, err := repo.FindByID(ctx, "p1")
		require.NoError(t, err, "FindByID should not return error")
		require.NotNil(t, p, "Product should not be nil")
		assert.Equal(t, "p1", p.ID)

		// Not Found
		mock.ExpectQuery(`SELECT .* FROM products`).WithArgs("unknown").WillReturnError(sql.ErrNoRows)
		p, err = repo.FindByID(ctx, "unknown")
		assert.NoError(t, err)
		assert.Nil(t, p)

		// Update
		upP := &domain.Product{ID: "p1", Name: "P1 Updated"}
		mock.ExpectQuery(`UPDATE products`).WithArgs(upP.CategoryID, upP.Name, upP.Description, upP.Barcode, upP.Unit, upP.CostPrice, upP.SellPrice, upP.UseGlobalTax, upP.TaxPercentage, upP.ImageURL, upP.IsActive, upP.ID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("p1", "P1 Updated"))
		res, _ := repo.Update(ctx, upP)
		assert.Equal(t, "P1 Updated", res.Name)

		// Not Found
		mock.ExpectQuery(`UPDATE products`).WillReturnError(sql.ErrNoRows)
		res, err = repo.Update(ctx, &domain.Product{ID: "unknown"})
		assert.NoError(t, err)
		assert.Nil(t, res)

		// SoftDelete
		mock.ExpectExec(`UPDATE products SET deleted_at`).WithArgs("p1").WillReturnResult(sqlmock.NewResult(1, 1))
		err = repo.SoftDelete(ctx, "p1")
		assert.NoError(t, err)

		// Not Found
		mock.ExpectExec(`UPDATE products SET deleted_at`).WillReturnResult(sqlmock.NewResult(0, 0))
		err = repo.SoftDelete(ctx, "unknown")
		assert.Error(t, err)
	})
}

func TestStockRepo(t *testing.T) {
	ctx := context.Background()

	t.Run("Movements", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()
		repo := NewStockRepo(sqlx.NewDb(db, "postgres"))

		f := dto.StockMovementFilter{StoreID: "st1"}
		f.Defaults()

		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM stock_movements sm`).WithArgs("st1").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectQuery(`SELECT .* FROM stock_movements sm`).WithArgs("st1", 20, 0).WillReturnRows(sqlmock.NewRows([]string{"id", "product_id"}).AddRow("sm1", "p1"))

		res, total, err := repo.FindMovements(ctx, f)
		assert.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, res, 1)
	})

	t.Run("Adjust", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()
		repo := NewStockRepo(sqlx.NewDb(db, "postgres"))

		input := domain.AdjustInput{ProductID: "p1", StoreID: "st1", Delta: 10.0}
		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO stock_movements`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectQuery(`INSERT INTO stock_levels .* RETURNING`).WithArgs("p1", "st1", 10.0).
			WillReturnRows(sqlmock.NewRows([]string{"id", "product_id", "store_id", "quantity", "min_quantity"}).AddRow("sl1", "p1", "st1", 10.0, 0.0))
		mock.ExpectCommit()
		mock.ExpectQuery(`SELECT name .* FROM products`).WithArgs("p1").WillReturnRows(sqlmock.NewRows([]string{"name", "sku", "unit"}).AddRow("P1", "S1", "pcs"))

		res, err := repo.Adjust(ctx, input)
		assert.NoError(t, err)
		assert.Equal(t, 10.0, res.Quantity)
	})

	t.Run("DeductStock", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()
		repo := NewStockRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO stock_movements`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`UPDATE stock_levels`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err = repo.DeductStock(ctx, "p1", "st1", 5.0, "ref1", "c1")
		assert.NoError(t, err)
	})

	t.Run("FindLevelByProduct", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()
		repo := NewStockRepo(sqlx.NewDb(db, "postgres"))

		// Not Found
		mock.ExpectQuery(`SELECT .* FROM stock_levels sl`).WithArgs("p_unknown", "st1").WillReturnError(sql.ErrNoRows)
		res, err := repo.FindLevelByProduct(ctx, "p_unknown", "st1")
		assert.NoError(t, err)
		assert.Nil(t, res)
	})
}
