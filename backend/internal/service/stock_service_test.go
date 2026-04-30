package service

import (
	"context"
	"errors"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
	repomocks "github.com/moedahpos/backend/internal/repository/mocks"
)

func TestStockService_GetStockLevels(t *testing.T) {
	stockRepo := new(repomocks.StockRepository)
	productRepo := new(repomocks.ProductRepository)
	logger := zerolog.Nop()
	svc := NewStockService(stockRepo, productRepo, logger)

	storeID := customerTestStoreID

	t.Run("success", func(t *testing.T) {
		levels := []*domain.StockLevel{
			{ProductID: "p1", ProductName: "P1", Quantity: 10},
		}
		stockRepo.On("FindLevelsByStore", mock.Anything, storeID, false).Return(levels, nil).Once()

		resp, err := svc.GetStockLevels(context.Background(), storeID, false)
		assert.NoError(t, err)
		assert.Len(t, resp, 1)
		assert.Equal(t, "P1", resp[0].ProductName)
		stockRepo.AssertExpectations(t)
	})

	t.Run("repo_error", func(t *testing.T) {
		stockRepo.On("FindLevelsByStore", mock.Anything, storeID, false).Return(nil, errors.New("db error")).Once()

		resp, err := svc.GetStockLevels(context.Background(), storeID, false)
		assert.Error(t, err)
		assert.Nil(t, resp)
		stockRepo.AssertExpectations(t)
	})
}

func TestStockService_GetProductStock(t *testing.T) {
	stockRepo := new(repomocks.StockRepository)
	productRepo := new(repomocks.ProductRepository)
	logger := zerolog.Nop()
	svc := NewStockService(stockRepo, productRepo, logger)

	productID := "p1"
	storeID := customerTestStoreID

	t.Run("success", func(t *testing.T) {
		level := &domain.StockLevel{ProductID: productID, ProductName: "P1", Quantity: 10}
		stockRepo.On("FindLevelByProduct", mock.Anything, productID, storeID).Return(level, nil).Once()

		resp, err := svc.GetProductStock(context.Background(), productID, storeID)
		assert.NoError(t, err)
		assert.Equal(t, 10.0, resp.Quantity)
		stockRepo.AssertExpectations(t)
	})

	t.Run("not_found", func(t *testing.T) {
		stockRepo.On("FindLevelByProduct", mock.Anything, productID, storeID).Return(nil, nil).Once()

		resp, err := svc.GetProductStock(context.Background(), productID, storeID)
		assert.ErrorIs(t, err, ErrStockLevelNotFound)
		assert.Nil(t, resp)
		stockRepo.AssertExpectations(t)
	})
}

func TestStockService_AdjustStock(t *testing.T) {
	stockRepo := new(repomocks.StockRepository)
	productRepo := new(repomocks.ProductRepository)
	logger := zerolog.Nop()
	svc := NewStockService(stockRepo, productRepo, logger)

	storeID := customerTestStoreID
	userID := "user-1"
	req := &dto.AdjustStockRequest{
		ProductID: "p1",
		Delta:     5,
		Notes:     "",
	}

	t.Run("success", func(t *testing.T) {
		product := &domain.Product{ID: "p1", StoreID: storeID}
		productRepo.On("FindByID", mock.Anything, "p1").Return(product, nil).Once()

		level := &domain.StockLevel{ProductID: "p1", Quantity: 15}
		stockRepo.On("Adjust", mock.Anything, mock.MatchedBy(func(in domain.AdjustInput) bool {
			return in.ProductID == "p1" && in.Delta == 5
		})).Return(level, nil).Once()

		resp, err := svc.AdjustStock(context.Background(), storeID, req, userID)
		assert.NoError(t, err)
		assert.Equal(t, 15.0, resp.Quantity)
		productRepo.AssertExpectations(t)
		stockRepo.AssertExpectations(t)
	})

	t.Run("product_not_in_store", func(t *testing.T) {
		product := &domain.Product{ID: "p1", StoreID: "other-store"}
		productRepo.On("FindByID", mock.Anything, "p1").Return(product, nil).Once()

		resp, err := svc.AdjustStock(context.Background(), storeID, req, userID)
		assert.ErrorIs(t, err, ErrProductNotInStore)
		assert.Nil(t, resp)
		productRepo.AssertExpectations(t)
	})

	t.Run("product_not_found", func(t *testing.T) {
		productRepo.On("FindByID", mock.Anything, "p1").Return(nil, nil).Once()

		resp, err := svc.AdjustStock(context.Background(), storeID, req, userID)
		assert.ErrorIs(t, err, ErrProductNotInStore)
		assert.Nil(t, resp)
		productRepo.AssertExpectations(t)
	})
}

func TestStockService_SetMinStock(t *testing.T) {
	stockRepo := new(repomocks.StockRepository)
	productRepo := new(repomocks.ProductRepository)
	logger := zerolog.Nop()
	svc := NewStockService(stockRepo, productRepo, logger)

	storeID := customerTestStoreID
	req := &dto.SetMinStockRequest{
		ProductID:   "p1",
		MinQuantity: 5,
	}

	t.Run("success", func(t *testing.T) {
		product := &domain.Product{ID: "p1", StoreID: storeID}
		productRepo.On("FindByID", mock.Anything, "p1").Return(product, nil).Once()
		stockRepo.On("SetMinQuantity", mock.Anything, "p1", storeID, 5.0).Return(nil).Once()

		err := svc.SetMinStock(context.Background(), storeID, req)
		assert.NoError(t, err)
		productRepo.AssertExpectations(t)
		stockRepo.AssertExpectations(t)
	})

	t.Run("unauthorized_store", func(t *testing.T) {
		product := &domain.Product{ID: "p1", StoreID: "other-store"}
		productRepo.On("FindByID", mock.Anything, "p1").Return(product, nil).Once()

		err := svc.SetMinStock(context.Background(), storeID, req)
		assert.ErrorIs(t, err, ErrProductNotInStore)
		productRepo.AssertExpectations(t)
	})
}

func TestStockService_GetMovements(t *testing.T) {
	stockRepo := new(repomocks.StockRepository)
	productRepo := new(repomocks.ProductRepository)
	logger := zerolog.Nop()
	svc := NewStockService(stockRepo, productRepo, logger)

	filter := dto.StockMovementFilter{StoreID: customerTestStoreID}

	t.Run("success", func(t *testing.T) {
		movements := []*domain.StockMovement{
			{ID: "m1", ProductID: "p1", QuantityDelta: 10},
		}
		stockRepo.On("FindMovements", mock.Anything, mock.Anything).Return(movements, 1, nil).Once()

		resp, meta, err := svc.GetMovements(context.Background(), filter)
		assert.NoError(t, err)
		assert.Len(t, resp, 1)
		assert.Equal(t, 1, meta.Total)
		stockRepo.AssertExpectations(t)
	})
}
