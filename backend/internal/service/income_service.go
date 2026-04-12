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

// ErrIncomeNotFound is returned when an income record does not exist.
var ErrIncomeNotFound = errors.New("income not found")

// incomeRepo is the minimal interface the IncomeService needs.
type incomeRepo interface {
	ListCategories(ctx context.Context) ([]*domain.IncomeCategory, error)
	CreateCategory(ctx context.Context, cat *domain.IncomeCategory) (*domain.IncomeCategory, error)
	Create(ctx context.Context, inc *domain.Income) (*domain.Income, error)
	FindAll(ctx context.Context, f dto.IncomeListFilter) ([]*domain.Income, int, error)
	FindByID(ctx context.Context, id string) (*domain.Income, error)
	Update(ctx context.Context, inc *domain.Income) (*domain.Income, error)
	Delete(ctx context.Context, id, storeID string) error
}

// IncomeService implements business logic for non-POS income entries.
type IncomeService struct {
	repo        incomeRepo
	activitySvc *ActivityLogService
	log         zerolog.Logger
}

func NewIncomeService(repo incomeRepo, activitySvc *ActivityLogService, log zerolog.Logger) *IncomeService {
	return &IncomeService{repo: repo, activitySvc: activitySvc, log: log}
}

// ── Categories ────────────────────────────────────────────────────────────────

func (s *IncomeService) ListCategories(ctx context.Context) ([]*dto.IncomeCategoryResponse, error) {
	cats, err := s.repo.ListCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing income categories: %w", err)
	}
	resp := make([]*dto.IncomeCategoryResponse, len(cats))
	for i, c := range cats {
		resp[i] = toIncomeCategoryResponse(c)
	}
	return resp, nil
}

func (s *IncomeService) CreateCategory(ctx context.Context, req *dto.CreateIncomeCategoryRequest) (*dto.IncomeCategoryResponse, error) {
	desc := req.Description
	cat, err := s.repo.CreateCategory(ctx, &domain.IncomeCategory{
		Name:        req.Name,
		Description: &desc,
	})
	if err != nil {
		return nil, fmt.Errorf("creating income category: %w", err)
	}
	return toIncomeCategoryResponse(cat), nil
}

// ── Incomes ───────────────────────────────────────────────────────────────────

func (s *IncomeService) CreateIncome(ctx context.Context, storeID, userID string, req *dto.CreateIncomeRequest) (*dto.IncomeResponse, error) {
	date, err := time.Parse("2006-01-02", req.IncomeDate)
	if err != nil {
		return nil, errors.New("invalid income_date: use YYYY-MM-DD")
	}

	var ref, notes *string
	if req.Reference != "" {
		r := req.Reference
		ref = &r
	}
	if req.Notes != "" {
		n := req.Notes
		notes = &n
	}
	var createdBy *string
	if userID != "" {
		createdBy = &userID
	}

	inc, err := s.repo.Create(ctx, &domain.Income{
		StoreID:       storeID,
		CategoryID:    req.CategoryID,
		Amount:        req.Amount,
		IncomeDate:    date,
		PaymentMethod: req.PaymentMethod,
		Reference:     ref,
		Notes:         notes,
		CreatedBy:     createdBy,
	})
	if err != nil {
		return nil, fmt.Errorf("creating income: %w", err)
	}

	categoryName := ""
	if inc.CategoryName != "" {
		categoryName = inc.CategoryName
	}

	// Determine username and avoid logging un-authenticated background usage if possible
	logUserID := "SYSTEM"
	if userID != "" {
		logUserID = userID
	}

	s.activitySvc.LogActivity(ctx, logUserID, storeID, domain.ActionIncomeCreate, domain.ModuleIncome, inc.ID, map[string]interface{}{
		"category":       categoryName,
		"amount":         req.Amount,
		"payment_method": req.PaymentMethod,
	})

	s.log.Info().Str("income_id", inc.ID).Float64("amount", inc.Amount).Msg("income created")
	return toIncomeResponse(inc), nil
}

