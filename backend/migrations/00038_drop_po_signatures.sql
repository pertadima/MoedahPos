-- +goose Up
ALTER TABLE purchase_orders DROP COLUMN IF EXISTS supplier_signature;
ALTER TABLE purchase_orders DROP COLUMN IF EXISTS buyer_signature;
ALTER TABLE purchase_orders DROP COLUMN IF EXISTS supplier_signed_at;
ALTER TABLE purchase_orders DROP COLUMN IF EXISTS buyer_signed_at;

-- +goose Down
ALTER TABLE purchase_orders ADD COLUMN supplier_signature TEXT;
ALTER TABLE purchase_orders ADD COLUMN buyer_signature TEXT;
ALTER TABLE purchase_orders ADD COLUMN supplier_signed_at TIMESTAMPTZ;
ALTER TABLE purchase_orders ADD COLUMN buyer_signed_at TIMESTAMPTZ;
