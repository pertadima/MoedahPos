-- +goose Up
-- +goose StatementBegin

-- Add table_id to transactions so restaurant draft orders can be linked to a table.
-- Nullable — retail transactions have no table.
ALTER TABLE transactions
  ADD COLUMN IF NOT EXISTS table_id UUID REFERENCES restaurant_tables(id) ON DELETE SET NULL;

-- Partial index: fast lookup of open (draft) orders per table
CREATE INDEX IF NOT EXISTS idx_transactions_table_draft
  ON transactions(table_id)
  WHERE status = 'draft' AND table_id IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_transactions_table_draft;
ALTER TABLE transactions DROP COLUMN IF EXISTS table_id;
-- +goose StatementEnd
