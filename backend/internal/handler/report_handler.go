package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/moedahpos/backend/internal/dto"
	"github.com/moedahpos/backend/internal/service"
	"github.com/moedahpos/backend/pkg/response"
	"github.com/rs/zerolog"
)

// ReportHandler handles analytics and reporting endpoints.
type ReportHandler struct {
	reportSvc *service.ReportService
	log       zerolog.Logger
}

func NewReportHandler(reportSvc *service.ReportService, log zerolog.Logger) *ReportHandler {
	return &ReportHandler{reportSvc: reportSvc, log: log}
}

func filterFromQuery(r *http.Request, storeID string) dto.ReportFilter {
	return dto.ReportFilter{
		StoreID:  storeID,
		DateFrom: r.URL.Query().Get("date_from"),
		DateTo:   r.URL.Query().Get("date_to"),
	}
}

// GET /stores/:storeId/reports/sales
func (h *ReportHandler) SalesSummary(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	result, err := h.reportSvc.SalesSummary(r.Context(), filterFromQuery(r, storeID))
	if err != nil {
		h.log.Error().Err(err).Msg("sales summary failed")
		response.InternalError(w)
		return
	}
	response.Success(w, result)
}

// GET /stores/:storeId/reports/sales/by-product
func (h *ReportHandler) SalesByProduct(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	rows, err := h.reportSvc.SalesByProduct(r.Context(), filterFromQuery(r, storeID))
	if err != nil {
		h.log.Error().Err(err).Msg("sales by product failed")
		response.InternalError(w)
		return
	}
	response.Success(w, rows)
}

// GET /stores/:storeId/reports/sales/by-cashier
func (h *ReportHandler) SalesByCashier(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	rows, err := h.reportSvc.SalesByCashier(r.Context(), filterFromQuery(r, storeID))
	if err != nil {
		h.log.Error().Err(err).Msg("sales by cashier failed")
		response.InternalError(w)
		return
	}
	response.Success(w, rows)
}

// GET /stores/:storeId/reports/stock-valuation
func (h *ReportHandler) StockValuation(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	result, err := h.reportSvc.StockValuation(r.Context(), storeID)
	if err != nil {
		h.log.Error().Err(err).Msg("stock valuation failed")
		response.InternalError(w)
		return
	}
	response.Success(w, result)
}

// unused errors suppressor
var _ = errors.New
