-- +goose Up
-- +goose StatementBegin

-- ─── Extensions ───────────────────────────────────────────────────────────────
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ─── Users ────────────────────────────────────────────────────────────────────
CREATE TABLE users (
    id            UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    name          VARCHAR(100) NOT NULL,
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    is_active     BOOLEAN     NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ
);

CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;

-- ─── Stores ───────────────────────────────────────────────────────────────────
CREATE TABLE stores (
    id         UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    name       VARCHAR(100) NOT NULL,
    address    TEXT,
    phone      VARCHAR(20),
    tax_number VARCHAR(50),
    currency   CHAR(3)     NOT NULL DEFAULT 'IDR',
    is_active  BOOLEAN     NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- ─── Roles ────────────────────────────────────────────────────────────────────
CREATE TABLE roles (
    id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─── Permissions ──────────────────────────────────────────────────────────────
CREATE TABLE permissions (
    id          UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        VARCHAR(100) NOT NULL UNIQUE,
    description TEXT
);

-- ─── Role ↔ Permissions ───────────────────────────────────────────────────────
CREATE TABLE role_permissions (
    role_id       UUID NOT NULL REFERENCES roles(id)       ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- ─── User ↔ Stores (with per-store role) ─────────────────────────────────────
CREATE TABLE user_stores (
    id        UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id   UUID        NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
    store_id  UUID        NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    role_id   UUID        NOT NULL REFERENCES roles(id),
    is_active BOOLEAN     NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, store_id)
);

CREATE INDEX idx_user_stores_user  ON user_stores(user_id);
CREATE INDEX idx_user_stores_store ON user_stores(store_id);

-- ─── Refresh Tokens ───────────────────────────────────────────────────────────
CREATE TABLE refresh_tokens (
    id         UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked    BOOLEAN     NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user    ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_hash    ON refresh_tokens(token_hash);

-- ─── Categories ───────────────────────────────────────────────────────────────
CREATE TABLE categories (
    id         UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    store_id   UUID        NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    name       VARCHAR(100) NOT NULL,
    parent_id  UUID        REFERENCES categories(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─── Products ─────────────────────────────────────────────────────────────────
CREATE TABLE products (
    id           UUID           PRIMARY KEY DEFAULT uuid_generate_v4(),
    store_id     UUID           NOT NULL REFERENCES stores(id)     ON DELETE CASCADE,
    category_id  UUID           REFERENCES categories(id),
    sku          VARCHAR(100)   NOT NULL,
    name         VARCHAR(200)   NOT NULL,
    description  TEXT,
    barcode      VARCHAR(100),
    unit         VARCHAR(20)    NOT NULL DEFAULT 'pcs',
    cost_price   NUMERIC(15,2)  NOT NULL DEFAULT 0,
    sell_price   NUMERIC(15,2)  NOT NULL DEFAULT 0,
    tax_rate     NUMERIC(5,2)   NOT NULL DEFAULT 0,
    image_url    TEXT,
    is_active    BOOLEAN        NOT NULL DEFAULT true,
    created_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ,
    UNIQUE(store_id, sku)
);

CREATE INDEX idx_products_store    ON products(store_id)   WHERE deleted_at IS NULL;
CREATE INDEX idx_products_barcode  ON products(barcode)    WHERE barcode IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX idx_products_category ON products(category_id);

-- ─── Stock Levels ─────────────────────────────────────────────────────────────
CREATE TABLE stock_levels (
    id           UUID           PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_id   UUID           NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    store_id     UUID           NOT NULL REFERENCES stores(id)   ON DELETE CASCADE,
    quantity     NUMERIC(15,3)  NOT NULL DEFAULT 0,
    min_quantity NUMERIC(15,3)  NOT NULL DEFAULT 0,
    updated_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    UNIQUE(product_id, store_id)
);

-- ─── Stock Movements (Audit Trail) ────────────────────────────────────────────
CREATE TABLE stock_movements (
    id             UUID           PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_id     UUID           NOT NULL REFERENCES products(id),
    store_id       UUID           NOT NULL REFERENCES stores(id),
    ref_type       VARCHAR(30)    NOT NULL, -- sale | purchase | adjustment | transfer
    ref_id         UUID,
    quantity_delta NUMERIC(15,3)  NOT NULL,
    notes          TEXT,
    created_by     UUID           NOT NULL REFERENCES users(id),
    created_at     TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_stock_movements_product ON stock_movements(product_id, store_id);
CREATE INDEX idx_stock_movements_ref     ON stock_movements(ref_type, ref_id);

-- ─── Suppliers ────────────────────────────────────────────────────────────────
CREATE TABLE suppliers (
    id           UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    name         VARCHAR(100) NOT NULL,
    contact_name VARCHAR(100),
    phone        VARCHAR(20),
    email        VARCHAR(255),
    address      TEXT,
    is_active    BOOLEAN     NOT NULL DEFAULT true,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─── Purchase Orders ──────────────────────────────────────────────────────────
CREATE TABLE purchase_orders (
    id           UUID           PRIMARY KEY DEFAULT uuid_generate_v4(),
    store_id     UUID           NOT NULL REFERENCES stores(id),
    supplier_id  UUID           REFERENCES suppliers(id),
    po_number    VARCHAR(50)    NOT NULL UNIQUE,
    status       VARCHAR(20)    NOT NULL DEFAULT 'draft', -- draft | ordered | received | cancelled
    total_amount NUMERIC(15,2)  NOT NULL DEFAULT 0,
    ordered_by   UUID           NOT NULL REFERENCES users(id),
    received_by  UUID           REFERENCES users(id),
    ordered_at   TIMESTAMPTZ,
    received_at  TIMESTAMPTZ,
    notes        TEXT,
    created_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_po_store  ON purchase_orders(store_id, status);
CREATE INDEX idx_po_number ON purchase_orders(po_number);

-- ─── Purchase Order Items ─────────────────────────────────────────────────────
CREATE TABLE purchase_order_items (
    id           UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
    po_id        UUID          NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
    product_id   UUID          NOT NULL REFERENCES products(id),
    quantity     NUMERIC(15,3) NOT NULL,
    unit_cost    NUMERIC(15,2) NOT NULL,
    received_qty NUMERIC(15,3) NOT NULL DEFAULT 0,
    subtotal     NUMERIC(15,2) NOT NULL
);

-- ─── Transactions ─────────────────────────────────────────────────────────────
CREATE TABLE transactions (
    id             UUID           PRIMARY KEY DEFAULT uuid_generate_v4(),
    store_id       UUID           NOT NULL REFERENCES stores(id),
    cashier_id     UUID           NOT NULL REFERENCES users(id),
    customer_name  VARCHAR(100),
    customer_phone VARCHAR(20),
    subtotal       NUMERIC(15,2)  NOT NULL DEFAULT 0,
    discount_amt   NUMERIC(15,2)  NOT NULL DEFAULT 0,
    tax_amt        NUMERIC(15,2)  NOT NULL DEFAULT 0,
    total          NUMERIC(15,2)  NOT NULL DEFAULT 0,
    payment_method VARCHAR(30)    NOT NULL, -- cash | card | qris | transfer
    payment_amount NUMERIC(15,2)  NOT NULL DEFAULT 0,
    change_amount  NUMERIC(15,2)  NOT NULL DEFAULT 0,
    status         VARCHAR(20)    NOT NULL DEFAULT 'completed', -- draft | completed | voided
    notes          TEXT,
    created_at     TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transactions_store    ON transactions(store_id, created_at DESC);
CREATE INDEX idx_transactions_cashier  ON transactions(cashier_id);
CREATE INDEX idx_transactions_status   ON transactions(status);

-- ─── Transaction Items ────────────────────────────────────────────────────────
CREATE TABLE transaction_items (
    id             UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
    transaction_id UUID          NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    product_id     UUID          REFERENCES products(id),
    product_name   VARCHAR(200)  NOT NULL, -- snapshot at time of sale
    sku            VARCHAR(100)  NOT NULL,
    quantity       NUMERIC(15,3) NOT NULL,
    unit_price     NUMERIC(15,2) NOT NULL,
    discount_pct   NUMERIC(5,2)  NOT NULL DEFAULT 0,
    tax_rate       NUMERIC(5,2)  NOT NULL DEFAULT 0,
    subtotal       NUMERIC(15,2) NOT NULL
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS transaction_items       CASCADE;
DROP TABLE IF EXISTS transactions            CASCADE;
DROP TABLE IF EXISTS purchase_order_items   CASCADE;
DROP TABLE IF EXISTS purchase_orders        CASCADE;
DROP TABLE IF EXISTS suppliers              CASCADE;
DROP TABLE IF EXISTS stock_movements        CASCADE;
DROP TABLE IF EXISTS stock_levels           CASCADE;
DROP TABLE IF EXISTS products               CASCADE;
DROP TABLE IF EXISTS categories             CASCADE;
DROP TABLE IF EXISTS refresh_tokens         CASCADE;
DROP TABLE IF EXISTS user_stores            CASCADE;
DROP TABLE IF EXISTS role_permissions       CASCADE;
DROP TABLE IF EXISTS permissions            CASCADE;
DROP TABLE IF EXISTS roles                  CASCADE;
DROP TABLE IF EXISTS stores                 CASCADE;
DROP TABLE IF EXISTS users                  CASCADE;
DROP EXTENSION IF EXISTS "uuid-ossp";
-- +goose StatementEnd
