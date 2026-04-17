package service

import (
	"context"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
	repomocks "github.com/moedahpos/backend/internal/repository/mocks"
)

func TestBatchStockService(t *testing.T) {
	repo := new(repomocks.BatchRepository)
	log := zerolog.Nop()
	svc := NewBatchStockService(repo, log)

	ctx := context.Background()

	t.Run("ReceivePurchaseOrder", func(t *testing.T) {
		items := []dto.POBatchItem{
			{ProductID: "p1", Quantity: 10, UnitCost: 100},
		}
		repo.On("CreateBatch", ctx, mock.MatchedBy(func(b *domain.StockBatch) bool {
			return b.ProductID == "p1" && b.QuantityRemaining == 10
		})).Return(nil).Once()

		err := svc.ReceivePurchaseOrder(ctx, "po1", "s1", items)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("DeductStockFIFO", func(t *testing.T) {
		repo.On("DeductFIFO", ctx, "p1", "s1", 5.0).Return(nil).Once()
		err := svc.DeductStockFIFO(ctx, "p1", "s1", 5.0)
		assert.NoError(t, err)
	})
}
