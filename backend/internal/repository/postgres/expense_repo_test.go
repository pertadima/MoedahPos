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

func strPtrExp(s string) *string {
	return &s
}

func TestExpenseRepo_Categories(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewExpenseRepo(sqlxDB)
	ctx := context.Background()

	t.Run("ListCategories", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "description", "is_active", "created_at", "updated_at"}).
			AddRow("c1", "Cat 1", "Desc 1", true, time.Now(), time.Now())
		mock.ExpectQuery(`(?is)SELECT .* FROM expense_categories`).WillReturnRows(rows)

		res, err := repo.ListCategories(ctx, false)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("CreateCategory", func(t *testing.T) {
		mock.ExpectQuery(`(?is)INSERT INTO expense_categories`).
			WithArgs("New Cat", "New Desc").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "is_active", "created_at", "updated_at"}).
				AddRow("c2", "New Cat", "New Desc", true, time.Now(), time.Now()))

		res, err := repo.CreateCategory(ctx, &domain.ExpenseCategory{Name: "New Cat", Description: "New Desc"})
		assert.NoError(t, err)
		assert.Equal(t, "c2", res.ID)
	})

	t.Run("GetCategoryByID", func(t *testing.T) {
		mock.ExpectQuery(`(?is)SELECT .* FROM expense_categories WHERE id = \$1`).
			WithArgs("c1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "is_active", "created_at", "updated_at"}).
				AddRow("c1", "Cat 1", "Desc 1", true, time.Now(), time.Now()))

		res, err := repo.GetCategoryByID(ctx, "c1")
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "c1", res.ID)

		// Not Found
		mock.ExpectQuery(`(?is)SELECT .* FROM expense_categories WHERE id = \$1`).
			WithArgs("unknown").
			WillReturnError(sql.ErrNoRows)
		res, err = repo.GetCategoryByID(ctx, "unknown")
		assert.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run("UpdateCategory", func(t *testing.T) {
		mock.ExpectQuery(`(?is)UPDATE expense_categories`).
			WithArgs("Updated", "Desc", true, "c1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "is_active", "created_at", "updated_at"}).
				AddRow("c1", "Updated", "Desc", true, time.Now(), time.Now()))

		res, err := repo.UpdateCategory(ctx, "c1", "Updated", "Desc", true)
		assert.NoError(t, err)
		assert.NotNil(t, res)

		// Not Found
		mock.ExpectQuery(`(?is)UPDATE expense_categories`).
			WillReturnError(sql.ErrNoRows)
		res, err = repo.UpdateCategory(ctx, "c1", "Updated", "Desc", true)
		assert.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run("SoftDeleteCategory", func(t *testing.T) {
		mock.ExpectExec(`(?is)UPDATE expense_categories .* deleted_at`).
			WithArgs("c1").
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.SoftDeleteCategory(ctx, "c1")
		assert.NoError(t, err)

		// Not Found
		mock.ExpectExec(`(?is)UPDATE expense_categories`).
			WillReturnResult(sqlmock.NewResult(0, 0))
		err = repo.SoftDeleteCategory(ctx, "c1")
		assert.Error(t, err)
	})
}

func TestExpenseRepo_Expenses_Basic(t *testing.T) {
	ctx := context.Background()

	t.Run("CreateExpense", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()
		repo := NewExpenseRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectQuery(`(?is)INSERT INTO expenses`).
			WithArgs("s1", "c1", 1000.0, sqlmock.AnyArg(), "note1", "pending", "u1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "store_id", "category_id"}).AddRow("e1", "s1", "c1"))

		mock.ExpectQuery(`(?is)SELECT name FROM expense_categories WHERE id = \$1`).
			WithArgs("c1").
			WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("Cat 1"))

		res, err := repo.CreateExpense(ctx, &domain.Expense{
			StoreID: "s1", CategoryID: "c1", Amount: 1000.0,
			ExpenseDate: time.Now(), Notes: "note1", PaymentStatus: "pending", CreatedBy: strPtrExp("u1"),
		})
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "Cat 1", res.CategoryName)
	})

	t.Run("FindAll", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()
		repo := NewExpenseRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectQuery(`(?is)SELECT COUNT\(\*\) FROM expenses e`).
			WithArgs("s1").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		mock.ExpectQuery(`(?is)SELECT .* FROM expenses e .* LIMIT \$2 OFFSET \$3`).
			WithArgs("s1", 20, 0).
			WillReturnRows(sqlmock.NewRows([]string{"id", "store_id", "category_id", "category_name", "amount"}).
				AddRow("e1", "s1", "c1", "Cat 1", 1000.0))

		res, total, err := repo.FindAll(ctx, dto.ExpenseListFilter{PaginationQuery: dto.PaginationQuery{PerPage: 20, Page: 1}, StoreID: "s1"})
		assert.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, res, 1)
	})

	t.Run("GetByID", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()
		repo := NewExpenseRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectQuery(`(?is)SELECT .* FROM expenses e .* WHERE e.id = \$1 AND e.store_id = \$2`).
			WithArgs("e1", "s1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "store_id", "category_id", "category_name", "amount"}).
				AddRow("e1", "s1", "c1", "Cat 1", 1000.0))

		res, err := repo.GetByID(ctx, "e1", "s1")
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "e1", res.ID)

		// Not Found
		mock.ExpectQuery(`(?is)SELECT .* FROM expenses e .* WHERE e.id = \$1`).
			WillReturnError(sql.ErrNoRows)
		res, err = repo.GetByID(ctx, "unknown", "s1")
		assert.NoError(t, err)
		assert.Nil(t, res)
	})
}

