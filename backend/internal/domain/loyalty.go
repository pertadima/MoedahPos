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

// LoyaltyLedger is a single point event (EARN or SPEND) for a customer.
// points_delta is positive for EARN and negative for SPEND.
type LoyaltyLedger struct {
	ID            string    `db:"id"`
	CustomerID    string    `db:"customer_id"`
	PointsDelta   float64   `db:"points_delta"`
	TransactionID *string   `db:"transaction_id"`
	Type          string    `db:"type"` // EARN | SPEND
	CreatedAt     time.Time `db:"created_at"`
}
