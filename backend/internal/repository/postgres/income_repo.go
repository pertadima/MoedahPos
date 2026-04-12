package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
)

// IncomeRepo handles income_categories and incomes persistence.
type IncomeRepo struct{ db *sqlx.DB }

func NewIncomeRepo(db *sqlx.DB) *IncomeRepo { return &IncomeRepo{db: db} }

// ─── Categories ───────────────────────────────────────────────────────────────

func (r *IncomeRepo) ListCategories(ctx context.Context, includeDeleted bool) ([]*domain.IncomeCategory, error) {
	q := `SELECT id, name, description, is_active, created_at, updated_at FROM income_categories`
	if !includeDeleted {
		q += ` WHERE deleted_at IS NULL AND is_active = true`
	}
	q += ` ORDER BY name ASC`
	var cats []*domain.IncomeCategory
	if err := r.db.SelectContext(ctx, &cats, q); err != nil {
		return nil, fmt.Errorf("IncomeRepo.ListCategories: %w", err)
	}
	return cats, nil
}

func (r *IncomeRepo) CreateCategory(ctx context.Context, cat *domain.IncomeCategory) (*domain.IncomeCategory, error) {
	const q = `
		INSERT INTO income_categories (name, description)
		VALUES ($1, $2)
		RETURNING id, name, description, is_active, created_at, updated_at`
	row := &domain.IncomeCategory{}
	if err := r.db.QueryRowxContext(ctx, q, cat.Name, cat.Description).StructScan(row); err != nil {
		return nil, fmt.Errorf("IncomeRepo.CreateCategory: %w", err)
	}
	return row, nil
}

func (r *IncomeRepo) GetCategoryByID(ctx context.Context, id string) (*domain.IncomeCategory, error) {
	const q = `SELECT id, name, description, is_active, created_at, updated_at FROM income_categories WHERE id = $1`
	row := &domain.IncomeCategory{}
	if err := r.db.QueryRowxContext(ctx, q, id).StructScan(row); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("IncomeRepo.GetCategoryByID: %w", err)
	}
	return row, nil
}

func (r *IncomeRepo) UpdateCategory(ctx context.Context, id string, name, desc string, isActive bool) (*domain.IncomeCategory, error) {
	const q = `
		UPDATE income_categories
		SET name = $1, description = $2, is_active = $3, updated_at = NOW()
		WHERE id = $4 AND deleted_at IS NULL
		RETURNING id, name, description, is_active, created_at, updated_at`
	row := &domain.IncomeCategory{}
	if err := r.db.QueryRowxContext(ctx, q, name, desc, isActive, id).StructScan(row); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("IncomeRepo.UpdateCategory: %w", err)
	}
	return row, nil
}

func (r *IncomeRepo) SoftDeleteCategory(ctx context.Context, id string) error {
	const q = `
		UPDATE income_categories
		SET deleted_at = NOW(), is_active = false, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL`
	res, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("IncomeRepo.SoftDeleteCategory: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("category not found or already deleted")
	}
	return nil
}

// ─── Incomes ──────────────────────────────────────────────────────────────────

func (r *IncomeRepo) Create(ctx context.Context, inc *domain.Income) (*domain.Income, error) {
	const q = `
		INSERT INTO incomes (store_id, category_id, amount, income_date, payment_method, reference, notes, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, store_id, category_id, amount, income_date, payment_method, reference, notes, created_by, created_at, updated_at`
	row := &domain.Income{}
	if err := r.db.QueryRowxContext(ctx, q,
		inc.StoreID, inc.CategoryID, inc.Amount, inc.IncomeDate,
		inc.PaymentMethod, inc.Reference, inc.Notes, inc.CreatedBy,
	).StructScan(row); err != nil {
		return nil, fmt.Errorf("IncomeRepo.Create: %w", err)
	}
	// re-fetch with joins for category_name / created_by_name
	return r.FindByID(ctx, row.ID)
}

