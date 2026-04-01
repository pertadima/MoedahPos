-- +goose Up
-- +goose StatementBegin

CREATE TABLE po_payments (
    id          UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
    po_id       UUID          NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
    store_id    UUID          NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    amount      NUMERIC(15,2) NOT NULL CHECK (amount > 0),
    note        TEXT,
    paid_by     UUID          NOT NULL REFERENCES users(id),
    paid_at     TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX po_payments_po_idx    ON po_payments (po_id);
CREATE INDEX po_payments_store_idx ON po_payments (store_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS po_payments;
-- +goose StatementEnd
