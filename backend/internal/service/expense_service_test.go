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
	service_mocks "github.com/moedahpos/backend/internal/service/mocks"
)

func TestExpenseService_CreateExpense(t *testing.T) {
	repo := mocks.NewExpenseRepository(t)
	activitySvc := service_mocks.NewActivityLogServiceInterface(t)
	svc := NewExpenseService(repo, activitySvc, zerolog.Nop())

	storeID := "st1"
	userID := "u1"
	req := &dto.CreateExpenseRequest{
		CategoryID:    "cat1",
		Amount:        100.0,
		ExpenseDate:   "2026-04-17",
		Notes:         "Lunch",
		PaymentStatus: "paid",
	}

	t.Run("success", func(t *testing.T) {
		repo.On("CreateExpense", mock.Anything, mock.AnythingOfType("*domain.Expense")).
			Return(&domain.Expense{
				ID:            "e1",
				StoreID:       storeID,
				CategoryID:    req.CategoryID,
				CategoryName:  "Food",
				Amount:        req.Amount,
				ExpenseDate:   time.Now(),
				Notes:         req.Notes,
				PaymentStatus: req.PaymentStatus,
				CreatedBy:     &userID,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}, nil).Once()

		activitySvc.On("LogActivity", mock.Anything, userID, storeID, domain.ActionExpenseCreate, domain.ModuleExpense, "e1", mock.Anything).
			Return().Once()

		resp, err := svc.CreateExpense(context.Background(), storeID, userID, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "e1", resp.ID)
		assert.Equal(t, "Food", resp.CategoryName)
	})

	t.Run("invalid date", func(t *testing.T) {
		invalidReq := &dto.CreateExpenseRequest{ExpenseDate: "invalid"}
		resp, err := svc.CreateExpense(context.Background(), storeID, userID, invalidReq)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestExpenseService_ListExpenses(t *testing.T) {
	repo := mocks.NewExpenseRepository(t)
	svc := NewExpenseService(repo, nil, zerolog.Nop())

	filter := dto.ExpenseListFilter{StoreID: "st1"}

	t.Run("success", func(t *testing.T) {
		repo.On("FindAll", mock.Anything, mock.Anything).
			Return([]*domain.Expense{
				{ID: "e1", StoreID: "st1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
			}, 1, nil).Once()

		resp, meta, err := svc.ListExpenses(context.Background(), filter)
		assert.NoError(t, err)
		assert.Len(t, resp, 1)
		assert.Equal(t, 1, meta.Total)
	})
}

func TestExpenseService_ProcessDueRecurringExpenses(t *testing.T) {
	repo := mocks.NewExpenseRepository(t)
	svc := NewExpenseService(repo, nil, zerolog.Nop())

	t.Run("success", func(t *testing.T) {
		repo.On("GetDueRecurringExpenses", mock.Anything).
			Return([]*domain.RecurringExpense{
				{
					ID:            "re1",
					StoreID:       "st1",
					CategoryID:    "cat1",
					Name:          "Monthly Rent",
					Amount:        1000.0,
					Interval:      "monthly",
					IntervalValue: 1,
					NextRunDate:   time.Now(),
				},
			}, nil).Once()

		repo.On("CreateExpense", mock.Anything, mock.MatchedBy(func(e *domain.Expense) bool {
			return e.CategoryID == "cat1" && e.Amount == 1000.0
		})).Return(&domain.Expense{ID: "e1"}, nil).Once()

		repo.On("BumpRecurringNextRun", mock.Anything, "re1", mock.Anything).Return(nil).Once()

		err := svc.ProcessDueRecurringExpenses(context.Background())
		assert.NoError(t, err)
	})
}
