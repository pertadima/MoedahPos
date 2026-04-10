-- +goose Up
-- +goose StatementBegin

-- Adds cart-level discount tracking.
-- Header: audit trail of what discount type/value was applied to whole cart.
-- Items: how much of that cart discount was allocated to this item (per unit).
-- unit_price already stores FINAL price after both item + cart discounts.
-- No changes required to any report queries.

ALTER TABLE transactions
  ADD COLUMN IF NOT EXISTS cart_discount_type  VARCHAR(20)   NOT NULL DEFAULT 'PERCENTAGE',
  ADD COLUMN IF NOT EXISTS cart_discount_value NUMERIC(15,4) NOT NULL DEFAULT 0;

ALTER TABLE transaction_items
  ADD COLUMN IF NOT EXISTS cart_discount_allocated NUMERIC(15,2) NOT NULL DEFAULT 0;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE transactions
  DROP COLUMN IF EXISTS cart_discount_type,
  DROP COLUMN IF EXISTS cart_discount_value;

ALTER TABLE transaction_items
  DROP COLUMN IF EXISTS cart_discount_allocated;
-- +goose StatementEnd
