package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/service/mocks"
)

func TestBatchStockHandler(t *testing.T) {
	svc := new(mocks.BatchStockServiceInterface)
	log := zerolog.Nop()
	h := NewBatchStockHandler(svc, log)

	t.Run("ListBatches", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/stores/s1/stock/batches", nil)
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		svc.On("GetBatchesByStore", mock.Anything, mock.Anything).Return([]*domain.StockBatch{
			{ID: "b1", ReceivedAt: time.Now(), CreatedAt: time.Now()},
		}, nil).Once()

		h.ListBatches(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("GetSummary", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/stores/s1/stock/batch-summary", nil)
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		svc.On("GetStockSummary", mock.Anything, "s1").Return([]*domain.BatchStockSummary{
			{ProductID: "p1", ProductName: "P1"},
		}, nil).Once()

		h.GetSummary(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
