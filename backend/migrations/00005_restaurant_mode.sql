-- +goose Up
-- +goose StatementBegin

-- ─── Store Type ───────────────────────────────────────────────────────────────
ALTER TABLE stores
  ADD COLUMN IF NOT EXISTS store_type VARCHAR(20) NOT NULL DEFAULT 'retail';

COMMENT ON COLUMN stores.store_type IS 'retail | restaurant';

-- ─── Restaurant Tables ────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS restaurant_tables (
    id           UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    store_id     UUID        NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    table_number VARCHAR(20) NOT NULL,
    capacity     INT         NOT NULL DEFAULT 4,
    status       VARCHAR(20) NOT NULL DEFAULT 'available', -- available | occupied | reserved
    notes        TEXT,
    is_active    BOOLEAN     NOT NULL DEFAULT true,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ,
    UNIQUE(store_id, table_number)
);

CREATE INDEX IF NOT EXISTS idx_tables_store  ON restaurant_tables(store_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tables_status ON restaurant_tables(store_id, status) WHERE deleted_at IS NULL;

-- ─── Menu Items ───────────────────────────────────────────────────────────────
-- A menu item is a composed dish: one sellable item made from multiple products/ingredients.
CREATE TABLE IF NOT EXISTS menu_items (
    id          UUID           PRIMARY KEY DEFAULT uuid_generate_v4(),
    store_id    UUID           NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    category_id UUID           REFERENCES categories(id),
    name        VARCHAR(200)   NOT NULL,
    description TEXT,
    sell_price  NUMERIC(15,2)  NOT NULL DEFAULT 0,
    tax_rate    NUMERIC(5,2)   NOT NULL DEFAULT 0,
    image_url   TEXT,
    is_active   BOOLEAN        NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_menu_items_store    ON menu_items(store_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_menu_items_category ON menu_items(category_id);

-- ─── Menu Item Ingredients (Recipe / BOM) ────────────────────────────────────
-- Links a menu item to raw product/ingredient stocks with quantity consumed per serving.
CREATE TABLE IF NOT EXISTS menu_item_ingredients (
    id           UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
    menu_item_id UUID          NOT NULL REFERENCES menu_items(id) ON DELETE CASCADE,
    product_id   UUID          NOT NULL REFERENCES products(id)   ON DELETE CASCADE,
    quantity     NUMERIC(15,3) NOT NULL,            -- qty of ingredient consumed per 1 serving
    UNIQUE(menu_item_id, product_id)
);

CREATE INDEX IF NOT EXISTS idx_ingredients_menu ON menu_item_ingredients(menu_item_id);

-- ─── Transactions — add table & order_type ───────────────────────────────────
ALTER TABLE transactions
  ADD COLUMN IF NOT EXISTS table_id   UUID        REFERENCES restaurant_tables(id),
  ADD COLUMN IF NOT EXISTS order_type VARCHAR(20) NOT NULL DEFAULT 'checkout'; -- checkout | dine_in | takeaway

COMMENT ON COLUMN transactions.order_type IS 'checkout (retail) | dine_in | takeaway';

-- ─── Transaction Items — link to menu item ───────────────────────────────────
ALTER TABLE transaction_items
  ADD COLUMN IF NOT EXISTS menu_item_id UUID REFERENCES menu_items(id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE transaction_items DROP COLUMN IF EXISTS menu_item_id;
ALTER TABLE transactions       DROP COLUMN IF EXISTS table_id;
ALTER TABLE transactions       DROP COLUMN IF EXISTS order_type;
DROP TABLE IF EXISTS menu_item_ingredients CASCADE;
DROP TABLE IF EXISTS menu_items            CASCADE;
DROP TABLE IF EXISTS restaurant_tables     CASCADE;
ALTER TABLE stores DROP COLUMN IF EXISTS store_type;
-- +goose StatementEnd
