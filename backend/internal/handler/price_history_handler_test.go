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

func TestPriceHistoryHandler(t *testing.T) {
	svc := new(mocks.PriceHistoryServiceInterface)
	log := zerolog.Nop()
	h := NewPriceHistoryHandler(svc, log)

	t.Run("ListByStore", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores/s1/price-history", nil)
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		svc.On("ListByStore", mock.Anything, "s1", mock.Anything).Return([]*dto.PriceHistoryRow{{ID: "h1"}}, dto.PaginationMeta{Total: 1}, nil).Once()

		h.ListByStore(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("ListByProduct", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores/s1/products/p1/price-history", nil)
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		rctx.URLParams.Add("productId", "p1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		svc.On("ListByProduct", mock.Anything, "p1", mock.Anything).Return([]*dto.PriceHistoryRow{{ID: "h1"}}, dto.PaginationMeta{Total: 1}, nil).Once()

		h.ListByProduct(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})
}
