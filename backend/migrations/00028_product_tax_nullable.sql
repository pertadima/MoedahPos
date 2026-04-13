-- +goose Up
-- +goose StatementBegin

ALTER TABLE products ALTER COLUMN tax_percentage DROP NOT NULL;
ALTER TABLE products ALTER COLUMN tax_percentage DROP DEFAULT;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE products ALTER COLUMN tax_percentage SET DEFAULT 0;
UPDATE products SET tax_percentage = 0 WHERE tax_percentage IS NULL;
ALTER TABLE products ALTER COLUMN tax_percentage SET NOT NULL;

-- +goose StatementEnd
