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

// PurchaseOrderHandler handles purchase-order lifecycle endpoints.
type PurchaseOrderHandler struct {
	poSvc     *service.PurchaseOrderService
	validator *validator.Validator
	log       zerolog.Logger
}

func NewPurchaseOrderHandler(poSvc *service.PurchaseOrderService, v *validator.Validator, log zerolog.Logger) *PurchaseOrderHandler {
	return &PurchaseOrderHandler{poSvc: poSvc, validator: v, log: log}
}

// GET /stores/:storeId/purchase-orders
func (h *PurchaseOrderHandler) List(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	filter := dto.POListFilter{StoreID: storeID}
	filter.Page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	filter.PerPage, _ = strconv.Atoi(r.URL.Query().Get("per_page"))
	filter.Status = r.URL.Query().Get("status")
	filter.DateFrom = r.URL.Query().Get("date_from")
	filter.DateTo = r.URL.Query().Get("date_to")

	pos, meta, err := h.poSvc.ListPOs(r.Context(), filter)
	if err != nil {
		h.log.Error().Err(err).Msg("list POs failed")
		response.InternalError(w)
		return
	}
	response.Success(w, dto.ListResponse{Data: pos, Meta: meta})
}

// POST /stores/:storeId/purchase-orders
func (h *PurchaseOrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	var req dto.CreatePORequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	if errs := h.validator.ValidateStruct(req); errs != nil {
		response.ValidationError(w, errs)
		return
	}
	result, err := h.poSvc.CreatePO(r.Context(), storeID, &req, userIDFromCtx(r))
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			response.NotFound(w, "Product")
			return
		}
		h.log.Error().Err(err).Msg("create PO failed")
		response.InternalError(w)
		return
	}
	response.Created(w, result)
}

// GET /stores/:storeId/purchase-orders/:poId
func (h *PurchaseOrderHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "poId")
	result, err := h.poSvc.GetPO(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrPONotFound) {
			response.NotFound(w, "Purchase Order")
			return
		}
		response.InternalError(w)
		return
	}
	response.Success(w, result)
}

// PUT /stores/:storeId/purchase-orders/:poId
func (h *PurchaseOrderHandler) Update(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	id := chi.URLParam(r, "poId")
	var req dto.UpdatePORequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	if errs := h.validator.ValidateStruct(req); errs != nil {
		response.ValidationError(w, errs)
		return
	}
	result, err := h.poSvc.UpdatePO(r.Context(), id, &req, storeID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPONotFound):
			response.NotFound(w, "Purchase Order")
		case errors.Is(err, service.ErrPONotEditable):
			response.Error(w, http.StatusConflict, "PO can only be edited when in draft status")
		default:
			response.InternalError(w)
		}
		return
	}
	response.Success(w, result)
}

// POST /stores/:storeId/purchase-orders/:poId/submit
func (h *PurchaseOrderHandler) Submit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "poId")
	if err := h.poSvc.SubmitPO(r.Context(), id, userIDFromCtx(r)); err != nil {
		switch {
		case errors.Is(err, service.ErrPONotFound):
			response.NotFound(w, "Purchase Order")
		case errors.Is(err, service.ErrPOCannotSubmit):
			response.Error(w, http.StatusConflict, "PO can only be submitted when in draft status")
		default:
			response.InternalError(w)
		}
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true, "message": "Purchase order submitted",
	})
}

// POST /stores/:storeId/purchase-orders/:poId/receive
func (h *PurchaseOrderHandler) Receive(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "poId")
	if err := h.poSvc.ReceivePO(r.Context(), id, userIDFromCtx(r)); err != nil {
		switch {
		case errors.Is(err, service.ErrPONotFound):
			response.NotFound(w, "Purchase Order")
		case errors.Is(err, service.ErrPOCannotReceive):
			response.Error(w, http.StatusConflict, "PO must be in ordered status to receive")
		default:
			h.log.Error().Err(err).Msg("receive PO failed")
			response.InternalError(w)
		}
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true, "message": "Purchase order received — stock levels updated",
	})
}

// DELETE /stores/:storeId/purchase-orders/:poId  (soft cancel)
func (h *PurchaseOrderHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "poId")
	if err := h.poSvc.CancelPO(r.Context(), id); err != nil {
		switch {
		case errors.Is(err, service.ErrPONotFound):
			response.NotFound(w, "Purchase Order")
		case errors.Is(err, service.ErrPOCannotCancel):
			response.Error(w, http.StatusConflict, "Cannot cancel a received or already cancelled PO")
		default:
			response.InternalError(w)
		}
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true, "message": "Purchase order cancelled",
	})
}

// GET /stores/:storeId/purchase-orders/payables
func (h *PurchaseOrderHandler) PayableSummary(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	summary, err := h.poSvc.PayableSummary(r.Context(), storeID)
	if err != nil {
		h.log.Error().Err(err).Msg("payable summary failed")
		response.InternalError(w)
		return
	}
	response.Success(w, summary)
}

// POST /stores/:storeId/purchase-orders/:poId/payments
func (h *PurchaseOrderHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	poID    := chi.URLParam(r, "poId")
	var req dto.POPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	if errs := h.validator.ValidateStruct(req); errs != nil {
		response.ValidationError(w, errs)
		return
	}
	result, err := h.poSvc.CreatePayment(r.Context(), poID, storeID, userIDFromCtx(r), req)
	if err != nil {
		if errors.Is(err, service.ErrPONotFound) {
			response.NotFound(w, "Purchase Order")
			return
		}
		response.Error(w, http.StatusConflict, err.Error())
		return
	}
	response.Created(w, result)
}

// GET /stores/:storeId/purchase-orders/:poId/payments
func (h *PurchaseOrderHandler) ListPayments(w http.ResponseWriter, r *http.Request) {
	poID := chi.URLParam(r, "poId")
	rows, err := h.poSvc.ListPayments(r.Context(), poID)
	if err != nil {
		h.log.Error().Err(err).Msg("list payments failed")
		response.InternalError(w)
		return
	}
	response.Success(w, rows)
}
