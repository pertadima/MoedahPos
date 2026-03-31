-- +goose Up
-- +goose StatementBegin

-- Snapshot cost-at-sale in transaction_items for accurate profit reporting.
-- Existing rows get cost_price = 0 (unknown historic cost — won't affect new sales).
ALTER TABLE transaction_items
  ADD COLUMN IF NOT EXISTS cost_price NUMERIC(15,2) NOT NULL DEFAULT 0;

-- Pre-fill existing rows from the current product cost_price as a best-effort estimate.
UPDATE transaction_items ti
  SET cost_price = p.cost_price
  FROM products p
  WHERE ti.product_id = p.id
    AND ti.cost_price = 0
    AND ti.product_id IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE transaction_items DROP COLUMN IF EXISTS cost_price;
-- +goose StatementEnd
