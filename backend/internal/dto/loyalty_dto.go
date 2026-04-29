package dto

// ─── Membership Tier ──────────────────────────────────────────────────────────

// MembershipTierResponse is the API representation of a membership tier.
type MembershipTierResponse struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Multiplier float64 `json:"multiplier"`
}

// ─── Loyalty Balance ─────────────────────────────────────────────────────────

// LoyaltyBalanceResponse returns a customer's current point balance and tier info.
type LoyaltyBalanceResponse struct {
	CustomerID string                  `json:"customer_id"`
	Balance    float64                 `json:"balance"`
	Tier       *MembershipTierResponse `json:"tier,omitempty"`
}

// ─── Point Events ────────────────────────────────────────────────────────────

// EarnPointsRequest issued after a completed checkout to credit loyalty points.
type EarnPointsRequest struct {
	TransactionID string  `json:"transaction_id" validate:"required,uuid"`
	Total         float64 `json:"total"          validate:"required,gte=0"`
}

// RedeemPointsRequest issued during checkout to debit loyalty points.
type RedeemPointsRequest struct {
	Points float64 `json:"points" validate:"required,gt=0"`
}

// AdjustPointsRequest for manual admin corrections (positive or negative delta).
type AdjustPointsRequest struct {
	Delta float64 `json:"delta" validate:"required"`
	Note  string  `json:"note"  validate:"required,max=500"`
}

// VoidPointsRequest issued when a transaction is voided to revoke its points.
type VoidPointsRequest struct {
	TransactionID  string  `json:"transaction_id"  validate:"required,uuid"`
	OriginalPoints float64 `json:"original_points" validate:"required,gt=0"`
}

// LoyaltyLedgerResponse is the API shape of a single ledger entry.
type LoyaltyLedgerResponse struct {
	ID              string   `json:"id"`
	CustomerID      string   `json:"customer_id"`
	PointsDelta     float64  `json:"points_delta"`
	TransactionID   *string  `json:"transaction_id,omitempty"`
	Type            string   `json:"type"` // EARN | SPEND | VOID | ADJUST
	BalanceSnapshot *float64 `json:"balance_snapshot,omitempty"`
	CreatedAt       string   `json:"created_at"`
}

// LoyaltyHistoryResponse wraps a paginated list of ledger entries.
type LoyaltyHistoryResponse struct {
	Data []*LoyaltyLedgerResponse `json:"data"`
	Meta PaginationMeta           `json:"meta"`
}

// AssignTierRequest assigns a loyalty tier to a customer.
type AssignTierRequest struct {
	TierID string `json:"tier_id" validate:"required,uuid"`
}
