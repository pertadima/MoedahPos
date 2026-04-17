package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/moedahpos/backend/internal/domain"
	repomocks "github.com/moedahpos/backend/internal/repository/mocks"
	"github.com/moedahpos/backend/internal/service/mocks"
)

func TestStockAdjustmentService(t *testing.T) {
	repo := new(repomocks.StockAdjustmentRepository)
	activitySvc := new(mocks.ActivityLogServiceInterface)
	svc := NewStockAdjustmentService(repo, activitySvc)

	ctx := context.Background()

	t.Run("CreateAdjustment_Success", func(t *testing.T) {
		input := domain.CreateAdjustmentInput{
			ProductID: "p1",
			Quantity:  5,
			Type:      "IN",
			Reason:    "MANUAL_CORRECTION",
			Notes:     "Correction",
		}

		repo.On("CreateAdjustment", ctx, "s1", "u1", input).Return(nil).Once()
		activitySvc.On("LogActivity", ctx, "u1", "s1", domain.ActionStockAdjustment, domain.ModuleInventory, "p1", mock.Anything).Return().Once()

		err := svc.CreateAdjustment(ctx, "s1", "u1", input)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("CreateAdjustment_InvalidReason", func(t *testing.T) {
		input := domain.CreateAdjustmentInput{
			ProductID: "p1",
			Quantity:  5,
			Type:      "IN",
			Reason:    "DAMAGED", // IN can only be MANUAL_CORRECTION
		}

		err := svc.CreateAdjustment(ctx, "s1", "u1", input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "can only be used with MANUAL_CORRECTION reason")
	})
}
