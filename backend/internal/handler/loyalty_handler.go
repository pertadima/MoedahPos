package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/moedahpos/backend/internal/dto"
	"github.com/moedahpos/backend/internal/service"
	"github.com/moedahpos/backend/internal/validator"
	"github.com/moedahpos/backend/pkg/response"
)

// LoyaltyHandler handles all loyalty-related HTTP endpoints.
type LoyaltyHandler struct {
	svc       service.LoyaltyServiceInterface
	validator *validator.Validator
	log       zerolog.Logger
}

// NewLoyaltyHandler constructs a LoyaltyHandler.
func NewLoyaltyHandler(svc service.LoyaltyServiceInterface, v *validator.Validator, log zerolog.Logger) *LoyaltyHandler {
	return &LoyaltyHandler{svc: svc, validator: v, log: log}
}

// ListTiers godoc
// GET /api/v1/stores/{storeId}/loyalty/tiers
func (h *LoyaltyHandler) ListTiers(w http.ResponseWriter, r *http.Request) {
	tiers, err := h.svc.ListTiers(r.Context())
	if err != nil {
		h.log.Error().Err(err).Msg("LoyaltyHandler.ListTiers failed")
		response.Error(w, http.StatusInternalServerError, "Failed to list tiers")
		return
	}
	response.Success(w, tiers)
}

// GetBalance godoc
// GET /api/v1/stores/{storeId}/customers/{customerId}/loyalty
func (h *LoyaltyHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	customerID := chi.URLParam(r, "customerId")
	if customerID == "" {
		response.Error(w, http.StatusBadRequest, "Missing customer ID")
		return
	}
	bal, err := h.svc.GetBalance(r.Context(), customerID)
	if err != nil {
		h.log.Error().Err(err).Str("customer_id", customerID).Msg("LoyaltyHandler.GetBalance failed")
		response.Error(w, http.StatusInternalServerError, "Failed to get loyalty balance")
		return
	}
	response.Success(w, bal)
}

// EarnPoints godoc
// POST /api/v1/stores/{storeId}/customers/{customerId}/loyalty/earn
func (h *LoyaltyHandler) EarnPoints(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	customerID := chi.URLParam(r, "customerId")
	if customerID == "" {
		response.Error(w, http.StatusBadRequest, "Missing customer ID")
		return
	}

	var req dto.EarnPointsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	if errs := h.validator.ValidateStruct(req); errs != nil {
		response.ValidationError(w, errs)
		return
	}

	txnID := req.TransactionID
	entry, err := h.svc.EarnPoints(r.Context(), storeID, customerID, &txnID, req.Total)
	if err != nil {
		h.log.Error().Err(err).Str("customer_id", customerID).Msg("LoyaltyHandler.EarnPoints failed")
		response.Error(w, http.StatusInternalServerError, "Failed to earn points")
		return
	}
	response.Success(w, entry)
}

// RedeemPoints godoc
// POST /api/v1/stores/{storeId}/customers/{customerId}/loyalty/redeem
func (h *LoyaltyHandler) RedeemPoints(w http.ResponseWriter, r *http.Request) {
	customerID := chi.URLParam(r, "customerId")
	if customerID == "" {
		response.Error(w, http.StatusBadRequest, "Missing customer ID")
		return
	}

	var req dto.RedeemPointsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	if errs := h.validator.ValidateStruct(req); errs != nil {
		response.ValidationError(w, errs)
		return
	}

	entry, err := h.svc.RedeemPoints(r.Context(), customerID, nil, req.Points)
	if err != nil {
		if isLoyaltyValidationError(err) {
			response.Error(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		h.log.Error().Err(err).Str("customer_id", customerID).Msg("LoyaltyHandler.RedeemPoints failed")
		response.Error(w, http.StatusInternalServerError, "Failed to redeem points")
		return
	}
	response.Success(w, entry)
}

// GetHistory godoc
// GET /api/v1/stores/{storeId}/customers/{customerId}/loyalty/history
func (h *LoyaltyHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	customerID := chi.URLParam(r, "customerId")
	if customerID == "" {
		response.Error(w, http.StatusBadRequest, "Missing customer ID")
		return
	}
	history, err := h.svc.GetHistory(r.Context(), customerID)
	if err != nil {
		h.log.Error().Err(err).Str("customer_id", customerID).Msg("LoyaltyHandler.GetHistory failed")
		response.Error(w, http.StatusInternalServerError, "Failed to get loyalty history")
		return
	}
	response.Success(w, history)
}

// AssignTier godoc
// PUT /api/v1/stores/{storeId}/customers/{customerId}/loyalty/tier
func (h *LoyaltyHandler) AssignTier(w http.ResponseWriter, r *http.Request) {
	customerID := chi.URLParam(r, "customerId")
	if customerID == "" {
		response.Error(w, http.StatusBadRequest, "Missing customer ID")
		return
	}

	var req dto.AssignTierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	if errs := h.validator.ValidateStruct(req); errs != nil {
		response.ValidationError(w, errs)
		return
	}

	if err := h.svc.AssignTier(r.Context(), customerID, req.TierID); err != nil {
		if errors.Is(err, service.ErrTierNotFound) {
			response.Error(w, http.StatusNotFound, "Tier not found")
			return
		}
		h.log.Error().Err(err).Str("customer_id", customerID).Msg("LoyaltyHandler.AssignTier failed")
		response.Error(w, http.StatusInternalServerError, "Failed to assign tier")
		return
	}
	response.Success(w, map[string]string{"status": "ok"})
}

// isLoyaltyValidationError returns true for domain validation errors that should be 422.
func isLoyaltyValidationError(err error) bool {
	return errors.Is(err, service.ErrInsufficientPoints) || errors.Is(err, service.ErrInvalidRedemption)
}
