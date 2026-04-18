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

	t.Run("Create Draft MenuItem Success", func(t *testing.T) {
		tRepo := new(repomocks.TransactionRepository)
		pRepo := new(repomocks.ProductRepository)
		mRepo := new(repomocks.MenuItemRepository)

		mid := "m1"
		iid := "i1"
		req := &dto.CreateDraftRequest{
			TableID: "table1",
			Items:   []dto.TxItemInput{{MenuItemID: mid, Quantity: 1}},
		}

		mRepo.On("FindByID", ctx, mid).Return(&domain.MenuItem{
			ID: mid, StoreID: "s1", Name: "M1", SellPrice: 150,
			Ingredients: []domain.MenuItemIngredient{{ProductID: iid, Quantity: 2}},
		}, nil)
		pRepo.On("FindByID", ctx, iid).Return(&domain.Product{ID: iid, CostPrice: 50}, nil)
		tRepo.On("Create", ctx, mock.Anything).Return(&domain.Transaction{ID: "d2", Status: "draft"}, nil)

		s := NewTransactionService(tRepo, pRepo, nil, mRepo, nil, nil, log)
		resp, err := s.CreateDraft(ctx, "s1", "u1", req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
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

func TestTransactionService_GetTransaction(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	tRepo := new(repomocks.TransactionRepository)
	s := NewTransactionService(tRepo, nil, nil, nil, nil, nil, log)

	tid := "t1"
	tRepo.On("FindByID", ctx, tid).Return(&domain.Transaction{ID: tid, Total: 100}, nil)

	resp, err := s.GetTransaction(ctx, tid)
	assert.NoError(t, err)
	assert.Equal(t, tid, resp.ID)
}

func TestTransactionService_UpdateDraftItems(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	tRepo := new(repomocks.TransactionRepository)
	pRepo := new(repomocks.ProductRepository)
	sRepo := new(repomocks.StockRepository)
	mRepo := new(repomocks.MenuItemRepository)
	s := NewTransactionService(tRepo, pRepo, sRepo, mRepo, nil, nil, log)

	tid := "d1"
	pid := "p1"
	req := &dto.UpdateDraftRequest{
		Items: []dto.TxItemInput{{ProductID: pid, Quantity: 2}},
	}

	pRepo.On("FindByID", ctx, pid).Return(&domain.Product{ID: pid, StoreID: "s1", SellPrice: 100, IsActive: true}, nil)
	tRepo.On("FindByID", ctx, tid).Return(&domain.Transaction{ID: tid, StoreID: "s1", Status: "draft"}, nil)
	tRepo.On("UpdateDraftItems", ctx, tid, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(&domain.Transaction{ID: tid, Status: "draft"}, nil)

	resp, err := s.UpdateDraftItems(ctx, "s1", tid, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestTransactionService_KDS(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	tRepo := new(repomocks.TransactionRepository)
	s := NewTransactionService(tRepo, nil, nil, nil, nil, nil, log)

	tRepo.On("GetKDSTickets", ctx, "s1").Return([]*domain.Transaction{{ID: "t1"}}, nil)
	resp, err := s.GetKDSTickets(ctx, "s1")
	assert.NoError(t, err)
	assert.Len(t, resp, 1)

	tRepo.On("UpdateKDSItemStatus", ctx, "item1", "completed").Return(nil)
	err = s.UpdateKDSItemStatus(ctx, "item1", &dto.UpdateKDSItemStatusRequest{Status: "completed"})
	assert.NoError(t, err)
}

func TestTransactionService_Checkout_MoreBranches(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	t.Run("Cart Discount Success", func(t *testing.T) {
		tRepo := new(repomocks.TransactionRepository)
		pRepo := new(repomocks.ProductRepository)
		sRepo := new(repomocks.StockRepository)
		bSvc := new(mocks.BatchStockServiceInterface)
		aSvc := new(mocks.ActivityLogServiceInterface)

		pid := "p1"
		req := &dto.CreateTransactionRequest{
			Items:             []dto.TxItemInput{{ProductID: pid, Quantity: 1}},
			CartDiscountType:  "FIXED",
			CartDiscountValue: 20,
			PaymentAmount:     100, // (100 - 20) = 80 total
		}

		pRepo.On("FindByID", ctx, pid).Return(&domain.Product{ID: pid, StoreID: "s1", SellPrice: 100, IsActive: true}, nil)
		sRepo.On("FindLevelByProduct", ctx, pid, "s1").Return(&domain.StockLevel{Quantity: 10}, nil)
		tRepo.On("Create", ctx, mock.Anything).Return(&domain.Transaction{ID: "t1"}, nil)
		bSvc.On("DeductStockFIFO", ctx, pid, "s1", 1.0).Return(nil)
		aSvc.On("LogActivity", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

		s := NewTransactionService(tRepo, pRepo, sRepo, nil, bSvc, aSvc, log)
		resp, err := s.Checkout(ctx, "s1", req, "u1")

		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("Inactive Product", func(t *testing.T) {
		pRepo := new(repomocks.ProductRepository)
		pid := "p1"
		req := &dto.CreateTransactionRequest{Items: []dto.TxItemInput{{ProductID: pid, Quantity: 1}}}
		pRepo.On("FindByID", ctx, pid).Return(&domain.Product{ID: pid, StoreID: "s1", IsActive: false, Name: "P1"}, nil)

		s := NewTransactionService(nil, pRepo, nil, nil, nil, nil, log)
		_, err := s.Checkout(ctx, "s1", req, "u1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "inactive")
	})

	t.Run("Insufficient Stock", func(t *testing.T) {
		pRepo := new(repomocks.ProductRepository)
		sRepo := new(repomocks.StockRepository)
		pid := "p1"
		req := &dto.CreateTransactionRequest{Items: []dto.TxItemInput{{ProductID: pid, Quantity: 10}}}
		pRepo.On("FindByID", ctx, pid).Return(&domain.Product{ID: pid, StoreID: "s1", IsActive: true, Name: "P1"}, nil)
		sRepo.On("FindLevelByProduct", ctx, pid, "s1").Return(&domain.StockLevel{Quantity: 5}, nil)

		s := NewTransactionService(nil, pRepo, sRepo, nil, nil, nil, log)
		_, err := s.Checkout(ctx, "s1", req, "u1")
		assert.ErrorIs(t, err, ErrInsufficientStock)
	})

	t.Run("MenuItem Ingredient Insufficient Stock", func(t *testing.T) {
		mRepo := new(repomocks.MenuItemRepository)
		sRepo := new(repomocks.StockRepository)
		mid := "m1"
		iid := "i1"
		req := &dto.CreateTransactionRequest{Items: []dto.TxItemInput{{MenuItemID: mid, Quantity: 1}}}
		mRepo.On("FindByID", ctx, mid).Return(&domain.MenuItem{
			ID: mid, StoreID: "s1", Name: "M1",
			Ingredients: []domain.MenuItemIngredient{{ProductID: iid, Quantity: 2, ProductName: "I1"}},
		}, nil)
		sRepo.On("FindLevelByProduct", ctx, iid, "s1").Return(&domain.StockLevel{Quantity: 1}, nil)

		s := NewTransactionService(nil, nil, sRepo, mRepo, nil, nil, log)
		_, err := s.Checkout(ctx, "s1", req, "u1")
		assert.ErrorIs(t, err, ErrInsufficientStock)
	})

	t.Run("Cart Discount Percentage Excess", func(t *testing.T) {
		tRepo := new(repomocks.TransactionRepository)
		pRepo := new(repomocks.ProductRepository)
		sRepo := new(repomocks.StockRepository)
		bSvc := new(mocks.BatchStockServiceInterface)
		aSvc := new(mocks.ActivityLogServiceInterface)

		pid := "p1"
		req := &dto.CreateTransactionRequest{
			Items:             []dto.TxItemInput{{ProductID: pid, Quantity: 1}},
			CartDiscountType:  "PERCENTAGE",
			CartDiscountValue: 150, // 150% discount
			PaymentAmount:     0,
		}

		pRepo.On("FindByID", ctx, pid).Return(&domain.Product{ID: pid, StoreID: "s1", SellPrice: 100, IsActive: true}, nil)
		sRepo.On("FindLevelByProduct", ctx, pid, "s1").Return(&domain.StockLevel{Quantity: 10}, nil)
		tRepo.On("Create", ctx, mock.Anything).Return(&domain.Transaction{ID: "t1"}, nil)
		bSvc.On("DeductStockFIFO", ctx, pid, "s1", 1.0).Return(nil)
		aSvc.On("LogActivity", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

		s := NewTransactionService(tRepo, pRepo, sRepo, nil, bSvc, aSvc, log)
		resp, err := s.Checkout(ctx, "s1", req, "u1")

		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})
}

func TestTransactionService_VoidTransaction_Errors(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	t.Run("Already Voided", func(t *testing.T) {
		tRepo := new(repomocks.TransactionRepository)
		tRepo.On("FindByID", ctx, "t1").Return(&domain.Transaction{ID: "t1", Status: "voided"}, nil)

		s := NewTransactionService(tRepo, nil, nil, nil, nil, nil, log)
		err := s.VoidTransaction(ctx, "t1", "u1")
		assert.ErrorIs(t, err, ErrTransactionAlreadyVoided)
	})

	t.Run("Repo Error", func(t *testing.T) {
		tRepo := new(repomocks.TransactionRepository)
		tRepo.On("FindByID", ctx, "t1").Return(&domain.Transaction{ID: "t1", Status: "completed"}, nil)
		tRepo.On("Void", ctx, "t1", "u1").Return(assert.AnError)

		s := NewTransactionService(tRepo, nil, nil, nil, nil, nil, log)
		err := s.VoidTransaction(ctx, "t1", "u1")
		assert.Error(t, err)
	})
}

func TestTransactionService_Drafts_Errors(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	t.Run("Pay Draft Insufficient Payment", func(t *testing.T) {
		tRepo := new(repomocks.TransactionRepository)
		tid := "d1"
		tRepo.On("FindByID", ctx, tid).Return(&domain.Transaction{ID: tid, StoreID: "s1", Status: "draft", Total: 100}, nil)

		s := NewTransactionService(tRepo, nil, nil, nil, nil, nil, log)
		_, err := s.PayDraft(ctx, "s1", tid, "u1", &dto.PayDraftRequest{PaymentAmount: 50})
		assert.ErrorIs(t, err, ErrInsuficientPayment)
	})

	t.Run("Pay Draft Not Found", func(t *testing.T) {
		tRepo := new(repomocks.TransactionRepository)
		tRepo.On("FindByID", ctx, "missing").Return(nil, nil)

		s := NewTransactionService(tRepo, nil, nil, nil, nil, nil, log)
		_, err := s.PayDraft(ctx, "s1", "missing", "u1", &dto.PayDraftRequest{})
		assert.ErrorIs(t, err, ErrDraftNotFound)
	})

	t.Run("Update Draft Items Not Found", func(t *testing.T) {
		tRepo := new(repomocks.TransactionRepository)
		tRepo.On("FindByID", ctx, "missing").Return(nil, nil)

		s := NewTransactionService(tRepo, nil, nil, nil, nil, nil, log)
		_, err := s.UpdateDraftItems(ctx, "s1", "missing", &dto.UpdateDraftRequest{})
		assert.ErrorIs(t, err, ErrDraftNotFound)
	})
}

func TestTransactionService_ListTransactions(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	t.Run("Success", func(t *testing.T) {
		tRepo := new(repomocks.TransactionRepository)
		tRepo.On("FindAll", ctx, mock.Anything).Return([]*domain.Transaction{{ID: "t1"}}, 1, nil)

		s := NewTransactionService(tRepo, nil, nil, nil, nil, nil, log)
		txns, meta, err := s.ListTransactions(ctx, dto.TransactionListFilter{})
		assert.NoError(t, err)
		assert.Len(t, txns, 1)
		assert.Equal(t, 1, meta.Total)
	})

	t.Run("Error", func(t *testing.T) {
		tRepo := new(repomocks.TransactionRepository)
		tRepo.On("FindAll", ctx, mock.Anything).Return(nil, 0, assert.AnError)

		s := NewTransactionService(tRepo, nil, nil, nil, nil, nil, log)
		_, _, err := s.ListTransactions(ctx, dto.TransactionListFilter{})
		assert.Error(t, err)
	})
}

func TestTransactionService_GetDraftByTable(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	t.Run("Found", func(t *testing.T) {
		tRepo := new(repomocks.TransactionRepository)
		tRepo.On("GetDraftByTable", ctx, "s1", "table1").Return(&domain.Transaction{ID: "d1"}, nil)

		s := NewTransactionService(tRepo, nil, nil, nil, nil, nil, log)
		resp, err := s.GetDraftByTable(ctx, "s1", "table1")
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("Not Found", func(t *testing.T) {
		tRepo := new(repomocks.TransactionRepository)
		tRepo.On("GetDraftByTable", ctx, "s1", "table2").Return(nil, nil)

		s := NewTransactionService(tRepo, nil, nil, nil, nil, nil, log)
		resp, err := s.GetDraftByTable(ctx, "s1", "table2")
		assert.NoError(t, err)
		assert.Nil(t, resp)
	})
}

func TestTransactionService_toTransactionResponse_Full(t *testing.T) {
	tableID := "table1"
	now := time.Now()
	txn := &domain.Transaction{
		ID:            "t1",
		TableID:       &tableID,
		TableNumber:   stringToPtr("T1"),
		CustomerName:  "John",
		CustomerPhone: "123",
		Status:        "completed",
		CreatedAt:     now,
		UpdatedAt:     now,
		Items: []domain.TransactionItem{
			{
				ID:          "ti1",
				ProductName: "P1",
				CompletedAt: &now,
			},
		},
	}

	resp := toTransactionResponse(txn)
	assert.Equal(t, "t1", resp.ID)
	assert.Equal(t, "table1", *resp.TableID)
	assert.Equal(t, "T1", *resp.TableNumber)
	assert.Equal(t, "John", resp.CustomerName)
	assert.Equal(t, "123", resp.CustomerPhone)
	assert.Len(t, resp.Items, 1)
	assert.NotNil(t, resp.Items[0].CompletedAt)
}
