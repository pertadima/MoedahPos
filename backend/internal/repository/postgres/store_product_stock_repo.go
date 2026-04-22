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

// StoreRepo is the PostgreSQL implementation of repository.StoreRepository.
type StoreRepo struct {
	db *sqlx.DB
}

func NewStoreRepo(db *sqlx.DB) *StoreRepo { return &StoreRepo{db: db} }

func (r *StoreRepo) Create(ctx context.Context, s *domain.Store) (*domain.Store, error) {
	const q = `
		INSERT INTO stores (name, address, phone, tax_number, currency, store_type, default_tax_percentage, loyalty_points_per_rupiah)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, name, address, phone, tax_number, currency, store_type, default_tax_percentage, loyalty_points_per_rupiah, is_active, created_at, updated_at, deleted_at`
	row := &domain.Store{}
	err := r.db.QueryRowxContext(ctx, q, s.Name, s.Address, s.Phone, s.TaxNumber, s.Currency, s.StoreType, s.DefaultTaxPercentage, s.LoyaltyPointsPerRupiah).StructScan(row)
	if err != nil {
		return nil, fmt.Errorf("StoreRepo.Create: %w", err)
	}
	return row, nil
}

func (r *StoreRepo) FindAll(ctx context.Context, filter dto.StoreListFilter) ([]*domain.Store, int, error) { //nolint:funlen
	args := []interface{}{}
	conds := []string{"s.deleted_at IS NULL"}
	i := 1

	if filter.Search != "" {
		conds = append(conds, fmt.Sprintf("s.name ILIKE $%d", i))
		args = append(args, "%"+filter.Search+"%")
		i++
	}
	if filter.IsActive != nil {
		conds = append(conds, fmt.Sprintf("s.is_active = $%d", i))
		args = append(args, *filter.IsActive)
		i++
	}

	where := "WHERE " + strings.Join(conds, " AND ")

	// Total count
	var total int
	countQ := fmt.Sprintf("SELECT COUNT(*) FROM stores s %s", where)
	if err := r.db.QueryRowxContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("StoreRepo.FindAll count: %w", err)
	}

	// Data
	args = append(args, filter.PerPage, filter.Offset())
	dataQ := fmt.Sprintf(`
		SELECT id, name, address, phone, tax_number, currency, store_type, default_tax_percentage, loyalty_points_per_rupiah, is_active, created_at, updated_at, deleted_at
		FROM stores s %s
		ORDER BY s.created_at DESC
		LIMIT $%d OFFSET $%d`, where, i, i+1)

	var stores []*domain.Store
	if err := r.db.SelectContext(ctx, &stores, dataQ, args...); err != nil {
		return nil, 0, fmt.Errorf("StoreRepo.FindAll: %w", err)
	}
	return stores, total, nil
}

func (r *StoreRepo) FindByID(ctx context.Context, id string) (*domain.Store, error) {
	const q = `
		SELECT id, name, address, phone, tax_number, currency, store_type, default_tax_percentage, loyalty_points_per_rupiah, is_active, created_at, updated_at, deleted_at
		FROM stores WHERE id = $1 AND deleted_at IS NULL`
	s := &domain.Store{}
	if err := r.db.QueryRowxContext(ctx, q, id).StructScan(s); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("StoreRepo.FindByID: %w", err)
	}
	return s, nil
}

func (r *StoreRepo) Update(ctx context.Context, s *domain.Store) (*domain.Store, error) {
	const q = `
		UPDATE stores
		SET name=$1, address=$2, phone=$3, tax_number=$4, currency=$5, store_type=$6, default_tax_percentage=$7, loyalty_points_per_rupiah=$8, is_active=$9, updated_at=NOW()
		WHERE id=$10 AND deleted_at IS NULL
		RETURNING id, name, address, phone, tax_number, currency, store_type, default_tax_percentage, loyalty_points_per_rupiah, is_active, created_at, updated_at, deleted_at`
	row := &domain.Store{}
	err := r.db.QueryRowxContext(ctx, q, s.Name, s.Address, s.Phone, s.TaxNumber, s.Currency, s.StoreType, s.DefaultTaxPercentage, s.LoyaltyPointsPerRupiah, s.IsActive, s.ID).StructScan(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("StoreRepo.Update: %w", err)
	}
	return row, nil
}

