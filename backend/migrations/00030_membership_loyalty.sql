-- +goose Up
-- +goose StatementBegin

-- Membership and Loyalty Migration

-- 1. Create membership_tiers table
CREATE TABLE IF NOT EXISTS membership_tiers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL UNIQUE,
    multiplier NUMERIC(5,2) NOT NULL DEFAULT 1.0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed default tiers
INSERT INTO membership_tiers (name, multiplier) VALUES
    ('Bronze', 1.0),
    ('Silver', 1.2),
    ('Gold', 1.5),
    ('Platinum', 2.0)
ON CONFLICT (name) DO UPDATE SET multiplier = EXCLUDED.multiplier;

-- 2. Add loyalty and sync fields to customers
ALTER TABLE customers ADD COLUMN IF NOT EXISTS loyalty_tier_id UUID REFERENCES membership_tiers(id) ON DELETE SET NULL;
ALTER TABLE customers ADD COLUMN IF NOT EXISTS server_updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE customers ADD COLUMN IF NOT EXISTS sync_version INT DEFAULT 1;

-- 3. Apply sync trigger to customers
CREATE OR REPLACE FUNCTION update_sync_fields()
RETURNS TRIGGER AS $$
BEGIN
    NEW.server_updated_at = CURRENT_TIMESTAMP;
    NEW.sync_version = COALESCE(OLD.sync_version, 0) + 1;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS customers_sync_trigger ON customers;
CREATE TRIGGER customers_sync_trigger BEFORE UPDATE ON customers FOR EACH ROW EXECUTE FUNCTION update_sync_fields();

-- 4. Create loyalty_ledger table
CREATE TABLE IF NOT EXISTS loyalty_ledger (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    points_delta NUMERIC(10,2) NOT NULL,
    transaction_id UUID REFERENCES transactions(id) ON DELETE SET NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('EARN', 'SPEND')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for quick ledger lookups
CREATE INDEX IF NOT EXISTS idx_loyalty_ledger_customer_id ON loyalty_ledger(customer_id);
CREATE INDEX IF NOT EXISTS idx_loyalty_ledger_created_at ON loyalty_ledger(created_at DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS loyalty_ledger CASCADE;

DROP TRIGGER IF EXISTS customers_sync_trigger ON customers;
ALTER TABLE customers DROP COLUMN IF EXISTS loyalty_tier_id;
ALTER TABLE customers DROP COLUMN IF EXISTS server_updated_at;
ALTER TABLE customers DROP COLUMN IF EXISTS sync_version;

DROP TABLE IF EXISTS membership_tiers CASCADE;
-- +goose StatementEnd
