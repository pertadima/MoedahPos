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
}
