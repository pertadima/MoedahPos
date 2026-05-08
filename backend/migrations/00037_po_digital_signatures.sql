-- +goose Up
-- +goose StatementBegin
ALTER TABLE purchase_orders
  ADD COLUMN IF NOT EXISTS supplier_signature TEXT,
  ADD COLUMN IF NOT EXISTS buyer_signature TEXT,
  ADD COLUMN IF NOT EXISTS supplier_signed_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS buyer_signed_at TIMESTAMPTZ;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE purchase_orders
  DROP COLUMN IF EXISTS supplier_signature,
  DROP COLUMN IF EXISTS buyer_signature,
  DROP COLUMN IF EXISTS supplier_signed_at,
  DROP COLUMN IF EXISTS buyer_signed_at;
-- +goose StatementEnd