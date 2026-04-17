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

func TestExpenseService(t *testing.T) {
	repo := new(repomocks.ExpenseRepository)
	activitySvc := new(mocks.ActivityLogServiceInterface)
	log := zerolog.Nop()
	svc := NewExpenseService(repo, activitySvc, log)

	ctx := context.Background()

	t.Run("CreateExpense", func(t *testing.T) {
		req := &dto.CreateExpenseRequest{
			CategoryID:    "c1",
			Amount:        500,
			ExpenseDate:   "2024-01-01",
			PaymentStatus: "paid",
		}

		repo.On("CreateExpense", ctx, mock.Anything).Return(&domain.Expense{
			ID: "e1", Amount: 500, PaymentStatus: "paid",
		}, nil).Once()

		activitySvc.On("LogActivity", ctx, "u1", "s1", domain.ActionExpenseCreate, domain.ModuleExpense, "e1", mock.Anything).Return().Once()

		resp, err := svc.CreateExpense(ctx, "s1", "u1", req)
		assert.NoError(t, err)
		assert.Equal(t, "e1", resp.ID)
	})

	t.Run("ProcessDueRecurringExpenses", func(t *testing.T) {
		due := []*domain.RecurringExpense{
			{
				ID: "re1", Name: "Rent", StoreID: "s1", Amount: 2000,
				Interval: "monthly", IntervalValue: 1,
				NextRunDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		}
		repo.On("GetDueRecurringExpenses", ctx).Return(due, nil).Once()
		repo.On("CreateExpense", ctx, mock.Anything).Return(&domain.Expense{ID: "e2"}, nil).Once()
		repo.On("BumpRecurringNextRun", ctx, "re1", "2024-02-01").Return(nil).Once()

		err := svc.ProcessDueRecurringExpenses(ctx)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})
}
