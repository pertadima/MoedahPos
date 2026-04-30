package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/moedahpos/backend/internal/dto"
	svcMocks "github.com/moedahpos/backend/internal/service/mocks"
	"github.com/moedahpos/backend/internal/validator"
)

const expenseTestStoreID = "st1"

func TestExpenseHandler_ListCategories(t *testing.T) {
	svc := svcMocks.NewExpenseServiceInterface(t)
	v := validator.New()
	h := NewExpenseHandler(svc, v, zerolog.Nop())

	t.Run("success", func(t *testing.T) {
		svc.On("ListCategories", mock.Anything, false).
			Return([]dto.ExpenseCategoryResponse{{ID: "cat1", Name: "Food"}}, nil).Once()

		req := httptest.NewRequestWithContext(context.Background(), "GET", "/expense-categories?include_deleted=false", nil)
		w := httptest.NewRecorder()

		h.ListCategories(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("CreateCategory Success", func(t *testing.T) {
		reqBody := dto.CreateExpenseCategoryRequest{Name: "New Cat"}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/expense-categories", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		svc.On("CreateCategory", mock.Anything, mock.Anything).Return(&dto.ExpenseCategoryResponse{ID: "cat2", Name: "New Cat"}, nil).Once()

		h.CreateCategory(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("UpdateCategory Success", func(t *testing.T) {
		reqBody := dto.UpdateExpenseCategoryRequest{Name: "Updated Cat"}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPut, "/expense-categories/cat1", bytes.NewBuffer(body))
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "cat1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		svc.On("UpdateCategory", mock.Anything, "cat1", mock.Anything).Return(&dto.ExpenseCategoryResponse{ID: "cat1", Name: "Updated Cat"}, nil).Once()

		h.UpdateCategory(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("DeleteCategory Success", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodDelete, "/expense-categories/cat1", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "cat1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		svc.On("SoftDeleteCategory", mock.Anything, "cat1").Return(nil).Once()

		h.DeleteCategory(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("CreateCategory Validation Error", func(t *testing.T) {
		reqBody := dto.CreateExpenseCategoryRequest{Name: ""}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/expense-categories", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		h.CreateCategory(w, req)
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("UpdateCategory Validation Error", func(t *testing.T) {
		reqBody := dto.UpdateExpenseCategoryRequest{Name: ""}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPut, "/expense-categories/cat1", bytes.NewBuffer(body))
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "cat1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		h.UpdateCategory(w, req)
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})
}

func TestExpenseHandler_ListExpenses(t *testing.T) {
	svc := svcMocks.NewExpenseServiceInterface(t)
	v := validator.New()
	h := NewExpenseHandler(svc, v, zerolog.Nop())

	t.Run("success", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores/s1/expenses", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		svc.On("ListExpenses", mock.Anything, mock.Anything).Return([]*dto.ExpenseResponse{{ID: "e1"}}, dto.PaginationMeta{Total: 1}, nil).Once()

		h.ListExpenses(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestExpenseHandler_CreateExpense(t *testing.T) {
	svc := svcMocks.NewExpenseServiceInterface(t)
	v := validator.New()
	h := NewExpenseHandler(svc, v, zerolog.Nop())

	storeID := expenseTestStoreID
	reqBody := dto.CreateExpenseRequest{
		CategoryID:    "550e8400-e29b-41d4-a716-446655440000",
		Amount:        100.0,
		ExpenseDate:   "2026-04-17",
		PaymentStatus: "paid",
	}
	body, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		svc.On("CreateExpense", mock.Anything, storeID, mock.Anything, mock.AnythingOfType("*dto.CreateExpenseRequest")).
			Return(&dto.ExpenseResponse{ID: "e1"}, nil).Once()

		req := httptest.NewRequestWithContext(context.Background(), "POST", "/stores/"+storeID+"/expenses", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		// Set storeId param
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", storeID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		h.CreateExpense(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("Validation Error", func(t *testing.T) {
		reqBodyEmpty := dto.CreateExpenseRequest{Amount: -10}
		bodyEmpty, _ := json.Marshal(reqBodyEmpty)
		req := httptest.NewRequestWithContext(context.Background(), "POST", "/stores/"+storeID+"/expenses", bytes.NewBuffer(bodyEmpty))
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", storeID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		h.CreateExpense(w, req)
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})
}

func TestExpenseHandler_DeleteExpense(t *testing.T) {
	svc := svcMocks.NewExpenseServiceInterface(t)
	v := validator.New()
	h := NewExpenseHandler(svc, v, zerolog.Nop())

	storeID := expenseTestStoreID
	id := "e1"

	t.Run("success", func(t *testing.T) {
		svc.On("DeleteExpense", mock.Anything, id, storeID).Return(nil).Once()

		req := httptest.NewRequestWithContext(context.Background(), "DELETE", "/stores/"+storeID+"/expenses/"+id, nil)
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", storeID)
		rctx.URLParams.Add("id", id)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		h.DeleteExpense(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestExpenseHandler_UpdateStatus(t *testing.T) {
	svc := svcMocks.NewExpenseServiceInterface(t)
	v := validator.New()
	h := NewExpenseHandler(svc, v, zerolog.Nop())

	t.Run("success", func(t *testing.T) {
		reqBody := dto.UpdateExpenseStatusRequest{PaymentStatus: "paid"}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPatch, "/stores/s1/expenses/e1/status", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		rctx.URLParams.Add("id", "e1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		svc.On("UpdatePaymentStatus", mock.Anything, "e1", "s1", mock.Anything).Return(&dto.ExpenseResponse{ID: "e1"}, nil).Once()

		h.UpdateExpenseStatus(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

//nolint:funlen
func TestExpenseHandler_Recurring(t *testing.T) {
	svc := svcMocks.NewExpenseServiceInterface(t)
	v := validator.New()
	h := NewExpenseHandler(svc, v, zerolog.Nop())

	t.Run("ListRecurring", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores/s1/recurring-expenses", nil)
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		svc.On("ListRecurringExpenses", mock.Anything, mock.Anything).Return([]*dto.RecurringExpenseResponse{{ID: "re1"}}, dto.PaginationMeta{Total: 1}, nil).Once()

		h.ListRecurringExpenses(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("CreateRecurring Success", func(t *testing.T) {
		reqBody := dto.CreateRecurringExpenseRequest{
			CategoryID:    "550e8400-e29b-41d4-a716-446655440000",
			Name:          "Rent",
			Amount:        100,
			Interval:      "monthly",
			IntervalValue: 1,
			StartDate:     "2024-01-01",
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/stores/s1/recurring-expenses", bytes.NewBuffer(body))
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		svc.On("CreateRecurringExpense", mock.Anything, "s1", mock.Anything, mock.Anything).Return(&dto.RecurringExpenseResponse{ID: "re1"}, nil).Once()

		h.CreateRecurringExpense(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("CreateRecurring Validation Error", func(t *testing.T) {
		reqBodyEmpty := dto.CreateRecurringExpenseRequest{Amount: -10}
		bodyEmpty, _ := json.Marshal(reqBodyEmpty)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/stores/s1/recurring-expenses", bytes.NewBuffer(bodyEmpty))
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		h.CreateRecurringExpense(w, req)
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("UpdateRecurring Success", func(t *testing.T) {
		reqBody := dto.UpdateRecurringExpenseRequest{
			CategoryID:    "550e8400-e29b-41d4-a716-446655440000",
			Name:          "Rent Updated",
			Amount:        200,
			Interval:      "monthly",
			IntervalValue: 1,
			StartDate:     "2024-01-01",
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPut, "/stores/s1/recurring-expenses/re1", bytes.NewBuffer(body))
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		rctx.URLParams.Add("id", "re1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		svc.On("UpdateRecurringExpense", mock.Anything, "re1", "s1", mock.Anything).Return(&dto.RecurringExpenseResponse{ID: "re1"}, nil).Once()

		h.UpdateRecurringExpense(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("DeleteRecurring Success", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodDelete, "/stores/s1/recurring-expenses/re1", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		rctx.URLParams.Add("id", "re1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		svc.On("DeleteRecurringExpense", mock.Anything, "re1", "s1").Return(nil).Once()

		h.DeleteRecurringExpense(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestExpenseHandler_UpdateExpense(t *testing.T) {
	svc := svcMocks.NewExpenseServiceInterface(t)
	v := validator.New()
	h := NewExpenseHandler(svc, v, zerolog.Nop())

	t.Run("success", func(t *testing.T) {
		reqBody := dto.UpdateExpenseRequest{
			CategoryID:  "550e8400-e29b-41d4-a716-446655440000",
			Amount:      150,
			ExpenseDate: "2024-01-01",
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPut, "/stores/s1/expenses/e1", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		rctx.URLParams.Add("id", "e1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		svc.On("UpdateExpense", mock.Anything, "e1", "s1", mock.Anything).Return(&dto.ExpenseResponse{ID: "e1"}, nil).Once()

		h.UpdateExpense(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Validation Error", func(t *testing.T) {
		reqBodyEmpty := dto.UpdateExpenseRequest{Amount: -10}
		bodyEmpty, _ := json.Marshal(reqBodyEmpty)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPut, "/stores/s1/expenses/e1", bytes.NewBuffer(bodyEmpty))
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		rctx.URLParams.Add("id", "e1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		h.UpdateExpense(w, req)
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})
}
