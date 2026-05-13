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

// TerminHandler exposes HTTP endpoints for PO installment (termin) management.
type TerminHandler struct {
	terminSvc service.TerminServiceInterface
	validator *validator.Validator
	log       zerolog.Logger
}

// NewTerminHandler creates a TerminHandler.
func NewTerminHandler(terminSvc service.TerminServiceInterface, v *validator.Validator, log zerolog.Logger) *TerminHandler {
	return &TerminHandler{terminSvc: terminSvc, validator: v, log: log}
}

// CreateSchedule handles POST /stores/:storeId/purchase-orders/:poId/termins
// Creates or replaces the installment schedule for a received PO.
func (h *TerminHandler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	poID := chi.URLParam(r, "poId")

	var req dto.CreateTerminScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	if errs := h.validator.ValidateStruct(req); errs != nil {
		response.ValidationError(w, errs)
		return
	}

	termins, err := h.terminSvc.CreateTerminSchedule(r.Context(), poID, req)
	if err != nil {
		h.handleErr(w, err)
		return
	}
	response.Created(w, termins)
}

// ListTermins handles GET /stores/:storeId/purchase-orders/:poId/termins
// Returns all termins for a PO with payment history.
func (h *TerminHandler) ListTermins(w http.ResponseWriter, r *http.Request) {
	poID := chi.URLParam(r, "poId")
	termins, err := h.terminSvc.GetTerminSchedule(r.Context(), poID)
	if err != nil {
		h.handleErr(w, err)
		return
	}
	response.Success(w, termins)
}

// RecordPayment handles POST /stores/:storeId/purchase-orders/:poId/termins/:terminId/payments
// Records a payment (full or partial) against a single termin.
func (h *TerminHandler) RecordPayment(w http.ResponseWriter, r *http.Request) {
	terminID := chi.URLParam(r, "terminId")

	var req dto.RecordPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	if errs := h.validator.ValidateStruct(req); errs != nil {
		response.ValidationError(w, errs)
		return
	}

	rec, err := h.terminSvc.RecordPayment(r.Context(), terminID, userIDFromCtx(r), req)
	if err != nil {
		h.handleErr(w, err)
		return
	}
	response.Created(w, rec)
}

// GetDebtSummary handles GET /stores/:storeId/purchase-orders/:poId/debt
// Returns aggregated debt metrics: total_paid, remaining, status, overdue count.
func (h *TerminHandler) GetDebtSummary(w http.ResponseWriter, r *http.Request) {
	poID := chi.URLParam(r, "poId")
	summary, err := h.terminSvc.CalculatePODebt(r.Context(), poID)
	if err != nil {
		h.handleErr(w, err)
		return
	}
	response.Success(w, summary)
}

// GetDocument handles GET /stores/:storeId/purchase-orders/:poId/document?type=invoice|receipt|termin_agreement
// Returns all data required by the frontend to render a printable document.
func (h *TerminHandler) GetDocument(w http.ResponseWriter, r *http.Request) {
	poID := chi.URLParam(r, "poId")
	docType := r.URL.Query().Get("type")
	if docType == "" {
		docType = "invoice"
	}
	switch docType {
	case "invoice", "receipt", "termin_agreement":
		// valid
	default:
		response.Error(w, http.StatusBadRequest, "type must be invoice, receipt, or termin_agreement")
		return
	}

	data, err := h.terminSvc.GenerateDocumentData(r.Context(), poID, docType)
	if err != nil {
		h.handleErr(w, err)
		return
	}
	response.Success(w, data)
}

// LogDocumentGenerate handles POST /stores/:storeId/purchase-orders/:poId/document-log
// Logs a document generation event to activity_logs (called once per click, not on doc page load).
func (h *TerminHandler) LogDocumentGenerate(w http.ResponseWriter, r *http.Request) {
	poID := chi.URLParam(r, "poId")
	storeID := chi.URLParam(r, "storeId")
	userID := userIDFromCtx(r)

	var body struct {
		DocumentType string `json:"document_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.DocumentType == "" {
		response.Error(w, http.StatusBadRequest, "document_type is required")
		return
	}

	if err := h.terminSvc.LogDocumentGenerate(r.Context(), poID, storeID, userID, body.DocumentType); err != nil {
		h.log.Error().Err(err).Msg("LogDocumentGenerate failed")
		response.InternalError(w)
		return
	}
	response.Success(w, map[string]string{"status": "logged"})
}

// handleErr maps service-layer sentinel errors to HTTP status codes.
func (h *TerminHandler) handleErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrPONotFound):
		response.NotFound(w, "Purchase Order")
	case errors.Is(err, service.ErrTerminNotFound):
		response.NotFound(w, "Termin")
	case errors.Is(err, service.ErrTerminOverpayment):
		response.Error(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrPONotReceived):
		response.Error(w, http.StatusBadRequest, "PO must be received before creating a termin schedule")
	default:
		h.log.Error().Err(err).Msg("termin handler error")
		response.InternalError(w)
	}
}
