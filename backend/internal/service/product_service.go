package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
	"github.com/moedahpos/backend/internal/repository"
)

// Product-specific sentinel errors.
var (
	ErrProductNotFound  = errors.New("product not found")
	ErrSKUAlreadyExists = errors.New("SKU is already used in this store")
	ErrCategoryNotFound = errors.New("category not found")
	ErrCategoryInUse    = errors.New("category still has active products")
)

// ProductService implements business logic for products and categories.
type ProductService struct {
	productRepo     repository.ProductRepository
	categoryRepo    repository.CategoryRepository
	stockRepo       repository.StockRepository
	priceHistorySvc *PriceHistoryService
	log             zerolog.Logger
}

func NewProductService(
	productRepo repository.ProductRepository,
	categoryRepo repository.CategoryRepository,
	stockRepo repository.StockRepository,
	priceHistorySvc *PriceHistoryService,
	log zerolog.Logger,
) *ProductService {
	return &ProductService{productRepo: productRepo, categoryRepo: categoryRepo, stockRepo: stockRepo, priceHistorySvc: priceHistorySvc, log: log}
}

// ─── Categories ───────────────────────────────────────────────────────────────

func (s *ProductService) ListCategories(ctx context.Context, storeID string) ([]*dto.CategoryResponse, error) {
	cats, err := s.categoryRepo.FindAllByStore(ctx, storeID)
	if err != nil {
		return nil, fmt.Errorf("listing categories: %w", err)
	}
	resp := make([]*dto.CategoryResponse, 0, len(cats))
	for _, c := range cats {
		resp = append(resp, toCategoryResponse(c))
	}
	return resp, nil
}

func (s *ProductService) CreateCategory(ctx context.Context, storeID string, req *dto.CreateCategoryRequest) (*dto.CategoryResponse, error) {
	cat, err := s.categoryRepo.Create(ctx, &domain.Category{
		StoreID: storeID, Name: req.Name, ParentID: req.ParentID,
	})
	if err != nil {
		return nil, fmt.Errorf("creating category: %w", err)
	}
	return toCategoryResponse(cat), nil
}

func (s *ProductService) UpdateCategory(ctx context.Context, id string, req *dto.UpdateCategoryRequest) (*dto.CategoryResponse, error) {
	cat, err := s.categoryRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("finding category: %w", err)
	}
	if cat == nil {
		return nil, ErrCategoryNotFound
	}
	cat.Name = req.Name
	cat.ParentID = req.ParentID
	updated, err := s.categoryRepo.Update(ctx, cat)
	if err != nil {
		return nil, fmt.Errorf("updating category: %w", err)
	}
	return toCategoryResponse(updated), nil
}

// DeleteCategory soft-deletes a category (sets deleted_at = NOW()).
func (s *ProductService) DeleteCategory(ctx context.Context, id string) error {
	cat, err := s.categoryRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("finding category: %w", err)
	}
	if cat == nil {
		return ErrCategoryNotFound
	}
	if err := s.categoryRepo.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("soft-deleting category: %w", err)
	}
	s.log.Info().Str("category_id", id).Msg("category soft-deleted")
	return nil
}

// ─── Products ─────────────────────────────────────────────────────────────────

func (s *ProductService) ListProducts(ctx context.Context, filter dto.ProductListFilter) ([]*dto.ProductResponse, dto.PaginationMeta, error) {
	filter.Defaults()
	filter.WithStock = true
	products, total, err := s.productRepo.FindAll(ctx, filter)
	if err != nil {
		return nil, dto.PaginationMeta{}, fmt.Errorf("listing products: %w", err)
	}
	resp := make([]*dto.ProductResponse, 0, len(products))
	for _, p := range products {
		resp = append(resp, toProductResponse(p))
	}
	return resp, dto.NewMeta(filter.PaginationQuery, total), nil
}

func (s *ProductService) GetProduct(ctx context.Context, id string) (*dto.ProductResponse, error) {
	p, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("finding product: %w", err)
	}
	if p == nil {
		return nil, ErrProductNotFound
	}
	return toProductResponse(p), nil
}

func (s *ProductService) GetProductByBarcode(ctx context.Context, storeID, barcode string) (*dto.ProductResponse, error) {
	p, err := s.productRepo.FindByBarcode(ctx, storeID, barcode)
	if err != nil {
		return nil, fmt.Errorf("finding product by barcode: %w", err)
	}
	if p == nil {
		return nil, ErrProductNotFound
	}
	return toProductResponse(p), nil
}

