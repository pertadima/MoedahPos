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

	t.Run("Receive PO Success", func(t *testing.T) {
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

	t.Run("Update PO - Error Not Found", func(t *testing.T) {
		poRepo.On("FindByID", ctx, "invalid").Return(nil, nil).Once()
		_, err := svc.UpdatePO(ctx, "invalid", &dto.UpdatePORequest{}, "s1")
		assert.ErrorIs(t, err, ErrPONotFound)
	})

	t.Run("Update PO - Error Not Editable", func(t *testing.T) {
		poRepo.On("FindByID", ctx, "po-rec").Return(&domain.PurchaseOrder{ID: "po-rec", Status: "received"}, nil).Once()
		_, err := svc.UpdatePO(ctx, "po-rec", &dto.UpdatePORequest{}, "s1")
		assert.ErrorIs(t, err, ErrPONotEditable)
	})

	t.Run("Cancel PO Success", func(t *testing.T) {
		poRepo.On("FindByID", ctx, "po1").Return(&domain.PurchaseOrder{ID: "po1", Status: "ordered"}, nil).Once()
		poRepo.On("Cancel", ctx, "po1").Return(nil).Once()
		err := svc.CancelPO(ctx, "po1")
		assert.NoError(t, err)
	})

	t.Run("Cancel PO - Error Cannot Cancel", func(t *testing.T) {
		poRepo.On("FindByID", ctx, "po1").Return(&domain.PurchaseOrder{ID: "po1", Status: "received"}, nil).Once()
		err := svc.CancelPO(ctx, "po1")
		assert.ErrorIs(t, err, ErrPOCannotCancel)
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

	t.Run("List Payments", func(t *testing.T) {
		payRepo.On("FindByPO", ctx, "po1").Return([]*domain.POPayment{{ID: "pay1", Amount: 1000}}, nil).Once()
		res, err := svc.ListPayments(ctx, "po1")
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("Payable Summary", func(t *testing.T) {
		payRepo.On("PayableSummary", ctx, "s1").Return(&dto.PayableSummary{TotalOutstanding: 5000}, nil).Once()
		res, err := svc.PayableSummary(ctx, "s1")
		assert.NoError(t, err)
		assert.Equal(t, 5000.0, res.TotalOutstanding)
	})
}

func TestPurchaseOrderService_Queries(t *testing.T) {
	poRepo := new(repomocks.PurchaseOrderRepository)
	payRepo := new(repomocks.POPaymentRepository)
	svc := NewPurchaseOrderService(poRepo, nil, payRepo, nil, nil, nil, nil, zerolog.Nop())
	ctx := context.Background()

	t.Run("Get PO Detail", func(t *testing.T) {
		poID := "po1"
		poRepo.On("FindByID", ctx, poID).Return(&domain.PurchaseOrder{ID: poID}, nil).Once()
		payRepo.On("AggregateByPO", ctx, poID, 0.0).Return(0.0, "unpaid", nil).Once()

		resp, err := svc.GetPO(ctx, poID)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("List POs", func(t *testing.T) {
		filter := dto.POListFilter{StoreID: "s1"}
		filter.Defaults()
		poRepo.On("FindAll", ctx, filter).Return([]*domain.PurchaseOrder{{ID: "po1"}}, 1, nil).Once()
		payRepo.On("PopulatePOPayments", ctx, mock.Anything).Return().Once()

		txns, meta, err := svc.ListPOs(ctx, filter)
		assert.NoError(t, err)
		assert.Len(t, txns, 1)
		assert.Equal(t, 1, meta.Total)
	})
}

func TestPurchaseOrderService_UpdatePO(t *testing.T) {
	poRepo := new(repomocks.PurchaseOrderRepository)
	prodRepo := new(repomocks.ProductRepository)
	svc := NewPurchaseOrderService(poRepo, prodRepo, nil, nil, nil, nil, nil, zerolog.Nop())
	ctx := context.Background()

	t.Run("Update Success", func(t *testing.T) {
		poID := "po1"
		req := &dto.UpdatePORequest{
			Items: []dto.POItemInput{{ProductID: "p1", Quantity: 5, UnitCost: 90}},
		}
		poRepo.On("FindByID", ctx, poID).Return(&domain.PurchaseOrder{ID: poID, StoreID: "s1", Status: "draft"}, nil).Once()
		prodRepo.On("FindByID", ctx, "p1").Return(&domain.Product{ID: "p1", StoreID: "s1"}, nil).Once()
		poRepo.On("Update", ctx, mock.Anything, mock.Anything).Return(&domain.PurchaseOrder{ID: poID}, nil).Once()

		resp, err := svc.UpdatePO(ctx, poID, req, "s1")
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})
}

func ptrToStringPtr(s string) *string {
	return &s
}
