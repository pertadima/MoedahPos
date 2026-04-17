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

func TestPurchaseOrderService_Lifecycle(t *testing.T) {
	poRepo := new(repomocks.PurchaseOrderRepository)
	prodRepo := new(repomocks.ProductRepository)
	payRepo := new(repomocks.POPaymentRepository)
	terminRepo := new(repomocks.TerminRepository)
	payRecRepo := new(repomocks.PaymentRecordRepository)
	priceSvc := new(mocks.PriceHistoryServiceInterface)
	actSvc := new(mocks.ActivityLogServiceInterface)
	log := zerolog.Nop()

	svc := NewPurchaseOrderService(poRepo, prodRepo, payRepo, terminRepo, payRecRepo, priceSvc, actSvc, log)
	ctx := context.Background()

	t.Run("Create PO", func(t *testing.T) {
		req := &dto.CreatePORequest{
			SupplierID: ptrToStringPtr("supp1"),
			Items: []dto.POItemInput{
				{ProductID: "p1", Quantity: 10, UnitCost: 100},
			},
		}
		prodRepo.On("FindByID", ctx, "p1").Return(&domain.Product{ID: "p1", StoreID: "s1", CostPrice: 90}, nil).Once()
		poRepo.On("Create", ctx, mock.MatchedBy(func(p *domain.PurchaseOrder) bool {
			return p.TotalAmount == 1000 && p.StoreID == "s1"
		}), mock.Anything).Return(&domain.PurchaseOrder{ID: "po1", PONumber: "PO-001", TotalAmount: 1000}, nil).Once()

		actSvc.On("LogActivity", ctx, "u1", "s1", domain.ActionPurchaseOrderCreate, domain.ModulePurchase, "po1", mock.Anything).Return().Once()

		resp, err := svc.CreatePO(ctx, "s1", req, "u1")
		assert.NoError(t, err)
		assert.Equal(t, "PO-001", resp.PONumber)
	})

	t.Run("Submit PO", func(t *testing.T) {
		poRepo.On("FindByID", ctx, "po1").Return(&domain.PurchaseOrder{ID: "po1", Status: "draft"}, nil).Once()
		poRepo.On("Submit", ctx, "po1", "u1").Return(nil).Once()

		err := svc.SubmitPO(ctx, "po1", "u1")
		assert.NoError(t, err)
	})

	t.Run("Receive PO", func(t *testing.T) {
		po := &domain.PurchaseOrder{
			ID: "po1", StoreID: "s1", Status: "ordered",
			Items: []domain.POItem{{ProductID: "p1", Quantity: 10, UnitCost: 110}},
		}
		poRepo.On("FindByID", ctx, "po1").Return(po, nil).Once()
		prodRepo.On("FindByID", ctx, "p1").Return(&domain.Product{ID: "p1", CostPrice: 100, SellPrice: 200}, nil).Once()
		poRepo.On("Receive", ctx, "po1", "u1").Return(nil).Once()
		priceSvc.On("RecordChange", ctx, "p1", "s1", "u1", 100.0, 110.0, 200.0, 200.0, "purchase_order", mock.Anything, mock.Anything).Return(nil).Once()

		err := svc.ReceivePO(ctx, "po1", "u1")
		assert.NoError(t, err)
	})
}

func TestPurchaseOrderService_Payments(t *testing.T) {
	poRepo := new(repomocks.PurchaseOrderRepository)
	prodRepo := new(repomocks.ProductRepository)
	payRepo := new(repomocks.POPaymentRepository)
	terminRepo := new(repomocks.TerminRepository)
	payRecRepo := new(repomocks.PaymentRecordRepository)
	priceSvc := new(mocks.PriceHistoryServiceInterface)
	actSvc := new(mocks.ActivityLogServiceInterface)
	log := zerolog.Nop()

	svc := NewPurchaseOrderService(poRepo, prodRepo, payRepo, terminRepo, payRecRepo, priceSvc, actSvc, log)
	ctx := context.Background()

	t.Run("Create Global Payment - Allocate to Termins", func(t *testing.T) {
		poID := "po1"
		userID := "u1"
		req := dto.POPaymentRequest{Amount: 1500}

		poRepo.On("FindByID", ctx, poID).Return(&domain.PurchaseOrder{ID: poID, PONumber: "PO-1", Status: "received"}, nil).Once()
		payRepo.On("Create", ctx, mock.Anything).Return(&domain.POPayment{ID: "pay1", POID: poID, Amount: 1500}, nil).Once()
		actSvc.On("LogActivity", ctx, userID, "s1", domain.ActionPurchaseOrderPayment, domain.ModulePurchase, poID, mock.Anything).Return().Once()

		// Allocation logic
		termins := []*domain.POTermin{
			{ID: "t1", Amount: 1000, AmountPaid: 0},
			{ID: "t2", Amount: 1000, AmountPaid: 0},
		}
		terminRepo.On("FindByPO", ctx, poID).Return(termins, nil).Once()
		// 1500 should pay t1 entirely (1000) and t2 partially (500)
		payRecRepo.On("Create", ctx, mock.MatchedBy(func(pr domain.PaymentRecord) bool {
			return pr.TerminID == "t1" && pr.AmountPaid == 1000
		})).Return(&domain.PaymentRecord{ID: "pr1"}, nil).Once()
		terminRepo.On("UpdateStatus", ctx, "t1").Return(nil).Once()

		payRecRepo.On("Create", ctx, mock.MatchedBy(func(pr domain.PaymentRecord) bool {
			return pr.TerminID == "t2" && pr.AmountPaid == 500
		})).Return(&domain.PaymentRecord{ID: "pr2"}, nil).Once()
		terminRepo.On("UpdateStatus", ctx, "t2").Return(nil).Once()

		resp, err := svc.CreatePayment(ctx, poID, "s1", userID, req)
		assert.NoError(t, err)
		assert.Equal(t, 1500.0, resp.Amount)
		
		terminRepo.AssertExpectations(t)
		payRecRepo.AssertExpectations(t)
	})
}

func ptrToStringPtr(s string) *string {
	return &s
}
