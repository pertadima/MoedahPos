package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
	"github.com/moedahpos/backend/internal/repository"
)

type ActivityLogService struct {
	repo repository.ActivityLogRepository
	log  zerolog.Logger
}

func NewActivityLogService(repo repository.ActivityLogRepository, log zerolog.Logger) *ActivityLogService {
	return &ActivityLogService{repo: repo, log: log}
}

func (s *ActivityLogService) LogActivity(ctx context.Context, userID, storeID string, action domain.ActionType, module domain.ActivityModule, refID string, metadata interface{}) {
	metaJSON, err := json.Marshal(metadata)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to marshal activity log metadata")
		metaJSON = []byte("{}")
	}

	var sidPtr *string
	if storeID != "" {
		sidPtr = &storeID
	}

	var ridPtr *string
	if refID != "" {
		ridPtr = &refID
	}

	activity := &domain.ActivityLog{
		UserID:      userID,
		StoreID:     sidPtr,
		ActionType:  action,
		Module:      module,
		ReferenceID: ridPtr,
		Metadata:    metaJSON,
	}

	if err := s.repo.Create(ctx, activity); err != nil {
		s.log.Error().Err(err).Msg("failed to save activity log")
	}
}

func (s *ActivityLogService) GetActivityLogs(ctx context.Context, storeID string, filter dto.ActivityLogFilter) ([]dto.ActivityLogResponse, dto.PaginationMeta, error) {
	filter.Defaults()
	logs, total, err := s.repo.FindAll(ctx, storeID, filter)
	if err != nil {
		return nil, dto.PaginationMeta{}, fmt.Errorf("listing activity logs: %w", err)
	}

	resp := make([]dto.ActivityLogResponse, 0, len(logs))
	for _, l := range logs {
		var meta interface{}
		_ = json.Unmarshal(l.Metadata, &meta)

		resp = append(resp, dto.ActivityLogResponse{
			ID:          l.ID,
			UserID:      l.UserID,
			UserName:    l.UserName,
			StoreID:     l.StoreID,
			ActionType:  l.ActionType,
			Module:      l.Module,
			ReferenceID: l.ReferenceID,
			Metadata:    meta,
			CreatedAt:   l.CreatedAt.Format(time.RFC3339),
		})
	}

	return resp, dto.NewMeta(filter.PaginationQuery, total), nil
}
