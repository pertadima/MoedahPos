-- +goose Up
-- +goose StatementBegin
ALTER TYPE action_type_enum ADD VALUE IF NOT EXISTS 'PURCHASE_ORDER_DOCUMENT_GENERATE';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- PostgreSQL enum value removals are non-trivial; keep no-op down migration.
-- +goose StatementEnd
