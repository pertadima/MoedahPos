-- +goose Up
-- +goose StatementBegin

-- ─── Scoped Reset for Duplicates ─────────────────────────────────────────────
-- Since we are in a duplicate state, we clear all dependent data to allow unique constraint addition.
-- This aligns with the user's request for a full data reset.
TRUNCATE 
    transaction_items, 
    transactions, 
    stock_movements, 
    stock_batches, 
    stock_levels, 
    purchase_order_items, 
    purchase_orders,
    menu_item_ingredients,
    menu_items,
    restaurant_tables,
    price_history,
    customers,
    po_payments,
    payment_records,
    purchase_order_termins
CASCADE;

-- ─── Cleanup Stores ──────────────────────────────────────────────────────────
DELETE FROM stores a
USING stores b
WHERE a.id > b.id
  AND a.name = b.name;

-- Add unique constraint
ALTER TABLE stores
ADD CONSTRAINT stores_name_key UNIQUE (name);

-- ─── Cleanup Suppliers ────────────────────────────────────────────────────────
DELETE FROM suppliers a
USING suppliers b
WHERE a.id > b.id
  AND a.name = b.name;

-- Add unique constraint
ALTER TABLE suppliers
ADD CONSTRAINT suppliers_name_key UNIQUE (name);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE suppliers DROP CONSTRAINT IF EXISTS suppliers_name_key;
ALTER TABLE stores    DROP CONSTRAINT IF EXISTS stores_name_key;
-- +goose StatementEnd
