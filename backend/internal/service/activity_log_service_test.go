package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
	"github.com/moedahpos/backend/mocks"
)

func TestActivityLogService(t *testing.T) {
	repo := new(mocks.ActivityLogRepository)
	log := zerolog.Nop()
	svc := NewActivityLogService(repo, log)

	ctx := context.Background()

	t.Run("LogActivity", func(t *testing.T) {
		metadata := map[string]string{"foo": "bar"}
		repo.On("Create", ctx, mock.MatchedBy(func(l *domain.ActivityLog) bool {
			return l.UserID == "u1" && l.ActionType == domain.ActionAuthLogin
		})).Return(nil).Once()

		svc.LogActivity(ctx, "u1", "s1", domain.ActionAuthLogin, "PRODUCT", "ref1", metadata)
		repo.AssertExpectations(t)
	})

	t.Run("GetActivityLogs", func(t *testing.T) {
		filter := dto.ActivityLogFilter{}
		logs := []*domain.ActivityLog{
			{
				ID: "l1", UserID: "u1", UserName: "User 1",
				ActionType: domain.ActionAuthLogin, Module: "PRODUCT",
				Metadata: json.RawMessage(`{"foo":"bar"}`),
				CreatedAt: time.Now(),
			},
		}
		repo.On("FindAll", ctx, "s1", mock.Anything).Return(logs, 1, nil).Once()

		resp, meta, err := svc.GetActivityLogs(ctx, "s1", filter)
		assert.NoError(t, err)
		assert.Len(t, resp, 1)
		assert.Equal(t, 1, meta.Total)
		assert.Equal(t, "User 1", resp[0].UserName)
		repo.AssertExpectations(t)
	})
}
