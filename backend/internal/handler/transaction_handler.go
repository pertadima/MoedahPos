package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/moedahpos/backend/internal/dto"
	"github.com/moedahpos/backend/internal/service"
	"github.com/moedahpos/backend/internal/validator"
	"github.com/moedahpos/backend/pkg/response"
	"github.com/rs/zerolog"
)

// TransactionHandler handles cashier-facing transaction endpoints.
type TransactionHandler struct {
	txnSvc    *service.TransactionService
	validator *validator.Validator
	log       zerolog.Logger
}

func NewTransactionHandler(txnSvc *service.TransactionService, v *validator.Validator, log zerolog.Logger) *TransactionHandler {
	return &TransactionHandler{txnSvc: txnSvc, validator: v, log: log}
}

// POST /stores/:storeId/transactions  – checkout / create sale
func (h *TransactionHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	var req dto.CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	if errs := h.validator.ValidateStruct(req); errs != nil {
		response.ValidationError(w, errs)
		return
	}

	result, err := h.txnSvc.Checkout(r.Context(), storeID, &req, userIDFromCtx(r))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrProductNotFound):
			response.NotFound(w, "Product")
		case errors.Is(err, service.ErrInsufficientStock):
			response.Error(w, http.StatusUnprocessableEntity, err.Error())
		case errors.Is(err, service.ErrInsuficientPayment):
			response.Error(w, http.StatusUnprocessableEntity, "Payment amount is less than the total")
		default:
			h.log.Error().Err(err).Msg("checkout failed")
			response.InternalError(w)
		}
		return
	}
	response.Created(w, result)
}

// GET /stores/:storeId/transactions
func (h *TransactionHandler) List(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	filter := dto.TransactionListFilter{StoreID: storeID}
	filter.Page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	filter.PerPage, _ = strconv.Atoi(r.URL.Query().Get("per_page"))
	filter.Status = r.URL.Query().Get("status")
	filter.CashierID = r.URL.Query().Get("cashier_id")
	filter.DateFrom = r.URL.Query().Get("date_from")
	filter.DateTo = r.URL.Query().Get("date_to")

	txns, meta, err := h.txnSvc.ListTransactions(r.Context(), filter)
	if err != nil {
		h.log.Error().Err(err).Msg("list transactions failed")
		response.InternalError(w)
		return
	}
	response.Success(w, dto.ListResponse{Data: txns, Meta: meta})
}

// GET /stores/:storeId/transactions/:txnId  – fetch receipt
func (h *TransactionHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "txnId")
	result, err := h.txnSvc.GetTransaction(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrTransactionNotFound) {
			response.NotFound(w, "Transaction")
			return
		}
		response.InternalError(w)
		return
	}
	response.Success(w, result)
}

// POST /stores/:storeId/transactions/:txnId/void
func (h *TransactionHandler) Void(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "txnId")
	if err := h.txnSvc.VoidTransaction(r.Context(), id, userIDFromCtx(r)); err != nil {
		switch {
		case errors.Is(err, service.ErrTransactionNotFound):
			response.NotFound(w, "Transaction")
		case errors.Is(err, service.ErrTransactionAlreadyVoided):
			response.Error(w, http.StatusConflict, "Transaction is already voided")
		default:
			h.log.Error().Err(err).Msg("void transaction failed")
			response.InternalError(w)
		}
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true, "message": "Transaction voided and stock restored",
	})
}
