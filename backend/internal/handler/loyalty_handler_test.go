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
	svcMocks "github.com/moedahpos/backend/internal/service/mocks"
	"github.com/moedahpos/backend/internal/validator"
)

func TestLoyaltyHandler(t *testing.T) {
	svc := new(svcMocks.LoyaltyServiceInterface)
	v := validator.New()
	h := NewLoyaltyHandler(svc, v, zerolog.Nop())

	t.Run("ListTiers Success", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/loyalty/tiers", nil)
		svc.On("ListTiers", mock.Anything).Return([]*dto.MembershipTierResponse{{ID: "t1"}}, nil).Once()

		w := httptest.NewRecorder()
		h.ListTiers(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("GetBalance Success", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores/s1/customers/c1/loyalty/balance", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("customerId", "c1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		svc.On("GetBalance", mock.Anything, "c1").Return(&dto.LoyaltyBalanceResponse{Balance: 100}, nil).Once()

		w := httptest.NewRecorder()
		h.GetBalance(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("EarnPoints Success", func(t *testing.T) {
		reqBody := dto.EarnPointsRequest{TransactionID: "550e8400-e29b-41d4-a716-446655440000", Total: 1000}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/stores/s1/customers/c1/loyalty/earn", bytes.NewBuffer(body))
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		rctx.URLParams.Add("customerId", "c1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		svc.On("EarnPoints", mock.Anything, "s1", "c1", mock.Anything, 1000.0).Return(&dto.LoyaltyLedgerResponse{ID: "l1"}, nil).Once()

		w := httptest.NewRecorder()
		h.EarnPoints(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("GetHistory Success", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores/s1/customers/c1/loyalty/history", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("customerId", "c1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		svc.On("GetHistory", mock.Anything, "c1").Return([]*dto.LoyaltyLedgerResponse{{PointsDelta: 10}}, nil).Once()

		w := httptest.NewRecorder()
		h.GetHistory(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("AssignTier Success", func(t *testing.T) {
		reqBody := dto.AssignTierRequest{TierID: "550e8400-e29b-41d4-a716-446655440000"}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/stores/s1/customers/c1/loyalty/tier", bytes.NewBuffer(body))
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("customerId", "c1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		svc.On("AssignTier", mock.Anything, "c1", "550e8400-e29b-41d4-a716-446655440000").Return(nil).Once()

		w := httptest.NewRecorder()
		h.AssignTier(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("EarnPoints Validation Error", func(t *testing.T) {
		reqBody := dto.EarnPointsRequest{Total: -10}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/stores/s1/customers/c1/loyalty/earn", bytes.NewBuffer(body))
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		rctx.URLParams.Add("customerId", "c1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		svc.On("EarnPoints", mock.Anything, "s1", "c1", mock.Anything, -10.0).Return(nil, service.ErrInsufficientPoints).Once()

		w := httptest.NewRecorder()
		h.EarnPoints(w, req)
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("RedeemPoints Validation Error", func(t *testing.T) {
		reqBody := dto.RedeemPointsRequest{Points: 100}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/stores/s1/customers/c1/loyalty/redeem", bytes.NewBuffer(body))
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("customerId", "c1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		svc.On("RedeemPoints", mock.Anything, "c1", mock.Anything, 100.0).Return(nil, service.ErrInvalidRedemption).Once()

		w := httptest.NewRecorder()
		h.RedeemPoints(w, req)
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})
}
