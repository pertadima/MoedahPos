-- +goose Up

ALTER TABLE income_categories
    ADD COLUMN is_active BOOLEAN DEFAULT true,
    ADD COLUMN updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ADD COLUMN deleted_at TIMESTAMP WITH TIME ZONE;

ALTER TABLE expense_categories
    ADD COLUMN is_active BOOLEAN DEFAULT true,
    ADD COLUMN updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ADD COLUMN deleted_at TIMESTAMP WITH TIME ZONE;

CREATE UNIQUE INDEX idx_income_categories_name_unique
    ON income_categories (LOWER(name))
    WHERE deleted_at IS NULL;

CREATE INDEX idx_income_categories_deleted_at
    ON income_categories (deleted_at);

CREATE UNIQUE INDEX idx_expense_categories_name_unique
    ON expense_categories (LOWER(name))
    WHERE deleted_at IS NULL;

CREATE INDEX idx_expense_categories_deleted_at
    ON expense_categories (deleted_at);

-- +goose Down

DROP INDEX IF EXISTS idx_expense_categories_deleted_at;
DROP INDEX IF EXISTS idx_expense_categories_name_unique;
DROP INDEX IF EXISTS idx_income_categories_deleted_at;
DROP INDEX IF EXISTS idx_income_categories_name_unique;

ALTER TABLE expense_categories
    DROP COLUMN is_active,
    DROP COLUMN updated_at,
    DROP COLUMN deleted_at;

ALTER TABLE income_categories
    DROP COLUMN is_active,
    DROP COLUMN updated_at,
    DROP COLUMN deleted_at;
