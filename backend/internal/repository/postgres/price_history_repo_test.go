package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
)

func TestPriceHistoryRepo(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer func() { _ = db.Close() }()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewPriceHistoryRepo(sqlxDB)

	t.Run("Record", func(t *testing.T) {
		h := domain.PriceHistory{
			ProductID: "p1",
			StoreID:   "s1",
			ChangedBy: "u1",
			OldCost:   1000,
			NewCost:   1200,
			OldSell:   2000,
			NewSell:   2200,
			Source:    "manual",
			Notes:     stringPtr("test"),
		}
		mock.ExpectExec(`(?is)INSERT INTO price_history`).
			WithArgs(h.ProductID, h.StoreID, h.ChangedBy, h.OldCost, h.NewCost, h.OldSell, h.NewSell, h.Source, h.RefID, h.Notes).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Record(context.Background(), h)
		assert.NoError(t, err)
	})

	t.Run("FindByProduct", func(t *testing.T) {
		mock.ExpectQuery(`(?is)SELECT COUNT\(\*\) FROM price_history WHERE product_id = \$1`).
			WithArgs("p1").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		rows := sqlmock.NewRows([]string{"id", "product_id", "product_name", "store_id", "changed_by", "changed_by_name", "old_cost", "new_cost", "old_sell", "new_sell", "source", "ref_id", "notes", "changed_at"}).
			AddRow("h1", "p1", "Product", "s1", "u1", "User", 1000, 1200, 2000, 2200, "manual", nil, nil, time.Now())

		mock.ExpectQuery(`(?is)SELECT .* FROM price_history .* WHERE ph.product_id = \$1`).
			WithArgs("p1", 10, 0).
			WillReturnRows(rows)

		res, total, err := repo.FindByProduct(context.Background(), "p1", dto.PriceHistoryFilter{
			PaginationQuery: dto.PaginationQuery{Page: 1, PerPage: 10},
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, res, 1)
	})

	t.Run("FindByStore", func(t *testing.T) {
		mock.ExpectQuery(`(?is)SELECT COUNT\(\*\) FROM price_history ph WHERE ph.store_id = \$1`).
			WithArgs("s1").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		rows := sqlmock.NewRows([]string{"id", "product_id", "product_name", "store_id", "changed_by", "changed_by_name", "old_cost", "new_cost", "old_sell", "new_sell", "source", "ref_id", "notes", "changed_at"}).
			AddRow("h1", "p1", "Product", "s1", "u1", "User", 1000, 1200, 2000, 2200, "manual", nil, nil, time.Now())

		mock.ExpectQuery(`(?is)SELECT .* FROM price_history .* WHERE ph.store_id = \$1`).
			WithArgs("s1", 10, 0).
			WillReturnRows(rows)

		res, total, err := repo.FindByStore(context.Background(), "s1", dto.PriceHistoryFilter{
			PaginationQuery: dto.PaginationQuery{Page: 1, PerPage: 10},
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, res, 1)

		// With Filters
		mock.ExpectQuery(`(?is)SELECT COUNT\(\*\) FROM price_history ph WHERE ph.store_id = \$1 AND ph.product_id = \$2 AND ph.source = \$3`).
			WithArgs("s1", "p1", "manual").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectQuery(`(?is)SELECT .* FROM price_history .* WHERE ph.store_id = \$1 AND ph.product_id = \$2 AND ph.source = \$3`).
			WithArgs("s1", "p1", "manual", 10, 0).
			WillReturnRows(rows)

		_, _, err = repo.FindByStore(context.Background(), "s1", dto.PriceHistoryFilter{
			ProductID:       "p1",
			Source:          "manual",
			PaginationQuery: dto.PaginationQuery{Page: 1, PerPage: 10},
		})
		assert.NoError(t, err)
	})
}
