package postgres

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
)

func TestExpenseRepo_CreateCategory(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewExpenseRepo(sqlxDB)

	c := &domain.ExpenseCategory{
		Name:        "Utilities",
		Description: "Office utilities",
	}

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "description", "is_active", "created_at", "updated_at"}).
			AddRow("cat1", c.Name, c.Description, true, time.Now(), time.Now())

		mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO expense_categories (name, description) VALUES ($1, $2) RETURNING id, name, description, is_active, created_at, updated_at")).
			WithArgs(c.Name, c.Description).
			WillReturnRows(rows)

		result, err := repo.CreateCategory(context.Background(), c)
		assert.NoError(t, err)
		assert.Equal(t, "cat1", result.ID)
	})
}

func TestExpenseRepo_CreateExpense(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewExpenseRepo(sqlxDB)

	userID := "u1"
	e := &domain.Expense{
		StoreID:       "st1",
		CategoryID:    "cat1",
		Amount:        100.0,
		ExpenseDate:   time.Now(),
		Notes:         "Lunch",
		PaymentStatus: "paid",
		CreatedBy:     &userID,
	}

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "store_id", "category_id", "amount", "expense_date", "notes", "payment_status", "created_by", "created_at", "updated_at"}).
			AddRow("e1", e.StoreID, e.CategoryID, e.Amount, e.ExpenseDate, e.Notes, e.PaymentStatus, e.CreatedBy, time.Now(), time.Now())

		mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO expenses (store_id, category_id, amount, expense_date, notes, payment_status, created_by) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, store_id, category_id, amount, expense_date, notes, payment_status, created_by, created_at, updated_at")).
			WithArgs(e.StoreID, e.CategoryID, e.Amount, e.ExpenseDate.Format("2006-01-02"), e.Notes, e.PaymentStatus, e.CreatedBy).
			WillReturnRows(rows)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT name FROM expense_categories WHERE id = $1")).
			WithArgs(e.CategoryID).
			WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("Utilities"))

		result, err := repo.CreateExpense(context.Background(), e)
		require.NoError(t, err)
		assert.Equal(t, "e1", result.ID)
		assert.Equal(t, "Utilities", result.CategoryName)
	})
}

func TestExpenseRepo_FindAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewExpenseRepo(sqlxDB)

	filter := dto.ExpenseListFilter{
		StoreID: "st1",
	}
	filter.PerPage = 10
	filter.Page = 1

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM expenses e WHERE e.store_id = $1")).
			WithArgs("st1").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		mock.ExpectQuery(regexp.QuoteMeta("SELECT e.id, e.store_id, e.category_id, ec.name AS category_name, e.amount, e.expense_date, e.notes, e.payment_status, e.created_by, e.created_at, e.updated_at FROM expenses e JOIN expense_categories ec ON ec.id = e.category_id WHERE e.store_id = $1 ORDER BY e.expense_date DESC, e.created_at DESC LIMIT $2 OFFSET $3")).
			WithArgs("st1", 10, 0).
			WillReturnRows(sqlmock.NewRows([]string{"id", "store_id", "category_id", "category_name", "amount", "expense_date", "notes", "payment_status", "created_by", "created_at", "updated_at"}).
				AddRow("e1", "st1", "cat1", "Utilities", 100.0, time.Now(), "Notes", "paid", "u1", time.Now(), time.Now()))

		expenses, total, err := repo.FindAll(context.Background(), filter)
		assert.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, expenses, 1)
	})
}
