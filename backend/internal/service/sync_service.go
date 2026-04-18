package service

import (
	"context"
	"fmt"
	"time"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
	"github.com/moedahpos/backend/internal/repository"
	"github.com/rs/zerolog"
)

type SyncService struct {
	categoryRepo    repository.CategoryRepository
	productRepo     repository.ProductRepository
	stockRepo       repository.StockRepository
	transactionRepo repository.TransactionRepository
	log             zerolog.Logger
}

func NewSyncService(
	catRepo repository.CategoryRepository,
	prodRepo repository.ProductRepository,
	stockRepo repository.StockRepository,
	txnRepo repository.TransactionRepository,
	log zerolog.Logger,
) *SyncService {
	return &SyncService{
		categoryRepo:    catRepo,
		productRepo:     prodRepo,
		stockRepo:       stockRepo,
		transactionRepo: txnRepo,
		log:             log,
	}
}

func (s *SyncService) Pull(ctx context.Context, storeID string, since time.Time) (*dto.SyncPullOutput, error) {
	out := &dto.SyncPullOutput{
		Categories:   []*domain.Category{},
		Products:     []*domain.Product{},
		StockLevels:  []*domain.StockLevel{},
		Transactions: []*domain.Transaction{},
	}

	cats, err := s.categoryRepo.GetModifiedSince(ctx, storeID, since)
	if err != nil {
		s.log.Error().Err(err).Str("store_id", storeID).Msg("SyncService.Pull failed for categories")
		return nil, fmt.Errorf("failed to sync categories: %w", err)
	}
	if cats != nil {
		out.Categories = cats
	}

	products, err := s.productRepo.GetModifiedSince(ctx, storeID, since)
	if err != nil {
		s.log.Error().Err(err).Str("store_id", storeID).Msg("SyncService.Pull failed for products")
		return nil, fmt.Errorf("failed to sync products: %w", err)
	}
	if products != nil {
		out.Products = products
	}

	stocks, err := s.stockRepo.GetModifiedSince(ctx, storeID, since)
	if err != nil {
		s.log.Error().Err(err).Str("store_id", storeID).Msg("SyncService.Pull failed for stock levels")
		return nil, fmt.Errorf("failed to sync stock levels: %w", err)
	}
	if stocks != nil {
		out.StockLevels = stocks
	}

	txns, err := s.transactionRepo.GetModifiedSince(ctx, storeID, since)
	if err != nil {
		s.log.Error().Err(err).Str("store_id", storeID).Msg("SyncService.Pull failed for transactions")
		return nil, fmt.Errorf("failed to sync transactions: %w", err)
	}
	if txns != nil {
		out.Transactions = txns
	}

	return out, nil
}
