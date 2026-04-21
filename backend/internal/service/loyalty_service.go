package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/rs/zerolog"

	"github.com/moedahpos/backend/internal/domain"
	"github.com/moedahpos/backend/internal/dto"
	"github.com/moedahpos/backend/internal/repository"
)

// Loyalty-specific sentinel errors.
var (
	// ErrInsufficientPoints is returned when a customer tries to spend more points than they have.
	ErrInsufficientPoints = errors.New("insufficient loyalty points")
	// ErrInvalidRedemption is returned when the redemption amount is zero or negative.
	ErrInvalidRedemption = errors.New("redemption amount must be greater than zero")
	// ErrTierNotFound is returned when the target tier does not exist.
	ErrTierNotFound = errors.New("membership tier not found")
)

// pointsPerUnit is the base: 1 point is earned for every 1000 IDR spent.
const pointsPerUnit = 1000.0

// LoyaltyService handles all membership tier and loyalty point business logic.
type LoyaltyService struct {
	loyaltyRepo repository.LoyaltyRepository
	tierRepo    repository.MembershipTierRepository
	log         zerolog.Logger
}

// NewLoyaltyService constructs a LoyaltyService.
func NewLoyaltyService(
	loyaltyRepo repository.LoyaltyRepository,
	tierRepo repository.MembershipTierRepository,
	_ repository.CustomerRepository, // reserved for future use
	log zerolog.Logger,
) *LoyaltyService {
	return &LoyaltyService{
		loyaltyRepo: loyaltyRepo,
		tierRepo:    tierRepo,
		log:         log,
	}
}

// ─── Pure Calculation ─────────────────────────────────────────────────────────

// CalculatePoints computes points to award for a given transaction total and multiplier.
// Formula: floor(total / pointsPerUnit) × multiplier.
// A future-dated total has no special casing — the formula is always applied as-is.
func CalculatePoints(total, multiplier float64) float64 {
	if total <= 0 || multiplier <= 0 {
		return 0
	}
	base := math.Floor(total / pointsPerUnit)
	return base * multiplier
}

// ─── Service Methods ──────────────────────────────────────────────────────────

// ListTiers returns all configured membership tiers.
func (s *LoyaltyService) ListTiers(ctx context.Context) ([]*dto.MembershipTierResponse, error) {
	tiers, err := s.tierRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("LoyaltyService.ListTiers: %w", err)
	}
	out := make([]*dto.MembershipTierResponse, 0, len(tiers))
	for _, t := range tiers {
		out = append(out, toTierResponse(t))
	}
	return out, nil
}

// GetBalance returns the current loyalty point balance and tier info for a customer.
func (s *LoyaltyService) GetBalance(ctx context.Context, customerID string) (*dto.LoyaltyBalanceResponse, error) {
	balance, err := s.loyaltyRepo.GetBalance(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("LoyaltyService.GetBalance: %w", err)
	}

	tier, err := s.loyaltyRepo.GetCustomerTier(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("LoyaltyService.GetBalance tier: %w", err)
	}

	resp := &dto.LoyaltyBalanceResponse{
		CustomerID: customerID,
		Balance:    balance,
	}
	if tier != nil {
		resp.Tier = toTierResponse(tier)
	}
	return resp, nil
}

// EarnPoints calculates and credits loyalty points after a completed transaction.
// transactionID is optional (nil for manual adjustments).
func (s *LoyaltyService) EarnPoints(ctx context.Context, customerID string, transactionID *string, total float64) (*dto.LoyaltyLedgerResponse, error) {
	// Resolve the customer's tier multiplier
	tier, err := s.loyaltyRepo.GetCustomerTier(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("LoyaltyService.EarnPoints: resolve tier: %w", err)
	}
	multiplier := 1.0
	if tier != nil {
		multiplier = tier.Multiplier
	}

	points := CalculatePoints(total, multiplier)
	if points <= 0 {
		s.log.Debug().
			Str("customer_id", customerID).
			Float64("total", total).
			Float64("multiplier", multiplier).
			Msg("LoyaltyService.EarnPoints: zero points, skipping ledger entry")
		// Return a zero-delta response without persisting
		return &dto.LoyaltyLedgerResponse{
			CustomerID:    customerID,
			PointsDelta:   0,
			TransactionID: transactionID,
			Type:          "EARN",
			CreatedAt:     time.Now().Format(time.RFC3339),
		}, nil
	}

	entry, err := s.loyaltyRepo.EarnPoints(ctx, customerID, transactionID, points)
	if err != nil {
		s.log.Error().Err(err).Str("customer_id", customerID).Msg("LoyaltyService.EarnPoints: db error")
		return nil, fmt.Errorf("LoyaltyService.EarnPoints: %w", err)
	}

	s.log.Info().
		Str("customer_id", customerID).
		Float64("points", points).
		Float64("multiplier", multiplier).
		Msg("loyalty points earned")

	return toLedgerResponse(entry), nil
}

