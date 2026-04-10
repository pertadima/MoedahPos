package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
)

var (
	ErrTableNotFound    = errors.New("table not found")
	ErrTableNumberTaken = errors.New("table number already in use")
	ErrMenuItemNotFound = errors.New("menu item not found")
)

// ─── Interfaces ───────────────────────────────────────────────────────────────

type tableRepo interface {
	FindAllByStore(ctx context.Context, storeID string) ([]*domain.RestaurantTable, error)
	FindByID(ctx context.Context, id string) (*domain.RestaurantTable, error)
	Create(ctx context.Context, t *domain.RestaurantTable) (*domain.RestaurantTable, error)
	Update(ctx context.Context, t *domain.RestaurantTable) (*domain.RestaurantTable, error)
	UpdateStatus(ctx context.Context, id string, status domain.TableStatus) error
	SoftDelete(ctx context.Context, id string) error
}

type menuItemRepo interface {
	FindAllByStore(ctx context.Context, storeID string) ([]*domain.MenuItem, error)
	FindByID(ctx context.Context, id string) (*domain.MenuItem, error)
	Create(ctx context.Context, item *domain.MenuItem) (*domain.MenuItem, error)
	Update(ctx context.Context, item *domain.MenuItem) (*domain.MenuItem, error)
	ReplaceIngredients(ctx context.Context, menuItemID string, ings []domain.MenuItemIngredient) error
	SoftDelete(ctx context.Context, id string) error
}

// ─── Table Service ────────────────────────────────────────────────────────────

type TableService struct {
	repo tableRepo
	log  zerolog.Logger
}

func NewTableService(repo tableRepo, log zerolog.Logger) *TableService {
	return &TableService{repo: repo, log: log}
}

func (s *TableService) List(ctx context.Context, storeID string) ([]*dto.TableResponse, error) {
	tables, err := s.repo.FindAllByStore(ctx, storeID)
	if err != nil {
		return nil, fmt.Errorf("listing tables: %w", err)
	}
	resp := make([]*dto.TableResponse, len(tables))
	for i, t := range tables {
		resp[i] = toTableResponse(t)
	}
	return resp, nil
}

func (s *TableService) Create(ctx context.Context, storeID string, req *dto.CreateTableRequest) (*dto.TableResponse, error) {
	t := &domain.RestaurantTable{
		StoreID:     storeID,
		TableNumber: req.TableNumber,
		Capacity:    req.Capacity,
		Status:      domain.TableAvailable,
		Notes:       req.Notes,
	}
	created, err := s.repo.Create(ctx, t)
	if err != nil {
		return nil, fmt.Errorf("creating table: %w", err)
	}
	s.log.Info().Str("table_id", created.ID).Str("number", created.TableNumber).Msg("table created")
	return toTableResponse(created), nil
}

func (s *TableService) Update(ctx context.Context, id string, req *dto.UpdateTableRequest) (*dto.TableResponse, error) {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("finding table: %w", err)
	}
	if existing == nil {
		return nil, ErrTableNotFound
	}
	existing.TableNumber = req.TableNumber
	existing.Capacity = req.Capacity
	existing.Notes = req.Notes
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}
	updated, err := s.repo.Update(ctx, existing)
	if err != nil {
		return nil, fmt.Errorf("updating table: %w", err)
	}
	return toTableResponse(updated), nil
}

func (s *TableService) UpdateStatus(ctx context.Context, id string, status domain.TableStatus) error {
	if err := s.repo.UpdateStatus(ctx, id, status); err != nil {
		return fmt.Errorf("updating table status: %w", err)
	}
	return nil
}

func (s *TableService) Delete(ctx context.Context, id string) error {
	if err := s.repo.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("deleting table: %w", err)
	}
	return nil
}

// ─── Menu Item Service ────────────────────────────────────────────────────────

type MenuItemService struct {
	repo menuItemRepo
	log  zerolog.Logger
}

func NewMenuItemService(repo menuItemRepo, log zerolog.Logger) *MenuItemService {
	return &MenuItemService{repo: repo, log: log}
}

func (s *MenuItemService) List(ctx context.Context, storeID string) ([]*dto.MenuItemResponse, error) {
	items, err := s.repo.FindAllByStore(ctx, storeID)
	if err != nil {
		return nil, fmt.Errorf("listing menu items: %w", err)
	}
	resp := make([]*dto.MenuItemResponse, len(items))
	for i, item := range items {
		resp[i] = toMenuItemResponse(item)
	}
	return resp, nil
}

