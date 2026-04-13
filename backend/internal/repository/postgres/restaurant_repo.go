package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/moedahpos/backend/internal/domain"
)

// ─── Table Repo ───────────────────────────────────────────────────────────────

type TableRepo struct{ db *sqlx.DB }

func NewTableRepo(db *sqlx.DB) *TableRepo { return &TableRepo{db: db} }

func (r *TableRepo) FindAllByStore(ctx context.Context, storeID string) ([]*domain.RestaurantTable, error) {
	const q = `
		SELECT id, store_id, table_number, capacity, status, notes, is_active, created_at, updated_at, deleted_at
		FROM restaurant_tables
		WHERE store_id = $1 AND deleted_at IS NULL
		ORDER BY table_number ASC`
	var tables []*domain.RestaurantTable
	if err := r.db.SelectContext(ctx, &tables, q, storeID); err != nil {
		return nil, fmt.Errorf("TableRepo.FindAllByStore: %w", err)
	}
	return tables, nil
}

func (r *TableRepo) FindByID(ctx context.Context, id string) (*domain.RestaurantTable, error) {
	const q = `
		SELECT id, store_id, table_number, capacity, status, notes, is_active, created_at, updated_at, deleted_at
		FROM restaurant_tables WHERE id = $1 AND deleted_at IS NULL`
	t := &domain.RestaurantTable{}
	if err := r.db.QueryRowxContext(ctx, q, id).StructScan(t); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("TableRepo.FindByID: %w", err)
	}
	return t, nil
}

func (r *TableRepo) Create(ctx context.Context, t *domain.RestaurantTable) (*domain.RestaurantTable, error) {
	const q = `
		INSERT INTO restaurant_tables (store_id, table_number, capacity, status, notes, is_active)
		VALUES ($1, $2, $3, $4, $5, true)
		RETURNING id, store_id, table_number, capacity, status, notes, is_active, created_at, updated_at, deleted_at`
	row := &domain.RestaurantTable{}
	err := r.db.QueryRowxContext(ctx, q,
		t.StoreID, t.TableNumber, t.Capacity, t.Status, t.Notes,
	).StructScan(row)
	if err != nil {
		return nil, fmt.Errorf("TableRepo.Create: %w", err)
	}
	return row, nil
}

func (r *TableRepo) Update(ctx context.Context, t *domain.RestaurantTable) (*domain.RestaurantTable, error) {
	const q = `
		UPDATE restaurant_tables
		SET table_number=$1, capacity=$2, notes=$3, is_active=$4, updated_at=NOW()
		WHERE id=$5 AND deleted_at IS NULL
		RETURNING id, store_id, table_number, capacity, status, notes, is_active, created_at, updated_at, deleted_at`
	row := &domain.RestaurantTable{}
	err := r.db.QueryRowxContext(ctx, q, t.TableNumber, t.Capacity, t.Notes, t.IsActive, t.ID).StructScan(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("TableRepo.Update: %w", err)
	}
	return row, nil
}

func (r *TableRepo) UpdateStatus(ctx context.Context, id string, status domain.TableStatus) error {
	const q = `UPDATE restaurant_tables SET status=$1, updated_at=NOW() WHERE id=$2 AND deleted_at IS NULL`
	res, err := r.db.ExecContext(ctx, q, status, id)
	if err != nil {
		return fmt.Errorf("TableRepo.UpdateStatus: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("table not found")
	}
	return nil
}

func (r *TableRepo) SoftDelete(ctx context.Context, id string) error {
	const q = `UPDATE restaurant_tables SET deleted_at=NOW(), updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL`
	res, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("TableRepo.SoftDelete: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("table not found")
	}
	return nil
}

// ─── Menu Item Repo ───────────────────────────────────────────────────────────

type MenuItemRepo struct{ db *sqlx.DB }

func NewMenuItemRepo(db *sqlx.DB) *MenuItemRepo { return &MenuItemRepo{db: db} }

func (r *MenuItemRepo) FindAllByStore(ctx context.Context, storeID string) ([]*domain.MenuItem, error) {
	const q = `
		SELECT mi.id, mi.store_id, mi.category_id, mi.name, mi.description,
		       mi.sell_price, mi.tax_rate, mi.image_url, mi.is_active,
		       mi.packaging_cost, mi.overhead_cost, mi.labor_cost,
		       mi.created_at, mi.updated_at, mi.deleted_at,
		       c.name AS category_name
		FROM menu_items mi
		LEFT JOIN categories c ON c.id = mi.category_id AND c.deleted_at IS NULL
		WHERE mi.store_id = $1 AND mi.deleted_at IS NULL
		ORDER BY mi.name ASC`
	var items []*domain.MenuItem
	if err := r.db.SelectContext(ctx, &items, q, storeID); err != nil {
		return nil, fmt.Errorf("MenuItemRepo.FindAllByStore: %w", err)
	}

	// Load ingredients for all items
	if len(items) == 0 {
		return items, nil
	}
	ids := make([]string, len(items))
	for i, item := range items {
		ids[i] = item.ID
	}
	ingMap, err := r.loadIngredients(ctx, ids)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		item.Ingredients = ingMap[item.ID]
	}
	return items, nil
}

