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

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/middleware"
	"github.com/moedahpos/backend/internal/service/mocks"
	"github.com/moedahpos/backend/internal/validator"
)

func TestStockAdjustmentHandler(t *testing.T) {
	svc := new(mocks.StockAdjustmentServiceInterface)
	v := validator.New()
	log := zerolog.Nop()
	h := NewStockAdjustmentHandler(svc, v, &log)

	t.Run("Create_Success", func(t *testing.T) {
		reqBody := domain.CreateAdjustmentInput{
			ProductID: "p1",
			Quantity:  5,
			Type:      "IN",
			Reason:    "MANUAL_CORRECTION",
			Notes:     "Correction",
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/stores/s1/stock/adjust", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, "u1"))

		svc.On("CreateAdjustment", mock.Anything, "s1", "u1", reqBody).Return(nil).Once()

		h.Create(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("GetHistory", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores/s1/stock/adjustments", nil)
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		svc.On("GetAdjustmentHistory", mock.Anything, "s1", mock.Anything).Return([]*domain.StockAdjustment{{ID: "a1"}}, nil).Once()

		h.GetHistory(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Create_Validation_Error", func(t *testing.T) {
		reqBody := domain.CreateAdjustmentInput{
			ProductID: "", // Invalid
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/stores/s1/stock/adjust", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, "u1"))

		h.Create(w, req)
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})
}
