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

// PriceHistoryService manages price-change audit logging.
type PriceHistoryService struct {
	priceHistoryRepo repository.PriceHistoryRepository
	log              zerolog.Logger
}

func NewPriceHistoryService(repo repository.PriceHistoryRepository, log zerolog.Logger) *PriceHistoryService {
	return &PriceHistoryService{priceHistoryRepo: repo, log: log}
}

// RecordChange records a price change for a product.
// It only logs when at least one price field actually changed.
func (s *PriceHistoryService) RecordChange(
	ctx context.Context,
	productID, storeID, changedBy string,
	oldCost, newCost, oldSell, newSell float64,
	source string,
	refID *string,
	notes *string,
) error {
	// Skip if nothing changed
	if oldCost == newCost && oldSell == newSell {
		return nil
	}
	return s.priceHistoryRepo.Record(ctx, domain.PriceHistory{
		ProductID: productID,
		StoreID:   storeID,
		ChangedBy: changedBy,
		OldCost:   oldCost,
		NewCost:   newCost,
		OldSell:   oldSell,
		NewSell:   newSell,
		Source:    source,
		RefID:     refID,
		Notes:     notes,
		ChangedAt: time.Now(),
	})
}

// ListByProduct returns paginated price history for one product.
func (s *PriceHistoryService) ListByProduct(ctx context.Context, productID string, f dto.PriceHistoryFilter) ([]*dto.PriceHistoryRow, dto.PaginationMeta, error) {
	f.Defaults()
	rows, total, err := s.priceHistoryRepo.FindByProduct(ctx, productID, f)
	if err != nil {
		return nil, dto.PaginationMeta{}, fmt.Errorf("listing price history: %w", err)
	}
	return toPriceHistoryRows(rows), dto.NewMeta(f.PaginationQuery, total), nil
}

// ListByStore returns paginated price history across all products in a store.
func (s *PriceHistoryService) ListByStore(ctx context.Context, storeID string, f dto.PriceHistoryFilter) ([]*dto.PriceHistoryRow, dto.PaginationMeta, error) {
	f.Defaults()
	rows, total, err := s.priceHistoryRepo.FindByStore(ctx, storeID, f)
	if err != nil {
		return nil, dto.PaginationMeta{}, fmt.Errorf("listing store price history: %w", err)
	}
	return toPriceHistoryRows(rows), dto.NewMeta(f.PaginationQuery, total), nil
}

func toPriceHistoryRows(rows []*domain.PriceHistory) []*dto.PriceHistoryRow {
	out := make([]*dto.PriceHistoryRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, &dto.PriceHistoryRow{
			ID:            r.ID,
			ProductID:     r.ProductID,
			ProductName:   r.ProductName,
			StoreID:       r.StoreID,
			ChangedBy:     r.ChangedBy,
			ChangedByName: r.ChangedByName,
			OldCost:       r.OldCost,
			NewCost:       r.NewCost,
			OldSell:       r.OldSell,
			NewSell:       r.NewSell,
			Source:        r.Source,
			RefID:         r.RefID,
			Notes:         r.Notes,
			ChangedAt:     r.ChangedAt.Format(time.RFC3339),
		})
	}
	return out
}
