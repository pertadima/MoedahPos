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
	svcMocks "github.com/moedahpos/backend/internal/service/mocks"
	"github.com/moedahpos/backend/internal/validator"
)

func TestSupplierHandler_Create(t *testing.T) {
	svc := svcMocks.NewSupplierServiceInterface(t)
	v := validator.New()
	h := NewSupplierHandler(svc, v, zerolog.Nop())

	reqBody := dto.CreateSupplierRequest{
		Name: "Supplier A",
	}
	body, _ := json.Marshal(reqBody)

	t.Run("success", func(t *testing.T) {
		svc.On("CreateSupplier", mock.Anything, mock.AnythingOfType("*dto.CreateSupplierRequest")).
			Return(&dto.SupplierResponse{ID: "s1", Name: "Supplier A"}, nil).Once()

		req := httptest.NewRequest("POST", "/suppliers", bytes.NewBuffer(body)).WithContext(context.Background()) // nolint:noctx
		w := httptest.NewRecorder()

		h.Create(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var resp map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "s1", resp["data"].(map[string]interface{})["id"])
	})

	t.Run("validation error", func(t *testing.T) {
		reqBodyEmpty := dto.CreateSupplierRequest{Name: ""}
		bodyEmpty, _ := json.Marshal(reqBodyEmpty)
		req := httptest.NewRequest("POST", "/suppliers", bytes.NewBuffer(bodyEmpty)).WithContext(context.Background()) // nolint:noctx
		w := httptest.NewRecorder()

		h.Create(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})
}

func TestSupplierHandler_List(t *testing.T) {
	svc := svcMocks.NewSupplierServiceInterface(t)
	v := validator.New()
	h := NewSupplierHandler(svc, v, zerolog.Nop())

	t.Run("success", func(t *testing.T) {
		svc.On("ListSuppliers", mock.Anything, mock.AnythingOfType("dto.SupplierListFilter")).
			Return([]*dto.SupplierResponse{{ID: "s1", Name: "S1"}}, dto.PaginationMeta{Total: 1}, nil).Once()

		req := httptest.NewRequest("GET", "/suppliers?page=1&per_page=10", nil).WithContext(context.Background()) // nolint:noctx
		w := httptest.NewRecorder()

		h.List(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestSupplierHandler_Get(t *testing.T) {
	svc := svcMocks.NewSupplierServiceInterface(t)
	v := validator.New()
	h := NewSupplierHandler(svc, v, zerolog.Nop())

	t.Run("success", func(t *testing.T) {
		svc.On("GetSupplier", mock.Anything, "s1").
			Return(&dto.SupplierResponse{ID: "s1", Name: "S1"}, nil).Once()

		req := httptest.NewRequest("GET", "/suppliers/s1", nil).WithContext(context.Background()) // nolint:noctx
		w := httptest.NewRecorder()

		// Set URL param
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("supplierId", "s1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		h.Get(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
