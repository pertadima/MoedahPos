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

func TestIncomeService_CreateIncome(t *testing.T) {
	repo := mocks.NewIncomeRepository(t)
	activitySvc := service_mocks.NewActivityLogServiceInterface(t)
	svc := NewIncomeService(repo, activitySvc, zerolog.Nop())

	storeID := "st1"
	userID := "u1"
	req := &dto.CreateIncomeRequest{
		CategoryID:    "cat1",
		Amount:        500.0,
		IncomeDate:    "2026-04-17",
		PaymentMethod: "cash",
	}

	t.Run("success", func(t *testing.T) {
		repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Income")).
			Return(&domain.Income{
				ID:            "inc1",
				StoreID:       storeID,
				CategoryID:    req.CategoryID,
				CategoryName:  "Service",
				Amount:        req.Amount,
				IncomeDate:    time.Now(),
				PaymentMethod: req.PaymentMethod,
				CreatedBy:     &userID,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}, nil).Once()

		activitySvc.On("LogActivity", mock.Anything, userID, storeID, domain.ActionIncomeCreate, domain.ModuleIncome, "inc1", mock.Anything).
			Return().Once()

		resp, err := svc.CreateIncome(context.Background(), storeID, userID, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "inc1", resp.ID)
	})

	t.Run("invalid date", func(t *testing.T) {
		invalidReq := &dto.CreateIncomeRequest{IncomeDate: "invalid"}
		resp, err := svc.CreateIncome(context.Background(), storeID, userID, invalidReq)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestIncomeService_ListIncomes(t *testing.T) {
	repo := mocks.NewIncomeRepository(t)
	svc := NewIncomeService(repo, nil, zerolog.Nop())

	filter := dto.IncomeListFilter{StoreID: "st1"}

	t.Run("success", func(t *testing.T) {
		repo.On("FindAll", mock.Anything, mock.Anything).
			Return([]*domain.Income{
				{ID: "inc1", StoreID: "st1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
			}, 1, nil).Once()

		resp, meta, err := svc.ListIncomes(context.Background(), filter)
		assert.NoError(t, err)
		assert.Len(t, resp, 1)
		assert.Equal(t, 1, meta.Total)
	})
}
