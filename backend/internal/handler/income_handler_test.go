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

func TestIncomeHandler_ListCategories(t *testing.T) {
	svc := svcMocks.NewIncomeServiceInterface(t)
	v := validator.New()
	h := NewIncomeHandler(svc, v, zerolog.Nop())

	t.Run("success", func(t *testing.T) {
		svc.On("ListCategories", mock.Anything, false).
			Return([]*dto.IncomeCategoryResponse{{ID: "cat1", Name: "Service"}}, nil).Once()

		req := httptest.NewRequestWithContext(context.Background(), "GET", "/income-categories?include_deleted=false", nil)
		w := httptest.NewRecorder()

		h.ListCategories(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestIncomeHandler_CreateIncome(t *testing.T) {
	svc := svcMocks.NewIncomeServiceInterface(t)
	v := validator.New()
	h := NewIncomeHandler(svc, v, zerolog.Nop())

	storeID := "st1"
	reqBody := dto.CreateIncomeRequest{
		CategoryID:    "550e8400-e29b-41d4-a716-446655440000",
		Amount:        500.0,
		IncomeDate:    "2026-04-17",
		PaymentMethod: "cash",
	}
	body, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		svc.On("CreateIncome", mock.Anything, storeID, mock.Anything, mock.AnythingOfType("*dto.CreateIncomeRequest")).
			Return(&dto.IncomeResponse{ID: "inc1"}, nil).Once()

		req := httptest.NewRequestWithContext(context.Background(), "POST", "/stores/"+storeID+"/incomes", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", storeID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		h.CreateIncome(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})
}

func TestIncomeHandler_List(t *testing.T) {
	svc := svcMocks.NewIncomeServiceInterface(t)
	v := validator.New()
	h := NewIncomeHandler(svc, v, zerolog.Nop())

	t.Run("success", func(t *testing.T) {
		svc.On("ListIncomes", mock.Anything, mock.Anything).Return([]*dto.IncomeResponse{{ID: "inc1"}}, dto.PaginationMeta{Total: 1}, nil).Once()

		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores/s1/incomes", nil)
		w := httptest.NewRecorder()
		h.ListIncomes(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestIncomeHandler_Update(t *testing.T) {
	svc := svcMocks.NewIncomeServiceInterface(t)
	v := validator.New()
	h := NewIncomeHandler(svc, v, zerolog.Nop())

	t.Run("success", func(t *testing.T) {
		catID := "550e8400-e29b-41d4-a716-446655440000"
		reqBody := dto.UpdateIncomeRequest{CategoryID: catID, Amount: 200, IncomeDate: "2024-01-01", PaymentMethod: "cash"}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPut, "/stores/s1/incomes/inc1", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		rctx.URLParams.Add("id", "inc1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		svc.On("UpdateIncome", mock.Anything, "inc1", "s1", mock.Anything).Return(&dto.IncomeResponse{ID: "inc1"}, nil).Once()

		h.UpdateIncome(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestIncomeHandler_Delete(t *testing.T) {
	svc := svcMocks.NewIncomeServiceInterface(t)
	v := validator.New()
	h := NewIncomeHandler(svc, v, zerolog.Nop())

	t.Run("success", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodDelete, "/stores/s1/incomes/inc1", nil)
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		rctx.URLParams.Add("id", "inc1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		svc.On("DeleteIncome", mock.Anything, "inc1", "s1").Return(nil).Once()

		h.DeleteIncome(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
