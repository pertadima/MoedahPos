package mocks

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/moedahpos/backend/internal/dto"
)

type SyncServiceInterface struct {
	mock.Mock
}

func (m *SyncServiceInterface) Pull(ctx context.Context, storeID string, since time.Time) (*dto.SyncPullOutput, error) {
	args := m.Called(ctx, storeID, since)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.SyncPullOutput), args.Error(1)
}
