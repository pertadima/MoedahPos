-- +goose Up
-- +goose StatementBegin

-- Add dynamic pricing columns to transaction_items.
-- original_price: snapshot of product's sell_price at time of sale (never changes)
-- discount_type: PERCENTAGE | FIXED | OVERRIDE
-- discount_value: the discount amount / percentage / override price
-- unit_price continues to store the FINAL selling price (backward-compatible with all reports)

ALTER TABLE transaction_items
  ADD COLUMN IF NOT EXISTS original_price NUMERIC(15,2) NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS discount_type  VARCHAR(20)   NOT NULL DEFAULT 'PERCENTAGE',
  ADD COLUMN IF NOT EXISTS discount_value NUMERIC(15,4) NOT NULL DEFAULT 0;

-- Back-fill existing rows:
-- original_price = unit_price (no discount was applied previously)
-- discount_value = discount_pct (matches old PERCENTAGE behavior)
UPDATE transaction_items
SET original_price = unit_price,
    discount_type  = 'PERCENTAGE',
    discount_value = discount_pct
WHERE original_price = 0;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE transaction_items
  DROP COLUMN IF EXISTS original_price,
  DROP COLUMN IF EXISTS discount_type,
  DROP COLUMN IF EXISTS discount_value;
-- +goose StatementEnd
