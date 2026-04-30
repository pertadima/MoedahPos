package service

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
	"github.com/moedahpos/backend/mocks"
)

func TestPriceHistoryService(t *testing.T) {
	repo := new(mocks.PriceHistoryRepository)
	log := zerolog.Nop()
	svc := NewPriceHistoryService(repo, log)

	ctx := context.Background()

	t.Run("RecordChange", func(t *testing.T) {
		repo.On("Record", ctx, mock.MatchedBy(func(h domain.PriceHistory) bool {
			return h.ProductID == "p1" && h.NewCost == 200.0
		})).Return(nil).Once()

		err := svc.RecordChange(ctx, "p1", "s1", "u1", 100.0, 200.0, 150.0, 250.0, "manual", nil, nil)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("ListByProduct", func(t *testing.T) {
		f := dto.PriceHistoryFilter{}
		rows := []*domain.PriceHistory{
			{ID: "h1", ProductID: "p1", ChangedAt: time.Now()},
		}
		repo.On("FindByProduct", ctx, "p1", mock.Anything).Return(rows, 1, nil).Once()

		resp, meta, err := svc.ListByProduct(ctx, "p1", f)
		assert.NoError(t, err)
		assert.Len(t, resp, 1)
		assert.Equal(t, 1, meta.Total)
		repo.AssertExpectations(t)
	})

	t.Run("ListByStore", func(t *testing.T) {
		f := dto.PriceHistoryFilter{}
		rows := []*domain.PriceHistory{
			{ID: "h1", StoreID: "s1", ChangedAt: time.Now()},
		}
		repo.On("FindByStore", ctx, "s1", mock.Anything).Return(rows, 1, nil).Once()

		resp, meta, err := svc.ListByStore(ctx, "s1", f)
		assert.NoError(t, err)
		assert.Len(t, resp, 1)
		assert.Equal(t, 1, meta.Total)
		repo.AssertExpectations(t)
	})
}
