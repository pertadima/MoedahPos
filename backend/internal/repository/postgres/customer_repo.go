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

// customerColsBase — always-present columns (safe before migration 00030).
const customerColsBase = `id, store_id, name, phone, email, address, notes, created_at, updated_at, deleted_at`

// customerColsFull — extended column list after migration 00030 (loyalty + sync fields).
// Only used in sync-specific queries where we know the migration has run.
const customerColsFull = customerColsBase + `, loyalty_tier_id, server_updated_at, sync_version`

// CustomerRepo is the PostgreSQL implementation for customers.
type CustomerRepo struct{ db *sqlx.DB }

func NewCustomerRepo(db *sqlx.DB) *CustomerRepo { return &CustomerRepo{db: db} }

func (r *CustomerRepo) Create(ctx context.Context, c *domain.Customer) (*domain.Customer, error) {
	const q = `
		INSERT INTO customers (store_id, name, phone, email, address, notes)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING ` + customerColsBase
	out := &domain.Customer{}
	if err := r.db.QueryRowxContext(ctx, q,
		c.StoreID, c.Name, c.Phone, c.Email, c.Address, c.Notes,
	).StructScan(out); err != nil {
		return nil, fmt.Errorf("CustomerRepo.Create: %w", err)
	}
	return out, nil
}

func (r *CustomerRepo) FindAll(ctx context.Context, f dto.CustomerListFilter) ([]*domain.Customer, int, error) {
	args := []interface{}{f.StoreID}
	conds := []string{"store_id = $1", "deleted_at IS NULL"}
	i := 2
	if f.Search != "" {
		conds = append(conds, fmt.Sprintf("(name ILIKE $%d OR phone ILIKE $%d)", i, i))
		args = append(args, "%"+f.Search+"%")
		i++
	}
	where := "WHERE " + strings.Join(conds, " AND ")

	var total int
	if err := r.db.QueryRowxContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM customers %s", where), args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("CustomerRepo.FindAll count: %w", err)
	}

	args = append(args, f.PerPage, f.Offset())
	q := fmt.Sprintf(`SELECT `+customerColsBase+` FROM customers %s ORDER BY name ASC LIMIT $%d OFFSET $%d`, where, i, i+1)
	var rows []*domain.Customer
	if err := r.db.SelectContext(ctx, &rows, q, args...); err != nil {
		return nil, 0, fmt.Errorf("CustomerRepo.FindAll: %w", err)
	}
	return rows, total, nil
}

func (r *CustomerRepo) FindByID(ctx context.Context, id string) (*domain.Customer, error) {
	q := `SELECT ` + customerColsBase + ` FROM customers WHERE id=$1 AND deleted_at IS NULL`
	c := &domain.Customer{}
	if err := r.db.QueryRowxContext(ctx, q, id).StructScan(c); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("CustomerRepo.FindByID: %w", err)
	}
	return c, nil
}

func (r *CustomerRepo) Update(ctx context.Context, c *domain.Customer) (*domain.Customer, error) {
	q := `
		UPDATE customers SET name=$1,phone=$2,email=$3,address=$4,notes=$5,updated_at=NOW()
		WHERE id=$6 AND deleted_at IS NULL
		RETURNING ` + customerColsBase
	out := &domain.Customer{}
	if err := r.db.QueryRowxContext(ctx, q, c.Name, c.Phone, c.Email, c.Address, c.Notes, c.ID).StructScan(out); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("CustomerRepo.Update: %w", err)
	}
	return out, nil
}

func (r *CustomerRepo) SoftDelete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `UPDATE customers SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("CustomerRepo.SoftDelete: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("customer not found")
	}
	return nil
}

// SearchByPhone looks up customers by phone prefix — useful for cashier quick-search.
func (r *CustomerRepo) SearchByPhone(ctx context.Context, storeID, phone string) ([]*domain.Customer, error) {
	q := `SELECT ` + customerColsBase + ` FROM customers WHERE store_id=$1 AND phone ILIKE $2 AND deleted_at IS NULL ORDER BY name LIMIT 10`
	var rows []*domain.Customer
	if err := r.db.SelectContext(ctx, &rows, q, storeID, phone+"%"); err != nil {
		return nil, fmt.Errorf("CustomerRepo.SearchByPhone: %w", err)
	}
	return rows, nil
}

// GetModifiedSince returns customers updated after the given timestamp (for offline sync delta).
// NOTE: requires migration 00030 — only called by the sync engine which runs post-migration.
func (r *CustomerRepo) GetModifiedSince(ctx context.Context, storeID string, since time.Time) ([]*domain.Customer, error) {
	q := `SELECT ` + customerColsFull + `
		FROM customers
		WHERE store_id = $1 AND server_updated_at > $2 AND deleted_at IS NULL
		ORDER BY server_updated_at ASC`
	var rows []*domain.Customer
	if err := r.db.SelectContext(ctx, &rows, q, storeID, since); err != nil {
		return nil, fmt.Errorf("CustomerRepo.GetModifiedSince: %w", err)
	}
	return rows, nil
}

