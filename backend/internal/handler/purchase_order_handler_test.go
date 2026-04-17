package handler

import (
	"bytes"
	"context"
	"encoding/json"
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

	t.Run("List POs", func(t *testing.T) {
		poSvc.On("ListPOs", mock.Anything, mock.Anything).Return([]*dto.POResponse{{ID: "po1"}}, dto.PaginationMeta{Total: 1}, nil)
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
		poSvc.On("CreatePO", mock.Anything, "s1", mock.Anything, "u123").Return(&dto.POResponse{ID: "po1"}, nil)

		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/stores/s1/purchase-orders", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
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

	t.Run("Submit PO", func(t *testing.T) {
		poSvc.On("SubmitPO", mock.Anything, "po1", "u123").Return(nil)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/stores/s1/purchase-orders/po1/submit", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Receive PO", func(t *testing.T) {
		poSvc.On("ReceivePO", mock.Anything, "po1", "u123").Return(nil)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/stores/s1/purchase-orders/po1/receive", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Cancel PO", func(t *testing.T) {
		poSvc.On("CancelPO", mock.Anything, "po1").Return(nil)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodDelete, "/stores/s1/purchase-orders/po1", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		// For DELETE, we need to register it in the router for ServeHTTP to pick it up
		r.Delete("/stores/{storeId}/purchase-orders/{poId}", h.Cancel)

		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Payable Summary", func(t *testing.T) {
		poSvc.On("PayableSummary", mock.Anything, "s1").Return(&dto.PayableSummary{TotalOutstanding: 5000}, nil)
		r.Get("/stores/{storeId}/purchase-orders/payables", h.PayableSummary)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores/s1/purchase-orders/payables", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Create Payment", func(t *testing.T) {
		reqBody := dto.POPaymentRequest{Amount: 1000}
		body, _ := json.Marshal(reqBody)
		poSvc.On("CreatePayment", mock.Anything, "po1", "s1", "u123", reqBody).Return(&dto.POPaymentResponse{ID: "pay1"}, nil)
		r.Post("/stores/{storeId}/purchase-orders/{poId}/payments", h.CreatePayment)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/stores/s1/purchase-orders/po1/payments", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
	})
}

func ptrToStringPtr(s string) *string {
	return &s
}