func (s *MenuItemService) Create(ctx context.Context, storeID string, req *dto.CreateMenuItemRequest) (*dto.MenuItemResponse, error) {
	desc := req.Description
	item := &domain.MenuItem{
		StoreID:     storeID,
		CategoryID:  req.CategoryID,
		Name:        req.Name,
		Description: &desc,
		SellPrice:   req.SellPrice,
		TaxRate:     req.TaxRate,
	}
	created, err := s.repo.Create(ctx, item)
	if err != nil {
		return nil, fmt.Errorf("creating menu item: %w", err)
	}

	// Set ingredients
	if len(req.Ingredients) > 0 {
		ings := make([]domain.MenuItemIngredient, len(req.Ingredients))
		for i, ing := range req.Ingredients {
			ings[i] = domain.MenuItemIngredient{
				MenuItemID: created.ID,
				ProductID:  ing.ProductID,
				Quantity:   ing.Quantity,
			}
		}
		if err := s.repo.ReplaceIngredients(ctx, created.ID, ings); err != nil {
			return nil, fmt.Errorf("setting ingredients: %w", err)
		}
	}

	full, _ := s.repo.FindByID(ctx, created.ID)
	s.log.Info().Str("menu_item_id", created.ID).Msg("menu item created")
	return toMenuItemResponse(full), nil
}

func (s *MenuItemService) Update(ctx context.Context, id string, req *dto.UpdateMenuItemRequest) (*dto.MenuItemResponse, error) {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("finding menu item: %w", err)
	}
	if existing == nil {
		return nil, ErrMenuItemNotFound
	}

	desc := req.Description
	existing.CategoryID = req.CategoryID
	existing.Name = req.Name
	existing.Description = &desc
	existing.SellPrice = req.SellPrice
	existing.TaxRate = req.TaxRate
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	updated, err := s.repo.Update(ctx, existing)
	if err != nil {
		return nil, fmt.Errorf("updating menu item: %w", err)
	}

	// Replace ingredients
	ings := make([]domain.MenuItemIngredient, len(req.Ingredients))
	for i, ing := range req.Ingredients {
		ings[i] = domain.MenuItemIngredient{
			MenuItemID: updated.ID,
			ProductID:  ing.ProductID,
			Quantity:   ing.Quantity,
		}
	}
	if err := s.repo.ReplaceIngredients(ctx, updated.ID, ings); err != nil {
		return nil, fmt.Errorf("replacing ingredients: %w", err)
	}

	full, _ := s.repo.FindByID(ctx, updated.ID)
	return toMenuItemResponse(full), nil
}

func (s *MenuItemService) Delete(ctx context.Context, id string) error {
	if err := s.repo.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("deleting menu item: %w", err)
	}
	return nil
}

// ─── Mappers ─────────────────────────────────────────────────────────────────

func toTableResponse(t *domain.RestaurantTable) *dto.TableResponse {
	r := &dto.TableResponse{
		ID:          t.ID,
		StoreID:     t.StoreID,
		TableNumber: t.TableNumber,
		Capacity:    t.Capacity,
		Status:      string(t.Status),
		IsActive:    t.IsActive,
		CreatedAt:   t.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   t.UpdatedAt.Format(time.RFC3339),
	}
	if t.Notes != nil {
		r.Notes = *t.Notes
	}
	return r
}

func toMenuItemResponse(item *domain.MenuItem) *dto.MenuItemResponse {
	if item == nil {
		return nil
	}
	desc := ""
	if item.Description != nil {
		desc = *item.Description
	}

	// Compute BOM cost (HPP) from ingredients
	var costPrice float64
	ings := make([]dto.IngredientResponse, len(item.Ingredients))
	for i, ing := range item.Ingredients {
		costPrice += ing.CostPrice * ing.Quantity
		ings[i] = dto.IngredientResponse{
			ID:          ing.ID,
			ProductID:   ing.ProductID,
			ProductName: ing.ProductName,
			ProductSKU:  ing.ProductSKU,
			Unit:        ing.Unit,
			Quantity:    ing.Quantity,
			CostPrice:   ing.CostPrice,
		}
	}

	r := &dto.MenuItemResponse{
		ID:           item.ID,
		StoreID:      item.StoreID,
		CategoryID:   item.CategoryID,
		CategoryName: item.CategoryName,
		Name:         item.Name,
		Description:  desc,
		SellPrice:    item.SellPrice,
		CostPrice:    costPrice,
		TaxRate:      item.TaxRate,
		IsActive:     item.IsActive,
		Ingredients:  ings,
		CreatedAt:    item.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    item.UpdatedAt.Format(time.RFC3339),
	}
	if item.ImageURL != nil {
		r.ImageURL = *item.ImageURL
	}
	return r
}
