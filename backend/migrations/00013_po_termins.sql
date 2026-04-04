-- +goose Up
-- +goose StatementBegin

-- ─── Purchase Order Termins ──────────────────────────────────────────────────
--
-- Each row represents one installment (termin) for a purchase order.
-- A PO can have 1..N termins totalling up to (or equal to) the PO's total_amount.
-- Payments against a termin are tracked in payment_records.
--
CREATE TABLE purchase_order_termins (
    id             UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
    po_id          UUID          NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
    termin_number  INT           NOT NULL,           -- sequential: 1, 2, 3…
    amount         NUMERIC(15,2) NOT NULL CHECK (amount > 0),
    due_date       DATE          NOT NULL,
    status         TEXT          NOT NULL DEFAULT 'unpaid'
                                 CHECK (status IN ('unpaid','partial','paid','overdue')),
    notes          TEXT          NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    -- Each PO can only have one termin per number
    UNIQUE (po_id, termin_number)
);

CREATE INDEX idx_po_termins_po        ON purchase_order_termins(po_id);
CREATE INDEX idx_po_termins_due_date  ON purchase_order_termins(due_date);
CREATE INDEX idx_po_termins_status    ON purchase_order_termins(status);

-- ─── Payment Records ─────────────────────────────────────────────────────────
--
-- Each row is one payment transaction against a specific termin.
-- A termin can accumulate multiple partial payments until it reaches 'paid' status.
-- Business rules (enforced in the service layer):
--   - sum(amount_paid) for a termin must not exceed termin.amount
--   - sum of all termin amounts must not exceed po.total_amount
--
CREATE TABLE payment_records (
    id              UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
    termin_id       UUID          NOT NULL REFERENCES purchase_order_termins(id) ON DELETE CASCADE,
    amount_paid     NUMERIC(15,2) NOT NULL CHECK (amount_paid > 0),
    payment_date    DATE          NOT NULL DEFAULT CURRENT_DATE,
    payment_method  TEXT          NOT NULL DEFAULT 'cash'
                                  CHECK (payment_method IN ('cash','transfer','check','other')),
    notes           TEXT          NOT NULL DEFAULT '',
    recorded_by     UUID          REFERENCES users(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payment_records_termin ON payment_records(termin_id);
CREATE INDEX idx_payment_records_date   ON payment_records(payment_date);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS payment_records CASCADE;
DROP TABLE IF EXISTS purchase_order_termins CASCADE;
-- +goose StatementEnd
