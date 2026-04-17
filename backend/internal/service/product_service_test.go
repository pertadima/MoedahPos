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
	"github.com/moedahpos/backend/internal/service/mocks"
)

func TestProductService_CreateProduct(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	t.Run("Success", func(t *testing.T) {
		pRepo := new(repomocks.ProductRepository)
		cRepo := new(repomocks.CategoryRepository)
		sRepo := new(repomocks.StockRepository)
		phSvc := new(mocks.PriceHistoryServiceInterface)

		req := &dto.CreateProductRequest{
			SKU:        "SKU1",
			Name:       "Product 1",
			SellPrice:  100,
			InitialQty: 10,
		}

		pRepo.On("ExistsBySKU", ctx, "s1", "SKU1", "").Return(false, nil)
		pRepo.On("Create", ctx, mock.Anything).Return(&domain.Product{
			ID: "p1", SKU: "SKU1", Name: "Product 1",
		}, nil)

		sRepo.On("Adjust", ctx, mock.MatchedBy(func(in domain.AdjustInput) bool {
			return in.ProductID == "p1" && in.Delta == 10
		})).Return(&domain.StockLevel{}, nil)

		s := NewProductService(pRepo, cRepo, sRepo, phSvc, log)
		resp, err := s.CreateProduct(ctx, "s1", req, "u1")

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "p1", resp.ID)
		pRepo.AssertExpectations(t)
		sRepo.AssertExpectations(t)
	})

	t.Run("SKU Exists", func(t *testing.T) {
		pRepo := new(repomocks.ProductRepository)
		pRepo.On("ExistsBySKU", ctx, "s1", "SKU1", "").Return(true, nil)

		s := NewProductService(pRepo, nil, nil, nil, log)
		resp, err := s.CreateProduct(ctx, "s1", &dto.CreateProductRequest{SKU: "SKU1"}, "u1")

		assert.ErrorIs(t, err, ErrSKUAlreadyExists)
		assert.Nil(t, resp)
	})
}

func TestProductService_UpdateProduct(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	t.Run("Success with Price Change", func(t *testing.T) {
		pRepo := new(repomocks.ProductRepository)
		phSvc := new(mocks.PriceHistoryServiceInterface)

		req := &dto.UpdateProductRequest{
			Name:      "Updated Product",
			CostPrice: 60,
			SellPrice: 120,
		}

		oldProduct := &domain.Product{
			ID: "p1", StoreID: "s1", CostPrice: 50, SellPrice: 100,
		}

		pRepo.On("FindByID", ctx, "p1").Return(oldProduct, nil)
		pRepo.On("Update", ctx, mock.Anything).Return(&domain.Product{
			ID: "p1", Name: "Updated Product", CostPrice: 60, SellPrice: 120,
		}, nil)

		phSvc.On("RecordChange", ctx, "p1", "s1", "u1", 50.0, 60.0, 100.0, 120.0, "manual", mock.Anything, mock.Anything).Return(nil)

		s := NewProductService(pRepo, nil, nil, phSvc, log)
		resp, err := s.UpdateProduct(ctx, "p1", req, "u1")

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		phSvc.AssertExpectations(t)
	})
}
