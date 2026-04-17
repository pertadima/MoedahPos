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

func TestTransactionService_Checkout(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	t.Run("Retail Success", func(t *testing.T) {
		tRepo := new(repomocks.TransactionRepository)
		pRepo := new(repomocks.ProductRepository)
		sRepo := new(repomocks.StockRepository)
		mRepo := new(repomocks.MenuItemRepository)
		bSvc := new(mocks.BatchStockServiceInterface)
		aSvc := new(mocks.ActivityLogServiceInterface)

		pid := "00000000-0000-0000-0000-000000000001"
		req := &dto.CreateTransactionRequest{
			Items: []dto.TxItemInput{
				{ProductID: pid, Quantity: 2},
			},
			PaymentAmount: 200,
			PaymentMethod: "cash",
		}

		pRepo.On("FindByID", ctx, pid).Return(&domain.Product{
			ID:        pid,
			StoreID:   "s1",
			Name:      "Product 1",
			SellPrice: 100,
			IsActive:  true,
		}, nil)
		sRepo.On("FindLevelByProduct", ctx, pid, "s1").Return(&domain.StockLevel{Quantity: 10}, nil)

		tRepo.On("Create", ctx, mock.Anything).Return(&domain.Transaction{
			ID:    "t1",
			Total: 200,
		}, nil)

		bSvc.On("DeductStockFIFO", ctx, pid, "s1", 2.0).Return(nil)
		aSvc.On("LogActivity", ctx, "u1", "s1", domain.ActionTransactionCreate, domain.ModuleTransaction, "t1", mock.Anything).Return()

		s := NewTransactionService(tRepo, pRepo, sRepo, mRepo, bSvc, aSvc, log)
		resp, err := s.Checkout(ctx, "s1", req, "u1")

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "t1", resp.ID)

		pRepo.AssertExpectations(t)
		sRepo.AssertExpectations(t)
		tRepo.AssertExpectations(t)
		bSvc.AssertExpectations(t)
	})

	t.Run("Insufficient Payment", func(t *testing.T) {
		tRepo := new(repomocks.TransactionRepository)
		pRepo := new(repomocks.ProductRepository)
		sRepo := new(repomocks.StockRepository)
		mRepo := new(repomocks.MenuItemRepository)
		bSvc := new(mocks.BatchStockServiceInterface)
		aSvc := new(mocks.ActivityLogServiceInterface)

		pid := "00000000-0000-0000-0000-000000000001"
		req := &dto.CreateTransactionRequest{
			Items: []dto.TxItemInput{
				{ProductID: pid, Quantity: 1},
			},
			PaymentAmount: 50, // Total is 100
		}

		pRepo.On("FindByID", ctx, pid).Return(&domain.Product{
			ID:        pid,
			StoreID:   "s1",
			SellPrice: 100,
			IsActive:  true,
		}, nil)
		sRepo.On("FindLevelByProduct", ctx, pid, "s1").Return(&domain.StockLevel{Quantity: 10}, nil)

		s := NewTransactionService(tRepo, pRepo, sRepo, mRepo, bSvc, aSvc, log)
		resp, err := s.Checkout(ctx, "s1", req, "u1")

		assert.ErrorIs(t, err, ErrInsuficientPayment)
		assert.Nil(t, resp)
	})

	t.Run("Restaurant Success", func(t *testing.T) {
		tRepo := new(repomocks.TransactionRepository)
		pRepo := new(repomocks.ProductRepository)
		sRepo := new(repomocks.StockRepository)
		mRepo := new(repomocks.MenuItemRepository)
		bSvc := new(mocks.BatchStockServiceInterface)
		aSvc := new(mocks.ActivityLogServiceInterface)

		mid := "00000000-0000-0000-0000-000000000001"
		iid := "00000000-0000-0000-0000-000000000002"
		req := &dto.CreateTransactionRequest{
			Items: []dto.TxItemInput{
				{MenuItemID: mid, Quantity: 1},
			},
			PaymentAmount: 150,
			PaymentMethod: "cash",
		}

		mRepo.On("FindByID", ctx, mid).Return(&domain.MenuItem{
			ID:        mid,
			StoreID:   "s1",
			Name:      "Menu 1",
			SellPrice: 150,
			Ingredients: []domain.MenuItemIngredient{
				{ProductID: iid, Quantity: 2, ProductName: "Ingredient 1"},
			},
		}, nil)

		sRepo.On("FindLevelByProduct", ctx, iid, "s1").Return(&domain.StockLevel{ProductID: iid, Quantity: 10}, nil)
		pRepo.On("FindByID", ctx, iid).Return(&domain.Product{ID: iid, CostPrice: 50}, nil)

		tRepo.On("Create", ctx, mock.Anything).Return(&domain.Transaction{
			ID:    "t2",
			Total: 150,
		}, nil)

		sRepo.On("DeductStock", ctx, iid, "s1", 2.0, "t2", "u1").Return(nil)
		bSvc.On("DeductStockFIFO", ctx, iid, "s1", 2.0).Return(nil)
		aSvc.On("LogActivity", ctx, "u1", "s1", mock.Anything, mock.Anything, "t2", mock.Anything).Return()

		s := NewTransactionService(tRepo, pRepo, sRepo, mRepo, bSvc, aSvc, log)
		resp, err := s.Checkout(ctx, "s1", req, "u1")

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		mRepo.AssertExpectations(t)
	})
}

