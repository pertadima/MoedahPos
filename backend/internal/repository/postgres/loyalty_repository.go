package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
	"github.com/moedahpos/backend/internal/repository"
)

// ─── MembershipTierRepo ───────────────────────────────────────────────────────

// MembershipTierRepo is the PostgreSQL implementation for membership tiers.
type MembershipTierRepo struct{ db *sqlx.DB }

// NewMembershipTierRepository creates a new MembershipTierRepository.
func NewMembershipTierRepository(db *sql.DB) repository.MembershipTierRepository {
	return &MembershipTierRepo{db: sqlx.NewDb(db, "postgres")}
}

func (r *MembershipTierRepo) FindAll(ctx context.Context) ([]*domain.MembershipTier, error) {
	const q = `SELECT id, name, multiplier, created_at, updated_at FROM membership_tiers ORDER BY multiplier ASC`
	var tiers []*domain.MembershipTier
	if err := r.db.SelectContext(ctx, &tiers, q); err != nil {
		return nil, fmt.Errorf("MembershipTierRepo.FindAll: %w", err)
	}
	return tiers, nil
}

func (r *MembershipTierRepo) FindByID(ctx context.Context, id string) (*domain.MembershipTier, error) {
	const q = `SELECT id, name, multiplier, created_at, updated_at FROM membership_tiers WHERE id = $1`
	t := &domain.MembershipTier{}
	if err := r.db.QueryRowxContext(ctx, q, id).StructScan(t); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("MembershipTierRepo.FindByID: %w", err)
	}
	return t, nil
}

// ─── LoyaltyRepo ─────────────────────────────────────────────────────────────

// LoyaltyRepo is the PostgreSQL implementation for the loyalty ledger.
type LoyaltyRepo struct{ db *sqlx.DB }

// NewLoyaltyRepository creates a new LoyaltyRepository.
func NewLoyaltyRepository(db *sql.DB) repository.LoyaltyRepository {
	return &LoyaltyRepo{db: sqlx.NewDb(db, "postgres")}
}

// GetBalance returns the algebraic sum of all points_delta entries for the customer.
func (r *LoyaltyRepo) GetBalance(ctx context.Context, customerID string) (float64, error) {
	const q = `SELECT COALESCE(SUM(points_delta), 0) FROM loyalty_ledger WHERE customer_id = $1`
	var balance float64
	if err := r.db.QueryRowxContext(ctx, q, customerID).Scan(&balance); err != nil {
		return 0, fmt.Errorf("LoyaltyRepo.GetBalance: %w", err)
	}
	return balance, nil
}

// EarnPoints inserts a positive EARN entry into the ledger with balance snapshot.
func (r *LoyaltyRepo) EarnPoints(ctx context.Context, customerID string, transactionID *string, points float64) (*domain.LoyaltyLedger, error) {
	const q = `
		INSERT INTO loyalty_ledger (customer_id, points_delta, transaction_id, type, balance_snapshot)
		VALUES ($1, $2, $3, 'EARN',
			COALESCE((SELECT SUM(points_delta) FROM loyalty_ledger WHERE customer_id = $1), 0) + $2)
		RETURNING id, customer_id, points_delta, transaction_id, type, balance_snapshot, created_at`
	entry := &domain.LoyaltyLedger{}
	if err := r.db.QueryRowxContext(ctx, q, customerID, points, transactionID).StructScan(entry); err != nil {
		return nil, fmt.Errorf("LoyaltyRepo.EarnPoints: %w", err)
	}
	return entry, nil
}

// SpendPoints inserts a negative SPEND entry into the ledger with balance snapshot.
// The caller is responsible for ensuring the balance is sufficient.
func (r *LoyaltyRepo) SpendPoints(ctx context.Context, customerID string, transactionID *string, points float64) (*domain.LoyaltyLedger, error) {
	const q = `
		INSERT INTO loyalty_ledger (customer_id, points_delta, transaction_id, type, balance_snapshot)
		VALUES ($1, $2, $3, 'SPEND',
			COALESCE((SELECT SUM(points_delta) FROM loyalty_ledger WHERE customer_id = $1), 0) + $2)
		RETURNING id, customer_id, points_delta, transaction_id, type, balance_snapshot, created_at`
	// Store as negative delta
	entry := &domain.LoyaltyLedger{}
	if err := r.db.QueryRowxContext(ctx, q, customerID, -points, transactionID).StructScan(entry); err != nil {
		return nil, fmt.Errorf("LoyaltyRepo.SpendPoints: %w", err)
	}
	return entry, nil
}

