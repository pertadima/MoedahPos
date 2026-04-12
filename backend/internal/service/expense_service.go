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
	activitySvc *ActivityLogService
	log         zerolog.Logger
}

func NewExpenseService(expenseRepo *postgres.ExpenseRepo, activitySvc *ActivityLogService, log zerolog.Logger) *ExpenseService {
	return &ExpenseService{expenseRepo: expenseRepo, activitySvc: activitySvc, log: log}
}

// ── Categories ────────────────────────────────────────────────────────────────

func (s *ExpenseService) ListCategories(ctx context.Context, includeDeleted bool) ([]dto.ExpenseCategoryResponse, error) {
	cats, err := s.expenseRepo.ListCategories(ctx, includeDeleted)
	if err != nil {
		return nil, fmt.Errorf("listing expense categories: %w", err)
	}

	resp := make([]dto.ExpenseCategoryResponse, 0, len(cats))
	for _, c := range cats {
		resp = append(resp, dto.ExpenseCategoryResponse{
			ID:          c.ID,
			Name:        c.Name,
			Description: c.Description,
			IsActive:    c.IsActive,
			CreatedAt:   c.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   c.UpdatedAt.Format(time.RFC3339),
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
		IsActive:    c.IsActive,
		CreatedAt:   c.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   c.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *ExpenseService) UpdateCategory(ctx context.Context, id string, req *dto.UpdateExpenseCategoryRequest) (*dto.ExpenseCategoryResponse, error) {
	c, err := s.expenseRepo.UpdateCategory(ctx, id, req.Name, req.Description, req.IsActive)
	if err != nil {
		return nil, fmt.Errorf("updating expense category: %w", err)
	}
	if c == nil {
		return nil, fmt.Errorf("expense category not found")
	}
	return &dto.ExpenseCategoryResponse{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		IsActive:    c.IsActive,
		CreatedAt:   c.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   c.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *ExpenseService) SoftDeleteCategory(ctx context.Context, id string) error {
	return s.expenseRepo.SoftDeleteCategory(ctx, id)
}

// ── Expenses ──────────────────────────────────────────────────────────────────

func (s *ExpenseService) CreateExpense(ctx context.Context, storeID, userID string, req *dto.CreateExpenseRequest) (*dto.ExpenseResponse, error) {
	expenseDate, err := time.Parse("2006-01-02", req.ExpenseDate)
	if err != nil {
		return nil, fmt.Errorf("invalid expense_date: %w", err)
	}

	e, err := s.expenseRepo.CreateExpense(ctx, &domain.Expense{
		StoreID:       storeID,
		CategoryID:    req.CategoryID,
		Amount:        req.Amount,
		ExpenseDate:   expenseDate,
		Notes:         req.Notes,
		PaymentStatus: req.PaymentStatus,
		CreatedBy:     &userID,
	})
	if err != nil {
		return nil, fmt.Errorf("creating expense: %w", err)
	}

	categoryName := ""
	if e.CategoryName != "" {
		categoryName = e.CategoryName
	}
	logUserID := "SYSTEM"
	if userID != "" {
		logUserID = userID
	}

	s.activitySvc.LogActivity(ctx, logUserID, storeID, domain.ActionExpenseCreate, domain.ModuleExpense, e.ID, map[string]interface{}{
		"category": categoryName,
		"amount":   req.Amount,
	})

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

	s.activitySvc.LogActivity(ctx, "SYSTEM", storeID, domain.ActionExpenseUpdate, domain.ModuleExpense, e.ID, map[string]interface{}{
		"category": e.CategoryName,
		"amount":   e.Amount,
	})

	return s.toExpenseResponse(e), nil
}

func (s *ExpenseService) DeleteExpense(ctx context.Context, id, storeID string) error {
	e, errGet := s.expenseRepo.GetByID(ctx, id, storeID)

	if err := s.expenseRepo.Delete(ctx, id, storeID); err != nil {
		return fmt.Errorf("deleting expense: %w", err)
	}

	if errGet == nil && e != nil {
		s.activitySvc.LogActivity(ctx, "SYSTEM", storeID, domain.ActionExpenseDelete, domain.ModuleExpense, id, map[string]interface{}{
			"category": e.CategoryName,
			"amount":   e.Amount,
		})
	}
	return nil
}

func (s *ExpenseService) toExpenseResponse(e *domain.Expense) *dto.ExpenseResponse {
	return &dto.ExpenseResponse{
		ID:            e.ID,
		StoreID:       e.StoreID,
		CategoryID:    e.CategoryID,
		CategoryName:  e.CategoryName,
		Amount:        e.Amount,
		ExpenseDate:   e.ExpenseDate.Format("2006-01-02"),
		Notes:         e.Notes,
		PaymentStatus: e.PaymentStatus,
		CreatedBy:     e.CreatedBy,
		CreatedAt:     e.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     e.UpdatedAt.Format(time.RFC3339),
	}
}

func (s *ExpenseService) UpdatePaymentStatus(ctx context.Context, id, storeID string, req *dto.UpdateExpenseStatusRequest) (*dto.ExpenseResponse, error) {
	e, err := s.expenseRepo.UpdatePaymentStatus(ctx, id, storeID, req.PaymentStatus)
	if err != nil {
		return nil, fmt.Errorf("updating expense payment status: %w", err)
	}
	if e == nil {
		return nil, fmt.Errorf("expense not found")
	}
	return s.toExpenseResponse(e), nil
}

// ── Recurring Expenses ─────────────────────────────────────────────────────────

func (s *ExpenseService) CreateRecurringExpense(ctx context.Context, storeID, userID string, req *dto.CreateRecurringExpenseRequest) (*dto.RecurringExpenseResponse, error) {
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date: %w", err)
	}
	var endDate *time.Time
	if req.EndDate != nil && *req.EndDate != "" {
		ed, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end_date: %w", err)
		}
		endDate = &ed
	}

	re, err := s.expenseRepo.CreateRecurringExpense(ctx, &domain.RecurringExpense{
		StoreID:       storeID,
		CategoryID:    req.CategoryID,
		Name:          req.Name,
		Amount:        req.Amount,
		Interval:      req.Interval,
		IntervalValue: req.IntervalValue,
		StartDate:     startDate,
		EndDate:       endDate,
		NextRunDate:   startDate, // initially run on start date
		Notes:         req.Notes,
		IsActive:      true,
		CreatedBy:     &userID,
	})
	if err != nil {
		return nil, fmt.Errorf("creating recurring expense: %w", err)
	}
	return s.toRecurringExpenseResponse(re), nil
}

func (s *ExpenseService) ListRecurringExpenses(ctx context.Context, filter dto.ExpenseListFilter) ([]*dto.RecurringExpenseResponse, dto.PaginationMeta, error) {
	filter.Defaults()
	expenses, total, err := s.expenseRepo.FindAllRecurring(ctx, filter)
	if err != nil {
		return nil, dto.PaginationMeta{}, fmt.Errorf("listing recurring expenses: %w", err)
	}

	resp := make([]*dto.RecurringExpenseResponse, 0, len(expenses))
	for _, e := range expenses {
		resp = append(resp, s.toRecurringExpenseResponse(e))
	}
	return resp, dto.NewMeta(filter.PaginationQuery, total), nil
}

func (s *ExpenseService) UpdateRecurringExpense(ctx context.Context, id, storeID string, req *dto.UpdateRecurringExpenseRequest) (*dto.RecurringExpenseResponse, error) {
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date: %w", err)
	}
	var endDate *time.Time
	if req.EndDate != nil && *req.EndDate != "" {
		ed, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end_date: %w", err)
		}
		endDate = &ed
	}

	re, err := s.expenseRepo.UpdateRecurring(ctx, &domain.RecurringExpense{
		ID:            id,
		StoreID:       storeID,
		CategoryID:    req.CategoryID,
		Name:          req.Name,
		Amount:        req.Amount,
		Interval:      req.Interval,
		IntervalValue: req.IntervalValue,
		StartDate:     startDate,
		EndDate:       endDate,
		IsActive:      req.IsActive,
		Notes:         req.Notes,
	})
	if err != nil {
		return nil, fmt.Errorf("updating recurring expense: %w", err)
	}
	if re == nil {
		return nil, fmt.Errorf("recurring expense not found")
	}
	return s.toRecurringExpenseResponse(re), nil
}

