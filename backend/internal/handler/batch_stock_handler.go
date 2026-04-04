package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
	"github.com/moedahpos/backend/internal/service"
	"github.com/moedahpos/backend/pkg/response"
)

// BatchStockHandler exposes HTTP endpoints for FIFO batch inventory data.
type BatchStockHandler struct {
	batchSvc *service.BatchStockService
	log      zerolog.Logger
}

// NewBatchStockHandler creates a BatchStockHandler with the injected service.
func NewBatchStockHandler(batchSvc *service.BatchStockService, log zerolog.Logger) *BatchStockHandler {
	return &BatchStockHandler{batchSvc: batchSvc, log: log}
}

// ListBatches handles GET /stores/:storeId/stock/batches
// Optional query param: product_id (UUID) — filters results to a single product.
// Response: array of StockBatchResponse ordered by product name then received_at ASC.
func (h *BatchStockHandler) ListBatches(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	f := dto.BatchListFilter{
		StoreID:   storeID,
		ProductID: r.URL.Query().Get("product_id"),
	}
	batches, err := h.batchSvc.GetBatchesByStore(r.Context(), f)
	if err != nil {
		h.log.Error().Err(err).Str("store_id", storeID).Msg("list batches failed")
		response.InternalError(w)
		return
	}
	out := make([]dto.StockBatchResponse, len(batches))
	for i, b := range batches {
		out[i] = batchToResponse(b)
	}
	response.Success(w, out)
}

// GetSummary handles GET /stores/:storeId/stock/batch-summary
// Returns total stock per product aggregated across all active batches,
// with average cost price for valuation purposes.
func (h *BatchStockHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	summary, err := h.batchSvc.GetStockSummary(r.Context(), storeID)
	if err != nil {
		h.log.Error().Err(err).Str("store_id", storeID).Msg("batch summary failed")
		response.InternalError(w)
		return
	}
	out := make([]dto.BatchStockSummaryResponse, len(summary))
	for i, s := range summary {
		out[i] = dto.BatchStockSummaryResponse{
			ProductID:    s.ProductID,
			ProductName:  s.ProductName,
			ProductSKU:   s.ProductSKU,
			Unit:         s.Unit,
			TotalQty:     s.TotalQty,
			BatchCount:   s.BatchCount,
			AvgCostPrice: s.AvgCostPrice,
		}
	}
	response.Success(w, out)
}

// batchToResponse converts a domain.StockBatch to its API representation.
func batchToResponse(b *domain.StockBatch) dto.StockBatchResponse {
	return dto.StockBatchResponse{
		ID:                b.ID,
		ProductID:         b.ProductID,
		ProductName:       b.ProductName,
		ProductSKU:        b.ProductSKU,
		Unit:              b.Unit,
		StoreID:           b.StoreID,
		POID:              b.POID,
		QuantityRemaining: b.QuantityRemaining,
		PurchasePrice:     b.PurchasePrice,
		ReceivedAt:        b.ReceivedAt.Format("2006-01-02T15:04:05Z07:00"),
		CreatedAt:         b.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
