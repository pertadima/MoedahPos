package service

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
	"github.com/moedahpos/backend/internal/repository/postgres"
)

type ExpenseService struct {
	expenseRepo *postgres.ExpenseRepo
	log         zerolog.Logger
}

func NewExpenseService(expenseRepo *postgres.ExpenseRepo, log zerolog.Logger) *ExpenseService {
	return &ExpenseService{expenseRepo: expenseRepo, log: log}
}

// ── Categories ────────────────────────────────────────────────────────────────

func (s *ExpenseService) ListCategories(ctx context.Context) ([]dto.ExpenseCategoryResponse, error) {
	cats, err := s.expenseRepo.ListCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing expense categories: %w", err)
	}

	resp := make([]dto.ExpenseCategoryResponse, 0, len(cats))
	for _, c := range cats {
		resp = append(resp, dto.ExpenseCategoryResponse{
			ID:          c.ID,
			Name:        c.Name,
			Description: c.Description,
			CreatedAt:   c.CreatedAt.Format(time.RFC3339),
		})
	}
	return resp, nil
}

func (s *ExpenseService) CreateCategory(ctx context.Context, req *dto.CreateExpenseCategoryRequest) (*dto.ExpenseCategoryResponse, error) {
	c, err := s.expenseRepo.CreateCategory(ctx, &domain.ExpenseCategory{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return nil, fmt.Errorf("creating expense category: %w", err)
	}
	return &dto.ExpenseCategoryResponse{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		CreatedAt:   c.CreatedAt.Format(time.RFC3339),
	}, nil
}

// ── Expenses ──────────────────────────────────────────────────────────────────

func (s *ExpenseService) CreateExpense(ctx context.Context, storeID, userID string, req *dto.CreateExpenseRequest) (*dto.ExpenseResponse, error) {
	expenseDate, err := time.Parse("2006-01-02", req.ExpenseDate)
	if err != nil {
		return nil, fmt.Errorf("invalid expense_date: %w", err)
	}

	e, err := s.expenseRepo.CreateExpense(ctx, &domain.Expense{
		StoreID:     storeID,
		CategoryID:  req.CategoryID,
		Amount:      req.Amount,
		ExpenseDate: expenseDate,
		Notes:       req.Notes,
		CreatedBy:   &userID,
	})
	if err != nil {
		return nil, fmt.Errorf("creating expense: %w", err)
	}
	return s.toExpenseResponse(e), nil
}

func (s *ExpenseService) ListExpenses(ctx context.Context, filter dto.ExpenseListFilter) ([]*dto.ExpenseResponse, dto.PaginationMeta, error) {
	filter.Defaults()
	expenses, total, err := s.expenseRepo.FindAll(ctx, filter)
	if err != nil {
		return nil, dto.PaginationMeta{}, fmt.Errorf("listing expenses: %w", err)
	}

	resp := make([]*dto.ExpenseResponse, 0, len(expenses))
	for _, e := range expenses {
		resp = append(resp, s.toExpenseResponse(e))
	}
	return resp, dto.NewMeta(filter.PaginationQuery, total), nil
}

func (s *ExpenseService) UpdateExpense(ctx context.Context, id, storeID string, req *dto.UpdateExpenseRequest) (*dto.ExpenseResponse, error) {
	expenseDate, err := time.Parse("2006-01-02", req.ExpenseDate)
	if err != nil {
		return nil, fmt.Errorf("invalid expense_date: %w", err)
	}

	e, err := s.expenseRepo.Update(ctx, &domain.Expense{
		ID:          id,
		StoreID:     storeID,
		CategoryID:  req.CategoryID,
		Amount:      req.Amount,
		ExpenseDate: expenseDate,
		Notes:       req.Notes,
	})
	if err != nil {
		return nil, fmt.Errorf("updating expense: %w", err)
	}
	if e == nil {
		return nil, fmt.Errorf("expense not found")
	}
	return s.toExpenseResponse(e), nil
}

func (s *ExpenseService) DeleteExpense(ctx context.Context, id, storeID string) error {
	if err := s.expenseRepo.Delete(ctx, id, storeID); err != nil {
		return fmt.Errorf("deleting expense: %w", err)
	}
	return nil
}

func (s *ExpenseService) toExpenseResponse(e *domain.Expense) *dto.ExpenseResponse {
	return &dto.ExpenseResponse{
		ID:           e.ID,
		StoreID:      e.StoreID,
		CategoryID:   e.CategoryID,
		CategoryName: e.CategoryName,
		Amount:       e.Amount,
		ExpenseDate:  e.ExpenseDate.Format("2006-01-02"),
		Notes:        e.Notes,
		CreatedBy:    e.CreatedBy,
		CreatedAt:    e.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    e.UpdatedAt.Format(time.RFC3339),
	}
}
