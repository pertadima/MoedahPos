package service

import (
	"context"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
	"github.com/moedahpos/backend/internal/service/mocks"
)

func TestTableService(t *testing.T) {
	repo := new(mocks.TableRepository)
	log := zerolog.Nop()
	svc := NewTableService(repo, log)

	ctx := context.Background()

	t.Run("List", func(t *testing.T) {
		repo.On("FindAllByStore", ctx, "s1").Return([]*domain.RestaurantTable{{ID: "t1", TableNumber: "1"}}, nil).Once()
		resp, err := svc.List(ctx, "s1")
		assert.NoError(t, err)
		assert.Len(t, resp, 1)
	})

	t.Run("Create", func(t *testing.T) {
		req := &dto.CreateTableRequest{TableNumber: "2", Capacity: 4}
		repo.On("Create", ctx, mock.Anything).Return(&domain.RestaurantTable{ID: "t2", TableNumber: "2"}, nil).Once()
		resp, err := svc.Create(ctx, "s1", req)
		assert.NoError(t, err)
		assert.Equal(t, "2", resp.TableNumber)
	})

	t.Run("UpdateStatus", func(t *testing.T) {
		repo.On("UpdateStatus", ctx, "t1", domain.TableStatus("occupied")).Return(nil).Once()
		err := svc.UpdateStatus(ctx, "t1", "occupied")
		assert.NoError(t, err)
	})

	t.Run("Delete", func(t *testing.T) {
		repo.On("SoftDelete", ctx, "t1").Return(nil).Once()
		err := svc.Delete(ctx, "t1")
		assert.NoError(t, err)
	})
}

func TestMenuItemService(t *testing.T) {
	repo := new(mocks.MenuItemRepository)
	log := zerolog.Nop()
	svc := NewMenuItemService(repo, log)

	ctx := context.Background()

	t.Run("List", func(t *testing.T) {
		repo.On("FindAllByStore", ctx, "s1").Return([]*domain.MenuItem{{ID: "m1", Name: "Nasi Goreng"}}, nil).Once()
		resp, err := svc.List(ctx, "s1")
		assert.NoError(t, err)
		assert.Len(t, resp, 1)
	})

	t.Run("Create", func(t *testing.T) {
		req := &dto.CreateMenuItemRequest{Name: "Mie Goreng", SellPrice: 15000}
		repo.On("Create", ctx, mock.Anything).Return(&domain.MenuItem{ID: "m2", Name: "Mie Goreng"}, nil).Once()
		repo.On("FindByID", ctx, "m2").Return(&domain.MenuItem{ID: "m2", Name: "Mie Goreng"}, nil).Once()

		resp, err := svc.Create(ctx, "s1", req)
		assert.NoError(t, err)
		assert.Equal(t, "Mie Goreng", resp.Name)
	})

	t.Run("Update", func(t *testing.T) {
		req := &dto.UpdateMenuItemRequest{Name: "Updated"}
		repo.On("FindByID", ctx, "m1").Return(&domain.MenuItem{ID: "m1", Name: "Old"}, nil).Once()
		repo.On("Update", ctx, mock.Anything).Return(&domain.MenuItem{ID: "m1", Name: "Updated"}, nil).Once()
		repo.On("ReplaceIngredients", ctx, "m1", mock.Anything).Return(nil).Once()
		repo.On("FindByID", ctx, "m1").Return(&domain.MenuItem{ID: "m1", Name: "Updated"}, nil).Once()

		resp, err := svc.Update(ctx, "m1", req)
		assert.NoError(t, err)
		assert.Equal(t, "Updated", resp.Name)
	})

	t.Run("Delete", func(t *testing.T) {
		repo.On("SoftDelete", ctx, "m1").Return(nil).Once()
		err := svc.Delete(ctx, "m1")
		assert.NoError(t, err)
	})
}
