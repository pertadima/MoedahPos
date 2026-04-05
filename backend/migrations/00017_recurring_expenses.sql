-- +goose Up
ALTER TABLE expenses 
ADD COLUMN payment_status VARCHAR(20) DEFAULT 'paid' CHECK (payment_status IN ('paid', 'unpaid', 'cancelled'));

CREATE TABLE recurring_expenses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES expense_categories(id),
    name VARCHAR(255) NOT NULL,
    amount DECIMAL(15, 2) NOT NULL CHECK (amount > 0),
    interval VARCHAR(50) NOT NULL CHECK (interval IN ('daily', 'weekly', 'monthly', 'yearly')),
    interval_value INT NOT NULL DEFAULT 1,
    start_date DATE NOT NULL,
    end_date DATE,
    next_run_date DATE NOT NULL,
    notes TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_generated_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_recurring_expenses_next_run ON recurring_expenses(next_run_date) WHERE is_active = true;

-- +goose Down
DROP TABLE recurring_expenses;
ALTER TABLE expenses DROP COLUMN payment_status;
