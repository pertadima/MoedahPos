package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
)

// Dummy Repositories
type mockCategoryRepo struct {
	cats []*domain.Category
	err  error
}

func (m *mockCategoryRepo) Create(_ context.Context, _ *domain.Category) (*domain.Category, error) {
	return nil, nil
}
func (m *mockCategoryRepo) FindAllByStore(_ context.Context, _ string) ([]*domain.Category, error) {
	return nil, nil
}
func (m *mockCategoryRepo) FindByID(_ context.Context, _ string) (*domain.Category, error) {
	return nil, nil
}
func (m *mockCategoryRepo) Update(_ context.Context, _ *domain.Category) (*domain.Category, error) {
	return nil, nil
}
func (m *mockCategoryRepo) SoftDelete(_ context.Context, _ string) error { return nil }
func (m *mockCategoryRepo) GetModifiedSince(_ context.Context, _ string, _ time.Time) ([]*domain.Category, error) {
	return m.cats, m.err
}

type mockProductRepo struct {
	prods []*domain.Product
	err   error
}

func (m *mockProductRepo) Create(_ context.Context, _ *domain.Product) (*domain.Product, error) {
	return nil, nil
}
func (m *mockProductRepo) FindAll(_ context.Context, _ dto.ProductListFilter) ([]*domain.Product, int, error) {
	return nil, 0, nil
}
func (m *mockProductRepo) FindByID(_ context.Context, _ string) (*domain.Product, error) {
	return nil, nil
}
func (m *mockProductRepo) FindByBarcode(_ context.Context, _ string, _ string) (*domain.Product, error) {
	return nil, nil
}
func (m *mockProductRepo) ExistsBySKU(_ context.Context, _ string, _ string, _ string) (bool, error) {
	return false, nil
}
func (m *mockProductRepo) Update(_ context.Context, _ *domain.Product) (*domain.Product, error) {
	return nil, nil
}
func (m *mockProductRepo) SoftDelete(_ context.Context, _ string) error { return nil }
func (m *mockProductRepo) GetModifiedSince(_ context.Context, _ string, _ time.Time) ([]*domain.Product, error) {
	return m.prods, m.err
}

type mockStockRepo struct {
	levels []*domain.StockLevel
	err    error
}

func (m *mockStockRepo) FindLevelsByStore(_ context.Context, _ string, _ bool) ([]*domain.StockLevel, error) {
	return nil, nil
}
func (m *mockStockRepo) FindLevelByProduct(_ context.Context, _ string, _ string) (*domain.StockLevel, error) {
	return nil, nil
}
func (m *mockStockRepo) SetMinQuantity(_ context.Context, _ string, _ string, _ float64) error {
	return nil
}
func (m *mockStockRepo) Adjust(_ context.Context, _ domain.AdjustInput) (*domain.StockLevel, error) {
	return nil, nil
}
func (m *mockStockRepo) FindMovements(_ context.Context, _ dto.StockMovementFilter) ([]*domain.StockMovement, int, error) {
	return nil, 0, nil
}
func (m *mockStockRepo) DeductStock(_ context.Context, _ string, _ string, _ float64, _ string, _ string) error {
	return nil
}
func (m *mockStockRepo) GetModifiedSince(_ context.Context, _ string, _ time.Time) ([]*domain.StockLevel, error) {
	return m.levels, m.err
}

type mockTxnRepo struct {
	txns []*domain.Transaction
	err  error
}

func (m *mockTxnRepo) Create(_ context.Context, _ domain.CreateTransactionInput) (*domain.Transaction, error) {
	return nil, nil
}
func (m *mockTxnRepo) FindAll(_ context.Context, _ dto.TransactionListFilter) ([]*domain.Transaction, int, error) {
	return nil, 0, nil
}
func (m *mockTxnRepo) FindByID(_ context.Context, _ string) (*domain.Transaction, error) {
	return nil, nil
}
func (m *mockTxnRepo) Void(_ context.Context, _ string, _ string) error { return nil }
func (m *mockTxnRepo) GetDraftByTable(_ context.Context, _ string, _ string) (*domain.Transaction, error) {
	return nil, nil
}
func (m *mockTxnRepo) UpdateDraftItems(_ context.Context, _ string, _ []domain.CreateTransactionItemInput, _ float64, _ float64, _ float64, _ float64, _ string, _ string) (*domain.Transaction, error) {
	return nil, nil
}
func (m *mockTxnRepo) PayDraft(_ context.Context, _ domain.PayDraftInput, _ string, _ string) (*domain.Transaction, error) {
	return nil, nil
}
func (m *mockTxnRepo) GetKDSTickets(_ context.Context, _ string) ([]*domain.Transaction, error) {
	return nil, nil
}
func (m *mockTxnRepo) UpdateKDSItemStatus(_ context.Context, _ string, _ string) error {
	return nil
}
func (m *mockTxnRepo) GetModifiedSince(_ context.Context, _ string, _ time.Time) ([]*domain.Transaction, error) {
	return m.txns, m.err
}

// mockCustomerSyncRepo implements repository.CustomerSyncRepository.
type mockCustomerSyncRepo struct {
	customers []*domain.Customer
	err       error
}

