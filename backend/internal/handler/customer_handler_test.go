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

const (
	customerTestStoreID = "store-1"
	testCustomerID      = "550e8400-e29b-41d4-a716-446655440000"
)

func TestCustomerHandler_List(t *testing.T) {
	svc := new(mocks.CustomerServiceInterface)
	v := validator.New()
	log := zerolog.Nop()
	h := NewCustomerHandler(svc, v, log)

	t.Run("success", func(t *testing.T) {
		customers := []*dto.CustomerResponse{{ID: "c1", Name: "Customer 1"}}
		svc.On("List", mock.Anything, mock.Anything).Return(customers, dto.PaginationMeta{Total: 1}, nil).Once()

		req, _ := http.NewRequestWithContext(context.Background(), "GET", "/stores/"+customerTestStoreID+"/customers", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", customerTestStoreID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		h.List(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})
}

func TestCustomerHandler_Create(t *testing.T) {
	svc := new(mocks.CustomerServiceInterface)
	v := validator.New()
	log := zerolog.Nop()
	h := NewCustomerHandler(svc, v, log)

	t.Run("success", func(t *testing.T) {
		reqBody := dto.CreateCustomerRequest{Name: "New Customer"}
		body, _ := json.Marshal(reqBody)

		svc.On("Create", mock.Anything, customerTestStoreID, mock.Anything).Return(&dto.CustomerResponse{ID: "550e8400-e29b-41d4-a716-446655440001", Name: "New Customer"}, nil).Once()

		req, _ := http.NewRequestWithContext(context.Background(), "POST", "/stores/"+customerTestStoreID+"/customers", bytes.NewBuffer(body))
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", customerTestStoreID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		h.Create(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("validation_error", func(t *testing.T) {
		reqBody := dto.CreateCustomerRequest{Name: ""} // Invalid
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequestWithContext(context.Background(), "POST", "/stores/"+customerTestStoreID+"/customers", bytes.NewBuffer(body))
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", customerTestStoreID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		h.Create(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})
}

func TestCustomerHandler_Get(t *testing.T) {
	svc := new(mocks.CustomerServiceInterface)
	v := validator.New()
	log := zerolog.Nop()
	h := NewCustomerHandler(svc, v, log)

	customerID := testCustomerID

	t.Run("success", func(t *testing.T) {
		svc.On("Get", mock.Anything, customerID).Return(&dto.CustomerResponse{ID: customerID, Name: "Customer 1"}, nil).Once()

		req, _ := http.NewRequestWithContext(context.Background(), "GET", "/stores/"+customerTestStoreID+"/customers/"+customerID, nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("customerId", customerID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		h.Get(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("not_found", func(t *testing.T) {
		svc.On("Get", mock.Anything, customerID).Return(nil, service.ErrCustomerNotFound).Once()

		req, _ := http.NewRequestWithContext(context.Background(), "GET", "/stores/"+customerTestStoreID+"/customers/"+customerID, nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("customerId", customerID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		h.Get(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		svc.AssertExpectations(t)
	})
}

func TestCustomerHandler_Delete(t *testing.T) {
	svc := new(mocks.CustomerServiceInterface)
	v := validator.New()
	log := zerolog.Nop()
	h := NewCustomerHandler(svc, v, log)

	customerID := testCustomerID

	t.Run("success", func(t *testing.T) {
		svc.On("Delete", mock.Anything, customerID).Return(nil).Once()

		req, _ := http.NewRequestWithContext(context.Background(), "DELETE", "/stores/"+customerTestStoreID+"/customers/"+customerID, nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("customerId", customerID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		h.Delete(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("service_error", func(t *testing.T) {
		svc.On("Delete", mock.Anything, customerID).Return(errors.New("error")).Once()

		req, _ := http.NewRequestWithContext(context.Background(), "DELETE", "/stores/"+customerTestStoreID+"/customers/"+customerID, nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("customerId", customerID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		h.Delete(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		svc.AssertExpectations(t)
	})
}

func TestCustomerHandler_Update(t *testing.T) {
	svc := new(mocks.CustomerServiceInterface)
	v := validator.New()
	log := zerolog.Nop()
	h := NewCustomerHandler(svc, v, log)

	customerID := testCustomerID

	t.Run("success", func(t *testing.T) {
		reqBody := dto.UpdateCustomerRequest{Name: "Updated Name"}
		body, _ := json.Marshal(reqBody)

		svc.On("Update", mock.Anything, customerID, mock.Anything).Return(&dto.CustomerResponse{ID: customerID, Name: "Updated Name"}, nil).Once()

		req, _ := http.NewRequestWithContext(context.Background(), "PUT", "/stores/"+customerTestStoreID+"/customers/"+customerID, bytes.NewBuffer(body))
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("customerId", customerID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		h.Update(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})
}

func TestCustomerHandler_Search(t *testing.T) {
	svc := new(mocks.CustomerServiceInterface)
	v := validator.New()
	log := zerolog.Nop()
	h := NewCustomerHandler(svc, v, log)

	t.Run("success", func(t *testing.T) {
		svc.On("Search", mock.Anything, mock.Anything).Return([]*dto.CustomerResponse{{ID: "c1", Name: "Customer 1"}}, nil).Once()

		req, _ := http.NewRequestWithContext(context.Background(), "GET", "/stores/store-1/customers/search?q=test", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", customerTestStoreID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		h.Search(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})
}
