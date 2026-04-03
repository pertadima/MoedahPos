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

// Supplier sentinel errors.
var (
	ErrSupplierNotFound = errors.New("supplier not found")
)

// SupplierService implements business logic for supplier management.
type SupplierService struct {
	supplierRepo repository.SupplierRepository
	log          zerolog.Logger
}

func NewSupplierService(supplierRepo repository.SupplierRepository, log zerolog.Logger) *SupplierService {
	return &SupplierService{supplierRepo: supplierRepo, log: log}
}

func (s *SupplierService) ListSuppliers(ctx context.Context, filter dto.SupplierListFilter) ([]*dto.SupplierResponse, dto.PaginationMeta, error) {
	filter.Defaults()
	suppliers, total, err := s.supplierRepo.FindAll(ctx, filter)
	if err != nil {
		return nil, dto.PaginationMeta{}, fmt.Errorf("listing suppliers: %w", err)
	}
	resp := make([]*dto.SupplierResponse, 0, len(suppliers))
	for _, s := range suppliers {
		resp = append(resp, toSupplierResponse(s))
	}
	return resp, dto.NewMeta(filter.PaginationQuery, total), nil
}

func (s *SupplierService) GetSupplier(ctx context.Context, id string) (*dto.SupplierResponse, error) {
	supplier, err := s.supplierRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("finding supplier: %w", err)
	}
	if supplier == nil {
		return nil, ErrSupplierNotFound
	}
	return toSupplierResponse(supplier), nil
}

func (s *SupplierService) CreateSupplier(ctx context.Context, req *dto.CreateSupplierRequest) (*dto.SupplierResponse, error) {
	supplier, err := s.supplierRepo.Create(ctx, &domain.Supplier{
		Name:        req.Name,
		ContactName: req.ContactName,
		Phone:       req.Phone,
		Email:       req.Email,
		Address:     req.Address,
		IsActive:    true,
	})
	if err != nil {
		return nil, fmt.Errorf("creating supplier: %w", err)
	}
	s.log.Info().Str("supplier_id", supplier.ID).Msg("supplier created")
	return toSupplierResponse(supplier), nil
}

func (s *SupplierService) UpdateSupplier(ctx context.Context, id string, req *dto.UpdateSupplierRequest) (*dto.SupplierResponse, error) {
	supplier, err := s.supplierRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("finding supplier: %w", err)
	}
	if supplier == nil {
		return nil, ErrSupplierNotFound
	}
	supplier.Name = req.Name
	supplier.ContactName = req.ContactName
	supplier.Phone = req.Phone
	supplier.Email = req.Email
	supplier.Address = req.Address
	if req.IsActive != nil {
		supplier.IsActive = *req.IsActive
	}
	updated, err := s.supplierRepo.Update(ctx, supplier)
	if err != nil {
		return nil, fmt.Errorf("updating supplier: %w", err)
	}
	return toSupplierResponse(updated), nil
}

// DeleteSupplier soft-deletes a supplier (sets deleted_at = NOW()).
func (s *SupplierService) DeleteSupplier(ctx context.Context, id string) error {
	supplier, err := s.supplierRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("finding supplier: %w", err)
	}
	if supplier == nil {
		return ErrSupplierNotFound
	}
	if err := s.supplierRepo.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("soft-deleting supplier: %w", err)
	}
	s.log.Info().Str("supplier_id", id).Msg("supplier soft-deleted")
	return nil
}

func toSupplierResponse(s *domain.Supplier) *dto.SupplierResponse {
	r := &dto.SupplierResponse{
		ID:          s.ID,
		Name:        s.Name,
		ContactName: s.ContactName,
		Phone:       s.Phone,
		Email:       s.Email,
		Address:     s.Address,
		IsActive:    s.IsActive,
		CreatedAt:   s.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   s.UpdatedAt.Format(time.RFC3339),
	}
	if s.DeletedAt != nil {
		t := s.DeletedAt.Format(time.RFC3339)
		r.DeletedAt = &t
	}
	return r
}
