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

	t.Run("Create Recurring Success", func(t *testing.T) {
		req := &dto.CreateRecurringExpenseRequest{
			Name: "Rent", Amount: 1000, StartDate: "2024-01-01",
		}
		repo.On("CreateRecurringExpense", ctx, mock.Anything).Return(&domain.RecurringExpense{
			ID: "re1", StartDate: time.Now(),
		}, nil).Once()
		resp, err := svc.CreateRecurringExpense(ctx, "s1", "u1", req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("List Recurring Success", func(t *testing.T) {
		filter := dto.ExpenseListFilter{StoreID: "s1"}
		repo.On("FindAllRecurring", ctx, mock.Anything).Return([]*domain.RecurringExpense{{
			ID: "re1", StartDate: time.Now(),
		}}, 1, nil).Once()
		resp, _, err := svc.ListRecurringExpenses(ctx, filter)
		assert.NoError(t, err)
		assert.Len(t, resp, 1)
	})

	t.Run("Category CRUD", func(t *testing.T) {
		repo.On("ListCategories", ctx, false).Return([]*domain.ExpenseCategory{{ID: "c1"}}, nil).Once()
		res, err := svc.ListCategories(ctx, false)
		assert.NoError(t, err)
		assert.Len(t, res, 1)

		repo.On("CreateCategory", ctx, mock.Anything).Return(&domain.ExpenseCategory{ID: "c2"}, nil).Once()
		res2, err := svc.CreateCategory(ctx, &dto.CreateExpenseCategoryRequest{Name: "Travel"})
		assert.NoError(t, err)
		assert.NotNil(t, res2)

		repo.On("UpdateCategory", ctx, "c1", "Updated", "", true).Return(&domain.ExpenseCategory{ID: "c1", Name: "Updated"}, nil).Once()
		res3, err := svc.UpdateCategory(ctx, "c1", &dto.UpdateExpenseCategoryRequest{Name: "Updated", IsActive: true})
		assert.NoError(t, err)
		assert.Equal(t, "Updated", res3.Name)

		repo.On("SoftDeleteCategory", ctx, "c1").Return(nil).Once()
		err = svc.SoftDeleteCategory(ctx, "c1")
		assert.NoError(t, err)
	})

	t.Run("ListExpenses Success", func(t *testing.T) {
		filter := dto.ExpenseListFilter{StoreID: "s1"}
		filter.Defaults()
		repo.On("FindAll", ctx, filter).Return([]*domain.Expense{{ID: "e1"}}, 1, nil).Once()
		resp, _, err := svc.ListExpenses(ctx, filter)
		assert.NoError(t, err)
		assert.Len(t, resp, 1)
	})

	t.Run("UpdateExpense Success", func(t *testing.T) {
		req := &dto.UpdateExpenseRequest{Amount: 600, ExpenseDate: "2024-01-01"}
		repo.On("Update", ctx, mock.Anything).Return(&domain.Expense{ID: "e1", Amount: 600, ExpenseDate: time.Now()}, nil).Once()
		activitySvc.On("LogActivity", ctx, "SYSTEM", "s1", domain.ActionExpenseUpdate, domain.ModuleExpense, "e1", mock.Anything).Return().Once()
		resp, err := svc.UpdateExpense(ctx, "e1", "s1", req)
		assert.NoError(t, err)
		assert.Equal(t, 600.0, resp.Amount)
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
