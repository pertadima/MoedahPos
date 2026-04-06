-- +goose Up
-- SQL in this section is executed when the migration is applied.

CREATE TYPE adjustment_type AS ENUM ('IN', 'OUT');
CREATE TYPE adjustment_reason AS ENUM ('DAMAGED', 'LOST', 'MANUAL_CORRECTION');

CREATE TABLE stock_adjustments (
    id UUID PRIMARY KEY,
    product_id UUID NOT NULL REFERENCES products(id),
    store_id UUID NOT NULL REFERENCES stores(id),
    type adjustment_type NOT NULL,
    reason adjustment_reason NOT NULL,
    quantity NUMERIC(15,3) NOT NULL CHECK (quantity > 0),
    notes TEXT,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Creating a link to track exactly which FIFO batches were touched by an adjustment
CREATE TABLE stock_adjustment_batches (
    id UUID PRIMARY KEY,
    adjustment_id UUID NOT NULL REFERENCES stock_adjustments(id) ON DELETE CASCADE,
    batch_id UUID NOT NULL REFERENCES stock_batches(id),
    deducted_qty NUMERIC(15,3) NOT NULL CHECK (deducted_qty > 0)
);

CREATE INDEX idx_stock_adjustments_store_product ON stock_adjustments(store_id, product_id);
CREATE INDEX idx_stock_adjustment_batches_adj ON stock_adjustment_batches(adjustment_id);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

DROP INDEX idx_stock_adjustment_batches_adj;
DROP INDEX idx_stock_adjustments_store_product;
DROP TABLE stock_adjustment_batches;
DROP TABLE stock_adjustments;
DROP TYPE adjustment_reason;
DROP TYPE adjustment_type;
