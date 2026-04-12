package handler

import (
	"encoding/json"
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

//nolint:goconst // Magic strings are okay here
type ExpenseHandler struct {
	expenseSvc *service.ExpenseService
	validate   *validator.Validator
	log        zerolog.Logger
}

func NewExpenseHandler(expenseSvc *service.ExpenseService, validate *validator.Validator, log zerolog.Logger) *ExpenseHandler {
	return &ExpenseHandler{expenseSvc: expenseSvc, validate: validate, log: log}
}

// ── Categories ────────────────────────────────────────────────────────────────

func (h *ExpenseHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	includeDeleted, _ := strconv.ParseBool(r.URL.Query().Get("include_deleted"))
	cats, err := h.expenseSvc.ListCategories(r.Context(), includeDeleted)
	if err != nil {
		h.log.Error().Err(err).Msg("list expense categories failed")
		response.InternalError(w)
		return
	}
	response.Success(w, cats)
}

func (h *ExpenseHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateExpenseCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if errs := h.validate.ValidateStruct(req); errs != nil {
		response.ValidationError(w, errs)
		return
	}

	cat, err := h.expenseSvc.CreateCategory(r.Context(), &req)
	if err != nil {
		h.log.Error().Err(err).Msg("create expense category failed")
		response.InternalError(w)
		return
	}
	response.Created(w, cat)
}

func (h *ExpenseHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req dto.UpdateExpenseCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if errs := h.validate.ValidateStruct(req); errs != nil {
		response.ValidationError(w, errs)
		return
	}

	cat, err := h.expenseSvc.UpdateCategory(r.Context(), id, &req)
	if err != nil {
		h.log.Error().Err(err).Msg("update expense category failed")
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(w, "Category")
			return
		}
		response.InternalError(w)
		return
	}
	response.Success(w, cat)
}

func (h *ExpenseHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.expenseSvc.SoftDeleteCategory(r.Context(), id); err != nil {
		h.log.Error().Err(err).Msg("delete expense category failed")
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(w, "Category")
			return
		}
		response.InternalError(w)
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{"success": true, "message": "Category deleted"})
}

// ── Expenses ──────────────────────────────────────────────────────────────────

func (h *ExpenseHandler) ListExpenses(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	f := dto.ExpenseListFilter{
		StoreID:    storeID,
		CategoryID: r.URL.Query().Get("category_id"),
		DateFrom:   r.URL.Query().Get("date_from"),
		DateTo:     r.URL.Query().Get("date_to"),
	}
	f.Page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	f.PerPage, _ = strconv.Atoi(r.URL.Query().Get("per_page"))

	expenses, meta, err := h.expenseSvc.ListExpenses(r.Context(), f)
	if err != nil {
		h.log.Error().Err(err).Msg("list expenses failed")
		response.InternalError(w)
		return
	}

	response.Success(w, dto.ListResponse{Data: expenses, Meta: meta})
}

func (h *ExpenseHandler) CreateExpense(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	userID := middleware.UserIDFromContext(r.Context())

	var req dto.CreateExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if errs := h.validate.ValidateStruct(req); errs != nil {
		response.ValidationError(w, errs)
		return
	}

	expense, err := h.expenseSvc.CreateExpense(r.Context(), storeID, userID, &req)
	if err != nil {
		h.log.Error().Err(err).Msg("create expense failed")
		if err.Error() == "invalid expense_date" {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		response.InternalError(w)
		return
	}
	response.Created(w, expense)
}

func (h *ExpenseHandler) UpdateExpense(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	id := chi.URLParam(r, "id")

	var req dto.UpdateExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if errs := h.validate.ValidateStruct(req); errs != nil {
		response.ValidationError(w, errs)
		return
	}

	expense, err := h.expenseSvc.UpdateExpense(r.Context(), id, storeID, &req)
	if err != nil {
		h.log.Error().Err(err).Msg("update expense failed")
		if strings.Contains(err.Error(), "expense not found") {
			response.NotFound(w, "Expense")
			return
		}
		if err.Error() == "invalid expense_date" {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		response.InternalError(w)
		return
	}
	response.Success(w, expense)
}

func (h *ExpenseHandler) UpdateExpenseStatus(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	id := chi.URLParam(r, "id")

	var req dto.UpdateExpenseStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if errs := h.validate.ValidateStruct(req); errs != nil {
		response.ValidationError(w, errs)
		return
	}

	expense, err := h.expenseSvc.UpdatePaymentStatus(r.Context(), id, storeID, &req)
	if err != nil {
		h.log.Error().Err(err).Msg("update expense status failed")
		if err.Error() == "updating expense payment status: expense not found" || err.Error() == "expense not found" {
			response.NotFound(w, "Expense")
			return
		}
		response.InternalError(w)
		return
	}
	response.Success(w, expense)
}

func (h *ExpenseHandler) DeleteExpense(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	id := chi.URLParam(r, "id")

	if err := h.expenseSvc.DeleteExpense(r.Context(), id, storeID); err != nil {
		h.log.Error().Err(err).Msg("delete expense failed")
		if err.Error() == "deleting expense: expense not found" || err.Error() == "expense not found" {
			response.NotFound(w, "Expense")
			return
		}
		response.InternalError(w)
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{"success": true, "message": "Expense deleted"})
}

// ── Recurring Expenses ─────────────────────────────────────────────────────────

func (h *ExpenseHandler) ListRecurringExpenses(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	f := dto.ExpenseListFilter{
		StoreID:    storeID,
		CategoryID: r.URL.Query().Get("category_id"),
	}
	f.Page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	f.PerPage, _ = strconv.Atoi(r.URL.Query().Get("per_page"))

	expenses, meta, err := h.expenseSvc.ListRecurringExpenses(r.Context(), f)
	if err != nil {
		h.log.Error().Err(err).Msg("list recurring expenses failed")
		response.InternalError(w)
		return
	}

	response.Success(w, dto.ListResponse{Data: expenses, Meta: meta})
}

func (h *ExpenseHandler) CreateRecurringExpense(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	userID := middleware.UserIDFromContext(r.Context())

	var req dto.CreateRecurringExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if errs := h.validate.ValidateStruct(req); errs != nil {
		response.ValidationError(w, errs)
		return
	}

	re, err := h.expenseSvc.CreateRecurringExpense(r.Context(), storeID, userID, &req)
	if err != nil {
		h.log.Error().Err(err).Msg("create recurring expense failed")
		response.InternalError(w)
		return
	}
	response.Created(w, re)
}

func (h *ExpenseHandler) UpdateRecurringExpense(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	id := chi.URLParam(r, "id")

	var req dto.UpdateRecurringExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if errs := h.validate.ValidateStruct(req); errs != nil {
		response.ValidationError(w, errs)
		return
	}

	re, err := h.expenseSvc.UpdateRecurringExpense(r.Context(), id, storeID, &req)
	if err != nil {
		h.log.Error().Err(err).Msg("update recurring expense failed")
		response.InternalError(w)
		return
	}
	response.Success(w, re)
}

func (h *ExpenseHandler) DeleteRecurringExpense(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	id := chi.URLParam(r, "id")

	if err := h.expenseSvc.DeleteRecurringExpense(r.Context(), id, storeID); err != nil {
		h.log.Error().Err(err).Msg("delete recurring expense failed")
		if err.Error() == "deleting recurring expense: recurring expense not found" || err.Error() == "recurring expense not found" {
			response.NotFound(w, "RecurringExpense")
			return
		}
		response.InternalError(w)
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{"success": true, "message": "Recurring expense deleted"})
}
