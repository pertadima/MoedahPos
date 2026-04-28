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

//nolint:funlen
func TestReportHandler(t *testing.T) {
	svc := new(mocks.ReportServiceInterface)
	log := zerolog.Nop()
	h := NewReportHandler(svc, log)

	t.Run("SalesSummary", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores/s1/reports/sales", nil)
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		svc.On("SalesSummary", mock.Anything, mock.Anything).Return(&dto.SalesSummaryResponse{TotalSales: 1000}, nil).Once()

		h.SalesSummary(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("StockValuation", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores/s1/reports/stock-valuation", nil)
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		svc.On("StockValuation", mock.Anything, "s1").Return(&dto.StockValuationResponse{GrandTotal: 2000}, nil).Once()

		h.StockValuation(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("CashFlow", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores/s1/reports/cash-flow", nil)
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		svc.On("CashFlow", mock.Anything, mock.Anything).Return(&dto.CashFlowResponse{TotalCashIn: 500}, nil).Once()

		h.CashFlow(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("SalesByProduct", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores/s1/reports/sales-by-product", nil)
		w := httptest.NewRecorder()
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		svc.On("SalesByProduct", mock.Anything, mock.Anything).Return([]dto.SalesByProductRow{}, nil).Once()
		h.SalesByProduct(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("SalesByCashier", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores/s1/reports/sales-by-cashier", nil)
		w := httptest.NewRecorder()
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		svc.On("SalesByCashier", mock.Anything, mock.Anything).Return([]dto.SalesByCashierRow{}, nil).Once()
		h.SalesByCashier(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("ProfitSummary", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores/s1/reports/profit?group_by=day", nil)
		w := httptest.NewRecorder()
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		svc.On("ProfitSummary", mock.Anything, mock.Anything, "day").Return(&dto.ProfitSummaryResponse{}, nil).Once()
		h.ProfitSummary(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("CashFlowDetail", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores/s1/reports/cash-flow/detail?date=2024-01-01", nil)
		w := httptest.NewRecorder()
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		svc.On("CashFlowDetail", mock.Anything, "s1", "2024-01-01").Return([]dto.CashFlowDetailEntry{}, nil).Once()
		h.CashFlowDetail(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
