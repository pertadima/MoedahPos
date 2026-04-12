-- +goose Up
-- +goose StatementBegin
ALTER TYPE action_type_enum ADD VALUE IF NOT EXISTS 'PURCHASE_ORDER_CREATE';
ALTER TYPE action_type_enum ADD VALUE IF NOT EXISTS 'PURCHASE_ORDER_UPDATE';
ALTER TYPE action_type_enum ADD VALUE IF NOT EXISTS 'PURCHASE_ORDER_PAYMENT';
ALTER TYPE action_type_enum ADD VALUE IF NOT EXISTS 'INCOME_CREATE';
ALTER TYPE action_type_enum ADD VALUE IF NOT EXISTS 'INCOME_UPDATE';
ALTER TYPE action_type_enum ADD VALUE IF NOT EXISTS 'INCOME_DELETE';
ALTER TYPE action_type_enum ADD VALUE IF NOT EXISTS 'EXPENSE_CREATE';
ALTER TYPE action_type_enum ADD VALUE IF NOT EXISTS 'EXPENSE_UPDATE';
ALTER TYPE action_type_enum ADD VALUE IF NOT EXISTS 'EXPENSE_DELETE';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- PostgreSQL does not support dropping enum values easily, so down migration usually does nothing for ALTER TYPE ADD VALUE.
-- +goose StatementEnd