func (r *MenuItemRepo) FindByID(ctx context.Context, id string) (*domain.MenuItem, error) {
	const q = `
		SELECT mi.id, mi.store_id, mi.category_id, mi.name, mi.description,
		       mi.sell_price, mi.tax_rate, mi.image_url, mi.is_active,
		       mi.packaging_cost, mi.overhead_cost, mi.labor_cost,
		       mi.created_at, mi.updated_at, mi.deleted_at,
		       c.name AS category_name
		FROM menu_items mi
		LEFT JOIN categories c ON c.id = mi.category_id AND c.deleted_at IS NULL
		WHERE mi.id = $1 AND mi.deleted_at IS NULL`
	item := &domain.MenuItem{}
	if err := r.db.QueryRowxContext(ctx, q, id).StructScan(item); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("MenuItemRepo.FindByID: %w", err)
	}
	ingMap, err := r.loadIngredients(ctx, []string{id})
	if err != nil {
		return nil, err
	}
	item.Ingredients = ingMap[id]
	return item, nil
}

func (r *MenuItemRepo) loadIngredients(ctx context.Context, ids []string) (map[string][]domain.MenuItemIngredient, error) {
	if len(ids) == 0 {
		return map[string][]domain.MenuItemIngredient{}, nil
	}
	query, args, err := sqlx.In(`
		SELECT mii.id, mii.menu_item_id, mii.product_id, mii.quantity,
		       p.name AS product_name, p.sku AS product_sku, p.unit, p.cost_price
		FROM menu_item_ingredients mii
		JOIN products p ON p.id = mii.product_id AND p.deleted_at IS NULL
		WHERE mii.menu_item_id IN (?)
		ORDER BY p.name ASC`, ids)
	if err != nil {
		return nil, fmt.Errorf("MenuItemRepo.loadIngredients build query: %w", err)
	}
	query = r.db.Rebind(query)
	var ings []domain.MenuItemIngredient
	if err := r.db.SelectContext(ctx, &ings, query, args...); err != nil {
		return nil, fmt.Errorf("MenuItemRepo.loadIngredients: %w", err)
	}
	m := map[string][]domain.MenuItemIngredient{}
	for _, ing := range ings {
		m[ing.MenuItemID] = append(m[ing.MenuItemID], ing)
	}
	return m, nil
}

func (r *MenuItemRepo) Create(ctx context.Context, item *domain.MenuItem) (*domain.MenuItem, error) {
	const q = `
		INSERT INTO menu_items (store_id, category_id, name, description, sell_price, tax_rate, packaging_cost, overhead_cost, labor_cost, image_url, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, true)
		RETURNING id, store_id, category_id, name, description, sell_price, tax_rate, packaging_cost, overhead_cost, labor_cost, image_url,
		          is_active, created_at, updated_at, deleted_at`
	row := &domain.MenuItem{}
	err := r.db.QueryRowxContext(ctx, q,
		item.StoreID, item.CategoryID, item.Name, item.Description,
		item.SellPrice, item.TaxRate, item.PackagingCost, item.OverheadCost, item.LaborCost, item.ImageURL,
	).StructScan(row)
	if err != nil {
		return nil, fmt.Errorf("MenuItemRepo.Create: %w", err)
	}
	return row, nil
}

func (r *MenuItemRepo) Update(ctx context.Context, item *domain.MenuItem) (*domain.MenuItem, error) {
	const q = `
		UPDATE menu_items
		SET category_id=$1, name=$2, description=$3, sell_price=$4, tax_rate=$5,
		    packaging_cost=$6, overhead_cost=$7, labor_cost=$8,
		    image_url=$9, is_active=$10, updated_at=NOW()
		WHERE id=$11 AND deleted_at IS NULL
		RETURNING id, store_id, category_id, name, description, sell_price, tax_rate, packaging_cost, overhead_cost, labor_cost, image_url,
		          is_active, created_at, updated_at, deleted_at`
	row := &domain.MenuItem{}
	err := r.db.QueryRowxContext(ctx, q,
		item.CategoryID, item.Name, item.Description, item.SellPrice, item.TaxRate,
		item.PackagingCost, item.OverheadCost, item.LaborCost,
		item.ImageURL, item.IsActive, item.ID,
	).StructScan(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("MenuItemRepo.Update: %w", err)
	}
	return row, nil
}

func (r *MenuItemRepo) ReplaceIngredients(ctx context.Context, menuItemID string, ings []domain.MenuItemIngredient) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("MenuItemRepo.ReplaceIngredients begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `DELETE FROM menu_item_ingredients WHERE menu_item_id = $1`, menuItemID); err != nil {
		return fmt.Errorf("MenuItemRepo.ReplaceIngredients delete: %w", err)
	}
	for _, ing := range ings {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO menu_item_ingredients (menu_item_id, product_id, quantity) VALUES ($1, $2, $3)
		`, menuItemID, ing.ProductID, ing.Quantity); err != nil {
			return fmt.Errorf("MenuItemRepo.ReplaceIngredients insert: %w", err)
		}
	}
	return tx.Commit()
}

func (r *MenuItemRepo) SoftDelete(ctx context.Context, id string) error {
	const q = `UPDATE menu_items SET deleted_at=NOW(), updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL`
	res, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("MenuItemRepo.SoftDelete: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("menu item not found")
	}
	return nil
}