// VoidPoints inserts a VOID entry that revokes points earned by a specific transaction.
func (r *LoyaltyRepo) VoidPoints(ctx context.Context, customerID string, transactionID *string, points float64) (*domain.LoyaltyLedger, error) {
	const q = `
		INSERT INTO loyalty_ledger (customer_id, points_delta, transaction_id, type, balance_snapshot)
		VALUES ($1, $2, $3, 'VOID',
			COALESCE((SELECT SUM(points_delta) FROM loyalty_ledger WHERE customer_id = $1), 0) + $2)
		RETURNING id, customer_id, points_delta, transaction_id, type, balance_snapshot, created_at`
	// Negative delta to revoke earned points
	entry := &domain.LoyaltyLedger{}
	if err := r.db.QueryRowxContext(ctx, q, customerID, -points, transactionID).StructScan(entry); err != nil {
		return nil, fmt.Errorf("LoyaltyRepo.VoidPoints: %w", err)
	}
	return entry, nil
}

// AdjustPoints inserts a manual ADJUST entry (positive or negative delta).
func (r *LoyaltyRepo) AdjustPoints(ctx context.Context, customerID string, delta float64, _ string) (*domain.LoyaltyLedger, error) {
	const q = `
		INSERT INTO loyalty_ledger (customer_id, points_delta, transaction_id, type, balance_snapshot)
		VALUES ($1, $2, NULL, 'ADJUST',
			COALESCE((SELECT SUM(points_delta) FROM loyalty_ledger WHERE customer_id = $1), 0) + $2)
		RETURNING id, customer_id, points_delta, transaction_id, type, balance_snapshot, created_at`
	entry := &domain.LoyaltyLedger{}
	if err := r.db.QueryRowxContext(ctx, q, customerID, delta).StructScan(entry); err != nil {
		return nil, fmt.Errorf("LoyaltyRepo.AdjustPoints: %w", err)
	}
	return entry, nil
}

// GetHistory returns all ledger entries for a customer, newest first.
func (r *LoyaltyRepo) GetHistory(ctx context.Context, customerID string) ([]*domain.LoyaltyLedger, error) {
	const q = `
		SELECT id, customer_id, points_delta, transaction_id, type, balance_snapshot, created_at
		FROM loyalty_ledger
		WHERE customer_id = $1
		ORDER BY created_at DESC`
	var entries []*domain.LoyaltyLedger
	if err := r.db.SelectContext(ctx, &entries, q, customerID); err != nil {
		return nil, fmt.Errorf("LoyaltyRepo.GetHistory: %w", err)
	}
	return entries, nil
}

// GetHistoryPaginated returns a paginated list of ledger entries for a customer.
func (r *LoyaltyRepo) GetHistoryPaginated(ctx context.Context, customerID string, page, perPage int) ([]*domain.LoyaltyLedger, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	var total int
	if err := r.db.QueryRowxContext(ctx,
		`SELECT COUNT(*) FROM loyalty_ledger WHERE customer_id = $1`, customerID,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("LoyaltyRepo.GetHistoryPaginated count: %w", err)
	}

	const q = `
		SELECT id, customer_id, points_delta, transaction_id, type, balance_snapshot, created_at
		FROM loyalty_ledger
		WHERE customer_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`
	var entries []*domain.LoyaltyLedger
	if err := r.db.SelectContext(ctx, &entries, q, customerID, perPage, offset); err != nil {
		return nil, 0, fmt.Errorf("LoyaltyRepo.GetHistoryPaginated: %w", err)
	}
	return entries, total, nil
}

