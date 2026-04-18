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
	repomocks "github.com/moedahpos/backend/internal/repository/mocks"
	"github.com/moedahpos/backend/internal/service/mocks"
)

func TestIncomeService(t *testing.T) {
	repo := new(repomocks.IncomeRepository)
	activitySvc := new(mocks.ActivityLogServiceInterface)
	log := zerolog.Nop()
	svc := NewIncomeService(repo, activitySvc, log)

	ctx := context.Background()

	t.Run("CreateIncome", func(t *testing.T) {
		req := &dto.CreateIncomeRequest{
			CategoryID:    "c1",
			Amount:        1000,
			IncomeDate:    "2024-01-01",
			PaymentMethod: "cash",
		}

		repo.On("Create", ctx, mock.MatchedBy(func(in *domain.Income) bool {
			return in.Amount == 1000 && in.PaymentMethod == "cash"
		})).Return(&domain.Income{ID: "i1", Amount: 1000, PaymentMethod: "cash"}, nil).Once()

		activitySvc.On("LogActivity", ctx, "u1", "s1", domain.ActionIncomeCreate, domain.ModuleIncome, "i1", mock.Anything).Return().Once()

		resp, err := svc.CreateIncome(ctx, "s1", "u1", req)
		assert.NoError(t, err)
		assert.Equal(t, "i1", resp.ID)
		repo.AssertExpectations(t)
		activitySvc.AssertExpectations(t)
	})

	t.Run("ListIncomes", func(t *testing.T) {
		repo.On("FindAll", ctx, mock.Anything).Return([]*domain.Income{{
			ID: "i1", IncomeDate: time.Now(), CreatedAt: time.Now(), UpdatedAt: time.Now(),
		}}, 1, nil).Once()
		resp, meta, err := svc.ListIncomes(ctx, dto.IncomeListFilter{})
		assert.NoError(t, err)
		assert.Len(t, resp, 1)
		assert.Equal(t, 1, meta.Total)
	})

	t.Run("Category CRUD", func(t *testing.T) {
		repo.On("ListCategories", ctx, false).Return([]*domain.IncomeCategory{{ID: "c1"}}, nil).Once()
		res, err := svc.ListCategories(ctx, false)
		assert.NoError(t, err)
		assert.Len(t, res, 1)

		repo.On("CreateCategory", ctx, mock.Anything).Return(&domain.IncomeCategory{ID: "c2"}, nil).Once()
		res2, err := svc.CreateCategory(ctx, &dto.CreateIncomeCategoryRequest{Name: "Gift"})
		assert.NoError(t, err)
		assert.NotNil(t, res2)
	})

	t.Run("Update/Delete Income", func(t *testing.T) {
		req := &dto.UpdateIncomeRequest{Amount: 2000, IncomeDate: "2024-01-01"}
		repo.On("Update", ctx, mock.Anything).Return(&domain.Income{
			ID: "i1", Amount: 2000, IncomeDate: time.Now(), CreatedAt: time.Now(), UpdatedAt: time.Now(),
		}, nil).Once()
		activitySvc.On("LogActivity", ctx, "SYSTEM", "s1", domain.ActionIncomeUpdate, mock.Anything, mock.Anything, mock.Anything).Return().Once()

		resp, err := svc.UpdateIncome(ctx, "i1", "s1", req)
		assert.NoError(t, err)
		assert.Equal(t, 2000.0, resp.Amount)

		repo.On("FindByID", ctx, "i1").Return(&domain.Income{ID: "i1"}, nil).Once()
		repo.On("Delete", ctx, "i1", "s1").Return(nil).Once()
		activitySvc.On("LogActivity", ctx, "SYSTEM", "s1", domain.ActionIncomeDelete, mock.Anything, mock.Anything, mock.Anything).Return().Once()
		err = svc.DeleteIncome(ctx, "i1", "s1")
		assert.NoError(t, err)
	})
}
