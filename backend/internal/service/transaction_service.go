package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
	"github.com/moedahpos/backend/internal/repository"
	"github.com/rs/zerolog"
)

// Transaction-specific sentinel errors.
var (
	ErrTransactionNotFound     = errors.New("transaction not found")
	ErrTransactionAlreadyVoided = errors.New("transaction is already voided")
	ErrInsufficientStock        = errors.New("insufficient stock for one or more items")
	ErrInsuficientPayment       = errors.New("payment amount is less than total")
)

// TransactionService implements cashier checkout logic.
type TransactionService struct {
	txnRepo     repository.TransactionRepository
	productRepo repository.ProductRepository
	stockRepo   repository.StockRepository
	log         zerolog.Logger
}

func NewTransactionService(
	txnRepo repository.TransactionRepository,
	productRepo repository.ProductRepository,
	stockRepo repository.StockRepository,
	log zerolog.Logger,
) *TransactionService {
	return &TransactionService{txnRepo: txnRepo, productRepo: productRepo, stockRepo: stockRepo, log: log}
}

// Checkout processes a sale: validates stock, calculates totals, persists atomically.
func (s *TransactionService) Checkout(ctx context.Context, storeID string, req *dto.CreateTransactionRequest, cashierID string) (*dto.TransactionResponse, error) {
	var (
		inputItems  []domain.CreateTransactionItemInput
		subtotal    float64
		discountAmt float64
		taxAmt      float64
	)

	for _, item := range req.Items {
		product, err := s.productRepo.FindByID(ctx, item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("finding product %s: %w", item.ProductID, err)
		}
		if product == nil || product.StoreID != storeID {
			return nil, fmt.Errorf("%w: product %s", ErrProductNotFound, item.ProductID)
		}
		if !product.IsActive {
			return nil, fmt.Errorf("product %s is inactive", product.Name)
		}

		// Stock check (advisory — race condition acceptable in Phase 3)
		level, err := s.stockRepo.FindLevelByProduct(ctx, item.ProductID, storeID)
		if err != nil {
			return nil, fmt.Errorf("checking stock: %w", err)
		}
		if level != nil && level.Quantity < item.Quantity {
			return nil, fmt.Errorf("%w: %s (have %.2f, need %.2f)",
				ErrInsufficientStock, product.Name, level.Quantity, item.Quantity)
		}

		// Price calculation
		lineGross := product.SellPrice * item.Quantity
		lineDiscount := lineGross * item.DiscountPct / 100
		lineNet := lineGross - lineDiscount
		lineTax := lineNet * product.TaxRate / 100
		lineSubtotal := lineNet + lineTax

		subtotal += lineNet
		discountAmt += lineDiscount
		taxAmt += lineTax

		pid := item.ProductID
		inputItems = append(inputItems, domain.CreateTransactionItemInput{
			ProductID:   &pid,
			ProductName: product.Name,
			SKU:         product.SKU,
			Quantity:    item.Quantity,
			UnitPrice:   product.SellPrice,
			DiscountPct: item.DiscountPct,
			TaxRate:     product.TaxRate,
			Subtotal:    lineSubtotal,
		})
	}

	total := subtotal + taxAmt
	if req.PaymentAmount < total {
		return nil, ErrInsuficientPayment
	}

	txn, err := s.txnRepo.Create(ctx, domain.CreateTransactionInput{
		StoreID:       storeID,
		CashierID:     cashierID,
		CustomerName:  req.CustomerName,
		CustomerPhone: req.CustomerPhone,
		PaymentMethod: req.PaymentMethod,
		PaymentAmount: req.PaymentAmount,
		ChangeAmount:  req.PaymentAmount - total,
		Notes:         req.Notes,
		Subtotal:      subtotal,
		DiscountAmt:   discountAmt,
		TaxAmt:        taxAmt,
		Total:         total,
		Items:         inputItems,
	})
	if err != nil {
		return nil, fmt.Errorf("creating transaction: %w", err)
	}

	s.log.Info().Str("txn_id", txn.ID).Float64("total", total).Msg("transaction completed")
	return toTransactionResponse(txn), nil
}

// ListTransactions returns a paginated list of transactions for a store.
func (s *TransactionService) ListTransactions(ctx context.Context, filter dto.TransactionListFilter) ([]*dto.TransactionResponse, dto.PaginationMeta, error) {
	filter.Defaults()
	txns, total, err := s.txnRepo.FindAll(ctx, filter)
	if err != nil {
		return nil, dto.PaginationMeta{}, fmt.Errorf("listing transactions: %w", err)
	}
	resp := make([]*dto.TransactionResponse, 0, len(txns))
	for _, t := range txns {
		resp = append(resp, toTransactionResponse(t))
	}
	return resp, dto.NewMeta(filter.PaginationQuery, total), nil
}

// GetTransaction returns a single transaction with all its items (receipt).
func (s *TransactionService) GetTransaction(ctx context.Context, id string) (*dto.TransactionResponse, error) {
	txn, err := s.txnRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("finding transaction: %w", err)
	}
	if txn == nil {
		return nil, ErrTransactionNotFound
	}
	return toTransactionResponse(txn), nil
}

// VoidTransaction reverses a completed transaction and restores stock.
func (s *TransactionService) VoidTransaction(ctx context.Context, id, userID string) error {
	txn, err := s.txnRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("finding transaction: %w", err)
	}
	if txn == nil {
		return ErrTransactionNotFound
	}
	if txn.Status == "voided" {
		return ErrTransactionAlreadyVoided
	}
	if err := s.txnRepo.Void(ctx, id, userID); err != nil {
		return fmt.Errorf("voiding transaction: %w", err)
	}
	s.log.Info().Str("txn_id", id).Str("voided_by", userID).Msg("transaction voided")
	return nil
}

// ─── Mapper ───────────────────────────────────────────────────────────────────

func toTransactionResponse(t *domain.Transaction) *dto.TransactionResponse {
	items := make([]dto.TransactionItemResponse, 0, len(t.Items))
	for _, ti := range t.Items {
		items = append(items, dto.TransactionItemResponse{
			ID:          ti.ID,
			ProductID:   ti.ProductID,
			ProductName: ti.ProductName,
			SKU:         ti.SKU,
			Quantity:    ti.Quantity,
			UnitPrice:   ti.UnitPrice,
			DiscountPct: ti.DiscountPct,
			TaxRate:     ti.TaxRate,
			Subtotal:    ti.Subtotal,
		})
	}
	return &dto.TransactionResponse{
		ID:            t.ID,
		StoreID:       t.StoreID,
		CashierID:     t.CashierID,
		CashierName:   t.CashierName,
		CustomerName:  t.CustomerName,
		CustomerPhone: t.CustomerPhone,
		Subtotal:      t.Subtotal,
		DiscountAmt:   t.DiscountAmt,
		TaxAmt:        t.TaxAmt,
		Total:         t.Total,
		PaymentMethod: t.PaymentMethod,
		PaymentAmount: t.PaymentAmount,
		ChangeAmount:  t.ChangeAmount,
		Status:        t.Status,
		Notes:         t.Notes,
		Items:         items,
		CreatedAt:     t.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     t.UpdatedAt.Format(time.RFC3339),
	}
}
