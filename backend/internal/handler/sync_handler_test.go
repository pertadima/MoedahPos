package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/moedahpos/backend/internal/dto"
	"github.com/moedahpos/backend/internal/service/mocks"
)

func TestSyncHandler(t *testing.T) {
	svc := new(mocks.SyncServiceInterface)
	log := zerolog.Nop()
	h := NewSyncHandler(svc, log)

	t.Run("Pull Success", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores/s1/sync/pull?last_synced_at=2024-01-01T00:00:00Z", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		svc.On("Pull", mock.Anything, "s1", mock.Anything).Return(&dto.SyncPullOutput{}, nil).Once()

		w := httptest.NewRecorder()
		h.Pull(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("Pull Invalid Date", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores/s1/sync/pull?last_synced_at=invalid", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		h.Pull(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