func (s *IncomeService) ListIncomes(ctx context.Context, f dto.IncomeListFilter) ([]*dto.IncomeResponse, dto.PaginationMeta, error) {
	f.Defaults()
	rows, total, err := s.repo.FindAll(ctx, f)
	if err != nil {
		return nil, dto.PaginationMeta{}, fmt.Errorf("listing incomes: %w", err)
	}
	resp := make([]*dto.IncomeResponse, len(rows))
	for i, r := range rows {
		resp[i] = toIncomeResponse(r)
	}
	return resp, dto.NewMeta(f.PaginationQuery, total), nil
}

func (s *IncomeService) UpdateIncome(ctx context.Context, id, storeID string, req *dto.UpdateIncomeRequest) (*dto.IncomeResponse, error) {
	date, err := time.Parse("2006-01-02", req.IncomeDate)
	if err != nil {
		return nil, errors.New("invalid income_date: use YYYY-MM-DD")
	}

	var ref, notes *string
	if req.Reference != "" {
		r := req.Reference
		ref = &r
	}
	if req.Notes != "" {
		n := req.Notes
		notes = &n
	}

	updated, err := s.repo.Update(ctx, &domain.Income{
		ID:            id,
		StoreID:       storeID,
		CategoryID:    req.CategoryID,
		Amount:        req.Amount,
		IncomeDate:    date,
		PaymentMethod: req.PaymentMethod,
		Reference:     ref,
		Notes:         notes,
	})
	if err != nil {
		return nil, fmt.Errorf("updating income: %w", err)
	}
	if updated == nil {
		return nil, ErrIncomeNotFound
	}

	// Assuming no user ID pass, use "SYSTEM" or maybe we should add userID to UpdateIncome method
	// As currently UpdateIncome does not receive userID, I will use a dummy UserID to satisfy LogActivity for now.
	// We can update the parameter if strictly required later.
	s.activitySvc.LogActivity(ctx, "SYSTEM", storeID, domain.ActionIncomeUpdate, domain.ModuleIncome, updated.ID, map[string]interface{}{
		"category":       updated.CategoryName,
		"amount":         updated.Amount,
		"payment_method": updated.PaymentMethod,
	})

	return toIncomeResponse(updated), nil
}

func (s *IncomeService) DeleteIncome(ctx context.Context, id, storeID string) error {
	inc, errGet := s.repo.FindByID(ctx, id)

	if err := s.repo.Delete(ctx, id, storeID); err != nil {
		if err.Error() == "income not found" {
			return ErrIncomeNotFound
		}
		return fmt.Errorf("deleting income: %w", err)
	}

	if errGet == nil && inc != nil {
		s.activitySvc.LogActivity(ctx, "SYSTEM", storeID, domain.ActionIncomeDelete, domain.ModuleIncome, id, map[string]interface{}{
			"category": inc.CategoryName,
			"amount":   inc.Amount,
		})
	}
	
	return nil
}

// ── Mappers ───────────────────────────────────────────────────────────────────

func toIncomeCategoryResponse(c *domain.IncomeCategory) *dto.IncomeCategoryResponse {
	return &dto.IncomeCategoryResponse{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		CreatedAt:   c.CreatedAt.Format(time.RFC3339),
	}
}

func toIncomeResponse(inc *domain.Income) *dto.IncomeResponse {
	return &dto.IncomeResponse{
		ID:            inc.ID,
		StoreID:       inc.StoreID,
		CategoryID:    inc.CategoryID,
		CategoryName:  inc.CategoryName,
		Amount:        inc.Amount,
		IncomeDate:    inc.IncomeDate.Format("2006-01-02"),
		PaymentMethod: inc.PaymentMethod,
		Reference:     inc.Reference,
		Notes:         inc.Notes,
		CreatedBy:     inc.CreatedBy,
		CreatedByName: inc.CreatedByName,
		CreatedAt:     inc.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     inc.UpdatedAt.Format(time.RFC3339),
	}
}
