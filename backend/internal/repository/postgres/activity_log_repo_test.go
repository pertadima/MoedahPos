package postgres

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
)

func TestActivityLogRepo(t *testing.T) {
	ctx := context.Background()

	t.Run("Create", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()
		repo := NewActivityLogRepo(sqlx.NewDb(db, "postgres"))

		mock.ExpectExec(`(?is)INSERT INTO activity_logs`).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err = repo.Create(ctx, &domain.ActivityLog{
			UserID: "u1", StoreID: strPtrExp("s1"), ActionType: "create", Module: "product",
		})
		assert.NoError(t, err)
	})

	t.Run("FindAll", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()
		repo := NewActivityLogRepo(sqlx.NewDb(db, "postgres"))

		// PrepareNamedContext is used twice. We need to match those prepared statements.
		// For sqlmock, we match the query during ExpectQuery.
		
		mock.ExpectQuery(`(?is)SELECT COUNT\(\*\)`).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		mock.ExpectQuery(`(?is)SELECT .* FROM activity_logs`).
			WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "user_name"}).AddRow("l1", "u1", "User 1"))

		res, total, err := repo.FindAll(ctx, "s1", dto.ActivityLogFilter{PaginationQuery: dto.PaginationQuery{PerPage: 10, Page: 1}})
		assert.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, res, 1)
	})
}
