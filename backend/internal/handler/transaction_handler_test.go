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

func TestTransactionHandler_List(t *testing.T) {
	svc := new(mocks.TransactionServiceInterface)
	h := NewTransactionHandler(svc, validator.New(), zerolog.Nop())

	t.Run("Success", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores/s1/transactions?page=1&per_page=10", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		svc.On("ListTransactions", mock.Anything, mock.Anything).Return([]*dto.TransactionResponse{}, dto.PaginationMeta{}, nil).Once()

		w := httptest.NewRecorder()
		h.List(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestTransactionHandler_Draft(t *testing.T) {
	svc := new(mocks.TransactionServiceInterface)
	h := NewTransactionHandler(svc, validator.New(), zerolog.Nop())

	t.Run("CreateDraft", func(t *testing.T) {
		reqBody := dto.CreateDraftRequest{
			TableID: "00000000-0000-0000-0000-000000000001",
			Items:   []dto.TxItemInput{{ProductID: "00000000-0000-0000-0000-000000000001", Quantity: 1}},
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/stores/s1/transactions/draft", bytes.NewBuffer(body))
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		svc.On("CreateDraft", mock.Anything, "s1", "", mock.Anything).Return(&dto.TransactionResponse{ID: "d1"}, nil).Once()

		w := httptest.NewRecorder()
		h.CreateDraft(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("GetDraftByTable", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores/s1/transactions/draft?table_id=tab1", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		svc.On("GetDraftByTable", mock.Anything, "s1", "tab1").Return(&dto.TransactionResponse{ID: "d1"}, nil).Once()

		w := httptest.NewRecorder()
		h.GetDraftByTable(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("UpdateDraft", func(t *testing.T) {
		reqBody := dto.UpdateDraftRequest{
			Items: []dto.TxItemInput{{ProductID: "00000000-0000-0000-0000-000000000001", Quantity: 2}},
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/stores/s1/transactions/draft/d1", bytes.NewBuffer(body))
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		rctx.URLParams.Add("txnId", "d1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		svc.On("UpdateDraftItems", mock.Anything, "s1", "d1", mock.Anything).Return(&dto.TransactionResponse{ID: "d1"}, nil).Once()

		w := httptest.NewRecorder()
		h.UpdateDraft(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("PayDraft", func(t *testing.T) {
		reqBody := dto.PayDraftRequest{
			PaymentMethod: "cash",
			PaymentAmount: 1000,
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/stores/s1/transactions/draft/d1/pay", bytes.NewBuffer(body))
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		rctx.URLParams.Add("txnId", "d1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		svc.On("PayDraft", mock.Anything, "s1", "d1", "", mock.Anything).Return(&dto.TransactionResponse{ID: "d1"}, nil).Once()

		w := httptest.NewRecorder()
		h.PayDraft(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
	})
}

func TestTransactionHandler_KDS(t *testing.T) {
	svc := new(mocks.TransactionServiceInterface)
	h := NewTransactionHandler(svc, validator.New(), zerolog.Nop())

	t.Run("GetTickets", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores/s1/kds/tickets", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		svc.On("GetKDSTickets", mock.Anything, "s1").Return([]*dto.TransactionResponse{}, nil).Once()

		w := httptest.NewRecorder()
		h.GetKDSTickets(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("UpdateItemStatus", func(t *testing.T) {
		reqBody := dto.UpdateKDSItemStatusRequest{Status: "completed"}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPut, "/stores/s1/kds/items/item1", bytes.NewBuffer(body))
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("itemId", "item1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		svc.On("UpdateKDSItemStatus", mock.Anything, "item1", mock.Anything).Return(nil).Once()

		w := httptest.NewRecorder()
		h.UpdateKDSItemStatus(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
