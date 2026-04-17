package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/moedahpos/backend/internal/domain"
	repomocks "github.com/moedahpos/backend/internal/repository/mocks"
	servicemocks "github.com/moedahpos/backend/internal/service/mocks"
)

func TestStockAdjustmentService_CreateAdjustment(t *testing.T) {
	repo := new(repomocks.StockAdjustmentRepository)
	activitySvc := new(servicemocks.ActivityLogServiceInterface)
	svc := NewStockAdjustmentService(repo, activitySvc)

	storeID := "store-1"
	userID := "user-1"

	t.Run("success_IN", func(t *testing.T) {
		input := domain.CreateAdjustmentInput{
			ProductID: "p1",
			Type:      "IN",
			Reason:    "MANUAL_CORRECTION",
			Quantity:  10.0,
			Notes:     "Manual",
		}
		repo.On("CreateAdjustment", mock.Anything, storeID, userID, input).Return(nil).Once()
		activitySvc.On("LogActivity", mock.Anything, userID, storeID, domain.ActionStockAdjustment, domain.ModuleInventory, "p1", mock.Anything).Return().Once()

		err := svc.CreateAdjustment(context.Background(), storeID, userID, input)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
		activitySvc.AssertExpectations(t)
	})

	t.Run("success_OUT", func(t *testing.T) {
		input := domain.CreateAdjustmentInput{
			ProductID: "p1",
			Type:      "OUT",
			Reason:    "DAMAGED",
			Quantity:  5.0,
		}
		repo.On("CreateAdjustment", mock.Anything, storeID, userID, input).Return(nil).Once()
		activitySvc.On("LogActivity", mock.Anything, userID, storeID, domain.ActionStockAdjustment, domain.ModuleInventory, "p1", mock.Anything).Return().Once()

		err := svc.CreateAdjustment(context.Background(), storeID, userID, input)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("invalid_OUT_reason", func(t *testing.T) {
		input := domain.CreateAdjustmentInput{
			ProductID: "p1",
			Type:      "OUT",
			Reason:    "SOLD", // Not permitted reason in service logic
			Quantity:  5.0,
		}

		err := svc.CreateAdjustment(context.Background(), storeID, userID, input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not permitted for type OUT")
	})

	t.Run("missing_notes_for_correction", func(t *testing.T) {
		input := domain.CreateAdjustmentInput{
			ProductID: "p1",
			Type:      "IN",
			Reason:    "MANUAL_CORRECTION",
			Quantity:  10.0,
			Notes:     "", // Empty
		}

		err := svc.CreateAdjustment(context.Background(), storeID, userID, input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "notes are required")
	})

	t.Run("repo_error", func(t *testing.T) {
		input := domain.CreateAdjustmentInput{
			ProductID: "p1",
			Type:      "OUT",
			Reason:    "LOST",
			Quantity:  1.0,
		}
		repo.On("CreateAdjustment", mock.Anything, storeID, userID, input).Return(errors.New("db error")).Once()

		err := svc.CreateAdjustment(context.Background(), storeID, userID, input)
		assert.Error(t, err)
		repo.AssertExpectations(t)
	})
}

func TestStockAdjustmentService_GetAdjustmentHistory(t *testing.T) {
	repo := new(repomocks.StockAdjustmentRepository)
	activitySvc := new(servicemocks.ActivityLogServiceInterface)
	svc := NewStockAdjustmentService(repo, activitySvc)

	storeID := "store-1"

	t.Run("success", func(t *testing.T) {
		history := []*domain.StockAdjustment{{ID: "adj-1"}}
		repo.On("GetStockAdjustmentHistory", mock.Anything, storeID, mock.Anything).Return(history, nil).Once()

		resp, err := svc.GetAdjustmentHistory(context.Background(), storeID, nil)
		assert.NoError(t, err)
		assert.Len(t, resp, 1)
		repo.AssertExpectations(t)
	})
}
