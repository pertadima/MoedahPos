package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/moedahpos/backend/internal/dto"
	"github.com/moedahpos/backend/internal/middleware"
	"github.com/moedahpos/backend/internal/service"
	"github.com/moedahpos/backend/internal/service/mocks"
	"github.com/moedahpos/backend/internal/validator"
	"github.com/moedahpos/backend/pkg/jwt"
)

func TestPurchaseOrderHandler_CRUD(t *testing.T) {
	poSvc := new(mocks.PurchaseOrderServiceInterface)
	v := validator.New()
	log := zerolog.Nop()
	h := NewPurchaseOrderHandler(poSvc, v, log)

	r := chi.NewRouter()
	jwtMgr := jwt.New("secret", time.Hour, time.Hour)
	token, _ := jwtMgr.GenerateAccessToken("u123", "test@test.com")
	authMiddleware := middleware.Authenticate(jwtMgr)
	r.Use(authMiddleware)

	r.Get("/stores/{storeId}/purchase-orders", h.List)
	r.Post("/stores/{storeId}/purchase-orders", h.Create)
	r.Get("/stores/{storeId}/purchase-orders/{poId}", h.Get)
	r.Put("/stores/{storeId}/purchase-orders/{poId}", h.Update)

	t.Run("List POs", func(t *testing.T) {
		poSvc.On("ListPOs", mock.Anything, mock.Anything).Return([]*dto.POResponse{{ID: "po1"}}, dto.PaginationMeta{Total: 1}, nil).Once()
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores/s1/purchase-orders", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Create PO", func(t *testing.T) {
		reqBody := dto.CreatePORequest{
			SupplierID: ptrToStringPtr("00000000-0000-0000-0000-000000000001"),
			Items: []dto.POItemInput{
				{ProductID: "00000000-0000-0000-0000-000000000010", Quantity: 10, UnitCost: 100},
			},
		}
		body, _ := json.Marshal(reqBody)
		poSvc.On("CreatePO", mock.Anything, "s1", mock.Anything, "u123").Return(&dto.POResponse{ID: "po1"}, nil).Once()

		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/stores/s1/purchase-orders", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("Get PO", func(t *testing.T) {
		poSvc.On("GetPO", mock.Anything, "po1").Return(&dto.POResponse{ID: "po1"}, nil).Once()
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores/s1/purchase-orders/po1", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Update PO", func(t *testing.T) {
		reqBody := dto.UpdatePORequest{
			Items: []dto.POItemInput{
				{ProductID: "00000000-0000-0000-0000-000000000010", Quantity: 20, UnitCost: 100},
			},
		}
		body, _ := json.Marshal(reqBody)
		poSvc.On("UpdatePO", mock.Anything, "po1", mock.Anything, "s1").Return(&dto.POResponse{ID: "po1"}, nil).Once()

		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPut, "/stores/s1/purchase-orders/po1", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Create PO Validation Error", func(t *testing.T) {
		reqBody := dto.CreatePORequest{}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/stores/s1/purchase-orders", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("Get PO Not Found", func(t *testing.T) {
		poSvc.On("GetPO", mock.Anything, "po-none").Return(nil, service.ErrPONotFound).Once()
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores/s1/purchase-orders/po-none", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Update PO Not Found", func(t *testing.T) {
		reqBody := dto.UpdatePORequest{Items: []dto.POItemInput{{ProductID: "00000000-0000-0000-0000-000000000010", Quantity: 1}}}
		body, _ := json.Marshal(reqBody)
		poSvc.On("UpdatePO", mock.Anything, "po-none", mock.Anything, "s1").Return(nil, service.ErrPONotFound).Once()

		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPut, "/stores/s1/purchase-orders/po-none", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Update PO Conflict", func(t *testing.T) {
		reqBody := dto.UpdatePORequest{Items: []dto.POItemInput{{ProductID: "00000000-0000-0000-0000-000000000010", Quantity: 1}}}
		body, _ := json.Marshal(reqBody)
		poSvc.On("UpdatePO", mock.Anything, "po1", mock.Anything, "s1").Return(nil, service.ErrPONotEditable).Once()

		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPut, "/stores/s1/purchase-orders/po1", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusConflict, w.Code)
	})
}

func TestPurchaseOrderHandler_Status(t *testing.T) {
	poSvc := new(mocks.PurchaseOrderServiceInterface)
	h := NewPurchaseOrderHandler(poSvc, validator.New(), zerolog.Nop())
	r := chi.NewRouter()

	jwtMgr := jwt.New("secret", time.Hour, time.Hour)
	token, _ := jwtMgr.GenerateAccessToken("u123", "test@test.com")
	r.Use(middleware.Authenticate(jwtMgr))

	r.Post("/stores/{storeId}/purchase-orders/{poId}/submit", h.Submit)
	r.Post("/stores/{storeId}/purchase-orders/{poId}/receive", h.Receive)
	r.Delete("/stores/{storeId}/purchase-orders/{poId}", h.Cancel)
	r.Get("/stores/{storeId}/purchase-orders/payables", h.PayableSummary)
	r.Post("/stores/{storeId}/purchase-orders/{poId}/payments", h.CreatePayment)
	r.Get("/stores/{storeId}/purchase-orders/{poId}/payments", h.ListPayments)

	t.Run("Submit PO", func(t *testing.T) {
		poSvc.On("SubmitPO", mock.Anything, "po1", "u123").Return(nil).Once()
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/stores/s1/purchase-orders/po1/submit", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Receive PO", func(t *testing.T) {
		poSvc.On("ReceivePO", mock.Anything, "po1", "u123").Return(nil).Once()
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/stores/s1/purchase-orders/po1/receive", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Cancel PO", func(t *testing.T) {
		poSvc.On("CancelPO", mock.Anything, "po1").Return(nil).Once()
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodDelete, "/stores/s1/purchase-orders/po1", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Payable Summary", func(t *testing.T) {
		poSvc.On("PayableSummary", mock.Anything, "s1").Return(&dto.PayableSummary{TotalOutstanding: 5000}, nil).Once()
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores/s1/purchase-orders/payables", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Create Payment", func(t *testing.T) {
		reqBody := dto.POPaymentRequest{Amount: 1000}
		body, _ := json.Marshal(reqBody)
		poSvc.On("CreatePayment", mock.Anything, "po1", "s1", "u123", reqBody).Return(&dto.POPaymentResponse{ID: "pay1"}, nil).Once()
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/stores/s1/purchase-orders/po1/payments", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("List Payments", func(t *testing.T) {
		poSvc.On("ListPayments", mock.Anything, "po1").Return([]*dto.POPaymentResponse{{ID: "pay1"}}, nil).Once()
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores/s1/purchase-orders/po1/payments", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestPurchaseOrderHandler_StatusErrors(t *testing.T) {
	poSvc := new(mocks.PurchaseOrderServiceInterface)
	h := NewPurchaseOrderHandler(poSvc, validator.New(), zerolog.Nop())
	r := chi.NewRouter()
	jwtMgr := jwt.New("secret", time.Hour, time.Hour)
	token, _ := jwtMgr.GenerateAccessToken("u123", "test@test.com")
	r.Use(middleware.Authenticate(jwtMgr))

	r.Post("/stores/{storeId}/purchase-orders/{poId}/submit", h.Submit)
	r.Post("/stores/{storeId}/purchase-orders/{poId}/receive", h.Receive)
	r.Delete("/stores/{storeId}/purchase-orders/{poId}", h.Cancel)
	r.Post("/stores/{storeId}/purchase-orders/{poId}/payments", h.CreatePayment)

	t.Run("Submit PO Conflict", func(t *testing.T) {
		poSvc.On("SubmitPO", mock.Anything, "po1", "u123").Return(service.ErrPOCannotSubmit).Once()
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/stores/s1/purchase-orders/po1/submit", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("Receive PO Conflict", func(t *testing.T) {
		poSvc.On("ReceivePO", mock.Anything, "po1", "u123").Return(service.ErrPOCannotReceive).Once()
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/stores/s1/purchase-orders/po1/receive", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("Cancel PO Conflict", func(t *testing.T) {
		poSvc.On("CancelPO", mock.Anything, "po1").Return(service.ErrPOCannotCancel).Once()
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodDelete, "/stores/s1/purchase-orders/po1", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("Create Payment Not Found", func(t *testing.T) {
		reqBody := dto.POPaymentRequest{Amount: 1000}
		body, _ := json.Marshal(reqBody)
		poSvc.On("CreatePayment", mock.Anything, "po-none", "s1", "u123", reqBody).Return(nil, service.ErrPONotFound).Once()
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/stores/s1/purchase-orders/po-none/payments", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Submit PO Internal Error", func(t *testing.T) {
		poSvc.On("SubmitPO", mock.Anything, "po1", "u123").Return(errors.New("db error")).Once()
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/stores/s1/purchase-orders/po1/submit", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func ptrToStringPtr(s string) *string {
	return &s
}
