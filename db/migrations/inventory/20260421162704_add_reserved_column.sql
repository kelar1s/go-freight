-- +goose Up
-- +goose StatementBegin
ALTER TABLE products
ADD COLUMN reserved INTEGER NOT NULL DEFAULT 0,
ADD CONSTRAINT check_reserved_not_exceed_quantity CHECK (reserved <= quantity),
ADD CONSTRAINT check_reserved_positive CHECK (reserved >= 0);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
ALTER TABLE products
DROP COLUMN reserved;

-- +goose StatementEnd