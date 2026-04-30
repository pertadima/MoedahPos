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

func TestActivityLogHandler(t *testing.T) {
	svc := new(mocks.ActivityLogServiceInterface)
	log := zerolog.Nop()
	h := NewActivityLogHandler(svc, log)

	t.Run("List", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores/s1/activity-logs", nil)
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		svc.On("GetActivityLogs", mock.Anything, "s1", mock.Anything).Return([]dto.ActivityLogResponse{{ID: "l1"}}, dto.PaginationMeta{Total: 1}, nil).Once()

		h.List(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})
}
