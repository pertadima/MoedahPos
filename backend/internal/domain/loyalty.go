package domain

import "time"

// MembershipTier defines a named loyalty tier with a multiplier applied when earning points.
type MembershipTier struct {
	ID         string    `db:"id"`
	Name       string    `db:"name"`
	Multiplier float64   `db:"multiplier"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

// LoyaltyLedger is an immutable point event for a customer.
// points_delta: positive = credit (EARN/ADJUST+), negative = debit (SPEND/VOID/ADJUST-).
type LoyaltyLedger struct {
	ID              string    `db:"id"`
	CustomerID      string    `db:"customer_id"`
	PointsDelta     float64   `db:"points_delta"`
	TransactionID   *string   `db:"transaction_id"`
	Type            string    `db:"type"`             // EARN | SPEND | VOID | ADJUST
	BalanceSnapshot *float64  `db:"balance_snapshot"` // balance after this entry
	CreatedAt       time.Time `db:"created_at"`
}

// LoyaltyLedgerType constants.
const (
	LedgerTypeEarn   = "EARN"
	LedgerTypeSpend  = "SPEND"
	LedgerTypeVoid   = "VOID"   // revoke points earned by a voided transaction
	LedgerTypeAdjust = "ADJUST" // manual correction by admin
)