// SoftDelete marks a store as deleted by setting deleted_at = NOW().
func (r *StoreRepo) SoftDelete(ctx context.Context, id string) error {
	const q = `UPDATE stores SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	res, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("StoreRepo.SoftDelete: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("store not found")
	}
	return nil
}

// ─── Members ──────────────────────────────────────────────────────────────────

func (r *StoreRepo) FindMember(ctx context.Context, userID, storeID string) (*domain.UserStore, error) {
	const q = `
		SELECT id, user_id, store_id, role_id, is_active, created_at
		FROM user_stores WHERE user_id = $1 AND store_id = $2`
	us := &domain.UserStore{}
	if err := r.db.QueryRowxContext(ctx, q, userID, storeID).StructScan(us); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("StoreRepo.FindMember: %w", err)
	}
	return us, nil
}

func (r *StoreRepo) ListMembers(ctx context.Context, storeID string) ([]*domain.StoreMember, error) {
	const q = `
		SELECT
			u.id         AS user_id,
			u.name       AS user_name,
			u.email      AS user_email,
			us.store_id,
			us.role_id,
			ro.name      AS role_name,
			us.is_active,
			us.created_at AS joined_at
		FROM user_stores us
		JOIN users u  ON u.id  = us.user_id  AND u.deleted_at IS NULL
		JOIN roles ro ON ro.id = us.role_id
		WHERE us.store_id = $1
		ORDER BY us.created_at DESC`
	var members []*domain.StoreMember
	if err := r.db.SelectContext(ctx, &members, q, storeID); err != nil {
		return nil, fmt.Errorf("StoreRepo.ListMembers: %w", err)
	}
	return members, nil
}

func (r *StoreRepo) AddMember(ctx context.Context, us *domain.UserStore) error {
	const q = `
		INSERT INTO user_stores (user_id, store_id, role_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, store_id) DO UPDATE SET role_id = EXCLUDED.role_id, is_active = true`
	if _, err := r.db.ExecContext(ctx, q, us.UserID, us.StoreID, us.RoleID); err != nil {
		return fmt.Errorf("StoreRepo.AddMember: %w", err)
	}
	return nil
}

func (r *StoreRepo) UpdateMemberRole(ctx context.Context, userID, storeID, roleID string) error {
	const q = `UPDATE user_stores SET role_id = $1 WHERE user_id = $2 AND store_id = $3`
	if _, err := r.db.ExecContext(ctx, q, roleID, userID, storeID); err != nil {
		return fmt.Errorf("StoreRepo.UpdateMemberRole: %w", err)
	}
	return nil
}

// DeactivateMember soft-deletes a membership by flagging is_active = false.
func (r *StoreRepo) DeactivateMember(ctx context.Context, userID, storeID string) error {
	const q = `UPDATE user_stores SET is_active = false WHERE user_id = $1 AND store_id = $2`
	res, err := r.db.ExecContext(ctx, q, userID, storeID)
	if err != nil {
		return fmt.Errorf("StoreRepo.DeactivateMember: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("member not found")
	}
	return nil
}

// ─── Category Repo ────────────────────────────────────────────────────────────

type CategoryRepo struct{ db *sqlx.DB }

func NewCategoryRepo(db *sqlx.DB) *CategoryRepo { return &CategoryRepo{db: db} }

func (r *CategoryRepo) Create(ctx context.Context, c *domain.Category) (*domain.Category, error) {
	const q = `
		INSERT INTO categories (store_id, name, parent_id)
		VALUES ($1, $2, $3)
		RETURNING id, store_id, name, parent_id, created_at, updated_at, deleted_at`
	row := &domain.Category{}
	if err := r.db.QueryRowxContext(ctx, q, c.StoreID, c.Name, c.ParentID).StructScan(row); err != nil {
		return nil, fmt.Errorf("CategoryRepo.Create: %w", err)
	}
	return row, nil
}

func (r *CategoryRepo) FindAllByStore(ctx context.Context, storeID string) ([]*domain.Category, error) {
	const q = `
		SELECT c.id, c.store_id, c.name, c.parent_id, c.created_at, c.updated_at, c.deleted_at,
		       p.name AS parent_name
		FROM categories c
		LEFT JOIN categories p ON p.id = c.parent_id AND p.deleted_at IS NULL
		WHERE c.store_id = $1 AND c.deleted_at IS NULL
		ORDER BY c.name`
	var cats []*domain.Category
	if err := r.db.SelectContext(ctx, &cats, q, storeID); err != nil {
		return nil, fmt.Errorf("CategoryRepo.FindAllByStore: %w", err)
	}
	return cats, nil
}

func (r *CategoryRepo) FindByID(ctx context.Context, id string) (*domain.Category, error) {
	const q = `
		SELECT id, store_id, name, parent_id, created_at, updated_at, deleted_at
		FROM categories WHERE id = $1 AND deleted_at IS NULL`
	c := &domain.Category{}
	if err := r.db.QueryRowxContext(ctx, q, id).StructScan(c); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("CategoryRepo.FindByID: %w", err)
	}
	return c, nil
}

func (r *CategoryRepo) Update(ctx context.Context, c *domain.Category) (*domain.Category, error) {
	const q = `
		UPDATE categories SET name=$1, parent_id=$2, updated_at=NOW()
		WHERE id=$3 AND deleted_at IS NULL
		RETURNING id, store_id, name, parent_id, created_at, updated_at, deleted_at`
	row := &domain.Category{}
	if err := r.db.QueryRowxContext(ctx, q, c.Name, c.ParentID, c.ID).StructScan(row); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("CategoryRepo.Update: %w", err)
	}
	return row, nil
}

func (r *CategoryRepo) SoftDelete(ctx context.Context, id string) error {
	const q = `UPDATE categories SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	res, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("CategoryRepo.SoftDelete: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("category not found")
	}
	return nil
}

