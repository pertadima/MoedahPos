-- +goose Up
-- +goose StatementBegin

-- ─── Stock Batches (FIFO Inventory Tracking) ──────────────────────────────────
--
-- Each row represents one batch of product inventory received from a PO.
-- FIFO deduction always processes the batch with the lowest received_at first,
-- ensuring oldest stock is sold before newer stock.
--
CREATE TABLE stock_batches (
    id                 UUID           PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_id         UUID           NOT NULL REFERENCES products(id)        ON DELETE CASCADE,
    store_id           UUID           NOT NULL REFERENCES stores(id)          ON DELETE CASCADE,
    po_id              UUID           REFERENCES purchase_orders(id)          ON DELETE SET NULL,
    quantity_remaining NUMERIC(15,3)  NOT NULL DEFAULT 0
                                      CONSTRAINT chk_batch_qty_non_negative CHECK (quantity_remaining >= 0),
    purchase_price     NUMERIC(15,2)  NOT NULL DEFAULT 0,
    received_at        TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    created_at         TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

-- FIFO index: partial index over non-empty batches ordered by received_at.
-- All FIFO deduction queries use this index via:
--   WHERE product_id=? AND store_id=? AND quantity_remaining>0 ORDER BY received_at ASC FOR UPDATE
CREATE INDEX idx_stock_batches_fifo
    ON stock_batches(product_id, store_id, received_at)
    WHERE quantity_remaining > 0;

-- General purpose lookup index for product/store queries.
CREATE INDEX idx_stock_batches_product
    ON stock_batches(product_id, store_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS stock_batches CASCADE;
-- +goose StatementEnd
