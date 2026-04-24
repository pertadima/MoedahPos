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
func TestProductService_Categories(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	t.Run("ListCategories", func(t *testing.T) {
		cRepo := new(repomocks.CategoryRepository)
		cats := []*domain.Category{{ID: "c1", Name: "Cat 1"}}
		cRepo.On("FindAllByStore", ctx, "s1").Return(cats, nil)

		s := NewProductService(nil, cRepo, nil, nil, log)
		resp, err := s.ListCategories(ctx, "s1")
		assert.NoError(t, err)
		assert.Len(t, resp, 1)
	})

	t.Run("CreateCategory", func(t *testing.T) {
		cRepo := new(repomocks.CategoryRepository)
		req := &dto.CreateCategoryRequest{Name: "New Cat"}
		cRepo.On("Create", ctx, mock.Anything).Return(&domain.Category{ID: "c1", Name: "New Cat"}, nil)

		s := NewProductService(nil, cRepo, nil, nil, log)
		resp, err := s.CreateCategory(ctx, "s1", req)
		assert.NoError(t, err)
		assert.Equal(t, "New Cat", resp.Name)
	})

	t.Run("UpdateCategory", func(t *testing.T) {
		cRepo := new(repomocks.CategoryRepository)
		req := &dto.UpdateCategoryRequest{Name: "Updated Cat"}
		cRepo.On("FindByID", ctx, "c1").Return(&domain.Category{ID: "c1"}, nil)
		cRepo.On("Update", ctx, mock.Anything).Return(&domain.Category{ID: "c1", Name: "Updated Cat"}, nil)

		s := NewProductService(nil, cRepo, nil, nil, log)
		resp, err := s.UpdateCategory(ctx, "c1", req)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Cat", resp.Name)
	})

	t.Run("DeleteCategory", func(t *testing.T) {
		cRepo := new(repomocks.CategoryRepository)
		cRepo.On("FindByID", ctx, "c1").Return(&domain.Category{ID: "c1"}, nil)
		cRepo.On("SoftDelete", ctx, "c1").Return(nil)

		s := NewProductService(nil, cRepo, nil, nil, log)
		err := s.DeleteCategory(ctx, "c1")
		assert.NoError(t, err)
	})
}

func TestProductService_Queries(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	t.Run("ListProducts", func(t *testing.T) {
		pRepo := new(repomocks.ProductRepository)
		filter := dto.ProductListFilter{StoreID: "s1"}
		filter.Defaults()
		filter.WithStock = true
		pRepo.On("FindAll", ctx, filter).Return([]*domain.Product{{ID: "p1", Name: "P1"}}, 1, nil).Once()

		s := NewProductService(pRepo, nil, nil, nil, log)
		resp, meta, err := s.ListProducts(ctx, filter)
		assert.NoError(t, err)
		assert.Len(t, resp, 1)
		assert.Equal(t, 1, meta.Total)
	})

	t.Run("GetProduct", func(t *testing.T) {
		pRepo := new(repomocks.ProductRepository)
		pRepo.On("FindByID", ctx, "p1").Return(&domain.Product{ID: "p1", Name: "P1"}, nil).Once()

		s := NewProductService(pRepo, nil, nil, nil, log)
		resp, err := s.GetProduct(ctx, "p1")
		assert.NoError(t, err)
		assert.Equal(t, "P1", resp.Name)
	})

	t.Run("GetProductByBarcode", func(t *testing.T) {
		pRepo := new(repomocks.ProductRepository)
		pRepo.On("FindByBarcode", ctx, "s1", "123").Return(&domain.Product{ID: "p1", Barcode: ptrString("123")}, nil).Once()

		s := NewProductService(pRepo, nil, nil, nil, log)
		resp, err := s.GetProductByBarcode(ctx, "s1", "123")
		assert.NoError(t, err)
		assert.NotNil(t, resp.Barcode)
		assert.Equal(t, "123", *resp.Barcode)
	})
}

func TestProductService_DeleteProduct(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	t.Run("Success", func(t *testing.T) {
		pRepo := new(repomocks.ProductRepository)
		pRepo.On("FindByID", ctx, "p1").Return(&domain.Product{ID: "p1"}, nil).Once()
		pRepo.On("SoftDelete", ctx, "p1").Return(nil).Once()

		s := NewProductService(pRepo, nil, nil, nil, log)
		err := s.DeleteProduct(ctx, "p1")
		assert.NoError(t, err)
		pRepo.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		pRepo := new(repomocks.ProductRepository)
		pRepo.On("FindByID", ctx, "p1").Return(nil, nil).Once()

		s := NewProductService(pRepo, nil, nil, nil, log)
		err := s.DeleteProduct(ctx, "p1")
		assert.ErrorIs(t, err, ErrProductNotFound)
	})
}

func ptrString(s string) *string {
	return &s
}
