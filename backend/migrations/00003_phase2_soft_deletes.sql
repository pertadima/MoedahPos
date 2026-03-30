-- +goose Up
-- +goose StatementBegin

-- Add soft-delete to categories (was missing from initial schema)
ALTER TABLE categories ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

-- Add soft-delete to suppliers
ALTER TABLE suppliers ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

-- Partial index for active categories
CREATE INDEX IF NOT EXISTS idx_categories_store
    ON categories(store_id) WHERE deleted_at IS NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE categories DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE suppliers  DROP COLUMN IF EXISTS deleted_at;
DROP INDEX IF EXISTS idx_categories_store;
-- +goose StatementEnd
