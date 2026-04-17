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
		repo.On("FindAll", ctx, mock.Anything).Return([]*domain.Income{{ID: "i1"}}, 1, nil).Once()
		resp, meta, err := svc.ListIncomes(ctx, dto.IncomeListFilter{})
		assert.NoError(t, err)
		assert.Len(t, resp, 1)
		assert.Equal(t, 1, meta.Total)
	})
}