func (m *mockCustomerSyncRepo) GetModifiedSince(_ context.Context, _ string, _ time.Time) ([]*domain.Customer, error) {
	return m.customers, m.err
}

//nolint:gocognit,cyclop,funlen
func TestSyncService_Pull(t *testing.T) {
	nopLog := zerolog.Nop()

	t.Run("success standard", func(t *testing.T) {
		catRepo := &mockCategoryRepo{cats: []*domain.Category{{ID: "c1"}}}
		prodRepo := &mockProductRepo{prods: []*domain.Product{{ID: "p1"}}}
		stockRepo := &mockStockRepo{levels: []*domain.StockLevel{{ID: "sl1"}}}
		txnRepo := &mockTxnRepo{txns: []*domain.Transaction{{ID: "tx1"}}}
		custRepo := &mockCustomerSyncRepo{customers: []*domain.Customer{{ID: "cu1"}}}

		svc := NewSyncService(catRepo, prodRepo, stockRepo, txnRepo, custRepo, nopLog)

		res, err := svc.Pull(context.Background(), "s1", time.Now())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(res.Categories) != 1 || res.Categories[0].ID != "c1" {
			t.Errorf("expected 1 category, got %d", len(res.Categories))
		}
		if len(res.Products) != 1 || res.Products[0].ID != "p1" {
			t.Errorf("expected 1 product, got %d", len(res.Products))
		}
		if len(res.StockLevels) != 1 || res.StockLevels[0].ID != "sl1" {
			t.Errorf("expected 1 stock level, got %d", len(res.StockLevels))
		}
		if len(res.Transactions) != 1 || res.Transactions[0].ID != "tx1" {
			t.Errorf("expected 1 transaction, got %d", len(res.Transactions))
		}
		if len(res.Customers) != 1 || res.Customers[0].ID != "cu1" {
			t.Errorf("expected 1 customer, got %d", len(res.Customers))
		}
	})

	t.Run("empty response handles gracefully", func(t *testing.T) {
		catRepo := &mockCategoryRepo{cats: nil}
		prodRepo := &mockProductRepo{prods: nil}
		stockRepo := &mockStockRepo{levels: nil}
		txnRepo := &mockTxnRepo{txns: nil}
		custRepo := &mockCustomerSyncRepo{customers: nil}

		svc := NewSyncService(catRepo, prodRepo, stockRepo, txnRepo, custRepo, nopLog)

		res, err := svc.Pull(context.Background(), "s1", time.Now())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if res.Categories == nil || len(res.Categories) != 0 {
			t.Error("expected empty categories array")
		}
		if res.Customers == nil || len(res.Customers) != 0 {
			t.Error("expected empty customers array")
		}
	})

	t.Run("future timestamp — still returns all data", func(t *testing.T) {
		futureTime := time.Now().Add(24 * time.Hour)
		custRepo := &mockCustomerSyncRepo{customers: []*domain.Customer{{ID: "c-future"}}}
		svc := NewSyncService(&mockCategoryRepo{}, &mockProductRepo{}, &mockStockRepo{}, &mockTxnRepo{}, custRepo, nopLog)
		res, err := svc.Pull(context.Background(), "s1", futureTime)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// The service passes the timestamp to the repo; the mock returns data regardless
		if res == nil {
			t.Error("expected non-nil response")
		}
	})

	t.Run("error in category repo", func(t *testing.T) {
		catRepo := &mockCategoryRepo{err: errors.New("db error")}
		svc := NewSyncService(catRepo, &mockProductRepo{}, &mockStockRepo{}, &mockTxnRepo{}, &mockCustomerSyncRepo{}, nopLog)
		_, err := svc.Pull(context.Background(), "s1", time.Now())
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("error in product repo", func(t *testing.T) {
		prodRepo := &mockProductRepo{err: errors.New("db error")}
		svc := NewSyncService(&mockCategoryRepo{}, prodRepo, &mockStockRepo{}, &mockTxnRepo{}, &mockCustomerSyncRepo{}, nopLog)
		_, err := svc.Pull(context.Background(), "s1", time.Now())
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("error in stock repo", func(t *testing.T) {
		stockRepo := &mockStockRepo{err: errors.New("db error")}
		svc := NewSyncService(&mockCategoryRepo{}, &mockProductRepo{}, stockRepo, &mockTxnRepo{}, &mockCustomerSyncRepo{}, nopLog)
		_, err := svc.Pull(context.Background(), "s1", time.Now())
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("error in transaction repo", func(t *testing.T) {
		txnRepo := &mockTxnRepo{err: errors.New("db error")}
		svc := NewSyncService(&mockCategoryRepo{}, &mockProductRepo{}, &mockStockRepo{}, txnRepo, &mockCustomerSyncRepo{}, nopLog)
		_, err := svc.Pull(context.Background(), "s1", time.Now())
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("error in customer repo", func(t *testing.T) {
		custRepo := &mockCustomerSyncRepo{err: errors.New("db error")}
		svc := NewSyncService(&mockCategoryRepo{}, &mockProductRepo{}, &mockStockRepo{}, &mockTxnRepo{}, custRepo, nopLog)
		_, err := svc.Pull(context.Background(), "s1", time.Now())
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}
