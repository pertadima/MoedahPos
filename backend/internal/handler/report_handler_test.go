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

func TestReportHandler_SalesSummary(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		rSvc := new(mocks.ReportServiceInterface)
		h := NewReportHandler(rSvc, zerolog.Nop())

		r := chi.NewRouter()
		r.Get("/stores/{storeId}/reports/sales", h.SalesSummary)

		rSvc.On("SalesSummary", mock.Anything, mock.Anything).Return(&dto.SalesSummaryResponse{TotalSales: 1000}, nil)

		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores/s1/reports/sales", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		rSvc.AssertExpectations(t)
	})
}

func TestReportHandler_StockValuation(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		rSvc := new(mocks.ReportServiceInterface)
		h := NewReportHandler(rSvc, zerolog.Nop())

		r := chi.NewRouter()
		r.Get("/stores/{storeId}/reports/stock-valuation", h.StockValuation)

		rSvc.On("StockValuation", mock.Anything, "s1").Return(&dto.StockValuationResponse{GrandTotal: 500}, nil)

		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores/s1/reports/stock-valuation", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