func (r *CategoryRepo) GetModifiedSince(ctx context.Context, storeID string, since time.Time) ([]*domain.Category, error) {
	const q = `
		SELECT id, store_id, name, parent_id, created_at, updated_at, deleted_at, server_updated_at, sync_version
		FROM categories
		WHERE store_id = $1 AND server_updated_at > $2
		ORDER BY server_updated_at ASC`
	var cats []*domain.Category
	if err := r.db.SelectContext(ctx, &cats, q, storeID, since); err != nil {
		return nil, fmt.Errorf("CategoryRepo.GetModifiedSince: %w", err)
	}
	return cats, nil
}

// ─── Product Repo ─────────────────────────────────────────────────────────────

type ProductRepo struct{ db *sqlx.DB }

func NewProductRepo(db *sqlx.DB) *ProductRepo { return &ProductRepo{db: db} }

func (r *ProductRepo) Create(ctx context.Context, p *domain.Product) (*domain.Product, error) {
	const q = `
		INSERT INTO products
		  (store_id, category_id, sku, name, description, barcode, unit, cost_price, sell_price, use_global_tax, tax_percentage, image_url)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING id, store_id, category_id, sku, name, description, barcode, unit,
		          cost_price, sell_price, use_global_tax, tax_percentage, image_url, is_active, created_at, updated_at, deleted_at`
	row := &domain.Product{}
	err := r.db.QueryRowxContext(ctx, q,
		p.StoreID, p.CategoryID, p.SKU, p.Name, p.Description,
		p.Barcode, p.Unit, p.CostPrice, p.SellPrice, p.UseGlobalTax, p.TaxPercentage, p.ImageURL,
	).StructScan(row)
	if err != nil {
		return nil, fmt.Errorf("ProductRepo.Create: %w", err)
	}
	return row, nil
}

func (r *ProductRepo) FindAll(ctx context.Context, f dto.ProductListFilter) ([]*domain.Product, int, error) { //nolint:funlen
	args := []interface{}{}
	conds := []string{"p.store_id = $1", "p.deleted_at IS NULL"}
	args = append(args, f.StoreID)
	i := 2

	if f.Search != "" {
		conds = append(conds, fmt.Sprintf("(p.name ILIKE $%d OR p.sku ILIKE $%d)", i, i))
		args = append(args, "%"+f.Search+"%")
		i++
	}
	if f.CategoryID != "" {
		conds = append(conds, fmt.Sprintf("p.category_id = $%d", i))
		args = append(args, f.CategoryID)
		i++
	}
	if f.IsActive != nil {
		conds = append(conds, fmt.Sprintf("p.is_active = $%d", i))
		args = append(args, *f.IsActive)
		i++
	}

	where := "WHERE " + strings.Join(conds, " AND ")

	var total int
	if err := r.db.QueryRowxContext(ctx,
		fmt.Sprintf("SELECT COUNT(*) FROM products p %s", where), args...,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("ProductRepo.FindAll count: %w", err)
	}

	args = append(args, f.PerPage, f.Offset())
	stockJoin := ""
	stockCol := "NULL::numeric AS stock_qty"
	if f.WithStock {
		stockJoin = "LEFT JOIN stock_levels sl ON sl.product_id = p.id AND sl.store_id = p.store_id"
		stockCol = "sl.quantity AS stock_qty"
	}

	dataQ := fmt.Sprintf(`
		SELECT p.id, p.store_id, p.category_id, p.sku, p.name, p.description, p.barcode, p.unit,
		       p.cost_price, p.sell_price, p.use_global_tax, p.tax_percentage, p.image_url, p.is_active,
		       p.created_at, p.updated_at, p.deleted_at,
		       c.name AS category_name, %s,
		       s.default_tax_percentage AS store_default_tax
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id AND c.deleted_at IS NULL
		INNER JOIN stores s ON s.id = p.store_id
		%s
		%s
		ORDER BY p.name ASC
		LIMIT $%d OFFSET $%d`, stockCol, stockJoin, where, i, i+1)

	var products []*domain.Product
	if err := r.db.SelectContext(ctx, &products, dataQ, args...); err != nil {
		return nil, 0, fmt.Errorf("ProductRepo.FindAll: %w", err)
	}
	return products, total, nil
}

