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

func TestTransactionHandler_Checkout(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		tSvc := new(mocks.TransactionServiceInterface)
		v := validator.New()
		log := zerolog.Nop()
		h := NewTransactionHandler(tSvc, v, log)

		// Setup JWT/Middleware for userID
		jwtMgr := jwt.New("secret", time.Hour, time.Hour)
		token, _ := jwtMgr.GenerateAccessToken("u123", "cashier@test.com")
		authMiddleware := middleware.Authenticate(jwtMgr)

		pid := "00000000-0000-0000-0000-000000000001"
		reqBody := dto.CreateTransactionRequest{
			Items:         []dto.TxItemInput{{ProductID: pid, Quantity: 1}},
			PaymentAmount: 100,
			PaymentMethod: "cash",
		}
		body, _ := json.Marshal(reqBody)

		// Router with URL param
		r := chi.NewRouter()
		r.Use(authMiddleware)
		r.Post("/stores/{storeId}/transactions", h.Checkout)

		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/stores/s1/transactions", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		tSvc.On("Checkout", mock.Anything, "s1", &reqBody, "u123").Return(&dto.TransactionResponse{ID: "t1"}, nil)

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		tSvc.AssertExpectations(t)
	})
}

func TestTransactionHandler_Get(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		tSvc := new(mocks.TransactionServiceInterface)
		h := NewTransactionHandler(tSvc, validator.New(), zerolog.Nop())

		r := chi.NewRouter()
		r.Get("/stores/{storeId}/transactions/{txnId}", h.Get)

		tSvc.On("GetTransaction", mock.Anything, "t1").Return(&dto.TransactionResponse{ID: "t1"}, nil)

		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores/s1/transactions/t1", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestTransactionHandler_Void(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		tSvc := new(mocks.TransactionServiceInterface)
		h := NewTransactionHandler(tSvc, validator.New(), zerolog.Nop())

		jwtMgr := jwt.New("secret", time.Hour, time.Hour)
		token, _ := jwtMgr.GenerateAccessToken("u123", "admin@test.com")
		authMiddleware := middleware.Authenticate(jwtMgr)

		r := chi.NewRouter()
		r.Use(authMiddleware)
		r.Post("/stores/{storeId}/transactions/{txnId}/void", h.Void)

		tSvc.On("VoidTransaction", mock.Anything, "t1", "u123").Return(nil)

		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/stores/s1/transactions/t1/void", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