func (s *ProductService) CreateProduct(ctx context.Context, storeID string, req *dto.CreateProductRequest, createdByID string) (*dto.ProductResponse, error) {
	// SKU uniqueness check.
	exists, err := s.productRepo.ExistsBySKU(ctx, storeID, req.SKU, "")
	if err != nil {
		return nil, fmt.Errorf("checking SKU: %w", err)
	}
	if exists {
		return nil, ErrSKUAlreadyExists
	}

	useGlobalTax := true
	if req.UseGlobalTax != nil {
		useGlobalTax = *req.UseGlobalTax
	}

	product, err := s.productRepo.Create(ctx, &domain.Product{
		StoreID: storeID, CategoryID: req.CategoryID, SKU: req.SKU, Name: req.Name,
		Description: &req.Description, Barcode: req.Barcode, Unit: req.Unit,
		CostPrice: req.CostPrice, SellPrice: req.SellPrice, UseGlobalTax: useGlobalTax, TaxPercentage: req.TaxPercentage,
		ImageURL: req.ImageURL, IsActive: true,
	})
	if err != nil {
		return nil, fmt.Errorf("creating product: %w", err)
	}

	// Initialize stock level if initial quantity provided.
	if req.InitialQty != 0 {
		_, err = s.stockRepo.Adjust(ctx, domain.AdjustInput{
			ProductID: product.ID, StoreID: storeID,
			Delta: req.InitialQty, RefType: "adjustment",
			Notes: "Initial stock on product creation", CreatedBy: createdByID,
		})
		if err != nil {
			s.log.Warn().Err(err).Str("product_id", product.ID).Msg("failed to set initial stock")
		}
	}

	s.log.Info().Str("product_id", product.ID).Str("sku", product.SKU).Msg("product created")
	return toProductResponse(product), nil
}

func (s *ProductService) UpdateProduct(ctx context.Context, id string, req *dto.UpdateProductRequest, changedBy string) (*dto.ProductResponse, error) {
	product, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("finding product: %w", err)
	}
	if product == nil {
		return nil, ErrProductNotFound
	}

	// Snapshot old prices before mutation
	oldCost := product.CostPrice
	oldSell := product.SellPrice

	product.CategoryID = req.CategoryID
	product.Name = req.Name
	product.Description = &req.Description
	product.Barcode = req.Barcode
	product.Unit = req.Unit
	product.CostPrice = req.CostPrice
	product.SellPrice = req.SellPrice
	if req.UseGlobalTax != nil {
		product.UseGlobalTax = *req.UseGlobalTax
	} else {
		product.UseGlobalTax = true // default fallback
	}
	product.TaxPercentage = req.TaxPercentage
	product.ImageURL = req.ImageURL
	if req.IsActive != nil {
		product.IsActive = *req.IsActive
	}

	updated, err := s.productRepo.Update(ctx, product)
	if err != nil {
		return nil, fmt.Errorf("updating product: %w", err)
	}

	// Record price change if anything changed (non-blocking)
	if s.priceHistorySvc != nil {
		source := "manual"
		if err2 := s.priceHistorySvc.RecordChange(ctx,
			product.ID, product.StoreID, changedBy,
			oldCost, req.CostPrice, oldSell, req.SellPrice,
			source, nil, nil,
		); err2 != nil {
			s.log.Warn().Err(err2).Str("product_id", product.ID).Msg("failed to record price history")
		}
	}

	return toProductResponse(updated), nil
}

// DeleteProduct soft-deletes a product (sets deleted_at = NOW()).
func (s *ProductService) DeleteProduct(ctx context.Context, id string) error {
	product, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("finding product: %w", err)
	}
	if product == nil {
		return ErrProductNotFound
	}
	if err := s.productRepo.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("soft-deleting product: %w", err)
	}
	s.log.Info().Str("product_id", id).Msg("product soft-deleted")
	return nil
}

// ─── Mappers ──────────────────────────────────────────────────────────────────

func toCategoryResponse(c *domain.Category) *dto.CategoryResponse {
	r := &dto.CategoryResponse{
		ID:         c.ID,
		StoreID:    c.StoreID,
		Name:       c.Name,
		ParentID:   c.ParentID,
		ParentName: c.ParentName,
		CreatedAt:  c.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  c.UpdatedAt.Format(time.RFC3339),
	}
	if c.DeletedAt != nil {
		t := c.DeletedAt.Format(time.RFC3339)
		r.DeletedAt = &t
	}
	return r
}

func toProductResponse(p *domain.Product) *dto.ProductResponse {
	desc := ""
	if p.Description != nil {
		desc = *p.Description
	}
	r := &dto.ProductResponse{
		ID:            p.ID,
		StoreID:       p.StoreID,
		CategoryID:    p.CategoryID,
		CategoryName:  p.CategoryName,
		SKU:           p.SKU,
		Name:          p.Name,
		Description:   desc,
		Barcode:       p.Barcode,
		Unit:          p.Unit,
		CostPrice:     p.CostPrice,
		SellPrice:     p.SellPrice,
		UseGlobalTax:  p.UseGlobalTax,
		TaxPercentage: p.TaxPercentage,
		TaxRate:       p.EffectiveTaxRate(),
		ImageURL:      p.ImageURL,
		IsActive:      p.IsActive,
		StockQty:      p.StockQty,
		CreatedAt:     p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     p.UpdatedAt.Format(time.RFC3339),
	}
	if p.DeletedAt != nil {
		t := p.DeletedAt.Format(time.RFC3339)
		r.DeletedAt = &t
	}
	return r
}
