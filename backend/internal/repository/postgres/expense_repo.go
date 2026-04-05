package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
)

type ExpenseRepo struct{ db *sqlx.DB }

func NewExpenseRepo(db *sqlx.DB) *ExpenseRepo { return &ExpenseRepo{db: db} }

// ── Categories ────────────────────────────────────────────────────────────────

func (r *ExpenseRepo) ListCategories(ctx context.Context) ([]*domain.ExpenseCategory, error) {
	const q = `SELECT id, name, description, created_at FROM expense_categories ORDER BY name ASC`
	var rows []*domain.ExpenseCategory
	if err := r.db.SelectContext(ctx, &rows, q); err != nil {
		return nil, fmt.Errorf("ExpenseRepo.ListCategories: %w", err)
	}
	return rows, nil
}

func (r *ExpenseRepo) CreateCategory(ctx context.Context, c *domain.ExpenseCategory) (*domain.ExpenseCategory, error) {
	const q = `
		INSERT INTO expense_categories (name, description)
		VALUES ($1, $2)
		RETURNING id, name, description, created_at`
	row := &domain.ExpenseCategory{}
	if err := r.db.QueryRowxContext(ctx, q, c.Name, c.Description).StructScan(row); err != nil {
		return nil, fmt.Errorf("ExpenseRepo.CreateCategory: %w", err)
	}
	return row, nil
}

func (r *ExpenseRepo) GetCategoryByID(ctx context.Context, id string) (*domain.ExpenseCategory, error) {
	const q = `SELECT id, name, description, created_at FROM expense_categories WHERE id = $1`
	row := &domain.ExpenseCategory{}
	if err := r.db.QueryRowxContext(ctx, q, id).StructScan(row); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("ExpenseRepo.GetCategoryByID: %w", err)
	}
	return row, nil
}

// ── Expenses ──────────────────────────────────────────────────────────────────

func (r *ExpenseRepo) CreateExpense(ctx context.Context, e *domain.Expense) (*domain.Expense, error) {
	const q = `
		INSERT INTO expenses (store_id, category_id, amount, expense_date, notes, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, store_id, category_id, amount, expense_date, notes, created_by, created_at, updated_at`
	row := &domain.Expense{}
	if err := r.db.QueryRowxContext(ctx, q, e.StoreID, e.CategoryID, e.Amount, e.ExpenseDate.Format("2006-01-02"), e.Notes, e.CreatedBy).StructScan(row); err != nil {
		return nil, fmt.Errorf("ExpenseRepo.CreateExpense: %w", err)
	}
	// Fetch CategoryName
	_ = r.db.QueryRowContext(ctx, `SELECT name FROM expense_categories WHERE id = $1`, row.CategoryID).Scan(&row.CategoryName)
	return row, nil
}

func (r *ExpenseRepo) FindAll(ctx context.Context, f dto.ExpenseListFilter) ([]*domain.Expense, int, error) {
	args := []interface{}{f.StoreID}
	conds := []string{"e.store_id = $1"}
	i := 2

	if f.CategoryID != "" {
		conds = append(conds, fmt.Sprintf("e.category_id = $%d", i))
		args = append(args, f.CategoryID)
		i++
	}
	if f.DateFrom != "" {
		conds = append(conds, fmt.Sprintf("e.expense_date >= $%d::date", i))
		args = append(args, f.DateFrom)
		i++
	}
	if f.DateTo != "" {
		conds = append(conds, fmt.Sprintf("e.expense_date <= $%d::date", i))
		args = append(args, f.DateTo)
		i++
	}

	where := "WHERE " + strings.Join(conds, " AND ")

	var total int
	if err := r.db.QueryRowxContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM expenses e %s", where), args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("ExpenseRepo.FindAll count: %w", err)
	}

	args = append(args, f.PerPage, f.Offset())
	dataQ := fmt.Sprintf(`
		SELECT e.id, e.store_id, e.category_id, ec.name AS category_name, e.amount, e.expense_date, e.notes, e.created_by, e.created_at, e.updated_at
		FROM expenses e
		JOIN expense_categories ec ON ec.id = e.category_id
		%s
		ORDER BY e.expense_date DESC, e.created_at DESC
		LIMIT $%d OFFSET $%d`, where, i, i+1)

	var rows []*domain.Expense
	if err := r.db.SelectContext(ctx, &rows, dataQ, args...); err != nil {
		return nil, 0, fmt.Errorf("ExpenseRepo.FindAll: %w", err)
	}
	return rows, total, nil
}

func (r *ExpenseRepo) GetByID(ctx context.Context, id, storeID string) (*domain.Expense, error) {
	const q = `
		SELECT e.id, e.store_id, e.category_id, ec.name AS category_name, e.amount, e.expense_date, e.notes, e.created_by, e.created_at, e.updated_at
		FROM expenses e
		JOIN expense_categories ec ON ec.id = e.category_id
		WHERE e.id = $1 AND e.store_id = $2`
	row := &domain.Expense{}
	if err := r.db.QueryRowxContext(ctx, q, id, storeID).StructScan(row); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("ExpenseRepo.GetByID: %w", err)
	}
	return row, nil
}

func (r *ExpenseRepo) Update(ctx context.Context, e *domain.Expense) (*domain.Expense, error) {
	const q = `
		UPDATE expenses SET category_id = $1, amount = $2, expense_date = $3, notes = $4, updated_at = NOW()
		WHERE id = $5 AND store_id = $6
		RETURNING id, store_id, category_id, amount, expense_date, notes, created_by, created_at, updated_at`
	row := &domain.Expense{}
	if err := r.db.QueryRowxContext(ctx, q, e.CategoryID, e.Amount, e.ExpenseDate.Format("2006-01-02"), e.Notes, e.ID, e.StoreID).StructScan(row); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("ExpenseRepo.Update: %w", err)
	}
	_ = r.db.QueryRowContext(ctx, `SELECT name FROM expense_categories WHERE id = $1`, row.CategoryID).Scan(&row.CategoryName)
	return row, nil
}

func (r *ExpenseRepo) Delete(ctx context.Context, id, storeID string) error {
	const q = `DELETE FROM expenses WHERE id = $1 AND store_id = $2`
	res, err := r.db.ExecContext(ctx, q, id, storeID)
	if err != nil {
		return fmt.Errorf("ExpenseRepo.Delete: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("expense not found")
	}
	return nil
}
