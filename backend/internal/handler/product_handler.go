package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/moedahpos/backend/internal/dto"
	"github.com/moedahpos/backend/internal/service"
	"github.com/moedahpos/backend/internal/validator"
	"github.com/moedahpos/backend/pkg/response"
)

// ProductHandler handles HTTP requests for categories and products.
type ProductHandler struct {
	productSvc service.ProductServiceInterface
	validator  *validator.Validator
	log        zerolog.Logger
}

func NewProductHandler(productSvc service.ProductServiceInterface, v *validator.Validator, log zerolog.Logger) *ProductHandler {
	return &ProductHandler{productSvc: productSvc, validator: v, log: log}
}

// ─── Categories ───────────────────────────────────────────────────────────────

// GET /stores/:storeId/categories
func (h *ProductHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	cats, err := h.productSvc.ListCategories(r.Context(), storeID)
	if err != nil {
		h.log.Error().Err(err).Msg("list categories failed")
		response.InternalError(w)
		return
	}
	response.Success(w, cats)
}

// POST /stores/:storeId/categories
func (h *ProductHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	var req dto.CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	if errs := h.validator.ValidateStruct(req); errs != nil {
		response.ValidationError(w, errs)
		return
	}
	result, err := h.productSvc.CreateCategory(r.Context(), storeID, &req)
	if err != nil {
		response.InternalError(w)
		return
	}
	response.Created(w, result)
}

// PUT /stores/:storeId/categories/:categoryId
func (h *ProductHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "categoryId")
	var req dto.UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	if errs := h.validator.ValidateStruct(req); errs != nil {
		response.ValidationError(w, errs)
		return
	}
	result, err := h.productSvc.UpdateCategory(r.Context(), id, &req)
	if err != nil {
		if errors.Is(err, service.ErrCategoryNotFound) {
			response.NotFound(w, "Category")
			return
		}
		response.InternalError(w)
		return
	}
	response.Success(w, result)
}

// DELETE /stores/:storeId/categories/:categoryId  (soft delete)
func (h *ProductHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "categoryId")
	if err := h.productSvc.DeleteCategory(r.Context(), id); err != nil {
		if errors.Is(err, service.ErrCategoryNotFound) {
			response.NotFound(w, "Category")
			return
		}
		response.InternalError(w)
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true, "message": "Category deleted successfully",
	})
}

// ─── Products ─────────────────────────────────────────────────────────────────

// GET /stores/:storeId/products
func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	filter := dto.ProductListFilter{StoreID: storeID}
	filter.Page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	filter.PerPage, _ = strconv.Atoi(r.URL.Query().Get("per_page"))
	filter.Search = r.URL.Query().Get("search")
	filter.CategoryID = r.URL.Query().Get("category_id")
	if v := r.URL.Query().Get("is_active"); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			filter.IsActive = &b
		}
	}

	products, meta, err := h.productSvc.ListProducts(r.Context(), filter)
	if err != nil {
		h.log.Error().Err(err).Msg("list products failed")
		response.InternalError(w)
		return
	}
	response.Success(w, dto.ListResponse{Data: products, Meta: meta})
}

// POST /stores/:storeId/products
func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	var req dto.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	if errs := h.validator.ValidateStruct(req); errs != nil {
		response.ValidationError(w, errs)
		return
	}
	result, err := h.productSvc.CreateProduct(r.Context(), storeID, &req, userIDFromCtx(r))
	if err != nil {
		if errors.Is(err, service.ErrSKUAlreadyExists) {
			response.ValidationError(w, []dto.FieldError{{Field: "sku", Message: "SKU is already used in this store"}})
			return
		}
		h.log.Error().Err(err).Msg("create product failed")
		response.InternalError(w)
		return
	}
	response.Created(w, result)
}

// GET /stores/:storeId/products/:productId
func (h *ProductHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "productId")
	result, err := h.productSvc.GetProduct(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			response.NotFound(w, "Product")
			return
		}
		response.InternalError(w)
		return
	}
	response.Success(w, result)
}

// GET /stores/:storeId/products/barcode/:barcode
func (h *ProductHandler) GetByBarcode(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	barcode := chi.URLParam(r, "barcode")
	result, err := h.productSvc.GetProductByBarcode(r.Context(), storeID, barcode)
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			response.NotFound(w, "Product")
			return
		}
		response.InternalError(w)
		return
	}
	response.Success(w, result)
}

// PUT /stores/:storeId/products/:productId
func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "productId")
	var req dto.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	if errs := h.validator.ValidateStruct(req); errs != nil {
		response.ValidationError(w, errs)
		return
	}
	result, err := h.productSvc.UpdateProduct(r.Context(), id, &req, userIDFromCtx(r))
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			response.NotFound(w, "Product")
			return
		}
		response.InternalError(w)
		return
	}
	response.Success(w, result)
}

// DELETE /stores/:storeId/products/:productId  (soft delete)
func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "productId")
	if err := h.productSvc.DeleteProduct(r.Context(), id); err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			response.NotFound(w, "Product")
			return
		}
		response.InternalError(w)
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true, "message": "Product deleted successfully",
	})
}
