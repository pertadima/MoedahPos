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
)

func TestIncomeRepo_CreateCategory(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewIncomeRepo(sqlxDB)

	c := &domain.IncomeCategory{
		Name:        "Service Fee",
		Description: stringPtr("Fees from services"),
	}

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "description", "is_active", "created_at", "updated_at"}).
			AddRow("cat1", c.Name, *c.Description, true, time.Now(), time.Now())

		mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO income_categories (name, description) VALUES ($1, $2) RETURNING id, name, description, is_active, created_at, updated_at")).
			WithArgs(c.Name, c.Description).
			WillReturnRows(rows)

		result, err := repo.CreateCategory(context.Background(), c)
		assert.NoError(t, err)
		assert.Equal(t, "cat1", result.ID)
	})
}

func TestIncomeRepo_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewIncomeRepo(sqlxDB)

	userID := "u1"
	inc := &domain.Income{
		StoreID:       "st1",
		CategoryID:    "cat1",
		Amount:        500.0,
		IncomeDate:    time.Now(),
		PaymentMethod: "cash",
		CreatedBy:     &userID,
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO incomes (store_id, category_id, amount, income_date, payment_method, reference, notes, created_by) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id")).
			WithArgs(inc.StoreID, inc.CategoryID, inc.Amount, inc.IncomeDate, inc.PaymentMethod, inc.Reference, inc.Notes, inc.CreatedBy).
			WillReturnRows(sqlmock.NewRows([]string{"id", "store_id", "category_id", "amount", "income_date", "payment_method", "reference", "notes", "created_by", "created_at", "updated_at"}).
				AddRow("inc1", inc.StoreID, inc.CategoryID, inc.Amount, inc.IncomeDate, inc.PaymentMethod, inc.Reference, inc.Notes, inc.CreatedBy, time.Now(), time.Now()))

		// Repo's Create calls FindByID
		mock.ExpectQuery(regexp.QuoteMeta("SELECT i.id, i.store_id, i.category_id, i.amount, i.income_date, i.payment_method, i.reference, i.notes, i.created_by, i.created_at, i.updated_at, c.name AS category_name, u.name AS created_by_name FROM incomes i JOIN income_categories c ON c.id = i.category_id LEFT JOIN users u ON u.id = i.created_by WHERE i.id = $1")).
			WithArgs("inc1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "store_id", "category_id", "amount", "income_date", "payment_method", "reference", "notes", "created_by", "created_at", "updated_at", "category_name", "created_by_name"}).
				AddRow("inc1", "st1", "cat1", 500.0, time.Now(), "cash", nil, nil, "u1", time.Now(), time.Now(), "Service", "User 1"))

		result, err := repo.Create(context.Background(), inc)
		assert.NoError(t, err)
		assert.Equal(t, "inc1", result.ID)
		assert.Equal(t, "Service", result.CategoryName)
	})
}

func stringPtr(s string) *string { return &s }
