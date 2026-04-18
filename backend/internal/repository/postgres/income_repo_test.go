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

func strPtr(s string) *string {
	return &s
}

func TestIncomeRepo_Categories(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewIncomeRepo(sqlxDB)
	ctx := context.Background()

	t.Run("ListCategories", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "description", "is_active", "created_at", "updated_at"}).
			AddRow("c1", "Cat 1", "Desc 1", true, time.Now(), time.Now())
		mock.ExpectQuery(`(?is)SELECT .* FROM income_categories`).WillReturnRows(rows)

		res, err := repo.ListCategories(ctx, false)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("CreateCategory", func(t *testing.T) {
		mock.ExpectQuery(`(?is)INSERT INTO income_categories`).
			WithArgs("New Cat", "New Desc").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "is_active", "created_at", "updated_at"}).
				AddRow("c2", "New Cat", strPtr("New Desc"), true, time.Now(), time.Now()))

		res, err := repo.CreateCategory(ctx, &domain.IncomeCategory{Name: "New Cat", Description: strPtr("New Desc")})
		assert.NoError(t, err)
		assert.Equal(t, "c2", res.ID)
	})

	t.Run("GetCategoryByID", func(t *testing.T) {
		mock.ExpectQuery(`(?is)SELECT .* FROM income_categories WHERE id = \$1`).
			WithArgs("c1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("c1", "Cat 1"))

		res, err := repo.GetCategoryByID(ctx, "c1")
		assert.NoError(t, err)
		assert.NotNil(t, res)

		// Not Found
		mock.ExpectQuery(`(?is)SELECT .* FROM income_categories WHERE id = \$1`).WithArgs("unknown").WillReturnError(sql.ErrNoRows)
		res, err = repo.GetCategoryByID(ctx, "unknown")
		assert.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run("UpdateCategory", func(t *testing.T) {
		mock.ExpectQuery(`(?is)UPDATE income_categories`).
			WithArgs("Updated", "Desc", true, "c1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("c1", "Updated"))

		res, err := repo.UpdateCategory(ctx, "c1", "Updated", "Desc", true)
		assert.NoError(t, err)
		assert.NotNil(t, res)

		// Not Found
		mock.ExpectQuery(`(?is)UPDATE income_categories`).WillReturnError(sql.ErrNoRows)
		res, err = repo.UpdateCategory(ctx, "unknown", "Updated", "Desc", true)
		assert.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run("SoftDeleteCategory", func(t *testing.T) {
		mock.ExpectExec(`(?is)UPDATE income_categories .* deleted_at`).
			WithArgs("c1").
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.SoftDeleteCategory(ctx, "c1")
		assert.NoError(t, err)

		// Not Found
		mock.ExpectExec(`(?is)UPDATE income_categories .* deleted_at`).WillReturnResult(sqlmock.NewResult(0, 0))
		err = repo.SoftDeleteCategory(ctx, "unknown")
		assert.Error(t, err)
	})
}

func TestIncomeRepo_Incomes(t *testing.T) {
	ctx := context.Background()

	t.Run("Create", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()
		repo := NewIncomeRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectQuery(`(?is)INSERT INTO incomes`).
			WithArgs("s1", "c1", 1000.0, sqlmock.AnyArg(), "cash", "ref1", "note1", "u1").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("i1"))

		mock.ExpectQuery(`(?is)SELECT .* FROM incomes i .* WHERE i.id = \$1`).
			WithArgs("i1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "amount"}).AddRow("i1", 1000.0))

		res, err := repo.Create(ctx, &domain.Income{
			StoreID: "s1", CategoryID: "c1", Amount: 1000.0,
			IncomeDate: time.Now(), PaymentMethod: "cash",
			Reference: strPtr("ref1"), Notes: strPtr("note1"), CreatedBy: strPtr("u1"),
		})
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("FindAll", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()
		repo := NewIncomeRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectQuery(`(?is)SELECT COUNT\(\*\) FROM incomes`).
			WithArgs("s1").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		mock.ExpectQuery(`(?is)SELECT .* FROM incomes i .* LIMIT \$2 OFFSET \$3`).
			WithArgs("s1", 20, 0).
			WillReturnRows(sqlmock.NewRows([]string{"id", "amount", "category_name", "created_by_name"}).
				AddRow("i1", 1000.0, "Cat 1", strPtr("User 1")))

		res, total, err := repo.FindAll(ctx, dto.IncomeListFilter{StoreID: "s1"})
		assert.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, res, 1)

		// With Filters
		filter := dto.IncomeListFilter{
			StoreID:    "s1",
			CategoryID: "c1",
			DateFrom:   "2024-01-01",
			DateTo:     "2024-01-31",
		}
		mock.ExpectQuery(`(?is)SELECT COUNT\(\*\) FROM incomes`).
			WithArgs("s1", "c1", "2024-01-01", "2024-01-31").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectQuery(`(?is)SELECT .* FROM incomes i .* LIMIT \$5 OFFSET \$6`).
			WithArgs("s1", "c1", "2024-01-01", "2024-01-31", 20, 0).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("i1"))

		_, _, err = repo.FindAll(ctx, filter)
		assert.NoError(t, err)
	})

	t.Run("FindByID", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()
		repo := NewIncomeRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectQuery(`(?is)SELECT .* FROM incomes i .* WHERE i.id = \$1`).
			WithArgs("i1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "amount", "category_name", "created_by_name"}).
				AddRow("i1", 1000.0, "Cat 1", strPtr("User 1")))

		res, err := repo.FindByID(ctx, "i1")
		assert.NoError(t, err)
		assert.NotNil(t, res)

		// Not Found
		mock.ExpectQuery(`(?is)SELECT .* FROM incomes i .* WHERE i.id = \$1`).
			WithArgs("unknown").
			WillReturnError(sql.ErrNoRows)

		res, err = repo.FindByID(ctx, "unknown")
		assert.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run("Update", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()
		repo := NewIncomeRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectQuery(`(?is)UPDATE incomes`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("i1"))

		mock.ExpectQuery(`(?is)SELECT .* FROM incomes i .* WHERE i.id = \$1`).
			WithArgs("i1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "amount"}).AddRow("i1", 1200.0))

		res, err := repo.Update(ctx, &domain.Income{ID: "i1", StoreID: "s1", Amount: 1200.0})
		assert.NoError(t, err)
		assert.NotNil(t, res)

		// Not Found
		mock.ExpectQuery(`(?is)UPDATE incomes`).WillReturnError(sql.ErrNoRows)
		res, err = repo.Update(ctx, &domain.Income{ID: "unknown", StoreID: "s1"})
		assert.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run("Delete", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()
		repo := NewIncomeRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectExec(`(?is)DELETE FROM incomes`).
			WithArgs("i1", "s1").
			WillReturnResult(sqlmock.NewResult(1, 1))

		err = repo.Delete(ctx, "i1", "s1")
		assert.NoError(t, err)

		// Not Found
		mock.ExpectExec(`(?is)DELETE FROM incomes`).WillReturnResult(sqlmock.NewResult(0, 0))
		err = repo.Delete(ctx, "unknown", "s1")
		assert.Error(t, err)
	})

	t.Run("SumByDateRange", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()
		repo := NewIncomeRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectQuery(`(?is)SELECT COALESCE\(SUM\(amount\), 0\) FROM incomes`).
			WithArgs("s1", "2024-01-01", "2024-01-02").
			WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(5000.0))

		total, sumErr := repo.SumByDateRange(ctx, "s1", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC))
		assert.NoError(t, sumErr)
		assert.Equal(t, 5000.0, total)
	})
}
