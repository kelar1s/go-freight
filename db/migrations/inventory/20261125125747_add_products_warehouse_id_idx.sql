-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_products_warehouse_id ON products (warehouse_id);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_products_warehouse_id;

-- +goose StatementEnd