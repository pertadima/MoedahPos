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
func TestProductHandler_Categories(t *testing.T) {
	t.Run("CreateCategory", func(t *testing.T) {
		pSvc := new(mocks.ProductServiceInterface)
		h := NewProductHandler(pSvc, validator.New(), zerolog.Nop())

		reqBody := dto.CreateCategoryRequest{Name: "New Cat"}
		body, _ := json.Marshal(reqBody)

		r := chi.NewRouter()
		r.Post("/stores/{storeId}/categories", h.CreateCategory)

		pSvc.On("CreateCategory", mock.Anything, "s1", &reqBody).Return(&dto.CategoryResponse{ID: "c1", Name: "New Cat"}, nil)

		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/stores/s1/categories", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("UpdateCategory", func(t *testing.T) {
		pSvc := new(mocks.ProductServiceInterface)
		h := NewProductHandler(pSvc, validator.New(), zerolog.Nop())

		reqBody := dto.UpdateCategoryRequest{Name: "Updated Cat"}
		body, _ := json.Marshal(reqBody)

		r := chi.NewRouter()
		r.Put("/stores/{storeId}/categories/{categoryId}", h.UpdateCategory)

		pSvc.On("UpdateCategory", mock.Anything, "c1", &reqBody).Return(&dto.CategoryResponse{ID: "c1", Name: "Updated Cat"}, nil)

		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPut, "/stores/s1/categories/c1", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("DeleteCategory", func(t *testing.T) {
		pSvc := new(mocks.ProductServiceInterface)
		h := NewProductHandler(pSvc, validator.New(), zerolog.Nop())

		r := chi.NewRouter()
		r.Delete("/stores/{storeId}/categories/{categoryId}", h.DeleteCategory)

		pSvc.On("DeleteCategory", mock.Anything, "c1").Return(nil)

		req, _ := http.NewRequestWithContext(context.Background(), http.MethodDelete, "/stores/s1/categories/c1", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestProductHandler_Products(t *testing.T) {
	pSvc := new(mocks.ProductServiceInterface)
	h := NewProductHandler(pSvc, validator.New(), zerolog.Nop())

	t.Run("List", func(t *testing.T) {
		r := chi.NewRouter()
		r.Get("/stores/{storeId}/products", h.List)

		pSvc.On("ListProducts", mock.Anything, mock.Anything).Return([]*dto.ProductResponse{}, dto.PaginationMeta{}, nil).Once()

		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores/s1/products?page=1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Get", func(t *testing.T) {
		r := chi.NewRouter()
		r.Get("/stores/{storeId}/products/{productId}", h.Get)

		pSvc.On("GetProduct", mock.Anything, "p1").Return(&dto.ProductResponse{ID: "p1"}, nil).Once()

		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores/s1/products/p1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("GetByBarcode", func(t *testing.T) {
		r := chi.NewRouter()
		r.Get("/stores/{storeId}/products/barcode/{barcode}", h.GetByBarcode)

		barcode := "123456"
		pSvc.On("GetProductByBarcode", mock.Anything, "s1", "123456").Return(&dto.ProductResponse{ID: "p1", Barcode: &barcode}, nil).Once()

		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/stores/s1/products/barcode/123456", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Update", func(t *testing.T) {
		reqBody := dto.UpdateProductRequest{Name: "New Name", Unit: "pcs"}
		body, _ := json.Marshal(reqBody)
		r := chi.NewRouter()
		r.Put("/stores/{storeId}/products/{productId}", h.Update)

		pSvc.On("UpdateProduct", mock.Anything, "p1", mock.Anything, mock.Anything).Return(&dto.ProductResponse{ID: "p1"}, nil).Once()

		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPut, "/stores/s1/products/p1", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Delete", func(t *testing.T) {
		r := chi.NewRouter()
		r.Delete("/stores/{storeId}/products/{productId}", h.Delete)

		pSvc.On("DeleteProduct", mock.Anything, "p1").Return(nil).Once()

		req, _ := http.NewRequestWithContext(context.Background(), http.MethodDelete, "/stores/s1/products/p1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
