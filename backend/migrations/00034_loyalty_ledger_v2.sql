-- +goose Up
-- +goose StatementBegin

-- Extends loyalty_ledger with:
--   1. VOID and ADJUST point types (for refunds and manual corrections)
--   2. balance_snapshot column (balance *after* this entry, for audit integrity)
--   3. Composite index on (customer_id, created_at) for fast paginated history

-- Step 1: Relax the CHECK constraint to allow new types
ALTER TABLE loyalty_ledger
    DROP CONSTRAINT IF EXISTS loyalty_ledger_type_check;

ALTER TABLE loyalty_ledger
    ADD CONSTRAINT loyalty_ledger_type_check
        CHECK (type IN ('EARN', 'SPEND', 'VOID', 'ADJUST'));

-- Step 2: Add balance_snapshot — balance *after* this ledger entry is applied.
-- Nullable so existing rows are not broken; new writes populate it.
ALTER TABLE loyalty_ledger
    ADD COLUMN IF NOT EXISTS balance_snapshot NUMERIC(15,2);

-- Step 3: Composite index for fast paginated history per customer.
CREATE INDEX IF NOT EXISTS idx_loyalty_ledger_customer_created
    ON loyalty_ledger (customer_id, created_at DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE loyalty_ledger
    DROP CONSTRAINT IF EXISTS loyalty_ledger_type_check;

ALTER TABLE loyalty_ledger
    ADD CONSTRAINT loyalty_ledger_type_check
        CHECK (type IN ('EARN', 'SPEND'));

ALTER TABLE loyalty_ledger
    DROP COLUMN IF EXISTS balance_snapshot;

DROP INDEX IF EXISTS idx_loyalty_ledger_customer_created;

-- +goose StatementEnd