func TestTransactionService_VoidTransaction(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	t.Run("Success", func(t *testing.T) {
		tRepo := new(repomocks.TransactionRepository)
		aSvc := new(mocks.ActivityLogServiceInterface)

		tRepo.On("FindByID", ctx, "t1").Return(&domain.Transaction{ID: "t1", StoreID: "s1", Status: "completed"}, nil)
		tRepo.On("Void", ctx, "t1", "u1").Return(nil)
		aSvc.On("LogActivity", ctx, "u1", "s1", domain.ActionTransactionCancel, domain.ModuleTransaction, "t1", mock.Anything).Return()

		s := NewTransactionService(tRepo, nil, nil, nil, nil, aSvc, log)
		err := s.VoidTransaction(ctx, "t1", "u1")

		assert.NoError(t, err)
		tRepo.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		tRepo := new(repomocks.TransactionRepository)
		tRepo.On("FindByID", ctx, "t2").Return(nil, nil)

		s := NewTransactionService(tRepo, nil, nil, nil, nil, nil, log)
		err := s.VoidTransaction(ctx, "t2", "u1")

		assert.ErrorIs(t, err, ErrTransactionNotFound)
	})
}

func TestTransactionService_Drafts(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	t.Run("Create Draft Success", func(t *testing.T) {
		tRepo := new(repomocks.TransactionRepository)
		pRepo := new(repomocks.ProductRepository)
		sRepo := new(repomocks.StockRepository)
		mRepo := new(repomocks.MenuItemRepository)
		bSvc := new(mocks.BatchStockServiceInterface)
		aSvc := new(mocks.ActivityLogServiceInterface)

		pid := "00000000-0000-0000-0000-000000000001"
		req := &dto.CreateDraftRequest{
			TableID: "00000000-0000-0000-0000-000000000002",
			Items:   []dto.TxItemInput{{ProductID: pid, Quantity: 1}},
		}

		pRepo.On("FindByID", ctx, pid).Return(&domain.Product{ID: pid, StoreID: "s1", SellPrice: 100, IsActive: true}, nil)
		tRepo.On("Create", ctx, mock.MatchedBy(func(in domain.CreateTransactionInput) bool {
			return in.Status == "draft"
		})).Return(&domain.Transaction{ID: "d1", Status: "draft"}, nil)

		s := NewTransactionService(tRepo, pRepo, sRepo, mRepo, bSvc, aSvc, log)
		resp, err := s.CreateDraft(ctx, "s1", "u1", req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "draft", resp.Status)
	})

	t.Run("Pay Draft Success", func(t *testing.T) {
		tRepo := new(repomocks.TransactionRepository)
		aSvc := new(mocks.ActivityLogServiceInterface)
		s := NewTransactionService(tRepo, nil, nil, nil, nil, aSvc, log)

		tid := "00000000-0000-0000-0000-000000000001"
		tRepo.On("FindByID", ctx, tid).Return(&domain.Transaction{ID: tid, StoreID: "s1", Status: "draft", Total: 100}, nil)
		tRepo.On("PayDraft", ctx, mock.Anything, "s1", "u1").Return(&domain.Transaction{ID: tid, Status: "completed"}, nil)
		aSvc.On("LogActivity", ctx, "u1", "s1", domain.ActionTransactionCreate, domain.ModuleTransaction, tid, mock.Anything).Return()

		resp, err := s.PayDraft(ctx, "s1", tid, "u1", &dto.PayDraftRequest{PaymentAmount: 100, PaymentMethod: "cash"})

		assert.NoError(t, err)
		assert.Equal(t, "completed", resp.Status)
	})
}
