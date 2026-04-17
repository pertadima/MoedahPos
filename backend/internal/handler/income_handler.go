package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/moedahpos/backend/internal/dto"
	"github.com/moedahpos/backend/internal/middleware"
	"github.com/moedahpos/backend/internal/service"
	"github.com/moedahpos/backend/internal/validator"
	"github.com/moedahpos/backend/pkg/response"
)

// IncomeHandler handles income-category and income endpoints.
type IncomeHandler struct {
	incomeSvc service.IncomeServiceInterface
	validate  *validator.Validator
	log       zerolog.Logger
}

func NewIncomeHandler(incomeSvc service.IncomeServiceInterface, validate *validator.Validator, log zerolog.Logger) *IncomeHandler {
	return &IncomeHandler{incomeSvc: incomeSvc, validate: validate, log: log}
}

// ── Categories ────────────────────────────────────────────────────────────────

// GET /income-categories
func (h *IncomeHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	includeDeleted, _ := strconv.ParseBool(r.URL.Query().Get("include_deleted"))
	cats, err := h.incomeSvc.ListCategories(r.Context(), includeDeleted)
	if err != nil {
		h.log.Error().Err(err).Msg("list income categories failed")
		response.InternalError(w)
		return
	}
	response.Success(w, cats)
}

// POST /income-categories
func (h *IncomeHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateIncomeCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if errs := h.validate.ValidateStruct(req); errs != nil {
		response.ValidationError(w, errs)
		return
	}
	cat, err := h.incomeSvc.CreateCategory(r.Context(), &req)
	if err != nil {
		h.log.Error().Err(err).Msg("create income category failed")
		response.InternalError(w)
		return
	}
	response.Created(w, cat)
}

// PUT /income-categories/:id
func (h *IncomeHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req dto.UpdateIncomeCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if errs := h.validate.ValidateStruct(req); errs != nil {
		response.ValidationError(w, errs)
		return
	}

	cat, err := h.incomeSvc.UpdateCategory(r.Context(), id, &req)
	if err != nil {
		h.log.Error().Err(err).Msg("update income category failed")
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(w, "Category")
			return
		}
		response.InternalError(w)
		return
	}
	response.Success(w, cat)
}

// DELETE /income-categories/:id
func (h *IncomeHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.incomeSvc.SoftDeleteCategory(r.Context(), id); err != nil {
		h.log.Error().Err(err).Msg("delete income category failed")
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(w, "Category")
			return
		}
		response.InternalError(w)
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{"success": true, "message": "Category deleted"})
}

// ── Incomes ───────────────────────────────────────────────────────────────────

// GET /stores/:storeId/incomes
func (h *IncomeHandler) ListIncomes(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	f := dto.IncomeListFilter{
		StoreID:    storeID,
		CategoryID: r.URL.Query().Get("category_id"),
		DateFrom:   r.URL.Query().Get("date_from"),
		DateTo:     r.URL.Query().Get("date_to"),
	}
	f.Page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	f.PerPage, _ = strconv.Atoi(r.URL.Query().Get("per_page"))

	incomes, meta, err := h.incomeSvc.ListIncomes(r.Context(), f)
	if err != nil {
		h.log.Error().Err(err).Msg("list incomes failed")
		response.InternalError(w)
		return
	}
	response.Success(w, dto.ListResponse{Data: incomes, Meta: meta})
}

// POST /stores/:storeId/incomes
func (h *IncomeHandler) CreateIncome(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	userID := middleware.UserIDFromContext(r.Context())

	var req dto.CreateIncomeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if errs := h.validate.ValidateStruct(req); errs != nil {
		response.ValidationError(w, errs)
		return
	}

	inc, err := h.incomeSvc.CreateIncome(r.Context(), storeID, userID, &req)
	if err != nil {
		h.log.Error().Err(err).Msg("create income failed")
		if err.Error() == "invalid income_date: use YYYY-MM-DD" {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		response.InternalError(w)
		return
	}
	response.Created(w, inc)
}

// PUT /stores/:storeId/incomes/:id
func (h *IncomeHandler) UpdateIncome(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	id := chi.URLParam(r, "id")

	var req dto.UpdateIncomeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if errs := h.validate.ValidateStruct(req); errs != nil {
		response.ValidationError(w, errs)
		return
	}

	inc, err := h.incomeSvc.UpdateIncome(r.Context(), id, storeID, &req)
	if err != nil {
		h.log.Error().Err(err).Msg("update income failed")
		if errors.Is(err, service.ErrIncomeNotFound) {
			response.NotFound(w, "Income")
			return
		}
		if err.Error() == "invalid income_date: use YYYY-MM-DD" {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		response.InternalError(w)
		return
	}
	response.Success(w, inc)
}

// DELETE /stores/:storeId/incomes/:id
func (h *IncomeHandler) DeleteIncome(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	id := chi.URLParam(r, "id")

	if err := h.incomeSvc.DeleteIncome(r.Context(), id, storeID); err != nil {
		h.log.Error().Err(err).Msg("delete income failed")
		if errors.Is(err, service.ErrIncomeNotFound) {
			response.NotFound(w, "Income")
			return
		}
		response.InternalError(w)
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{"success": true, "message": "Income deleted"})
}
