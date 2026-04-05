-- +goose Up
-- +goose StatementBegin
ALTER TABLE transaction_items
ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'pending',
ADD COLUMN completed_at TIMESTAMP WITH TIME ZONE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE transaction_items
DROP COLUMN status,
DROP COLUMN completed_at;
-- +goose StatementEnd