func (r *ProductRepo) FindByID(ctx context.Context, id string) (*domain.Product, error) {
	const q = `
		SELECT p.id, p.store_id, p.category_id, p.sku, p.name, p.description, p.barcode, p.unit,
		       p.cost_price, p.sell_price, p.use_global_tax, p.tax_percentage, p.image_url, p.is_active,
		       p.created_at, p.updated_at, p.deleted_at, c.name AS category_name,
		       s.default_tax_percentage AS store_default_tax
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id AND c.deleted_at IS NULL
		INNER JOIN stores s ON s.id = p.store_id
		WHERE p.id = $1 AND p.deleted_at IS NULL`
	p := &domain.Product{}
	if err := r.db.QueryRowxContext(ctx, q, id).StructScan(p); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("ProductRepo.FindByID: %w", err)
	}
	return p, nil
}

func (r *ProductRepo) FindByBarcode(ctx context.Context, storeID, barcode string) (*domain.Product, error) {
	const q = `
		SELECT p.id, p.store_id, p.category_id, p.sku, p.name, p.description, p.barcode, p.unit,
		       p.cost_price, p.sell_price, p.use_global_tax, p.tax_percentage, p.image_url, p.is_active,
		       p.created_at, p.updated_at, p.deleted_at, c.name AS category_name,
		       s.default_tax_percentage AS store_default_tax
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id AND c.deleted_at IS NULL
		INNER JOIN stores s ON s.id = p.store_id
		WHERE p.store_id = $1 AND p.barcode = $2 AND p.deleted_at IS NULL AND p.is_active = true`
	p := &domain.Product{}
	if err := r.db.QueryRowxContext(ctx, q, storeID, barcode).StructScan(p); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("ProductRepo.FindByBarcode: %w", err)
	}
	return p, nil
}

func (r *ProductRepo) ExistsBySKU(ctx context.Context, storeID, sku, excludeID string) (bool, error) {
	q := `SELECT EXISTS(SELECT 1 FROM products WHERE store_id=$1 AND sku=$2 AND deleted_at IS NULL`
	args := []interface{}{storeID, sku}
	if excludeID != "" {
		q += " AND id != $3"
		args = append(args, excludeID)
	}
	q += ")"
	var exists bool
	if err := r.db.QueryRowxContext(ctx, q, args...).Scan(&exists); err != nil {
		return false, fmt.Errorf("ProductRepo.ExistsBySKU: %w", err)
	}
	return exists, nil
}

func (r *ProductRepo) Update(ctx context.Context, p *domain.Product) (*domain.Product, error) {
	const q = `
		UPDATE products
		SET category_id=$1, name=$2, description=$3, barcode=$4, unit=$5,
		    cost_price=$6, sell_price=$7, use_global_tax=$8, tax_percentage=$9, image_url=$10, is_active=$11, updated_at=NOW()
		WHERE id=$12 AND deleted_at IS NULL
		RETURNING id, store_id, category_id, sku, name, description, barcode, unit,
		          cost_price, sell_price, use_global_tax, tax_percentage, image_url, is_active, created_at, updated_at, deleted_at`
	row := &domain.Product{}
	err := r.db.QueryRowxContext(ctx, q,
		p.CategoryID, p.Name, p.Description, p.Barcode, p.Unit,
		p.CostPrice, p.SellPrice, p.UseGlobalTax, p.TaxPercentage, p.ImageURL, p.IsActive, p.ID,
	).StructScan(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("ProductRepo.Update: %w", err)
	}
	return row, nil
}