func (s *ExpenseService) DeleteRecurringExpense(ctx context.Context, id, storeID string) error {
	if err := s.expenseRepo.DeleteRecurring(ctx, id, storeID); err != nil {
		return fmt.Errorf("deleting recurring expense: %w", err)
	}
	return nil
}

func (s *ExpenseService) ProcessDueRecurringExpenses(ctx context.Context) error {
	due, err := s.expenseRepo.GetDueRecurringExpenses(ctx)
	if err != nil {
		return fmt.Errorf("getting due recurring expenses: %w", err)
	}

	for _, re := range due {
		// Create the corresponding expense as 'unpaid'
		notes := fmt.Sprintf("Auto-generated from recurring template: %s\n%s", re.Name, re.Notes)
		_, err := s.expenseRepo.CreateExpense(ctx, &domain.Expense{
			StoreID:       re.StoreID,
			CategoryID:    re.CategoryID,
			Amount:        re.Amount,
			ExpenseDate:   time.Now(),
			Notes:         notes,
			PaymentStatus: "unpaid",
			CreatedBy:     re.CreatedBy,
		})
		if err != nil {
			s.log.Error().Err(err).Str("recurring_id", re.ID).Msg("Failed to auto-generate expense record")
			continue
		}

		// Calculate Next Run Date
		var nextRun time.Time
		switch re.Interval {
		case "daily":
			nextRun = re.NextRunDate.AddDate(0, 0, re.IntervalValue)
		case "weekly":
			nextRun = re.NextRunDate.AddDate(0, 0, 7*re.IntervalValue)
		case "monthly":
			nextRun = re.NextRunDate.AddDate(0, re.IntervalValue, 0)
		case "yearly":
			nextRun = re.NextRunDate.AddDate(re.IntervalValue, 0, 0)
		}

		if err := s.expenseRepo.BumpRecurringNextRun(ctx, re.ID, nextRun.Format("2006-01-02")); err != nil {
			s.log.Error().Err(err).Str("recurring_id", re.ID).Msg("Failed to bump next_run_date")
		} else {
			s.log.Info().Str("recurring_id", re.ID).Msg("Successfully processed recurring expense")
		}
	}
	return nil
}

func (s *ExpenseService) toRecurringExpenseResponse(e *domain.RecurringExpense) *dto.RecurringExpenseResponse {
	var endDate, lastGen *string
	if e.EndDate != nil {
		v := e.EndDate.Format("2006-01-02")
		endDate = &v
	}
	if e.LastGeneratedAt != nil {
		v := e.LastGeneratedAt.Format(time.RFC3339)
		lastGen = &v
	}

	return &dto.RecurringExpenseResponse{
		ID:              e.ID,
		StoreID:         e.StoreID,
		CategoryID:      e.CategoryID,
		CategoryName:    e.CategoryName,
		Name:            e.Name,
		Amount:          e.Amount,
		Interval:        e.Interval,
		IntervalValue:   e.IntervalValue,
		StartDate:       e.StartDate.Format("2006-01-02"),
		EndDate:         endDate,
		NextRunDate:     e.NextRunDate.Format("2006-01-02"),
		Notes:           e.Notes,
		IsActive:        e.IsActive,
		CreatedBy:       e.CreatedBy,
		CreatedAt:       e.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       e.UpdatedAt.Format(time.RFC3339),
		LastGeneratedAt: lastGen,
	}
}
