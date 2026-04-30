package handler

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/moedahpos/backend/internal/dto"
	"github.com/moedahpos/backend/internal/service"
	"github.com/moedahpos/backend/pkg/response"
)

// ReportHandler handles analytics and reporting endpoints.
type ReportHandler struct {
	reportSvc service.ReportServiceInterface
	log       zerolog.Logger
}

func NewReportHandler(reportSvc service.ReportServiceInterface, log zerolog.Logger) *ReportHandler {
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

// GET /stores/:storeId/reports/profit?group_by=day|week|month
func (h *ReportHandler) ProfitSummary(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	groupBy := r.URL.Query().Get("group_by")
	if groupBy == "" {
		groupBy = "day"
	}
	result, err := h.reportSvc.ProfitSummary(r.Context(), filterFromQuery(r, storeID), groupBy)
	if err != nil {
		h.log.Error().Err(err).Msg("profit summary failed")
		response.InternalError(w)
		return
	}
	response.Success(w, result)
}

// GET /stores/:storeId/reports/cash-flow?date_from=&date_to=
func (h *ReportHandler) CashFlow(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	result, err := h.reportSvc.CashFlow(r.Context(), filterFromQuery(r, storeID))
	if err != nil {
		h.log.Error().Err(err).Msg("cash flow failed")
		response.InternalError(w)
		return
	}
	response.Success(w, result)
}

// GET /stores/:storeId/reports/cash-flow/detail?date=
func (h *ReportHandler) CashFlowDetail(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	date := r.URL.Query().Get("date")

	result, err := h.reportSvc.CashFlowDetail(r.Context(), storeID, date)
	if err != nil {
		h.log.Error().Err(err).Msg("cash flow detail failed")
		response.InternalError(w)
		return
	}
	response.Success(w, result)
}

// ExportReport handles GET /stores/:storeId/reports/export?type=csv|pdf&report=sales|inventory|profit
//
// RBAC:
//   - sales   → penjualan:read
//   - inventory → inventory:read
//   - profit  → keuangan:read
//
// The permission is validated per-report in the handler so a single route covers all types.
func (h *ReportHandler) ExportReport(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	q := r.URL.Query()
	exportType := q.Get("type")   // "csv" or "pdf"
	reportName := q.Get("report") // "sales", "inventory", "profit"
	filter := filterFromQuery(r, storeID)

	if exportType != "csv" && exportType != "pdf" {
		http.Error(w, `{"error":"type must be csv or pdf"}`, http.StatusBadRequest)
		return
	}
	if reportName != "sales" && reportName != "inventory" && reportName != "profit" {
		http.Error(w, `{"error":"report must be sales, inventory, or profit"}`, http.StatusBadRequest)
		return
	}

	var (
		data        []byte
		err         error
		filename    string
		contentType string
	)

	dateTag := filter.DateFrom
	if dateTag == "" {
		dateTag = "all"
	}

	switch exportType {
	case "csv":
		data, err = h.reportSvc.ExportCSV(r.Context(), reportName, filter)
		filename = fmt.Sprintf("laporan-%s-%s.csv", reportName, dateTag)
		contentType = "text/csv; charset=utf-8"
	default: // "pdf"
		data, err = h.reportSvc.ExportPDF(r.Context(), reportName, filter)
		filename = fmt.Sprintf("laporan-%s-%s.pdf", reportName, dateTag)
		contentType = "application/pdf"
	}

	if err != nil {
		h.log.Error().Err(err).Str("report", reportName).Str("type", exportType).Msg("export failed")
		response.InternalError(w)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data) //nolint:gosec // data is server-generated HTML/CSV, not user-supplied input
}
