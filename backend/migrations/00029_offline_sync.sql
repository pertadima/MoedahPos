-- +goose Up
-- +goose StatementBegin
-- Offline Sync Fields Migration

-- 1. Add fields to products
ALTER TABLE products ADD COLUMN IF NOT EXISTS server_updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE products ADD COLUMN IF NOT EXISTS sync_version INT DEFAULT 1;

-- 2. Add fields to categories
ALTER TABLE categories ADD COLUMN IF NOT EXISTS server_updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE categories ADD COLUMN IF NOT EXISTS sync_version INT DEFAULT 1;

-- 3. Add fields to stock_levels
ALTER TABLE stock_levels ADD COLUMN IF NOT EXISTS server_updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE stock_levels ADD COLUMN IF NOT EXISTS sync_version INT DEFAULT 1;

-- 4. Add fields to transactions
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS server_updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS sync_version INT DEFAULT 1;

-- 5. Set up the generic trigger function for sync_version
CREATE OR REPLACE FUNCTION update_sync_fields()
RETURNS TRIGGER AS $$
BEGIN
    NEW.server_updated_at = CURRENT_TIMESTAMP;
    NEW.sync_version = COALESCE(OLD.sync_version, 0) + 1;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 6. Apply triggers to all 4 tables
-- Products
DROP TRIGGER IF EXISTS products_sync_trigger ON products;
CREATE TRIGGER products_sync_trigger BEFORE UPDATE ON products FOR EACH ROW EXECUTE FUNCTION update_sync_fields();

-- Categories
DROP TRIGGER IF EXISTS categories_sync_trigger ON categories;
CREATE TRIGGER categories_sync_trigger BEFORE UPDATE ON categories FOR EACH ROW EXECUTE FUNCTION update_sync_fields();

-- Stock Levels
DROP TRIGGER IF EXISTS stock_levels_sync_trigger ON stock_levels;
CREATE TRIGGER stock_levels_sync_trigger BEFORE UPDATE ON stock_levels FOR EACH ROW EXECUTE FUNCTION update_sync_fields();

-- Transactions
DROP TRIGGER IF EXISTS transactions_sync_trigger ON transactions;
CREATE TRIGGER transactions_sync_trigger BEFORE UPDATE ON transactions FOR EACH ROW EXECUTE FUNCTION update_sync_fields();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS transactions_sync_trigger ON transactions;
DROP TRIGGER IF EXISTS stock_levels_sync_trigger ON stock_levels;
DROP TRIGGER IF EXISTS categories_sync_trigger ON categories;
DROP TRIGGER IF EXISTS products_sync_trigger ON products;

DROP FUNCTION IF EXISTS update_sync_fields();

ALTER TABLE transactions DROP COLUMN IF EXISTS server_updated_at;
ALTER TABLE transactions DROP COLUMN IF EXISTS sync_version;

ALTER TABLE stock_levels DROP COLUMN IF EXISTS server_updated_at;
ALTER TABLE stock_levels DROP COLUMN IF EXISTS sync_version;

ALTER TABLE categories DROP COLUMN IF EXISTS server_updated_at;
ALTER TABLE categories DROP COLUMN IF EXISTS sync_version;

ALTER TABLE products DROP COLUMN IF EXISTS server_updated_at;
ALTER TABLE products DROP COLUMN IF EXISTS sync_version;
-- +goose StatementEnd
