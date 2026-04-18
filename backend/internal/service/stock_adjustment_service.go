package service

import (
	"context"
	"fmt"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/repository"
)

type StockAdjustmentService struct {
	repo        repository.StockAdjustmentRepository
	activitySvc ActivityLogServiceInterface
}

func NewStockAdjustmentService(repo repository.StockAdjustmentRepository, activitySvc ActivityLogServiceInterface) *StockAdjustmentService {
	return &StockAdjustmentService{repo: repo, activitySvc: activitySvc}
}

func (s *StockAdjustmentService) CreateAdjustment(ctx context.Context, storeID, userID string, input domain.CreateAdjustmentInput) error {
	const reasonManualCorrection = "MANUAL_CORRECTION"

	// Business validation
	if input.Type == "OUT" && (input.Reason != "DAMAGED" && input.Reason != "LOST" && input.Reason != reasonManualCorrection) {
		return fmt.Errorf("invalid mapping: reason %s is not permitted for type OUT", input.Reason)
	}

	if input.Type == "IN" && input.Reason != reasonManualCorrection {
		return fmt.Errorf("invalid mapping: IN type can only be used with MANUAL_CORRECTION reason")
	}

	if input.Reason == reasonManualCorrection && input.Notes == "" {
		return fmt.Errorf("notes are required for manual corrections")
	}

	if err := s.repo.CreateAdjustment(ctx, storeID, userID, input); err != nil {
		return err
	}

	s.activitySvc.LogActivity(ctx, userID, storeID, domain.ActionStockAdjustment, domain.ModuleInventory, input.ProductID, map[string]interface{}{
		"quantity": input.Quantity,
		"type":     input.Type, // IN / OUT
		"reason":   input.Reason,
		"notes":    input.Notes,
	})

	return nil
}

func (s *StockAdjustmentService) GetAdjustmentHistory(ctx context.Context, storeID string, productID *string) ([]*domain.StockAdjustment, error) {
	return s.repo.GetStockAdjustmentHistory(ctx, storeID, productID)
}