func TestExpenseRepo_Expenses_Update(t *testing.T) {
	ctx := context.Background()

	t.Run("Update", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()
		repo := NewExpenseRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectQuery(`(?is)UPDATE expenses`).
			WithArgs("c1", 1200.0, sqlmock.AnyArg(), "updated note", "e1", "s1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "store_id", "category_id", "amount"}).AddRow("e1", "s1", "c1", 1200.0))

		mock.ExpectQuery(`(?is)SELECT name FROM expense_categories WHERE id = \$1`).
			WithArgs("c1").
			WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("Cat 1"))

		res, err := repo.Update(ctx, &domain.Expense{ID: "e1", StoreID: "s1", CategoryID: "c1", Amount: 1200.0, ExpenseDate: time.Now(), Notes: "updated note"})
		assert.NoError(t, err)
		assert.NotNil(t, res)

		// Not Found
		mock.ExpectQuery(`(?is)UPDATE expenses`).
			WillReturnError(sql.ErrNoRows)
		res, err = repo.Update(ctx, &domain.Expense{ID: "unknown", StoreID: "s1"})
		assert.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run("Delete", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()
		repo := NewExpenseRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectExec(`(?is)DELETE FROM expenses WHERE id = \$1 AND store_id = \$2`).
			WithArgs("e1", "s1").
			WillReturnResult(sqlmock.NewResult(1, 1))

		err = repo.Delete(ctx, "e1", "s1")
		assert.NoError(t, err)

		// Not Found
		mock.ExpectExec(`(?is)DELETE FROM expenses`).
			WillReturnResult(sqlmock.NewResult(0, 0))
		err = repo.Delete(ctx, "unknown", "s1")
		assert.Error(t, err)
	})

	t.Run("UpdatePaymentStatus", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()
		repo := NewExpenseRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectQuery(`(?is)UPDATE expenses SET payment_status = \$1`).
			WithArgs("paid", "e1", "s1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "payment_status"}).AddRow("e1", "paid"))

		mock.ExpectQuery(`(?is)SELECT name FROM expense_categories`).
			WithArgs(sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("Cat 1"))

		res, err := repo.UpdatePaymentStatus(ctx, "e1", "s1", "paid")
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "paid", res.PaymentStatus)

		// Not Found
		mock.ExpectQuery(`(?is)UPDATE expenses SET payment_status = \$1`).
			WillReturnError(sql.ErrNoRows)
		res, err = repo.UpdatePaymentStatus(ctx, "unknown", "s1", "paid")
		assert.NoError(t, err)
		assert.Nil(t, res)
	})
}