// RedeemPoints deducts loyalty points for a customer during checkout.
// Returns ErrInsufficientPoints if balance < points requested.
// Returns ErrInvalidRedemption if points <= 0.
func (s *LoyaltyService) RedeemPoints(ctx context.Context, customerID string, transactionID *string, points float64) (*dto.LoyaltyLedgerResponse, error) {
	if points <= 0 {
		return nil, ErrInvalidRedemption
	}

	balance, err := s.loyaltyRepo.GetBalance(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("LoyaltyService.RedeemPoints: get balance: %w", err)
	}
	if balance < points {
		return nil, fmt.Errorf("%w: balance=%.2f, requested=%.2f", ErrInsufficientPoints, balance, points)
	}

	entry, err := s.loyaltyRepo.SpendPoints(ctx, customerID, transactionID, points)
	if err != nil {
		s.log.Error().Err(err).Str("customer_id", customerID).Float64("points", points).Msg("LoyaltyService.RedeemPoints: db error")
		return nil, fmt.Errorf("LoyaltyService.RedeemPoints: %w", err)
	}

	s.log.Info().
		Str("customer_id", customerID).
		Float64("points_spent", points).
		Float64("remaining_balance", balance-points).
		Msg("loyalty points redeemed")

	return toLedgerResponse(entry), nil
}

// GetHistory returns the full ledger for a customer (newest first).
func (s *LoyaltyService) GetHistory(ctx context.Context, customerID string) ([]*dto.LoyaltyLedgerResponse, error) {
	entries, err := s.loyaltyRepo.GetHistory(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("LoyaltyService.GetHistory: %w", err)
	}
	out := make([]*dto.LoyaltyLedgerResponse, 0, len(entries))
	for _, e := range entries {
		out = append(out, toLedgerResponse(e))
	}
	return out, nil
}

// AssignTier links a customer to a membership tier.
func (s *LoyaltyService) AssignTier(ctx context.Context, customerID, tierID string) error {
	// Validate tier exists
	tier, err := s.tierRepo.FindByID(ctx, tierID)
	if err != nil {
		return fmt.Errorf("LoyaltyService.AssignTier: find tier: %w", err)
	}
	if tier == nil {
		return ErrTierNotFound
	}

	if err := s.loyaltyRepo.AssignTier(ctx, customerID, tierID); err != nil {
		return fmt.Errorf("LoyaltyService.AssignTier: %w", err)
	}

	s.log.Info().
		Str("customer_id", customerID).
		Str("tier_id", tierID).
		Str("tier_name", tier.Name).
		Msg("membership tier assigned")
	return nil
}

// ─── Mappers ──────────────────────────────────────────────────────────────────

func toTierResponse(t *domain.MembershipTier) *dto.MembershipTierResponse {
	return &dto.MembershipTierResponse{
		ID:         t.ID,
		Name:       t.Name,
		Multiplier: t.Multiplier,
	}
}

func toLedgerResponse(e *domain.LoyaltyLedger) *dto.LoyaltyLedgerResponse {
	return &dto.LoyaltyLedgerResponse{
		ID:            e.ID,
		CustomerID:    e.CustomerID,
		PointsDelta:   e.PointsDelta,
		TransactionID: e.TransactionID,
		Type:          e.Type,
		CreatedAt:     e.CreatedAt.Format(time.RFC3339),
	}
}
