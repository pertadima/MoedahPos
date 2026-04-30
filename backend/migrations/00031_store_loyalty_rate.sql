-- +goose Up
-- +goose StatementBegin
ALTER TABLE stores
  ADD COLUMN IF NOT EXISTS loyalty_points_per_rupiah NUMERIC(10,2) NOT NULL DEFAULT 1000;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE stores DROP COLUMN IF EXISTS loyalty_points_per_rupiah;
-- +goose StatementEnd
