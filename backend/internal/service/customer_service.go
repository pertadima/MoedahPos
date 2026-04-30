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

var ErrCustomerNotFound = errors.New("customer not found")

// CustomerService handles business logic for customer master data.
type CustomerService struct {
	repo repository.CustomerRepository
	log  zerolog.Logger
}

func NewCustomerService(repo repository.CustomerRepository, log zerolog.Logger) *CustomerService {
	return &CustomerService{repo: repo, log: log}
}

func (s *CustomerService) List(ctx context.Context, f dto.CustomerListFilter) ([]*dto.CustomerResponse, dto.PaginationMeta, error) {
	f.Defaults()
	rows, total, err := s.repo.FindAll(ctx, f)
	if err != nil {
		return nil, dto.PaginationMeta{}, fmt.Errorf("listing customers: %w", err)
	}
	resp := make([]*dto.CustomerResponse, 0, len(rows))
	for _, c := range rows {
		resp = append(resp, toCustomerResponse(c))
	}
	return resp, dto.NewMeta(f.PaginationQuery, total), nil
}

func (s *CustomerService) Get(ctx context.Context, id string) (*dto.CustomerResponse, error) {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("finding customer: %w", err)
	}
	if c == nil {
		return nil, ErrCustomerNotFound
	}
	return toCustomerResponse(c), nil
}

func (s *CustomerService) Create(ctx context.Context, storeID string, req dto.CreateCustomerRequest) (*dto.CustomerResponse, error) {
	c := &domain.Customer{
		StoreID: storeID,
		Name:    req.Name,
	}
	if req.Phone != "" {
		c.Phone = &req.Phone
	}
	if req.Email != "" {
		c.Email = &req.Email
	}
	if req.Address != "" {
		c.Address = &req.Address
	}
	if req.Notes != "" {
		c.Notes = &req.Notes
	}

	out, err := s.repo.Create(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("creating customer: %w", err)
	}
	s.log.Info().Str("id", out.ID).Str("name", out.Name).Msg("customer created")
	return toCustomerResponse(out), nil
}

func (s *CustomerService) Update(ctx context.Context, id string, req dto.UpdateCustomerRequest) (*dto.CustomerResponse, error) {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("finding customer: %w", err)
	}
	if existing == nil {
		return nil, ErrCustomerNotFound
	}
	existing.Name = req.Name
	if req.Phone != "" {
		existing.Phone = &req.Phone
	} else {
		existing.Phone = nil
	}
	if req.Email != "" {
		existing.Email = &req.Email
	} else {
		existing.Email = nil
	}
	if req.Address != "" {
		existing.Address = &req.Address
	} else {
		existing.Address = nil
	}
	if req.Notes != "" {
		existing.Notes = &req.Notes
	} else {
		existing.Notes = nil
	}

	out, err := s.repo.Update(ctx, existing)
	if err != nil {
		return nil, fmt.Errorf("updating customer: %w", err)
	}
	if out == nil {
		return nil, ErrCustomerNotFound
	}
	return toCustomerResponse(out), nil
}

func (s *CustomerService) Delete(ctx context.Context, id string) error {
	if err := s.repo.SoftDelete(ctx, id); err != nil {
		if err.Error() == "customer not found" {
			return ErrCustomerNotFound
		}
		return fmt.Errorf("deleting customer: %w", err)
	}
	return nil
}

// Search is used by the cashier quick-lookup.
func (s *CustomerService) Search(ctx context.Context, f dto.CustomerListFilter) ([]*dto.CustomerResponse, error) {
	f.Defaults()
	rows, _, err := s.repo.FindAll(ctx, f)
	if err != nil {
		return nil, fmt.Errorf("searching customers: %w", err)
	}
	resp := make([]*dto.CustomerResponse, 0, len(rows))
	for _, c := range rows {
		resp = append(resp, toCustomerResponse(c))
	}
	return resp, nil
}

func toCustomerResponse(c *domain.Customer) *dto.CustomerResponse {
	res := &dto.CustomerResponse{
		ID:             c.ID,
		StoreID:        c.StoreID,
		Name:           c.Name,
		Phone:          c.Phone,
		Email:          c.Email,
		Address:        c.Address,
		Notes:          c.Notes,
		CreatedAt:      c.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      c.UpdatedAt.Format(time.RFC3339),
		LoyaltyBalance: c.LoyaltyBalance,
	}

	if c.LoyaltyTierID != nil && c.LoyaltyTierName != nil && c.LoyaltyTierMult != nil {
		res.LoyaltyTierID = c.LoyaltyTierID
		res.LoyaltyTier = &dto.MembershipTierResponse{
			ID:         *c.LoyaltyTierID,
			Name:       *c.LoyaltyTierName,
			Multiplier: *c.LoyaltyTierMult,
		}
	}

	return res
}
