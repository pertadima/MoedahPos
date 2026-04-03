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

// SupplierHandler handles supplier CRUD endpoints.
type SupplierHandler struct {
	supplierSvc *service.SupplierService
	validator   *validator.Validator
	log         zerolog.Logger
}

func NewSupplierHandler(supplierSvc *service.SupplierService, v *validator.Validator, log zerolog.Logger) *SupplierHandler {
	return &SupplierHandler{supplierSvc: supplierSvc, validator: v, log: log}
}

// GET /api/v1/suppliers
func (h *SupplierHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := dto.SupplierListFilter{}
	filter.Page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	filter.PerPage, _ = strconv.Atoi(r.URL.Query().Get("per_page"))
	filter.Search = r.URL.Query().Get("search")
	if v := r.URL.Query().Get("is_active"); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			filter.IsActive = &b
		}
	}

	suppliers, meta, err := h.supplierSvc.ListSuppliers(r.Context(), filter)
	if err != nil {
		h.log.Error().Err(err).Msg("list suppliers failed")
		response.InternalError(w)
		return
	}
	response.Success(w, dto.ListResponse{Data: suppliers, Meta: meta})
}

// POST /api/v1/suppliers
func (h *SupplierHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateSupplierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	if errs := h.validator.ValidateStruct(req); errs != nil {
		response.ValidationError(w, errs)
		return
	}
	result, err := h.supplierSvc.CreateSupplier(r.Context(), &req)
	if err != nil {
		h.log.Error().Err(err).Msg("create supplier failed")
		response.InternalError(w)
		return
	}
	response.Created(w, result)
}

// GET /api/v1/suppliers/:supplierId
func (h *SupplierHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "supplierId")
	result, err := h.supplierSvc.GetSupplier(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrSupplierNotFound) {
			response.NotFound(w, "Supplier")
			return
		}
		response.InternalError(w)
		return
	}
	response.Success(w, result)
}

// PUT /api/v1/suppliers/:supplierId
func (h *SupplierHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "supplierId")
	var req dto.UpdateSupplierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	if errs := h.validator.ValidateStruct(req); errs != nil {
		response.ValidationError(w, errs)
		return
	}
	result, err := h.supplierSvc.UpdateSupplier(r.Context(), id, &req)
	if err != nil {
		if errors.Is(err, service.ErrSupplierNotFound) {
			response.NotFound(w, "Supplier")
			return
		}
		response.InternalError(w)
		return
	}
	response.Success(w, result)
}

// DELETE /api/v1/suppliers/:supplierId  (soft delete)
func (h *SupplierHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "supplierId")
	if err := h.supplierSvc.DeleteSupplier(r.Context(), id); err != nil {
		if errors.Is(err, service.ErrSupplierNotFound) {
			response.NotFound(w, "Supplier")
			return
		}
		response.InternalError(w)
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true, "message": "Supplier deleted successfully",
	})
}
