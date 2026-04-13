-- +goose Up
-- +goose StatementBegin

ALTER TABLE menu_items
  ADD COLUMN IF NOT EXISTS packaging_cost NUMERIC(15,2) NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS overhead_cost  NUMERIC(15,2) NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS labor_cost     NUMERIC(15,2) NOT NULL DEFAULT 0;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE menu_items
  DROP COLUMN IF EXISTS packaging_cost,
  DROP COLUMN IF EXISTS overhead_cost,
  DROP COLUMN IF EXISTS labor_cost;

-- +goose StatementEnd
