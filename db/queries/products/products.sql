-- name: CreateProduct :one
INSERT INTO products(warehouse_id, name, quantity) VALUES($1, $2, $3) RETURNING *;

-- name: GetProduct :one
SELECT * FROM products WHERE id = $1 LIMIT 1;

-- name: ListProductsByWarehouse :many
SELECT * FROM products WHERE warehouse_id = $1 ORDER BY id;

-- name: SetProductQuantity :exec
UPDATE products SET quantity = $2 WHERE id = $1;

-- name: AddProductQuantity :exec
UPDATE products SET quantity = quantity + $2 WHERE id = $1;

-- name: DeleteProduct :exec
DELETE FROM products WHERE id = $1;