func (r *IncomeRepo) FindAll(ctx context.Context, f dto.IncomeListFilter) ([]*domain.Income, int, error) {
	args := []interface{}{}
	conds := []string{"i.store_id = $1"}
	args = append(args, f.StoreID)
	i := 2

	if f.CategoryID != "" {
		conds = append(conds, fmt.Sprintf("i.category_id = $%d", i))
		args = append(args, f.CategoryID)
		i++
	}
	if f.DateFrom != "" {
		conds = append(conds, fmt.Sprintf("i.income_date >= $%d", i))
		args = append(args, f.DateFrom)
		i++
	}
	if f.DateTo != "" {
		conds = append(conds, fmt.Sprintf("i.income_date <= $%d", i))
		args = append(args, f.DateTo)
		i++
	}

	where := "WHERE " + strings.Join(conds, " AND ")

	var total int
	if err := r.db.QueryRowxContext(ctx,
		fmt.Sprintf("SELECT COUNT(*) FROM incomes i %s", where), args...,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("IncomeRepo.FindAll count: %w", err)
	}

	f.Defaults()
	args = append(args, f.PerPage, f.Offset())
	dataQ := fmt.Sprintf(`
		SELECT i.id, i.store_id, i.category_id, i.amount, i.income_date,
		       i.payment_method, i.reference, i.notes, i.created_by,
		       i.created_at, i.updated_at,
		       c.name  AS category_name,
		       u.name  AS created_by_name
		FROM incomes i
		JOIN income_categories c ON c.id = i.category_id
		LEFT JOIN users u ON u.id  = i.created_by
		%s
		ORDER BY i.income_date DESC, i.created_at DESC
		LIMIT $%d OFFSET $%d`, where, i, i+1)

	var rows []*domain.Income
	if err := r.db.SelectContext(ctx, &rows, dataQ, args...); err != nil {
		return nil, 0, fmt.Errorf("IncomeRepo.FindAll: %w", err)
	}
	return rows, total, nil
}

func (r *IncomeRepo) FindByID(ctx context.Context, id string) (*domain.Income, error) {
	const q = `
		SELECT i.id, i.store_id, i.category_id, i.amount, i.income_date,
		       i.payment_method, i.reference, i.notes, i.created_by,
		       i.created_at, i.updated_at,
		       c.name  AS category_name,
		       u.name  AS created_by_name
		FROM incomes i
		JOIN income_categories c ON c.id = i.category_id
		LEFT JOIN users u ON u.id  = i.created_by
		WHERE i.id = $1`
	row := &domain.Income{}
	if err := r.db.QueryRowxContext(ctx, q, id).StructScan(row); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("IncomeRepo.FindByID: %w", err)
	}
	return row, nil
}

func (r *IncomeRepo) Update(ctx context.Context, inc *domain.Income) (*domain.Income, error) {
	const q = `
		UPDATE incomes
		SET category_id=$1, amount=$2, income_date=$3, payment_method=$4,
		    reference=$5, notes=$6, updated_at=NOW()
		WHERE id=$7 AND store_id=$8
		RETURNING id`
	var id string
	if err := r.db.QueryRowxContext(ctx, q,
		inc.CategoryID, inc.Amount, inc.IncomeDate, inc.PaymentMethod,
		inc.Reference, inc.Notes, inc.ID, inc.StoreID,
	).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("IncomeRepo.Update: %w", err)
	}
	return r.FindByID(ctx, id)
}

func (r *IncomeRepo) Delete(ctx context.Context, id, storeID string) error {
	const q = `DELETE FROM incomes WHERE id=$1 AND store_id=$2`
	res, err := r.db.ExecContext(ctx, q, id, storeID)
	if err != nil {
		return fmt.Errorf("IncomeRepo.Delete: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("income not found")
	}
	return nil
}

// SumByDateRange returns the total income amount for a store in [from, to).
func (r *IncomeRepo) SumByDateRange(ctx context.Context, storeID string, from, to time.Time) (float64, error) {
	const q = `
		SELECT COALESCE(SUM(amount), 0)
		FROM incomes
		WHERE store_id=$1 AND income_date >= $2 AND income_date < $3`
	var total float64
	if err := r.db.QueryRowxContext(ctx, q, storeID, from.Format("2006-01-02"), to.Format("2006-01-02")).Scan(&total); err != nil {
		return 0, fmt.Errorf("IncomeRepo.SumByDateRange: %w", err)
	}
	return total, nil
}
