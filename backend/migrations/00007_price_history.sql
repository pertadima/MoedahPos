-- +goose Up
-- +goose StatementBegin

CREATE TABLE price_history (
    id           UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_id   UUID          NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    store_id     UUID          NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    changed_by   UUID          NOT NULL REFERENCES users(id),
    old_cost     NUMERIC(15,2) NOT NULL DEFAULT 0,
    new_cost     NUMERIC(15,2) NOT NULL DEFAULT 0,
    old_sell     NUMERIC(15,2) NOT NULL DEFAULT 0,
    new_sell     NUMERIC(15,2) NOT NULL DEFAULT 0,
    source       VARCHAR(30)   NOT NULL DEFAULT 'manual', -- manual | purchase_order
    ref_id       UUID,          -- PO id when source = purchase_order
    notes        TEXT,
    changed_at   TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX price_history_product_idx ON price_history (product_id, changed_at DESC);
CREATE INDEX price_history_store_idx   ON price_history (store_id,   changed_at DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS price_history;
-- +goose StatementEnd
