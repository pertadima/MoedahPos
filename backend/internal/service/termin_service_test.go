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

func TestTerminService_CreateTerminSchedule(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	tRepo := new(repomocks.TerminRepository)
	payRepo := new(repomocks.PaymentRecordRepository)
	poRepo := new(repomocks.PurchaseOrderRepository)
	sRepo := new(repomocks.StoreRepository)
	aSvc := new(mocks.ActivityLogServiceInterface)

	svc := NewTerminService(tRepo, payRepo, poRepo, sRepo, aSvc, log)

	poID := "po1"
	req := dto.CreateTerminScheduleRequest{
		Termins: []dto.TerminInput{
			{TerminNumber: 1, Amount: 1000, DueDate: "2026-05-01", Notes: "N1"},
		},
	}

	poRepo.On("FindByID", ctx, poID).Return(&domain.PurchaseOrder{
		ID:          poID,
		Status:      "received",
		TotalAmount: 1000,
	}, nil)

	tRepo.On("CreateSchedule", ctx, poID, mock.Anything).Return(nil)
	tRepo.On("FindByPO", ctx, poID).Return([]*domain.POTermin{
		{ID: "t1", POID: poID, TerminNumber: 1, Amount: 1000, DueDate: time.Date(2026, 5, 1, 0, 0, 0, 0, time.Local), Status: "unpaid"},
	}, nil)
	payRepo.On("FindByTermin", ctx, "t1").Return([]*domain.PaymentRecord{}, nil)

	res, err := svc.CreateTerminSchedule(ctx, poID, req)
	assert.NoError(t, err)
	assert.Len(t, res, 1)
	assert.Equal(t, 1000.0, res[0].Amount)
	poRepo.AssertExpectations(t)
	tRepo.AssertExpectations(t)
}

func TestTerminService_RecordPayment(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	tRepo := new(repomocks.TerminRepository)
	payRepo := new(repomocks.PaymentRecordRepository)
	poRepo := new(repomocks.PurchaseOrderRepository)
	sRepo := new(repomocks.StoreRepository)
	aSvc := new(mocks.ActivityLogServiceInterface)

	svc := NewTerminService(tRepo, payRepo, poRepo, sRepo, aSvc, log)

	tid := "t1"
	userID := "u1"
	req := dto.RecordPaymentRequest{
		AmountPaid:    500,
		PaymentDate:   "2026-05-01",
		PaymentMethod: "cash",
	}

	tRepo.On("FindByID", ctx, tid).Return(&domain.POTermin{
		ID:        tid,
		POID:      "po1",
		AmountDue: 1000,
	}, nil)

	payRepo.On("Create", ctx, mock.Anything).Return(&domain.PaymentRecord{
		ID:         "pr1",
		TerminID:   tid,
		AmountPaid: 500,
	}, nil)

	tRepo.On("UpdateStatus", ctx, tid).Return(nil)
	poRepo.On("FindByID", ctx, "po1").Return(&domain.PurchaseOrder{ID: "po1", PONumber: "PO-001", StoreID: "s1"}, nil)
	aSvc.On("LogActivity", ctx, userID, "s1", domain.ActionPurchaseOrderPayment, domain.ModulePurchase, "po1", mock.Anything).Return()

	resp, err := svc.RecordPayment(ctx, tid, userID, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 500.0, resp.AmountPaid)

	// Test overpayment
	req.AmountPaid = 2000
	resp, err = svc.RecordPayment(ctx, tid, userID, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "payment would exceed termin amount")
}

func TestTerminService_CalculatePODebt(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	tRepo := new(repomocks.TerminRepository)
	poRepo := new(repomocks.PurchaseOrderRepository)
	svc := NewTerminService(tRepo, nil, poRepo, nil, nil, log)

	poID := "po1"
	poRepo.On("FindByID", ctx, poID).Return(&domain.PurchaseOrder{ID: poID, PONumber: "PO-001", TotalAmount: 1000}, nil)
	tRepo.On("DebtSummary", ctx, poID, 1000.0).Return(&domain.PODebtSummary{
		TotalAmount:   1000,
		TotalTermin:   1000,
		TotalPaid:     400,
		RemainingDebt: 600,
		Status:        "partial",
	}, nil)

	res, err := svc.CalculatePODebt(ctx, poID)
	assert.NoError(t, err)
	assert.Equal(t, 600.0, res.RemainingDebt)
	assert.Equal(t, "partial", res.Status)
}
