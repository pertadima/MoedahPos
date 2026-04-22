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

// Store-specific sentinel errors.
var (
	ErrStoreNotFound       = errors.New("store not found")
	ErrMemberAlreadyExists = errors.New("user is already a member of this store")
	ErrMemberNotFound      = errors.New("member not found in store")
	ErrUserNotFoundInStore = errors.New("user not found")
)

// StoreService implements business logic for stores and store membership.
type StoreService struct {
	storeRepo repository.StoreRepository
	userRepo  repository.UserRepository
	log       zerolog.Logger
}

func NewStoreService(storeRepo repository.StoreRepository, userRepo repository.UserRepository, log zerolog.Logger) *StoreService {
	return &StoreService{storeRepo: storeRepo, userRepo: userRepo, log: log}
}

func (s *StoreService) ListStores(ctx context.Context, filter dto.StoreListFilter) ([]*dto.StoreResponse, dto.PaginationMeta, error) {
	filter.Defaults()
	stores, total, err := s.storeRepo.FindAll(ctx, filter)
	if err != nil {
		return nil, dto.PaginationMeta{}, fmt.Errorf("listing stores: %w", err)
	}
	resp := make([]*dto.StoreResponse, 0, len(stores))
	for _, st := range stores {
		resp = append(resp, toStoreResponse(st))
	}
	return resp, dto.NewMeta(filter.PaginationQuery, total), nil
}

func (s *StoreService) GetStore(ctx context.Context, id string) (*dto.StoreResponse, error) {
	store, err := s.storeRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("finding store: %w", err)
	}
	if store == nil {
		return nil, ErrStoreNotFound
	}
	return toStoreResponse(store), nil
}

func (s *StoreService) CreateStore(ctx context.Context, req *dto.CreateStoreRequest) (*dto.StoreResponse, error) {
	if req.Currency == "" {
		req.Currency = "IDR"
	}
	if req.StoreType == "" {
		req.StoreType = "retail"
	}
	store, err := s.storeRepo.Create(ctx, &domain.Store{
		Name: req.Name, Address: req.Address, Phone: req.Phone,
		TaxNumber: req.TaxNumber, Currency: req.Currency,
		StoreType: req.StoreType, DefaultTaxPercentage: req.DefaultTaxPercentage,
		LoyaltyPointsPerRupiah: req.LoyaltyPointsPerRupiah, IsActive: true,
	})
	if err != nil {
		return nil, fmt.Errorf("creating store: %w", err)
	}
	s.log.Info().Str("store_id", store.ID).Msg("store created")
	return toStoreResponse(store), nil
}

func (s *StoreService) UpdateStore(ctx context.Context, id string, req *dto.UpdateStoreRequest) (*dto.StoreResponse, error) {
	store, err := s.storeRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("finding store: %w", err)
	}
	if store == nil {
		return nil, ErrStoreNotFound
	}
	store.Name = req.Name
	store.Address = req.Address
	store.Phone = req.Phone
	store.TaxNumber = req.TaxNumber
	store.DefaultTaxPercentage = req.DefaultTaxPercentage
	if req.LoyaltyPointsPerRupiah > 0 {
		store.LoyaltyPointsPerRupiah = req.LoyaltyPointsPerRupiah
	}
	if req.Currency != "" {
		store.Currency = req.Currency
	}
	if req.StoreType != "" {
		store.StoreType = req.StoreType
	}
	if req.IsActive != nil {
		store.IsActive = *req.IsActive
	}
	updated, err := s.storeRepo.Update(ctx, store)
	if err != nil {
		return nil, fmt.Errorf("updating store: %w", err)
	}
	return toStoreResponse(updated), nil
}

// DeleteStore soft-deletes a store by setting deleted_at.
func (s *StoreService) DeleteStore(ctx context.Context, id string) error {
	store, err := s.storeRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("finding store: %w", err)
	}
	if store == nil {
		return ErrStoreNotFound
	}
	if err := s.storeRepo.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("soft-deleting store: %w", err)
	}
	s.log.Info().Str("store_id", id).Msg("store soft-deleted")
	return nil
}

// ─── Members ──────────────────────────────────────────────────────────────────

func (s *StoreService) ListMembers(ctx context.Context, storeID string) ([]*dto.MemberResponse, error) {
	if _, err := s.GetStore(ctx, storeID); err != nil {
		return nil, err
	}
	members, err := s.storeRepo.ListMembers(ctx, storeID)
	if err != nil {
		return nil, fmt.Errorf("listing members: %w", err)
	}
	resp := make([]*dto.MemberResponse, 0, len(members))
	for _, m := range members {
		resp = append(resp, toMemberResponse(m))
	}
	return resp, nil
}

func (s *StoreService) AddMember(ctx context.Context, storeID string, req *dto.AddMemberRequest) error {
	if _, err := s.GetStore(ctx, storeID); err != nil {
		return err
	}
	existing, err := s.storeRepo.FindMember(ctx, req.UserID, storeID)
	if err != nil {
		return fmt.Errorf("checking membership: %w", err)
	}
	if existing != nil && existing.IsActive {
		return ErrMemberAlreadyExists
	}
	return s.storeRepo.AddMember(ctx, &domain.UserStore{
		UserID: req.UserID, StoreID: storeID, RoleID: req.RoleID,
	})
}

func (s *StoreService) UpdateMemberRole(ctx context.Context, storeID, userID string, req *dto.UpdateMemberRoleRequest) error {
	m, err := s.storeRepo.FindMember(ctx, userID, storeID)
	if err != nil {
		return fmt.Errorf("finding member: %w", err)
	}
	if m == nil {
		return ErrMemberNotFound
	}
	return s.storeRepo.UpdateMemberRole(ctx, userID, storeID, req.RoleID)
}

// RemoveMember soft-removes a user from a store by flagging is_active = false.
func (s *StoreService) RemoveMember(ctx context.Context, storeID, userID string) error {
	m, err := s.storeRepo.FindMember(ctx, userID, storeID)
	if err != nil {
		return fmt.Errorf("finding member: %w", err)
	}
	if m == nil {
		return ErrMemberNotFound
	}
	return s.storeRepo.DeactivateMember(ctx, userID, storeID)
}

// ─── Mappers ──────────────────────────────────────────────────────────────────

func toStoreResponse(s *domain.Store) *dto.StoreResponse {
	storeType := s.StoreType
	if storeType == "" {
		storeType = "retail"
	}
	r := &dto.StoreResponse{
		ID:                     s.ID,
		Name:                   s.Name,
		Address:                s.Address,
		Phone:                  s.Phone,
		TaxNumber:              s.TaxNumber,
		Currency:               s.Currency,
		StoreType:              storeType,
		DefaultTaxPercentage:   s.DefaultTaxPercentage,
		LoyaltyPointsPerRupiah: s.LoyaltyPointsPerRupiah,
		IsActive:               s.IsActive,
		CreatedAt:              s.CreatedAt.Format(time.RFC3339),
		UpdatedAt:              s.UpdatedAt.Format(time.RFC3339),
	}
	if s.DeletedAt != nil {
		t := s.DeletedAt.Format(time.RFC3339)
		r.DeletedAt = &t
	}
	return r
}

func toMemberResponse(m *domain.StoreMember) *dto.MemberResponse {
	return &dto.MemberResponse{
		UserID:    m.UserID,
		UserName:  m.UserName,
		UserEmail: m.UserEmail,
		RoleID:    m.RoleID,
		RoleName:  m.RoleName,
		IsActive:  m.IsActive,
		JoinedAt:  m.JoinedAt.Format(time.RFC3339),
	}
}
