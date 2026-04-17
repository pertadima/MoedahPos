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

		req := httptest.NewRequest("GET", "/income-categories?include_deleted=false", nil).WithContext(context.Background()) // nolint:noctx
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

		req := httptest.NewRequest("POST", "/stores/"+storeID+"/incomes", bytes.NewBuffer(body)).WithContext(context.Background()) // nolint:noctx
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", storeID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		h.CreateIncome(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})
}
