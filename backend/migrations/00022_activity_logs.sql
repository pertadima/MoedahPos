-- +goose Up
-- +goose StatementBegin

-- ─── Activity Log Enum ────────────────────────────────────────────────────────
CREATE TYPE action_type_enum AS ENUM (
    'AUTH_LOGIN',
    'AUTH_LOGOUT',
    'TRANSACTION_CREATE',
    'TRANSACTION_CANCEL',
    'DISCOUNT_ITEM',
    'DISCOUNT_CART',
    'PRICE_OVERRIDE',
    'STOCK_ADJUSTMENT'
);

-- ─── Activity Logs Table ──────────────────────────────────────────────────────
CREATE TABLE activity_logs (
    id           UUID            PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id      UUID            NOT NULL REFERENCES users(id),
    store_id     UUID            REFERENCES stores(id), -- Nullable for global auth events
    action_type  action_type_enum NOT NULL,
    module       VARCHAR(50)     NOT NULL, -- AUTH | TRANSACTION | DISCOUNT | INVENTORY
    reference_id UUID,                     -- transaction_id, product_id, etc.
    metadata     JSONB           NOT NULL DEFAULT '{}',
    created_at   TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- ─── Performance Indices ──────────────────────────────────────────────────────
CREATE INDEX idx_activity_logs_user    ON activity_logs(user_id);
CREATE INDEX idx_activity_logs_store   ON activity_logs(store_id);
CREATE INDEX idx_activity_logs_module  ON activity_logs(module);
CREATE INDEX idx_activity_logs_created ON activity_logs(created_at DESC);

-- ─── Permissions ──────────────────────────────────────────────────────────────
INSERT INTO permissions (id, name, description) VALUES
    (uuid_generate_v4(), 'reports.audit', 'View comprehensive user activity logs');

-- superadmin & admin get the new permission
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name IN ('superadmin', 'admin')
  AND p.name = 'reports.audit';

-- manager also gets it
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'manager'
  AND p.name = 'reports.audit';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM role_permissions WHERE permission_id IN (SELECT id FROM permissions WHERE name = 'reports.audit');
DELETE FROM permissions WHERE name = 'reports.audit';
DROP TABLE IF EXISTS activity_logs;
DROP TYPE IF EXISTS action_type_enum;
-- +goose StatementEnd
