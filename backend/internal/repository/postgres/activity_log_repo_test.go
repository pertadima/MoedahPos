package postgres

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
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
		defer func() { _ = db.Close() }()
		repo := NewActivityLogRepository(db)

		mock.ExpectExec(`(?is)INSERT INTO activity_logs`).
			WithArgs("u1", "s1", "create", "product", sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err = repo.Create(ctx, &domain.ActivityLog{
			UserID: "u1", StoreID: ptr("s1"), ActionType: "create", Module: "product",
		})
		assert.NoError(t, err)
	})

	t.Run("FindAll", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		require.NoError(t, err)
		defer func() { _ = db.Close() }()
		repo := NewActivityLogRepository(db)

		// Case 1: Minimal store_id
		mock.ExpectPrepare(`(?is)SELECT COUNT\(\*\) FROM activity_logs`).
			ExpectQuery().
			WithArgs("s1").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		mock.ExpectPrepare(`(?is)SELECT .* FROM activity_logs al JOIN users u`).
			ExpectQuery().
			WithArgs("s1", 10, 0).
			WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "user_name"}).AddRow("l1", "u1", "User 1"))

		res, total, err := repo.FindAll(ctx, "s1", dto.ActivityLogFilter{PaginationQuery: dto.PaginationQuery{PerPage: 10, Page: 1}})
		assert.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, res, 1)

		// Case 2: All filters
		filter := dto.ActivityLogFilter{
			UserID:          "u1",
			Module:          "product",
			ActionType:      "create",
			StartDate:       "2024-01-01",
			EndDate:         "2024-01-31",
			PaginationQuery: dto.PaginationQuery{PerPage: 10, Page: 1},
		}

		// The order of map iteration is random, but PrepareNamed usually succeeds as long as placeholders match.
		// However, sqlmock with ExpectQuery().WithArgs() needs exact order.
		// Named queries in sqlx use :name which are translated to $1, $2 etc.
		// Let's use AnyArg to be safe if order varies, or just check repo logic.
		mock.ExpectPrepare(`(?is)SELECT COUNT\(\*\) FROM activity_logs`).
			ExpectQuery().
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		mock.ExpectPrepare(`(?is)SELECT .* FROM activity_logs al JOIN users u`).
			ExpectQuery().
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("l1"))

		_, _, err = repo.FindAll(ctx, "s1", filter)
		assert.NoError(t, err)
	})
}
