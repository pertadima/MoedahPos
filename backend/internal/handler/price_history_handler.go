package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/moedahpos/backend/internal/dto"
	"github.com/moedahpos/backend/internal/service"
	"github.com/moedahpos/backend/pkg/response"
	"github.com/rs/zerolog"
)

// PriceHistoryHandler handles HTTP requests for price history.
type PriceHistoryHandler struct {
	svc *service.PriceHistoryService
	log zerolog.Logger
}

func NewPriceHistoryHandler(svc *service.PriceHistoryService, log zerolog.Logger) *PriceHistoryHandler {
	return &PriceHistoryHandler{svc: svc, log: log}
}

// GET /stores/:storeId/price-history?product_id=&source=&page=&per_page=
func (h *PriceHistoryHandler) ListByStore(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	f := dto.PriceHistoryFilter{
		StoreID:   storeID,
		ProductID: r.URL.Query().Get("product_id"),
		Source:    r.URL.Query().Get("source"),
	}
	f.Page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	f.PerPage, _ = strconv.Atoi(r.URL.Query().Get("per_page"))

	rows, meta, err := h.svc.ListByStore(r.Context(), storeID, f)
	if err != nil {
		h.log.Error().Err(err).Msg("price history list failed")
		response.InternalError(w)
		return
	}
	response.Success(w, dto.ListResponse{Data: rows, Meta: meta})
}

// GET /stores/:storeId/products/:productId/price-history
func (h *PriceHistoryHandler) ListByProduct(w http.ResponseWriter, r *http.Request) {
	productID := chi.URLParam(r, "productId")
	storeID := chi.URLParam(r, "storeId")
	f := dto.PriceHistoryFilter{StoreID: storeID}
	f.Page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	f.PerPage, _ = strconv.Atoi(r.URL.Query().Get("per_page"))

	rows, meta, err := h.svc.ListByProduct(r.Context(), productID, f)
	if err != nil {
		h.log.Error().Err(err).Msg("price history by product failed")
		response.InternalError(w)
		return
	}
	response.Success(w, dto.ListResponse{Data: rows, Meta: meta})
}