func (r *ProductRepo) SoftDelete(ctx context.Context, id string) error {
	const q = `UPDATE products SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	res, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("ProductRepo.SoftDelete: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("product not found")
	}
	return nil
}

func (r *ProductRepo) GetModifiedSince(ctx context.Context, storeID string, since time.Time) ([]*domain.Product, error) {
	const q = `
		SELECT id, store_id, category_id, sku, name, description, barcode, unit,
		       cost_price, sell_price, use_global_tax, tax_percentage, image_url, is_active,
		       created_at, updated_at, deleted_at, server_updated_at, sync_version
		FROM products
		WHERE store_id = $1 AND server_updated_at > $2
		ORDER BY server_updated_at ASC`
	var products []*domain.Product
	if err := r.db.SelectContext(ctx, &products, q, storeID, since); err != nil {
		return nil, fmt.Errorf("ProductRepo.GetModifiedSince: %w", err)
	}
	return products, nil
}

// ─── Stock Repo ───────────────────────────────────────────────────────────────

type StockRepo struct{ db *sqlx.DB }

func NewStockRepo(db *sqlx.DB) *StockRepo { return &StockRepo{db: db} }

func (r *StockRepo) FindLevelsByStore(ctx context.Context, storeID string, lowStockOnly bool) ([]*domain.StockLevel, error) {
	q := `
		SELECT sl.id, sl.product_id, sl.store_id, sl.quantity, sl.min_quantity, sl.updated_at,
		       p.name AS product_name, p.sku AS product_sku, p.unit, p.cost_price
		FROM stock_levels sl
		JOIN products p ON p.id = sl.product_id AND p.deleted_at IS NULL AND p.is_active = true
		WHERE sl.store_id = $1`
	if lowStockOnly {
		q += " AND sl.quantity <= sl.min_quantity"
	}
	q += " ORDER BY p.name ASC"

	var levels []*domain.StockLevel
	if err := r.db.SelectContext(ctx, &levels, q, storeID); err != nil {
		return nil, fmt.Errorf("StockRepo.FindLevelsByStore: %w", err)
	}
	return levels, nil
}

func (r *StockRepo) FindLevelByProduct(ctx context.Context, productID, storeID string) (*domain.StockLevel, error) {
	const q = `
		SELECT sl.id, sl.product_id, sl.store_id, sl.quantity, sl.min_quantity, sl.updated_at,
		       p.name AS product_name, p.sku AS product_sku, p.unit
		FROM stock_levels sl
		JOIN products p ON p.id = sl.product_id
		WHERE sl.product_id = $1 AND sl.store_id = $2`
	sl := &domain.StockLevel{}
	if err := r.db.QueryRowxContext(ctx, q, productID, storeID).StructScan(sl); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("StockRepo.FindLevelByProduct: %w", err)
	}
	return sl, nil
}

func (r *StockRepo) SetMinQuantity(ctx context.Context, productID, storeID string, min float64) error {
	const q = `
		INSERT INTO stock_levels (product_id, store_id, min_quantity, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (product_id, store_id) DO UPDATE SET min_quantity = $3, updated_at = NOW()`
	if _, err := r.db.ExecContext(ctx, q, productID, storeID, min); err != nil {
		return fmt.Errorf("StockRepo.SetMinQuantity: %w", err)
	}
	return nil
}

// Adjust atomically inserts a stock_movement and upserts stock_levels in a DB transaction.
func (r *StockRepo) Adjust(ctx context.Context, input domain.AdjustInput) (*domain.StockLevel, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("StockRepo.Adjust begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// 1. Insert movement record
	const mvQ = `
		INSERT INTO stock_movements (product_id, store_id, ref_type, ref_id, quantity_delta, notes, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	if _, err := tx.ExecContext(ctx, mvQ,
		input.ProductID, input.StoreID, input.RefType, input.RefID,
		input.Delta, input.Notes, input.CreatedBy,
	); err != nil {
		return nil, fmt.Errorf("StockRepo.Adjust insert movement: %w", err)
	}

	// 2. Upsert stock level
	const slQ = `
		INSERT INTO stock_levels (product_id, store_id, quantity, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (product_id, store_id)
		DO UPDATE SET quantity = stock_levels.quantity + $3, updated_at = NOW()
		RETURNING id, product_id, store_id, quantity, min_quantity, updated_at`
	sl := &domain.StockLevel{}
	if err := tx.QueryRowxContext(ctx, slQ, input.ProductID, input.StoreID, input.Delta).StructScan(sl); err != nil {
		return nil, fmt.Errorf("StockRepo.Adjust upsert level: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("StockRepo.Adjust commit: %w", err)
	}

	// Populate product info
	const pQ = `SELECT name AS product_name, sku AS product_sku, unit FROM products WHERE id = $1`
	_ = r.db.QueryRowxContext(ctx, pQ, input.ProductID).StructScan(sl)

	return sl, nil
}

func (r *StockRepo) FindMovements(ctx context.Context, f dto.StockMovementFilter) ([]*domain.StockMovement, int, error) {
	args := []interface{}{}
	conds := []string{"sm.store_id = $1"}
	args = append(args, f.StoreID)
	i := 2

	if f.ProductID != "" {
		conds = append(conds, fmt.Sprintf("sm.product_id = $%d", i))
		args = append(args, f.ProductID)
		i++
	}
	if f.RefType != "" {
		conds = append(conds, fmt.Sprintf("sm.ref_type = $%d", i))
		args = append(args, f.RefType)
		i++
	}

	where := "WHERE " + strings.Join(conds, " AND ")

	var total int
	if err := r.db.QueryRowxContext(ctx,
		fmt.Sprintf("SELECT COUNT(*) FROM stock_movements sm %s", where), args...,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("StockRepo.FindMovements count: %w", err)
	}

	args = append(args, f.PerPage, f.Offset())
	dataQ := fmt.Sprintf(`
		SELECT sm.id, sm.product_id, sm.store_id, sm.ref_type, sm.ref_id,
		       sm.quantity_delta, sm.notes, sm.created_by, sm.created_at,
		       p.name AS product_name, u.name AS created_by_name
		FROM stock_movements sm
		JOIN products p ON p.id = sm.product_id
		JOIN users    u ON u.id = sm.created_by
		%s
		ORDER BY sm.created_at DESC
		LIMIT $%d OFFSET $%d`, where, i, i+1)

	var movements []*domain.StockMovement
	if err := r.db.SelectContext(ctx, &movements, dataQ, args...); err != nil {
		return nil, 0, fmt.Errorf("StockRepo.FindMovements: %w", err)
	}
	return movements, total, nil
}

// DeductStock subtracts qty from stock_levels and records a 'sale' stock_movement.
// Used for menu item ingredient deduction in restaurant checkouts.
func (r *StockRepo) DeductStock(ctx context.Context, productID, storeID string, qty float64, refID, cashierID string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("StockRepo.DeductStock begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	const mvQ = `
		INSERT INTO stock_movements (product_id, store_id, ref_type, ref_id, quantity_delta, notes, created_by)
		VALUES ($1, $2, 'sale', $3, $4, 'Penjualan menu restoran', $5)`
	if _, err := tx.ExecContext(ctx, mvQ, productID, storeID, refID, -qty, cashierID); err != nil {
		return fmt.Errorf("StockRepo.DeductStock insert movement: %w", err)
	}

	const slQ = `
		UPDATE stock_levels
		SET quantity = GREATEST(0, quantity - $1), updated_at = NOW()
		WHERE product_id = $2 AND store_id = $3`
	if _, err := tx.ExecContext(ctx, slQ, qty, productID, storeID); err != nil {
		return fmt.Errorf("StockRepo.DeductStock update level: %w", err)
	}

	return tx.Commit()
}

func (r *StockRepo) GetModifiedSince(ctx context.Context, storeID string, since time.Time) ([]*domain.StockLevel, error) {
	const q = `
		SELECT id, product_id, store_id, quantity, min_quantity, updated_at, server_updated_at, sync_version
		FROM stock_levels
		WHERE store_id = $1 AND server_updated_at > $2
		ORDER BY server_updated_at ASC`
	var levels []*domain.StockLevel
	if err := r.db.SelectContext(ctx, &levels, q, storeID, since); err != nil {
		return nil, fmt.Errorf("StockRepo.GetModifiedSince: %w", err)
	}
	return levels, nil
}

// Ensure unused import does not cause issues
var _ = time.Now
