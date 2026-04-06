-- +goose Up
-- +goose StatementBegin
CREATE TABLE
   warehouses (
      id SERIAL PRIMARY KEY,
      name TEXT NOT NULL,
      location TEXT NOT NULL,
      created_at TIMESTAMPTZ DEFAULT now () NOT NULL
   );

CREATE TABLE
   products (
      id SERIAL PRIMARY KEY,
      warehouse_id INTEGER REFERENCES warehouses (id) NOT NULL,
      name TEXT NOT NULL,
      quantity INTEGER DEFAULT 0 NOT NULL CHECK (quantity >= 0),
      created_at TIMESTAMPTZ DEFAULT now () NOT NULL
   );

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE products;

DROP TABLE warehouses;

-- +goose StatementEnd