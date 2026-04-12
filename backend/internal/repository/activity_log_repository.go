package repository

import (
	"context"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
)

type ActivityLogRepository interface {
	Create(ctx context.Context, log *domain.ActivityLog) error
	FindAll(ctx context.Context, storeID string, filter dto.ActivityLogFilter) ([]*domain.ActivityLog, int, error)
}
