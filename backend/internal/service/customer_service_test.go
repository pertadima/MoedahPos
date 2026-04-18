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

const customerTestStoreID = "store-1"

func TestCustomerService_List(t *testing.T) {
	repo := new(repomocks.CustomerRepository)
	log := zerolog.Nop()
	svc := NewCustomerService(repo, log)

	storeID := customerTestStoreID
	filter := dto.CustomerListFilter{StoreID: storeID}

	t.Run("success", func(t *testing.T) {
		customers := []*domain.Customer{{ID: "c1", Name: "Customer 1"}}
		repo.On("FindAll", mock.Anything, mock.Anything).Return(customers, 1, nil).Once()

		resp, meta, err := svc.List(context.Background(), filter)
		assert.NoError(t, err)
		assert.Len(t, resp, 1)
		assert.Equal(t, 1, meta.Total)
		repo.AssertExpectations(t)
	})
}

func TestCustomerService_Create(t *testing.T) {
	repo := new(repomocks.CustomerRepository)
	log := zerolog.Nop()
	svc := NewCustomerService(repo, log)

	storeID := customerTestStoreID
	req := dto.CreateCustomerRequest{Name: "New Customer"}

	t.Run("success", func(t *testing.T) {
		repo.On("Create", mock.Anything, mock.Anything).Return(&domain.Customer{ID: "c1", Name: req.Name}, nil).Once()

		resp, err := svc.Create(context.Background(), storeID, req)
		assert.NoError(t, err)
		assert.Equal(t, "c1", resp.ID)
		repo.AssertExpectations(t)
	})
}

func TestCustomerService_Update(t *testing.T) {
	repo := new(repomocks.CustomerRepository)
	log := zerolog.Nop()
	svc := NewCustomerService(repo, log)

	id := "c1"
	req := dto.UpdateCustomerRequest{Name: "Updated Name"}

	t.Run("success", func(t *testing.T) {
		existing := &domain.Customer{ID: id, Name: "Old Name"}
		repo.On("FindByID", mock.Anything, id).Return(existing, nil).Once()
		repo.On("Update", mock.Anything, mock.Anything).Return(&domain.Customer{ID: id, Name: req.Name}, nil).Once()

		resp, err := svc.Update(context.Background(), id, req)
		assert.NoError(t, err)
		assert.Equal(t, req.Name, resp.Name)
		repo.AssertExpectations(t)
	})

	t.Run("not_found", func(t *testing.T) {
		repo.On("FindByID", mock.Anything, id).Return(nil, nil).Once()

		resp, err := svc.Update(context.Background(), id, req)
		assert.ErrorIs(t, err, ErrCustomerNotFound)
		assert.Nil(t, resp)
		repo.AssertExpectations(t)
	})
}

func TestCustomerService_Delete(t *testing.T) {
	repo := new(repomocks.CustomerRepository)
	log := zerolog.Nop()
	svc := NewCustomerService(repo, log)

	id := "c1"

	t.Run("success", func(t *testing.T) {
		repo.On("SoftDelete", mock.Anything, id).Return(nil).Once()

		err := svc.Delete(context.Background(), id)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("not_found", func(t *testing.T) {
		repo.On("SoftDelete", mock.Anything, id).Return(errors.New("customer not found")).Once()

		err := svc.Delete(context.Background(), id)
		assert.ErrorIs(t, err, ErrCustomerNotFound)
		repo.AssertExpectations(t)
	})
}
