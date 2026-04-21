package domain

import "time"

// Customer is a registered buyer associated with a store.
type Customer struct {
	ID              string     `db:"id"`
	StoreID         string     `db:"store_id"`
	Name            string     `db:"name"`
	Phone           *string    `db:"phone"`
	Email           *string    `db:"email"`
	Address         *string    `db:"address"`
	Notes           *string    `db:"notes"`
	LoyaltyTierID   *string    `db:"loyalty_tier_id"` // FK → membership_tiers.id (nullable)
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
	DeletedAt       *time.Time `db:"deleted_at"`
	ServerUpdatedAt time.Time  `db:"server_updated_at"`
	SyncVersion     int        `db:"sync_version"`
}