// AssignTier updates the customer's loyalty_tier_id and bumps updated_at.
func (r *LoyaltyRepo) AssignTier(ctx context.Context, customerID, tierID string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE customers SET loyalty_tier_id = $1, updated_at = NOW() WHERE id = $2 AND deleted_at IS NULL`,
		tierID, customerID,
	)
	if err != nil {
		return fmt.Errorf("LoyaltyRepo.AssignTier: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("customer %s not found", customerID)
	}
	return nil
}

// GetCustomerTier fetches the tier for a given customer (nil if none assigned).
func (r *LoyaltyRepo) GetCustomerTier(ctx context.Context, customerID string) (*domain.MembershipTier, error) {
	const q = `
		SELECT mt.id, mt.name, mt.multiplier, mt.created_at, mt.updated_at
		FROM membership_tiers mt
		JOIN customers c ON c.loyalty_tier_id = mt.id
		WHERE c.id = $1 AND c.deleted_at IS NULL`
	t := &domain.MembershipTier{}
	if err := r.db.QueryRowxContext(ctx, q, customerID).StructScan(t); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("LoyaltyRepo.GetCustomerTier: %w", err)
	}
	return t, nil
}

// Ensure LoyaltyRepo satisfies the interface at compile time.
var _ interface {
	GetBalance(context.Context, string) (float64, error)
	EarnPoints(context.Context, string, *string, float64) (*domain.LoyaltyLedger, error)
	SpendPoints(context.Context, string, *string, float64) (*domain.LoyaltyLedger, error)
	VoidPoints(context.Context, string, *string, float64) (*domain.LoyaltyLedger, error)
	AdjustPoints(context.Context, string, float64, string) (*domain.LoyaltyLedger, error)
	GetHistory(context.Context, string) ([]*domain.LoyaltyLedger, error)
	GetHistoryPaginated(context.Context, string, int, int) ([]*domain.LoyaltyLedger, int, error)
	AssignTier(context.Context, string, string) error
	GetCustomerTier(context.Context, string) (*domain.MembershipTier, error)
	GetTopCustomersByBalance(context.Context, string, int) ([]dto.TopCustomerLoyalty, error)
	GetPointsSummary(context.Context, string, time.Time, time.Time) (float64, float64, error)
} = (*LoyaltyRepo)(nil)

// GetTopCustomersByBalance returns the top N customers by current loyalty balance for a store.
func (r *LoyaltyRepo) GetTopCustomersByBalance(ctx context.Context, storeID string, limit int) ([]dto.TopCustomerLoyalty, error) {
	const q = `
		SELECT
			c.id                     AS customer_id,
			c.name                   AS customer_name,
			COALESCE(bal.balance, 0) AS balance,
			mt.name                  AS tier_name,
			mt.multiplier            AS tier_multiplier
		FROM customers c
		LEFT JOIN LATERAL (
			SELECT COALESCE(SUM(points_delta), 0) AS balance
			FROM loyalty_ledger
			WHERE customer_id = c.id
		) bal ON true
		LEFT JOIN membership_tiers mt ON mt.id = c.loyalty_tier_id
		WHERE c.store_id = $1
		  AND c.deleted_at IS NULL
		  AND COALESCE(bal.balance, 0) > 0
		ORDER BY balance DESC
		LIMIT $2`

	type row struct {
		CustomerID   string   `db:"customer_id"`
		CustomerName string   `db:"customer_name"`
		Balance      float64  `db:"balance"`
		TierName     *string  `db:"tier_name"`
		TierMult     *float64 `db:"tier_multiplier"`
	}
	var rows []row
	if err := r.db.SelectContext(ctx, &rows, q, storeID, limit); err != nil {
		return nil, fmt.Errorf("LoyaltyRepo.GetTopCustomersByBalance: %w", err)
	}
	out := make([]dto.TopCustomerLoyalty, 0, len(rows))
	for _, row := range rows {
		out = append(out, dto.TopCustomerLoyalty{
			CustomerID:   row.CustomerID,
			CustomerName: row.CustomerName,
			Balance:      row.Balance,
			TierName:     row.TierName,
			TierMult:     row.TierMult,
		})
	}
	return out, nil
}

// GetPointsSummary returns total earned and total used (absolute) within [from, to) for a store.
func (r *LoyaltyRepo) GetPointsSummary(ctx context.Context, storeID string, from, to time.Time) (earned, used float64, err error) {
	const q = `
		SELECT
			COALESCE(SUM(CASE WHEN ll.points_delta > 0 THEN ll.points_delta  ELSE 0 END), 0) AS earned,
			COALESCE(SUM(CASE WHEN ll.points_delta < 0 THEN -ll.points_delta ELSE 0 END), 0) AS used
		FROM loyalty_ledger ll
		JOIN customers c ON c.id = ll.customer_id
		WHERE c.store_id = $1
		  AND ll.created_at >= $2
		  AND ll.created_at < $3`

	var res struct {
		Earned float64 `db:"earned"`
		Used   float64 `db:"used"`
	}
	if err = r.db.QueryRowxContext(ctx, q, storeID, from, to).StructScan(&res); err != nil {
		return 0, 0, fmt.Errorf("LoyaltyRepo.GetPointsSummary: %w", err)
	}
	return res.Earned, res.Used, nil
}
