package handler

import (
	"context"
	"fmt"
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

// ── ExportReport handler tests ────────────────────────────────────────────────

//nolint:funlen
func TestExportReportHandler(t *testing.T) {
	makeReq := func(t *testing.T, query string) (*http.Request, *httptest.ResponseRecorder) {
		t.Helper()
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet,
			"/stores/s1/reports/export"+query, nil)
		w := httptest.NewRecorder()
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		return req, w
	}

	t.Run("CSV sales success", func(t *testing.T) {
		svc := new(mocks.ReportServiceInterface)
		h := NewReportHandler(svc, zerolog.Nop())

		csvBytes := []byte("Tanggal,Jml Transaksi\n2024-01-01,5\n")
		svc.On("ExportCSV", mock.Anything, "sales", mock.Anything).Return(csvBytes, nil).Once()

		req, w := makeReq(t, "?type=csv&report=sales&date_from=2024-01-01")
		h.ExportReport(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), "text/csv")
		assert.Contains(t, w.Header().Get("Content-Disposition"), "laporan-sales")
		svc.AssertExpectations(t)
	})

	t.Run("PDF inventory success", func(t *testing.T) {
		svc := new(mocks.ReportServiceInterface)
		h := NewReportHandler(svc, zerolog.Nop())

		htmlBytes := []byte("<!DOCTYPE html><html><body>Inventory</body></html>")
		svc.On("ExportPDF", mock.Anything, "inventory", mock.Anything).Return(htmlBytes, nil).Once()

		req, w := makeReq(t, "?type=pdf&report=inventory")
		h.ExportReport(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), "text/html")
		svc.AssertExpectations(t)
	})

	t.Run("invalid type returns 400", func(t *testing.T) {
		svc := new(mocks.ReportServiceInterface)
		h := NewReportHandler(svc, zerolog.Nop())

		req, w := makeReq(t, "?type=xlsx&report=sales")
		h.ExportReport(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid report returns 400", func(t *testing.T) {
		svc := new(mocks.ReportServiceInterface)
		h := NewReportHandler(svc, zerolog.Nop())

		req, w := makeReq(t, "?type=csv&report=unknown")
		h.ExportReport(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error returns 500", func(t *testing.T) {
		svc := new(mocks.ReportServiceInterface)
		h := NewReportHandler(svc, zerolog.Nop())

		svc.On("ExportCSV", mock.Anything, "profit", mock.Anything).
			Return(nil, fmt.Errorf("db error")).Once()

		req, w := makeReq(t, "?type=csv&report=profit")
		h.ExportReport(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("PDF profit with date_from in filename", func(t *testing.T) {
		svc := new(mocks.ReportServiceInterface)
		h := NewReportHandler(svc, zerolog.Nop())

		htmlBytes := []byte("<!DOCTYPE html><html></html>")
		svc.On("ExportPDF", mock.Anything, "profit", mock.Anything).Return(htmlBytes, nil).Once()

		req, w := makeReq(t, "?type=pdf&report=profit&date_from=2024-06-01")
		h.ExportReport(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Disposition"), "2024-06-01")
		svc.AssertExpectations(t)
	})
}
