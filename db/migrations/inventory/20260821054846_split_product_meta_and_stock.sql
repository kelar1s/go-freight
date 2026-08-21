-- +goose Up
-- +goose StatementBegin
CREATE TABLE
   product_stocks (
      product_id BIGINT PRIMARY KEY REFERENCES products (id) ON DELETE CASCADE,
      quantity BIGINT NOT NULL DEFAULT 0 CHECK (quantity >= 0),
      reserved BIGINT NOT NULL DEFAULT 0 CHECK (reserved >= 0),
      CONSTRAINT check_reserved_not_exceed_quantity CHECK (reserved <= quantity)
   );

INSERT INTO
   product_stocks (product_id, quantity, reserved)
SELECT
   id,
   quantity,
   reserved
FROM
   products;

ALTER TABLE products
DROP COLUMN quantity;

ALTER TABLE products
DROP COLUMN reserved;

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
ALTER TABLE products
ADD COLUMN quantity BIGINT NOT NULL DEFAULT 0 CHECK (quantity >= 0),
ADD COLUMN reserved BIGINT NOT NULL DEFAULT 0 CHECK (reserved >= 0);

UPDATE products p
SET
   quantity = s.quantity,
   reserved = s.reserved
FROM
   product_stocks s
WHERE
   p.id = s.product_id;

ALTER TABLE products ADD CONSTRAINT check_reserved_not_exceed_quantity CHECK (reserved <= quantity);

DROP TABLE product_stocks;

-- +goose StatementEnd