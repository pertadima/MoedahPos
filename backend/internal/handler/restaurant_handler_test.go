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
}
