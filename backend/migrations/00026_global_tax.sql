-- +goose Up
-- +goose StatementBegin

-- 1. Add default tax percentage to stores
ALTER TABLE stores
ADD COLUMN IF NOT EXISTS default_tax_percentage NUMERIC(5,2) NOT NULL DEFAULT 0;

-- 2. Add use_global_tax to menu_items
ALTER TABLE menu_items
ADD COLUMN IF NOT EXISTS use_global_tax BOOLEAN NOT NULL DEFAULT true;

-- 3. Rename tax_rate to tax_percentage in menu_items
ALTER TABLE menu_items RENAME COLUMN tax_rate TO tax_percentage;
ALTER TABLE menu_items ALTER COLUMN tax_percentage DROP NOT NULL;
ALTER TABLE menu_items ALTER COLUMN tax_percentage DROP DEFAULT;

-- Ensure existing records with a non-zero tax_percentage are preserved as custom overrides
UPDATE menu_items SET use_global_tax = false WHERE tax_percentage > 0;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE menu_items RENAME COLUMN tax_percentage TO tax_rate;
ALTER TABLE menu_items ALTER COLUMN tax_rate SET DEFAULT 0;
ALTER TABLE menu_items ALTER COLUMN tax_rate SET NOT NULL;
UPDATE menu_items SET tax_rate = 0 WHERE tax_rate IS NULL;

ALTER TABLE menu_items DROP COLUMN IF EXISTS use_global_tax;
ALTER TABLE stores DROP COLUMN IF EXISTS default_tax_percentage;

-- +goose StatementEnd
