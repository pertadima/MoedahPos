-- +goose Up
CREATE TABLE expense_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Seed default categories
INSERT INTO expense_categories (name, description) VALUES
    ('Rent', 'Store rent and building lease'),
    ('Electricity', 'Electricity utility bills'),
    ('Salary', 'Employee salaries and wages'),
    ('Maintenance', 'Equipment and building maintenance'),
    ('Marketing', 'Advertising and promotions'),
    ('Supplies', 'Non-inventory operational supplies'),
    ('Other', 'Miscellaneous operational expenses');

CREATE TABLE expenses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES expense_categories(id),
    amount DECIMAL(15, 2) NOT NULL CHECK (amount > 0),
    expense_date DATE NOT NULL,
    notes TEXT,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_expenses_store_date ON expenses(store_id, expense_date);
CREATE INDEX idx_expenses_category ON expenses(category_id);

-- +goose Down
DROP TABLE expenses;
DROP TABLE expense_categories;
