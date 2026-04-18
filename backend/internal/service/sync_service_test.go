package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
	"github.com/rs/zerolog"
)

// Dummy Repositories
type mockCategoryRepo struct {
	cats []*domain.Category
	err  error
}

func (m *mockCategoryRepo) Create(ctx context.Context, cat *domain.Category) (*domain.Category, error) { return nil, nil }
func (m *mockCategoryRepo) FindAllByStore(ctx context.Context, storeID string) ([]*domain.Category, error) { return nil, nil }
func (m *mockCategoryRepo) FindByID(ctx context.Context, id string) (*domain.Category, error) { return nil, nil }
func (m *mockCategoryRepo) Update(ctx context.Context, cat *domain.Category) (*domain.Category, error) { return nil, nil }
func (m *mockCategoryRepo) SoftDelete(ctx context.Context, id string) error { return nil }
func (m *mockCategoryRepo) GetModifiedSince(ctx context.Context, storeID string, since time.Time) ([]*domain.Category, error) {
	return m.cats, m.err
}

type mockProductRepo struct {
	prods []*domain.Product
	err   error
}

func (m *mockProductRepo) Create(ctx context.Context, p *domain.Product) (*domain.Product, error) { return nil, nil }
func (m *mockProductRepo) FindAll(ctx context.Context, filter dto.ProductListFilter) ([]*domain.Product, int, error) { return nil, 0, nil }
func (m *mockProductRepo) FindByID(ctx context.Context, id string) (*domain.Product, error) { return nil, nil }
func (m *mockProductRepo) FindByBarcode(ctx context.Context, storeID, barcode string) (*domain.Product, error) { return nil, nil }
func (m *mockProductRepo) ExistsBySKU(ctx context.Context, storeID, sku string, excludeID string) (bool, error) { return false, nil }
func (m *mockProductRepo) Update(ctx context.Context, p *domain.Product) (*domain.Product, error) { return nil, nil }
func (m *mockProductRepo) SoftDelete(ctx context.Context, id string) error { return nil }
func (m *mockProductRepo) GetModifiedSince(ctx context.Context, storeID string, since time.Time) ([]*domain.Product, error) {
	return m.prods, m.err
}

type mockStockRepo struct {
	levels []*domain.StockLevel
	err    error
}

func (m *mockStockRepo) FindLevelsByStore(ctx context.Context, storeID string, lowStockOnly bool) ([]*domain.StockLevel, error) { return nil, nil }
func (m *mockStockRepo) FindLevelByProduct(ctx context.Context, productID, storeID string) (*domain.StockLevel, error) { return nil, nil }
func (m *mockStockRepo) SetMinQuantity(ctx context.Context, productID, storeID string, min float64) error { return nil }
func (m *mockStockRepo) Adjust(ctx context.Context, input domain.AdjustInput) (*domain.StockLevel, error) { return nil, nil }
func (m *mockStockRepo) FindMovements(ctx context.Context, filter dto.StockMovementFilter) ([]*domain.StockMovement, int, error) { return nil, 0, nil }
func (m *mockStockRepo) DeductStock(ctx context.Context, productID, storeID string, qty float64, refID, cashierID string) error { return nil }
func (m *mockStockRepo) GetModifiedSince(ctx context.Context, storeID string, since time.Time) ([]*domain.StockLevel, error) {
	return m.levels, m.err
}

type mockTxnRepo struct {
	txns []*domain.Transaction
	err  error
}

func (m *mockTxnRepo) Create(ctx context.Context, input domain.CreateTransactionInput) (*domain.Transaction, error) { return nil, nil }
func (m *mockTxnRepo) FindAll(ctx context.Context, filter dto.TransactionListFilter) ([]*domain.Transaction, int, error) { return nil, 0, nil }
func (m *mockTxnRepo) FindByID(ctx context.Context, id string) (*domain.Transaction, error) { return nil, nil }
func (m *mockTxnRepo) Void(ctx context.Context, txnID, userID string) error { return nil }
func (m *mockTxnRepo) GetDraftByTable(ctx context.Context, storeID, tableID string) (*domain.Transaction, error) { return nil, nil }
func (m *mockTxnRepo) UpdateDraftItems(ctx context.Context, txnID string, items []domain.CreateTransactionItemInput, subtotal, discountAmt, taxAmt, total float64, customerName, notes string) (*domain.Transaction, error) { return nil, nil }
func (m *mockTxnRepo) PayDraft(ctx context.Context, input domain.PayDraftInput, storeID, cashierID string) (*domain.Transaction, error) { return nil, nil }
func (m *mockTxnRepo) GetKDSTickets(ctx context.Context, storeID string) ([]*domain.Transaction, error) { return nil, nil }
func (m *mockTxnRepo) UpdateKDSItemStatus(ctx context.Context, itemID, status string) error { return nil }
func (m *mockTxnRepo) GetModifiedSince(ctx context.Context, storeID string, since time.Time) ([]*domain.Transaction, error) {
	return m.txns, m.err
}


func TestSyncService_Pull(t *testing.T) {
	nopLog := zerolog.Nop()

	t.Run("success standard", func(t *testing.T) {
		catRepo := &mockCategoryRepo{cats: []*domain.Category{{ID: "c1"}}}
		prodRepo := &mockProductRepo{prods: []*domain.Product{{ID: "p1"}}}
		stockRepo := &mockStockRepo{levels: []*domain.StockLevel{{ID: "sl1"}}}
		txnRepo := &mockTxnRepo{txns: []*domain.Transaction{{ID: "tx1"}}}

		svc := NewSyncService(catRepo, prodRepo, stockRepo, txnRepo, nopLog)
		
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
	})

	t.Run("empty response handles gracefully", func(t *testing.T) {
		catRepo := &mockCategoryRepo{cats: nil}
		prodRepo := &mockProductRepo{prods: nil}
		stockRepo := &mockStockRepo{levels: nil}
		txnRepo := &mockTxnRepo{txns: nil}

		svc := NewSyncService(catRepo, prodRepo, stockRepo, txnRepo, nopLog)
		
		res, err := svc.Pull(context.Background(), "s1", time.Now())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if res.Categories == nil || len(res.Categories) != 0 {
			t.Error("expected empty categories array")
		}
	})

	t.Run("error in category repo", func(t *testing.T) {
		catRepo := &mockCategoryRepo{err: errors.New("db error")}
		svc := NewSyncService(catRepo, &mockProductRepo{}, &mockStockRepo{}, &mockTxnRepo{}, nopLog)
		_, err := svc.Pull(context.Background(), "s1", time.Now())
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("error in product repo", func(t *testing.T) {
		prodRepo := &mockProductRepo{err: errors.New("db error")}
		svc := NewSyncService(&mockCategoryRepo{}, prodRepo, &mockStockRepo{}, &mockTxnRepo{}, nopLog)
		_, err := svc.Pull(context.Background(), "s1", time.Now())
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("error in stock repo", func(t *testing.T) {
		stockRepo := &mockStockRepo{err: errors.New("db error")}
		svc := NewSyncService(&mockCategoryRepo{}, &mockProductRepo{}, stockRepo, &mockTxnRepo{}, nopLog)
		_, err := svc.Pull(context.Background(), "s1", time.Now())
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("error in transaction repo", func(t *testing.T) {
		txnRepo := &mockTxnRepo{err: errors.New("db error")}
		svc := NewSyncService(&mockCategoryRepo{}, &mockProductRepo{}, &mockStockRepo{}, txnRepo, nopLog)
		_, err := svc.Pull(context.Background(), "s1", time.Now())
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}
