package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/moedahpos/backend/internal/domain"
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

// EarnPoints inserts a positive EARN entry into the ledger.
func (r *LoyaltyRepo) EarnPoints(ctx context.Context, customerID string, transactionID *string, points float64) (*domain.LoyaltyLedger, error) {
	const q = `
		INSERT INTO loyalty_ledger (customer_id, points_delta, transaction_id, type)
		VALUES ($1, $2, $3, 'EARN')
		RETURNING id, customer_id, points_delta, transaction_id, type, created_at`
	entry := &domain.LoyaltyLedger{}
	if err := r.db.QueryRowxContext(ctx, q, customerID, points, transactionID).StructScan(entry); err != nil {
		return nil, fmt.Errorf("LoyaltyRepo.EarnPoints: %w", err)
	}
	return entry, nil
}

// SpendPoints inserts a negative SPEND entry into the ledger.
// The caller is responsible for ensuring the balance is sufficient.
func (r *LoyaltyRepo) SpendPoints(ctx context.Context, customerID string, transactionID *string, points float64) (*domain.LoyaltyLedger, error) {
	const q = `
		INSERT INTO loyalty_ledger (customer_id, points_delta, transaction_id, type)
		VALUES ($1, $2, $3, 'SPEND')
		RETURNING id, customer_id, points_delta, transaction_id, type, created_at`
	// Store as negative delta
	entry := &domain.LoyaltyLedger{}
	if err := r.db.QueryRowxContext(ctx, q, customerID, -points, transactionID).StructScan(entry); err != nil {
		return nil, fmt.Errorf("LoyaltyRepo.SpendPoints: %w", err)
	}
	return entry, nil
}

// GetHistory returns all ledger entries for a customer, newest first.
func (r *LoyaltyRepo) GetHistory(ctx context.Context, customerID string) ([]*domain.LoyaltyLedger, error) {
	const q = `
		SELECT id, customer_id, points_delta, transaction_id, type, created_at
		FROM loyalty_ledger
		WHERE customer_id = $1
		ORDER BY created_at DESC`
	var entries []*domain.LoyaltyLedger
	if err := r.db.SelectContext(ctx, &entries, q, customerID); err != nil {
		return nil, fmt.Errorf("LoyaltyRepo.GetHistory: %w", err)
	}
	return entries, nil
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
	GetHistory(context.Context, string) ([]*domain.LoyaltyLedger, error)
	AssignTier(context.Context, string, string) error
	GetCustomerTier(context.Context, string) (*domain.MembershipTier, error)
} = (*LoyaltyRepo)(nil)
