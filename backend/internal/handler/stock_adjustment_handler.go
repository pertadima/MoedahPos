package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/middleware"
	"github.com/moedahpos/backend/internal/service"
	"github.com/moedahpos/backend/internal/validator"
	"github.com/moedahpos/backend/pkg/response"
)

type StockAdjustmentHandler struct {
	svc      service.StockAdjustmentServiceInterface
	validate *validator.Validator
	log      *zerolog.Logger
}

func NewStockAdjustmentHandler(svc service.StockAdjustmentServiceInterface, validate *validator.Validator, log *zerolog.Logger) *StockAdjustmentHandler {
	return &StockAdjustmentHandler{
		svc:      svc,
		validate: validate,
		log:      log,
	}
}

func (h *StockAdjustmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	if storeID == "" {
		response.Error(w, http.StatusBadRequest, "store_id represents a required parameter")
		return
	}

	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		response.Unauthorized(w, "unauthorized")
		return
	}

	var req domain.CreateAdjustmentInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	if errs := h.validate.ValidateStruct(req); errs != nil {
		response.ValidationError(w, errs)
		return
	}

	if err := h.svc.CreateAdjustment(r.Context(), storeID, userID, req); err != nil {
		h.log.Error().Err(err).Msg("failed to execute stock adjustment")
		// Ideally we wrap domain errors, but for simplicity we return 400 with message
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(w, map[string]string{"status": "Stock adjustment successful"})
}

func (h *StockAdjustmentHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	if storeID == "" {
		response.Error(w, http.StatusBadRequest, "store_id required")
		return
	}

	productID := r.URL.Query().Get("product_id")
	var pID *string
	if productID != "" {
		pID = &productID
	}

	adjustments, err := h.svc.GetAdjustmentHistory(r.Context(), storeID, pID)
	if err != nil {
		h.log.Error().Err(err).Msg("failed to load stock adjustments")
		response.InternalError(w)
		return
	}

	response.Success(w, adjustments)
}
