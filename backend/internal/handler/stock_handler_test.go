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

	"github.com/moedahpos/backend/internal/dto"
	"github.com/moedahpos/backend/internal/service"
	"github.com/moedahpos/backend/internal/service/mocks"
	"github.com/moedahpos/backend/internal/validator"
)

func TestStockHandler_GetLevels(t *testing.T) {
	svc := new(mocks.StockServiceInterface)
	v := validator.New()
	log := zerolog.Nop()
	h := NewStockHandler(svc, v, log)

	t.Run("success", func(t *testing.T) {
		levels := []*dto.StockLevelResponse{{ProductID: "p1", ProductName: "P1"}}
		svc.On("GetStockLevels", mock.Anything, "store-1", false).Return(levels, nil).Once()

		req, _ := http.NewRequestWithContext(context.Background(), "GET", "/stores/store-1/stock", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "store-1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		h.GetLevels(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("service_error", func(t *testing.T) {
		svc.On("GetStockLevels", mock.Anything, "store-1", false).Return(nil, errors.New("error")).Once()

		req, _ := http.NewRequestWithContext(context.Background(), "GET", "/stores/store-1/stock", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "store-1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		h.GetLevels(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		svc.AssertExpectations(t)
	})
}

func TestStockHandler_Adjust(t *testing.T) {
	svc := new(mocks.StockServiceInterface)
	v := validator.New()
	log := zerolog.Nop()
	h := NewStockHandler(svc, v, log)

	productID := "550e8400-e29b-41d4-a716-446655440000"

	t.Run("success", func(t *testing.T) {
		reqBody := dto.AdjustStockRequest{
			ProductID: productID,
			Delta:     5.0,
		}
		body, _ := json.Marshal(reqBody)

		level := &dto.StockLevelResponse{ProductID: productID, Quantity: 15}
		svc.On("AdjustStock", mock.Anything, "store-1", mock.Anything, "").Return(level, nil).Once()

		req, _ := http.NewRequestWithContext(context.Background(), "POST", "/stores/store-1/stock/adjust", bytes.NewBuffer(body))
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "store-1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		h.Adjust(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("validation_error", func(t *testing.T) {
		reqBody := dto.AdjustStockRequest{
			ProductID: "invalid-uuid", // Invalid
			Delta:     5.0,
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequestWithContext(context.Background(), "POST", "/stores/store-1/stock/adjust", bytes.NewBuffer(body))
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "store-1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		h.Adjust(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("product_not_found", func(t *testing.T) {
		reqBody := dto.AdjustStockRequest{
			ProductID: productID,
			Delta:     5.0,
		}
		body, _ := json.Marshal(reqBody)

		svc.On("AdjustStock", mock.Anything, "store-1", mock.Anything, "").Return(nil, service.ErrProductNotInStore).Once()

		req, _ := http.NewRequestWithContext(context.Background(), "POST", "/stores/store-1/stock/adjust", bytes.NewBuffer(body))
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "store-1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		h.Adjust(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		svc.AssertExpectations(t)
	})
}

func TestStockHandler_GetProductStock(t *testing.T) {
	svc := new(mocks.StockServiceInterface)
	v := validator.New()
	log := zerolog.Nop()
	h := NewStockHandler(svc, v, log)

	t.Run("success", func(t *testing.T) {
		level := &dto.StockLevelResponse{ProductID: "p1", Quantity: 10}
		svc.On("GetProductStock", mock.Anything, "p1", "store-1").Return(level, nil).Once()

		req, _ := http.NewRequestWithContext(context.Background(), "GET", "/stores/store-1/stock/p1", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "store-1")
		rctx.URLParams.Add("productId", "p1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		h.GetProductStock(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("not_found", func(t *testing.T) {
		svc.On("GetProductStock", mock.Anything, "p1", "store-1").Return(nil, service.ErrStockLevelNotFound).Once()

		req, _ := http.NewRequestWithContext(context.Background(), "GET", "/stores/store-1/stock/p1", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "store-1")
		rctx.URLParams.Add("productId", "p1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		h.GetProductStock(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		svc.AssertExpectations(t)
	})
}
