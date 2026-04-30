package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/moedahpos/backend/internal/dto"
	"github.com/moedahpos/backend/internal/service"
	"github.com/moedahpos/backend/internal/validator"
	"github.com/moedahpos/backend/pkg/response"
)

// StockHandler handles HTTP requests for stock management endpoints.
type StockHandler struct {
	stockSvc  service.StockServiceInterface
	validator *validator.Validator
	log       zerolog.Logger
}

func NewStockHandler(stockSvc service.StockServiceInterface, v *validator.Validator, log zerolog.Logger) *StockHandler {
	return &StockHandler{stockSvc: stockSvc, validator: v, log: log}
}

// GET /stores/:storeId/stock
func (h *StockHandler) GetLevels(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	levels, err := h.stockSvc.GetStockLevels(r.Context(), storeID, false)
	if err != nil {
		h.log.Error().Err(err).Msg("get stock levels failed")
		response.InternalError(w)
		return
	}
	response.Success(w, levels)
}

// GET /stores/:storeId/stock/low
func (h *StockHandler) GetLowStock(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	levels, err := h.stockSvc.GetStockLevels(r.Context(), storeID, true)
	if err != nil {
		h.log.Error().Err(err).Msg("get low stock failed")
		response.InternalError(w)
		return
	}
	response.Success(w, levels)
}

// GET /stores/:storeId/stock/:productId
func (h *StockHandler) GetProductStock(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	productID := chi.URLParam(r, "productId")
	level, err := h.stockSvc.GetProductStock(r.Context(), productID, storeID)
	if err != nil {
		if errors.Is(err, service.ErrStockLevelNotFound) {
			response.NotFound(w, "Stock level")
			return
		}
		response.InternalError(w)
		return
	}
	response.Success(w, level)
}

// POST /stores/:storeId/stock/adjust
func (h *StockHandler) Adjust(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	var req dto.AdjustStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	if errs := h.validator.ValidateStruct(req); errs != nil {
		response.ValidationError(w, errs)
		return
	}
	result, err := h.stockSvc.AdjustStock(r.Context(), storeID, &req, userIDFromCtx(r))
	if err != nil {
		if errors.Is(err, service.ErrProductNotInStore) {
			response.NotFound(w, "Product")
			return
		}
		h.log.Error().Err(err).Msg("adjust stock failed")
		response.InternalError(w)
		return
	}
	response.Success(w, result)
}

// PUT /stores/:storeId/stock/min
func (h *StockHandler) SetMinStock(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	var req dto.SetMinStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	if errs := h.validator.ValidateStruct(req); errs != nil {
		response.ValidationError(w, errs)
		return
	}
	if err := h.stockSvc.SetMinStock(r.Context(), storeID, &req); err != nil {
		if errors.Is(err, service.ErrProductNotInStore) {
			response.NotFound(w, "Product")
			return
		}
		response.InternalError(w)
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true, "message": "Minimum stock threshold updated",
	})
}

// GET /stores/:storeId/stock/movements
func (h *StockHandler) GetMovements(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	filter := dto.StockMovementFilter{StoreID: storeID}
	filter.Page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	filter.PerPage, _ = strconv.Atoi(r.URL.Query().Get("per_page"))
	filter.ProductID = r.URL.Query().Get("product_id")
	filter.RefType = r.URL.Query().Get("ref_type")

	movements, meta, err := h.stockSvc.GetMovements(r.Context(), filter)
	if err != nil {
		h.log.Error().Err(err).Msg("get movements failed")
		response.InternalError(w)
		return
	}
	response.Success(w, dto.ListResponse{Data: movements, Meta: meta})
}
