-- +goose Up
-- +goose StatementBegin

CREATE TABLE customers (
    id         UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    store_id   UUID         NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    name       VARCHAR(150) NOT NULL,
    phone      VARCHAR(30),
    email      VARCHAR(150),
    address    TEXT,
    notes      TEXT,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX customers_store_idx ON customers (store_id) WHERE deleted_at IS NULL;
CREATE INDEX customers_phone_idx ON customers (store_id, phone) WHERE deleted_at IS NULL;

-- Link transactions to the customers table (optional FK)
ALTER TABLE transactions ADD COLUMN customer_id UUID REFERENCES customers(id) ON DELETE SET NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE transactions DROP COLUMN IF EXISTS customer_id;
DROP TABLE IF EXISTS customers;
-- +goose StatementEnd
