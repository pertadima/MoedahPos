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

func TestProductHandler_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		pSvc := new(mocks.ProductServiceInterface)
		h := NewProductHandler(pSvc, validator.New(), zerolog.Nop())

		jwtMgr := jwt.New("secret", time.Hour, time.Hour)
		token, _ := jwtMgr.GenerateAccessToken("u1", "admin@test.com")
		authMiddleware := middleware.Authenticate(jwtMgr)

		reqBody := dto.CreateProductRequest{
			SKU:       "SKU1",
			Name:      "Product 1",
			SellPrice: 100,
			Unit:      "pcs",
		}
		body, _ := json.Marshal(reqBody)

		r := chi.NewRouter()
		r.Use(authMiddleware)
		r.Post("/stores/{storeId}/products", h.Create)

		pSvc.On("CreateProduct", mock.Anything, "s1", &reqBody, "u1").Return(&dto.ProductResponse{ID: "p1"}, nil)

		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/stores/s1/products", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		pSvc.AssertExpectations(t)
	})
}

func TestProductHandler_ListCategories(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		pSvc := new(mocks.ProductServiceInterface)
		h := NewProductHandler(pSvc, validator.New(), zerolog.Nop())

		r := chi.NewRouter()
		r.Get("/stores/{storeId}/categories", h.ListCategories)

		pSvc.On("ListCategories", mock.Anything, "s1").Return([]*dto.CategoryResponse{{ID: "c1", Name: "Cat 1"}}, nil)

		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores/s1/categories", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
