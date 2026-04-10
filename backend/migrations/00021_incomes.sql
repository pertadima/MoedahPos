-- +goose Up

CREATE TABLE income_categories (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Seed default categories
INSERT INTO income_categories (name, description) VALUES
    ('Pendapatan Lain-lain', 'Pendapatan di luar operasional utama'),
    ('Injeksi Modal',        'Penambahan modal dari pemilik atau investor'),
    ('Pengembalian Supplier','Refund atau kelebihan bayar dari supplier'),
    ('Piutang Diterima',     'Penerimaan pembayaran dari piutang pelanggan'),
    ('Lainnya',              'Penerimaan kas lainnya yang tidak terklasifikasi');

CREATE TABLE incomes (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    store_id       UUID         NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    category_id    UUID         NOT NULL REFERENCES income_categories(id),
    amount         DECIMAL(15,2) NOT NULL CHECK (amount > 0),
    income_date    DATE         NOT NULL,
    payment_method VARCHAR(50)  NOT NULL DEFAULT 'cash',
    reference      VARCHAR(255),
    notes          TEXT,
    created_by     UUID         REFERENCES users(id) ON DELETE SET NULL,
    created_at     TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at     TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_incomes_store_date ON incomes(store_id, income_date);
CREATE INDEX idx_incomes_category   ON incomes(category_id);

-- +goose Down
DROP TABLE incomes;
DROP TABLE income_categories;
