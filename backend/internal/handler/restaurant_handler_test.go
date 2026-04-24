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

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
	"github.com/moedahpos/backend/internal/handler/mocks"
	"github.com/moedahpos/backend/internal/validator"
)

func TestRestaurantHandler(t *testing.T) {
	tableSvc := new(mocks.TableServiceInterface)
	menuSvc := new(mocks.MenuItemServiceInterface)
	v := validator.New()
	log := zerolog.Nop()
	h := NewRestaurantHandler(tableSvc, menuSvc, v, log)

	t.Run("ListTables", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores/s1/tables", nil)
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		tableSvc.On("List", mock.Anything, "s1").Return([]*dto.TableResponse{{ID: "t1"}}, nil).Once()

		h.ListTables(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		tableSvc.AssertExpectations(t)
	})

	t.Run("CreateTable", func(t *testing.T) {
		reqBody := dto.CreateTableRequest{TableNumber: "5", Capacity: 6}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/stores/s1/tables", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		tableSvc.On("Create", mock.Anything, "s1", mock.Anything).Return(&dto.TableResponse{ID: "t5"}, nil).Once()

		h.CreateTable(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("UpdateTable", func(t *testing.T) {
		reqBody := dto.UpdateTableRequest{TableNumber: "5B", Capacity: 4}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPut, "/stores/s1/tables/t1", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		rctx.URLParams.Add("tableId", "t1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		tableSvc.On("Update", mock.Anything, "t1", mock.Anything).Return(&dto.TableResponse{ID: "t1"}, nil).Once()

		h.UpdateTable(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("UpdateTableStatus", func(t *testing.T) {
		reqBody := dto.UpdateTableStatusRequest{Status: "occupied"}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPut, "/stores/s1/tables/t1/status", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		rctx.URLParams.Add("tableId", "t1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		tableSvc.On("UpdateStatus", mock.Anything, "t1", domain.TableStatus("occupied")).Return(nil).Once()

		h.UpdateTableStatus(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("DeleteTable", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodDelete, "/stores/s1/tables/t1", nil)
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		rctx.URLParams.Add("tableId", "t1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		tableSvc.On("Delete", mock.Anything, "t1").Return(nil).Once()

		h.DeleteTable(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("ListMenuItems", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores/s1/menu-items", nil)
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		menuSvc.On("List", mock.Anything, "s1").Return([]*dto.MenuItemResponse{{ID: "m1"}}, nil).Once()

		h.ListMenuItems(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("CreateMenuItem", func(t *testing.T) {
		reqBody := dto.CreateMenuItemRequest{Name: "Menu Item 1", SellPrice: 10000}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/stores/s1/menu-items", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		menuSvc.On("Create", mock.Anything, "s1", mock.Anything).Return(&dto.MenuItemResponse{ID: "m1"}, nil).Once()

		h.CreateMenuItem(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("UpdateMenuItem", func(t *testing.T) {
		reqBody := dto.UpdateMenuItemRequest{Name: "Updated Menu Item", SellPrice: 12000}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPut, "/stores/s1/menu-items/m1", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		rctx.URLParams.Add("menuItemId", "m1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		menuSvc.On("Update", mock.Anything, "m1", mock.Anything).Return(&dto.MenuItemResponse{ID: "m1"}, nil).Once()

		h.UpdateMenuItem(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("DeleteMenuItem", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodDelete, "/stores/s1/menu-items/m1", nil)
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		rctx.URLParams.Add("menuItemId", "m1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		menuSvc.On("Delete", mock.Anything, "m1").Return(nil).Once()

		h.DeleteMenuItem(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("CreateTable Validation Error", func(t *testing.T) {
		reqBody := dto.CreateTableRequest{TableNumber: ""}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/stores/s1/tables", bytes.NewBuffer(body))
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("storeId", "s1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		h.CreateTable(w, req)
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})
}
