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

// Stock-specific sentinel errors.
var (
	ErrProductNotInStore  = errors.New("product does not belong to this store")
	ErrStockLevelNotFound = errors.New("stock level not found for this product")
)

// StockService implements business logic for stock levels and movements.
type StockService struct {
	stockRepo   repository.StockRepository
	productRepo repository.ProductRepository
	log         zerolog.Logger
}

func NewStockService(stockRepo repository.StockRepository, productRepo repository.ProductRepository, log zerolog.Logger) *StockService {
	return &StockService{stockRepo: stockRepo, productRepo: productRepo, log: log}
}

// GetStockLevels returns all stock levels for a store, optionally filtered to low-stock items.
func (s *StockService) GetStockLevels(ctx context.Context, storeID string, lowStockOnly bool) ([]*dto.StockLevelResponse, error) {
	levels, err := s.stockRepo.FindLevelsByStore(ctx, storeID, lowStockOnly)
	if err != nil {
		return nil, fmt.Errorf("finding stock levels: %w", err)
	}
	resp := make([]*dto.StockLevelResponse, 0, len(levels))
	for _, l := range levels {
		resp = append(resp, toStockLevelResponse(l))
	}
	return resp, nil
}

// GetProductStock returns the stock level for a specific product in a store.
func (s *StockService) GetProductStock(ctx context.Context, productID, storeID string) (*dto.StockLevelResponse, error) {
	level, err := s.stockRepo.FindLevelByProduct(ctx, productID, storeID)
	if err != nil {
		return nil, fmt.Errorf("finding stock level: %w", err)
	}
	if level == nil {
		return nil, ErrStockLevelNotFound
	}
	return toStockLevelResponse(level), nil
}

// AdjustStock creates a stock movement and atomically updates the stock level.
func (s *StockService) AdjustStock(ctx context.Context, storeID string, req *dto.AdjustStockRequest, userID string) (*dto.StockLevelResponse, error) {
	// Verify product belongs to store.
	product, err := s.productRepo.FindByID(ctx, req.ProductID)
	if err != nil {
		return nil, fmt.Errorf("finding product: %w", err)
	}
	if product == nil || product.StoreID != storeID {
		return nil, ErrProductNotInStore
	}

	level, err := s.stockRepo.Adjust(ctx, domain.AdjustInput{
		ProductID: req.ProductID,
		StoreID:   storeID,
		Delta:     req.Delta,
		RefType:   "adjustment",
		Notes:     req.Notes,
		CreatedBy: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("adjusting stock: %w", err)
	}

	s.log.Info().
		Str("product_id", req.ProductID).
		Str("store_id", storeID).
		Float64("delta", req.Delta).
		Msg("stock adjusted")

	return toStockLevelResponse(level), nil
}

// SetMinStock sets the minimum stock threshold for a product (for low-stock alerts).
func (s *StockService) SetMinStock(ctx context.Context, storeID string, req *dto.SetMinStockRequest) error {
	product, err := s.productRepo.FindByID(ctx, req.ProductID)
	if err != nil {
		return fmt.Errorf("finding product: %w", err)
	}
	if product == nil || product.StoreID != storeID {
		return ErrProductNotInStore
	}
	return s.stockRepo.SetMinQuantity(ctx, req.ProductID, storeID, req.MinQuantity)
}

// GetMovements returns paginated stock movement history.
func (s *StockService) GetMovements(ctx context.Context, filter dto.StockMovementFilter) ([]*dto.StockMovementResponse, dto.PaginationMeta, error) {
	filter.Defaults()
	movements, total, err := s.stockRepo.FindMovements(ctx, filter)
	if err != nil {
		return nil, dto.PaginationMeta{}, fmt.Errorf("finding movements: %w", err)
	}
	resp := make([]*dto.StockMovementResponse, 0, len(movements))
	for _, m := range movements {
		resp = append(resp, toMovementResponse(m))
	}
	return resp, dto.NewMeta(filter.PaginationQuery, total), nil
}

// ─── Mappers ──────────────────────────────────────────────────────────────────

func toStockLevelResponse(l *domain.StockLevel) *dto.StockLevelResponse {
	return &dto.StockLevelResponse{
		ProductID:   l.ProductID,
		ProductName: l.ProductName,
		ProductSKU:  l.ProductSKU,
		Unit:        l.Unit,
		StoreID:     l.StoreID,
		Quantity:    l.Quantity,
		MinQuantity: l.MinQuantity,
		IsLowStock:  l.Quantity <= l.MinQuantity,
		UpdatedAt:   l.UpdatedAt.Format(time.RFC3339),
	}
}

func toMovementResponse(m *domain.StockMovement) *dto.StockMovementResponse {
	return &dto.StockMovementResponse{
		ID:            m.ID,
		ProductID:     m.ProductID,
		ProductName:   m.ProductName,
		StoreID:       m.StoreID,
		RefType:       m.RefType,
		RefID:         m.RefID,
		QuantityDelta: m.QuantityDelta,
		Notes:         m.Notes,
		CreatedBy:     m.CreatedBy,
		CreatedByName: m.CreatedByName,
		CreatedAt:     m.CreatedAt.Format(time.RFC3339),
	}
}
