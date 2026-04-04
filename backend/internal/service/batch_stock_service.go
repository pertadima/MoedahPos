package service

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
	"github.com/moedahpos/backend/internal/repository"
)

// BatchStockService orchestrates FIFO batch-level inventory tracking.
//
// It sits on top of BatchRepository and provides three categories of operations:
//  1. Inbound  — ReceivePurchaseOrder: create batches when a PO is received.
//  2. Outbound — DeductStockFIFO: consume oldest batches when a sale is made.
//  3. Query    — GetStockSummary / GetBatchesByProduct / GetBatchesByStore.
type BatchStockService struct {
	batchRepo repository.BatchRepository
	log       zerolog.Logger
}

// NewBatchStockService creates a BatchStockService with the given repository and logger.
func NewBatchStockService(batchRepo repository.BatchRepository, log zerolog.Logger) *BatchStockService {
	return &BatchStockService{batchRepo: batchRepo, log: log}
}

// ─── Inbound ──────────────────────────────────────────────────────────────────

// ReceivePurchaseOrder creates one stock batch per item in a received PO.
// The batch's received_at is set to the current time, which determines
// its position in the FIFO queue relative to other batches of the same product.
func (s *BatchStockService) ReceivePurchaseOrder(ctx context.Context, poID, storeID string, items []dto.POBatchItem) error {
	for _, item := range items {
		batch := &domain.StockBatch{
			ProductID:         item.ProductID,
			StoreID:           storeID,
			POID:              &poID,
			QuantityRemaining: item.Quantity,
			PurchasePrice:     item.UnitCost,
			ReceivedAt:        time.Now(), // determines FIFO order
		}
		if err := s.batchRepo.CreateBatch(ctx, batch); err != nil {
			return fmt.Errorf("BatchStockService.ReceivePurchaseOrder product=%s: %w", item.ProductID, err)
		}
	}
	s.log.Info().
		Str("po_id", poID).
		Str("store_id", storeID).
		Int("items", len(items)).
		Msg("stock batches created for received PO")
	return nil
}

// ─── Outbound (FIFO) ──────────────────────────────────────────────────────────

// DeductStockFIFO removes qty from the oldest available batch(es) for a product.
// It delegates to BatchRepository.DeductFIFO which uses SELECT FOR UPDATE to
// guarantee no two concurrent transactions deduct the same stock.
//
// If the total batch stock is insufficient, an error is returned and no
// batches are modified (the underlying transaction is rolled back).
func (s *BatchStockService) DeductStockFIFO(ctx context.Context, productID, storeID string, qty float64) error {
	if err := s.batchRepo.DeductFIFO(ctx, productID, storeID, qty); err != nil {
		return fmt.Errorf("BatchStockService.DeductStockFIFO: %w", err)
	}
	s.log.Debug().
		Str("product_id", productID).
		Str("store_id", storeID).
		Float64("qty", qty).
		Msg("FIFO batch stock deducted")
	return nil
}

// ─── Query ────────────────────────────────────────────────────────────────────

// GetStockSummary returns aggregate batch inventory (total qty, batch count, avg cost)
// per product for the given store.
func (s *BatchStockService) GetStockSummary(ctx context.Context, storeID string) ([]*domain.BatchStockSummary, error) {
	return s.batchRepo.GetStockSummary(ctx, storeID)
}

// GetBatchesByProduct returns all active batches for one product at one store,
// ordered oldest-first (FIFO order), for audit and debugging purposes.
func (s *BatchStockService) GetBatchesByProduct(ctx context.Context, productID, storeID string) ([]*domain.StockBatch, error) {
	return s.batchRepo.GetBatchesByProduct(ctx, productID, storeID)
}

// GetBatchesByStore returns active batches for a store.
// Providing a non-empty productID in the filter restricts results to that product.
func (s *BatchStockService) GetBatchesByStore(ctx context.Context, f dto.BatchListFilter) ([]*domain.StockBatch, error) {
	return s.batchRepo.GetBatchesByStore(ctx, f)
}
