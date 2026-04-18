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
	"github.com/moedahpos/backend/internal/repository/mocks"
)

func TestSupplierService_CreateSupplier(t *testing.T) {
	repo := mocks.NewSupplierRepository(t)
	svc := NewSupplierService(repo, zerolog.Nop())

	req := &dto.CreateSupplierRequest{
		Name:        "Supplier A",
		ContactName: "John",
		Phone:       "123",
		Email:       "a@a.com",
		Address:     "Addr",
	}

	t.Run("success", func(t *testing.T) {
		repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Supplier")).
			Return(&domain.Supplier{
				ID:          "s1",
				Name:        req.Name,
				ContactName: req.ContactName,
				Phone:       req.Phone,
				Email:       req.Email,
				Address:     req.Address,
				IsActive:    true,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}, nil).Once()

		resp, err := svc.CreateSupplier(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "s1", resp.ID)
	})
}

func TestSupplierService_ListSuppliers(t *testing.T) {
	repo := mocks.NewSupplierRepository(t)
	svc := NewSupplierService(repo, zerolog.Nop())

	filter := dto.SupplierListFilter{}

	t.Run("success", func(t *testing.T) {
		repo.On("FindAll", mock.Anything, mock.Anything).
			Return([]*domain.Supplier{
				{ID: "s1", Name: "S1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
			}, 1, nil).Once()

		resp, meta, err := svc.ListSuppliers(context.Background(), filter)
		assert.NoError(t, err)
		assert.Len(t, resp, 1)
		assert.Equal(t, 1, meta.Total)
	})
}

func TestSupplierService_GetSupplier(t *testing.T) {
	repo := mocks.NewSupplierRepository(t)
	svc := NewSupplierService(repo, zerolog.Nop())

	id := "s1"

	t.Run("success", func(t *testing.T) {
		repo.On("FindByID", mock.Anything, id).
			Return(&domain.Supplier{ID: id, Name: "S1", CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil).Once()

		resp, err := svc.GetSupplier(context.Background(), id)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, id, resp.ID)
	})

	t.Run("not found", func(t *testing.T) {
		repo.On("FindByID", mock.Anything, id).Return(nil, nil).Once()

		resp, err := svc.GetSupplier(context.Background(), id)
		assert.Error(t, err)
		assert.IsType(t, ErrSupplierNotFound, err)
		assert.Nil(t, resp)
	})
}

func TestSupplierService_UpdateSupplier(t *testing.T) {
	repo := mocks.NewSupplierRepository(t)
	svc := NewSupplierService(repo, zerolog.Nop())

	id := "s1"
	req := &dto.UpdateSupplierRequest{
		Name: "Updated S1",
	}

	t.Run("success", func(t *testing.T) {
		repo.On("FindByID", mock.Anything, id).
			Return(&domain.Supplier{ID: id, Name: "S1", CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil).Once()

		repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Supplier")).
			Return(&domain.Supplier{ID: id, Name: "Updated S1", CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil).Once()

		resp, err := svc.UpdateSupplier(context.Background(), id, req)
		assert.NoError(t, err)
		assert.Equal(t, "Updated S1", resp.Name)
	})
}

func TestSupplierService_DeleteSupplier(t *testing.T) {
	repo := mocks.NewSupplierRepository(t)
	svc := NewSupplierService(repo, zerolog.Nop())

	id := "s1"

	t.Run("success", func(t *testing.T) {
		repo.On("FindByID", mock.Anything, id).
			Return(&domain.Supplier{ID: id}, nil).Once()
		repo.On("SoftDelete", mock.Anything, id).Return(nil).Once()

		err := svc.DeleteSupplier(context.Background(), id)
		assert.NoError(t, err)
	})
}
