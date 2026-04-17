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

func TestExpenseHandler_ListCategories(t *testing.T) {
	svc := svcMocks.NewExpenseServiceInterface(t)
	v := validator.New()
	h := NewExpenseHandler(svc, v, zerolog.Nop())

	t.Run("success", func(t *testing.T) {
		svc.On("ListCategories", mock.Anything, false).
			Return([]dto.ExpenseCategoryResponse{{ID: "cat1", Name: "Food"}}, nil).Once()

		req := httptest.NewRequest("GET", "/expense-categories?include_deleted=false", nil).WithContext(context.Background()) // nolint:noctx
		w := httptest.NewRecorder()

		h.ListCategories(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestExpenseHandler_CreateExpense(t *testing.T) {
	svc := svcMocks.NewExpenseServiceInterface(t)
	v := validator.New()
	h := NewExpenseHandler(svc, v, zerolog.Nop())

	storeID := "st1"
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

		req := httptest.NewRequest("POST", "/stores/"+storeID+"/expenses", bytes.NewBuffer(body)).WithContext(context.Background()) // nolint:noctx
		w := httptest.NewRecorder()

		// Set storeId param
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", storeID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		h.CreateExpense(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})
}

func TestExpenseHandler_DeleteExpense(t *testing.T) {
	svc := svcMocks.NewExpenseServiceInterface(t)
	v := validator.New()
	h := NewExpenseHandler(svc, v, zerolog.Nop())

	storeID := "st1"
	id := "e1"

	t.Run("success", func(t *testing.T) {
		svc.On("DeleteExpense", mock.Anything, id, storeID).Return(nil).Once()

		req := httptest.NewRequest("DELETE", "/stores/"+storeID+"/expenses/"+id, nil).WithContext(context.Background()) // nolint:noctx
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", storeID)
		rctx.URLParams.Add("id", id)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		h.DeleteExpense(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