func TestExpenseRepo_Recurring_Basic(t *testing.T) {
	ctx := context.Background()

	t.Run("CreateRecurring", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()
		repo := NewExpenseRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectQuery(`(?is)INSERT INTO recurring_expenses`).
			WithArgs("s1", "c1", "Recur", 500.0, "monthly", 1, "2024-01-01", nil, "2024-02-01", "notes", true, strPtrExp("u1")).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "category_id"}).AddRow("re1", "Recur", "c1"))

		mock.ExpectQuery(`(?is)SELECT name FROM expense_categories`).
			WithArgs("c1").
			WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("Cat 1"))

		res, err := repo.CreateRecurringExpense(ctx, &domain.RecurringExpense{
			StoreID: "s1", CategoryID: "c1", Name: "Recur", Amount: 500.0,
			Interval: "monthly", IntervalValue: 1, StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			NextRunDate: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
			Notes:       "notes", IsActive: true, CreatedBy: strPtrExp("u1"),
		})
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("FindAllRecurring", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()
		repo := NewExpenseRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectQuery(`(?is)SELECT COUNT\(\*\) FROM recurring_expenses re`).
			WithArgs("s1").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		mock.ExpectQuery(`(?is)SELECT .* FROM recurring_expenses re`).
			WithArgs("s1", 10, 0).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "category_name"}).AddRow("re1", "Recur", "Cat 1"))

		res, total, err := repo.FindAllRecurring(ctx, dto.ExpenseListFilter{StoreID: "s1", PaginationQuery: dto.PaginationQuery{PerPage: 10, Page: 1}})
		assert.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, res, 1)
	})

	t.Run("GetRecurringByID", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()
		repo := NewExpenseRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectQuery(`(?is)SELECT .* FROM recurring_expenses re .* WHERE re.id = \$1`).
			WithArgs("re1", "s1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("re1", "Recur"))

		res, err := repo.GetRecurringByID(ctx, "re1", "s1")
		assert.NoError(t, err)
		assert.NotNil(t, res)

		// Not Found
		mock.ExpectQuery(`(?is)SELECT .* FROM recurring_expenses re .* WHERE re.id = \$1`).
			WillReturnError(sql.ErrNoRows)
		res, err = repo.GetRecurringByID(ctx, "unknown", "s1")
		assert.NoError(t, err)
		assert.Nil(t, res)
	})
}

func TestExpenseRepo_Recurring_Manage(t *testing.T) {
	ctx := context.Background()

	t.Run("UpdateRecurring", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()
		repo := NewExpenseRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectQuery(`(?is)UPDATE recurring_expenses`).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "category_id"}).AddRow("re1", "Updated", "c1"))

		mock.ExpectQuery(`(?is)SELECT name FROM expense_categories`).
			WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("Cat 1"))

		res, err := repo.UpdateRecurring(ctx, &domain.RecurringExpense{ID: "re1", StoreID: "s1", CategoryID: "c1", Name: "Updated", StartDate: time.Now()})
		assert.NoError(t, err)
		assert.NotNil(t, res)

		// Not Found
		mock.ExpectQuery(`(?is)UPDATE recurring_expenses`).
			WillReturnError(sql.ErrNoRows)
		res, err = repo.UpdateRecurring(ctx, &domain.RecurringExpense{ID: "unknown", StoreID: "s1"})
		assert.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run("DeleteRecurring", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()
		repo := NewExpenseRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectExec(`(?is)DELETE FROM recurring_expenses`).
			WithArgs("re1", "s1").
			WillReturnResult(sqlmock.NewResult(1, 1))

		err = repo.DeleteRecurring(ctx, "re1", "s1")
		assert.NoError(t, err)

		// Not Found
		mock.ExpectExec(`(?is)DELETE FROM recurring_expenses`).
			WillReturnResult(sqlmock.NewResult(0, 0))
		err = repo.DeleteRecurring(ctx, "unknown", "s1")
		assert.Error(t, err)
	})

	t.Run("GetDueRecurringExpenses", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()
		repo := NewExpenseRepo(sqlx.NewDb(db, "postgres"))

		cols := []string{"id", "store_id", "category_id", "category_name", "name", "amount", "interval", "interval_value", "start_date", "end_date", "next_run_date", "notes", "is_active", "created_by", "created_at", "updated_at", "last_generated_at"}
		mock.ExpectQuery(`(?is)SELECT .* FROM recurring_expenses re`).
			WillReturnRows(sqlmock.NewRows(cols).AddRow("re1", "s1", "c1", "Cat 1", "Due Recur", 500.0, "daily", 1, time.Now(), nil, time.Now(), "", true, strPtrExp("u1"), time.Now(), time.Now(), nil))

		res, err := repo.GetDueRecurringExpenses(ctx)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("BumpRecurringNextRun", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = db.Close() }()
		repo := NewExpenseRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectExec(`(?is)UPDATE recurring_expenses SET next_run_date`).
			WithArgs("2024-03-01", "re1").
			WillReturnResult(sqlmock.NewResult(1, 1))

		err = repo.BumpRecurringNextRun(ctx, "re1", "2024-03-01")
		assert.NoError(t, err)
	})
}
