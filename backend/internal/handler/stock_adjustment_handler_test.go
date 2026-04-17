package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/middleware"
	"github.com/moedahpos/backend/internal/service/mocks"
	"github.com/moedahpos/backend/internal/validator"
)

func TestStockAdjustmentHandler_Create(t *testing.T) {
	svc := new(mocks.StockAdjustmentServiceInterface)
	v := validator.New()
	log := zerolog.Nop()
	h := NewStockAdjustmentHandler(svc, v, &log)

	t.Run("success", func(t *testing.T) {
		reqBody := domain.CreateAdjustmentInput{
			ProductID: "p1",
			Type:      "IN",
			Reason:    "MANUAL_CORRECTION",
			Quantity:  10,
			Notes:     "Correction",
		}
		body, _ := json.Marshal(reqBody)

		svc.On("CreateAdjustment", mock.Anything, "store-1", "user-1", reqBody).Return(nil).Once()

		req, _ := http.NewRequestWithContext(context.Background(), "POST", "/stores/store-1/stock/adjustments", bytes.NewBuffer(body))
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "store-1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, "user-1"))

		w := httptest.NewRecorder()
		h.Create(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("validation_error", func(t *testing.T) {
		reqBody := domain.CreateAdjustmentInput{
			ProductID: "", // Invalid
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequestWithContext(context.Background(), "POST", "/stores/store-1/stock/adjustments", bytes.NewBuffer(body))
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "store-1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, "user-1"))

		w := httptest.NewRecorder()
		h.Create(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("service_error", func(t *testing.T) {
		reqBody := domain.CreateAdjustmentInput{
			ProductID: "p1",
			Type:      "IN",
			Reason:    "MANUAL_CORRECTION",
			Quantity:  10,
			Notes:     "Correction",
		}
		body, _ := json.Marshal(reqBody)

		svc.On("CreateAdjustment", mock.Anything, "store-1", "user-1", reqBody).Return(errors.New("service error")).Once()

		req, _ := http.NewRequestWithContext(context.Background(), "POST", "/stores/store-1/stock/adjustments", bytes.NewBuffer(body))
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "store-1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, "user-1"))

		w := httptest.NewRecorder()
		h.Create(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		svc.AssertExpectations(t)
	})
}

func TestStockAdjustmentHandler_GetHistory(t *testing.T) {
	svc := new(mocks.StockAdjustmentServiceInterface)
	v := validator.New()
	log := zerolog.Nop()
	h := NewStockAdjustmentHandler(svc, v, &log)

	t.Run("success", func(t *testing.T) {
		svc.On("GetAdjustmentHistory", mock.Anything, "store-1", mock.Anything).Return([]*domain.StockAdjustment{}, nil).Once()

		req, _ := http.NewRequestWithContext(context.Background(), "GET", "/stores/store-1/stock/adjustments", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "store-1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		h.GetHistory(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})
}
