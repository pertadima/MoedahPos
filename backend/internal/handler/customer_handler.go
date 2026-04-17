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

// CustomerHandler handles customer CRUD endpoints.
type CustomerHandler struct {
	svc       service.CustomerServiceInterface
	validator *validator.Validator
	log       zerolog.Logger
}

func NewCustomerHandler(svc service.CustomerServiceInterface, v *validator.Validator, log zerolog.Logger) *CustomerHandler {
	return &CustomerHandler{svc: svc, validator: v, log: log}
}

// GET /stores/:storeId/customers
func (h *CustomerHandler) List(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	f := dto.CustomerListFilter{StoreID: storeID}
	f.Page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	f.PerPage, _ = strconv.Atoi(r.URL.Query().Get("per_page"))
	f.Search = r.URL.Query().Get("search")

	customers, meta, err := h.svc.List(r.Context(), f)
	if err != nil {
		h.log.Error().Err(err).Msg("list customers failed")
		response.InternalError(w)
		return
	}
	response.Success(w, dto.ListResponse{Data: customers, Meta: meta})
}

// GET /stores/:storeId/customers/search  (lightweight — for cashier picker)
func (h *CustomerHandler) Search(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	f := dto.CustomerListFilter{
		StoreID:         storeID,
		Search:          r.URL.Query().Get("q"),
		PaginationQuery: dto.PaginationQuery{PerPage: 10},
	}
	rows, err := h.svc.Search(r.Context(), f)
	if err != nil {
		response.InternalError(w)
		return
	}
	response.Success(w, rows)
}

// POST /stores/:storeId/customers
func (h *CustomerHandler) Create(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	var req dto.CreateCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	if errs := h.validator.ValidateStruct(req); errs != nil {
		response.ValidationError(w, errs)
		return
	}
	result, err := h.svc.Create(r.Context(), storeID, req)
	if err != nil {
		h.log.Error().Err(err).Msg("create customer failed")
		response.InternalError(w)
		return
	}
	response.Created(w, result)
}

// GET /stores/:storeId/customers/:customerId
func (h *CustomerHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "customerId")
	result, err := h.svc.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrCustomerNotFound) {
			response.NotFound(w, "Customer")
			return
		}
		response.InternalError(w)
		return
	}
	response.Success(w, result)
}

// PUT /stores/:storeId/customers/:customerId
func (h *CustomerHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "customerId")
	var req dto.UpdateCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	if errs := h.validator.ValidateStruct(req); errs != nil {
		response.ValidationError(w, errs)
		return
	}
	result, err := h.svc.Update(r.Context(), id, req)
	if err != nil {
		if errors.Is(err, service.ErrCustomerNotFound) {
			response.NotFound(w, "Customer")
			return
		}
		response.InternalError(w)
		return
	}
	response.Success(w, result)
}

// DELETE /stores/:storeId/customers/:customerId
func (h *CustomerHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "customerId")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		if errors.Is(err, service.ErrCustomerNotFound) {
			response.NotFound(w, "Customer")
			return
		}
		response.InternalError(w)
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{"success": true, "message": "Customer deleted"})
}
