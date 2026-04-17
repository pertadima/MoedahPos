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
	"github.com/moedahpos/backend/internal/service"
	"github.com/moedahpos/backend/internal/service/mocks"
	"github.com/moedahpos/backend/internal/validator"
)

func TestTransactionHandler_Checkout(t *testing.T) {
	svc := new(mocks.TransactionServiceInterface)
	v := validator.New()
	h := NewTransactionHandler(svc, v, zerolog.Nop())

	t.Run("Success", func(t *testing.T) {
		reqBody := dto.CreateTransactionRequest{
			Items:         []dto.TxItemInput{{ProductID: "00000000-0000-0000-0000-000000000001", Quantity: 1}},
			PaymentAmount: 100,
			PaymentMethod: "cash",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/stores/s1/transactions", bytes.NewBuffer(body))
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		svc.On("Checkout", mock.Anything, "s1", mock.Anything, "").Return(&dto.TransactionResponse{ID: "t1"}, nil).Once()

		w := httptest.NewRecorder()
		h.Checkout(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("Insufficient Payment", func(t *testing.T) {
		reqBody := dto.CreateTransactionRequest{
			Items:         []dto.TxItemInput{{ProductID: "p1", Quantity: 1}},
			PaymentAmount: 10,
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/stores/s1/transactions", bytes.NewBuffer(body))

		svc.On("Checkout", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, service.ErrInsuficientPayment)

		w := httptest.NewRecorder()
		h.Checkout(w, req)
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})
}

func TestTransactionHandler_Get(t *testing.T) {
	svc := new(mocks.TransactionServiceInterface)
	v := validator.New()
	h := NewTransactionHandler(svc, v, zerolog.Nop())

	t.Run("Success", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores/s1/transactions/t1", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("txnId", "t1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		svc.On("GetTransaction", mock.Anything, "t1").Return(&dto.TransactionResponse{ID: "t1"}, nil).Once()

		w := httptest.NewRecorder()
		h.Get(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestTransactionHandler_Void(t *testing.T) {
	svc := new(mocks.TransactionServiceInterface)
	v := validator.New()
	h := NewTransactionHandler(svc, v, zerolog.Nop())

	t.Run("Success", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/stores/s1/transactions/t1/void", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("txnId", "t1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		svc.On("VoidTransaction", mock.Anything, "t1", "").Return(nil).Once()

		w := httptest.NewRecorder()
		h.Void(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
