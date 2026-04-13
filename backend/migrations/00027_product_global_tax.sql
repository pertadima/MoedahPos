-- +goose Up
-- Add use_global_tax
ALTER TABLE products ADD COLUMN if not exists use_global_tax BOOLEAN NOT NULL DEFAULT true;

-- Rename tax_rate to tax_percentage
ALTER TABLE products RENAME COLUMN tax_rate TO tax_percentage;

-- If there are distinct settings, we fallback via use_global_tax=false
UPDATE products SET use_global_tax = false WHERE tax_percentage > 0;

-- +goose Down
-- Revert the column rename
ALTER TABLE products RENAME COLUMN tax_percentage TO tax_rate;

-- Drop the use_global_tax column
ALTER TABLE products DROP COLUMN if exists use_global_tax;
