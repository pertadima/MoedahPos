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
	"github.com/moedahpos/backend/internal/service"
	"github.com/moedahpos/backend/internal/service/mocks"
	"github.com/moedahpos/backend/internal/validator"
	"github.com/moedahpos/backend/pkg/jwt"
)

//nolint:funlen
func TestTerminHandler(t *testing.T) {
	tSvc := new(mocks.TerminServiceInterface)
	v := validator.New()
	log := zerolog.Nop()
	h := NewTerminHandler(tSvc, v, log)

	jwtMgr := jwt.New("secret", time.Hour, time.Hour)
	token, _ := jwtMgr.GenerateAccessToken("u123", "admin@test.com")
	authMiddleware := middleware.Authenticate(jwtMgr)

	r := chi.NewRouter()
	r.Use(authMiddleware)
	r.Post("/stores/{storeId}/purchase-orders/{poId}/termins", h.CreateSchedule)
	r.Get("/stores/{storeId}/purchase-orders/{poId}/termins", h.ListTermins)
	r.Post("/stores/{storeId}/purchase-orders/{poId}/termins/{terminId}/payments", h.RecordPayment)
	r.Get("/stores/{storeId}/purchase-orders/{poId}/debt", h.GetDebtSummary)
	r.Get("/stores/{storeId}/purchase-orders/{poId}/document", h.GetDocument)

	t.Run("Create Schedule Success", func(t *testing.T) {
		reqBody := dto.CreateTerminScheduleRequest{
			Termins: []dto.TerminInput{{TerminNumber: 1, Amount: 100, DueDate: "2026-01-01"}},
		}
		body, _ := json.Marshal(reqBody)

		tSvc.On("CreateTerminSchedule", mock.Anything, "po1", mock.Anything).Return([]dto.TerminResponse{{ID: "t1"}}, nil)

		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/stores/s1/purchase-orders/po1/termins", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("Record Payment Success", func(t *testing.T) {
		reqBody := dto.RecordPaymentRequest{AmountPaid: 50, PaymentDate: "2026-01-01", PaymentMethod: "cash"}
		body, _ := json.Marshal(reqBody)

		tSvc.On("RecordPayment", mock.Anything, "t1", "u123", mock.Anything).Return(&dto.PaymentRecordResponse{ID: "pr1"}, nil)

		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/stores/s1/purchase-orders/po1/termins/t1/payments", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("Get Debt Summary", func(t *testing.T) {
		tSvc.On("CalculatePODebt", mock.Anything, "po1").Return(&dto.PODebtSummaryResponse{RemainingDebt: 100}, nil)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores/s1/purchase-orders/po1/debt", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("List Termins Success", func(t *testing.T) {
		tSvc.On("GetTerminSchedule", mock.Anything, "po1").Return([]dto.TerminResponse{{ID: "t1"}}, nil)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores/s1/purchase-orders/po1/termins", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Get Document Success", func(t *testing.T) {
		tSvc.On("GenerateDocumentData", mock.Anything, "po1", "invoice").Return(&dto.PODocumentData{DocType: "invoice"}, nil)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores/s1/purchase-orders/po1/document", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("List Termins Error", func(t *testing.T) {
		tSvc.On("GetTerminSchedule", mock.Anything, "po2").Return(nil, service.ErrPONotFound)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores/s1/purchase-orders/po2/termins", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
