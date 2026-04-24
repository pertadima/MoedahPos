-- +goose Up
-- +goose StatementBegin
ALTER TABLE transactions
ADD COLUMN IF NOT EXISTS points_redeemed NUMERIC(15,2) NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS points_discount NUMERIC(15,2) NOT NULL DEFAULT 0;

ALTER TABLE stores
ADD COLUMN IF NOT EXISTS loyalty_rupiah_per_point NUMERIC(10,2) NOT NULL DEFAULT 1;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE stores
DROP COLUMN IF EXISTS loyalty_rupiah_per_point;

ALTER TABLE transactions
DROP COLUMN IF EXISTS points_redeemed,
DROP COLUMN IF EXISTS points_discount;
-- +goose StatementEnd
