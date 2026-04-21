-- name: CreateProduct :one
INSERT INTO products(warehouse_id, name, quantity, reserved) VALUES($1, $2, $3, 0) RETURNING *;

-- name: GetProduct :one
SELECT * FROM products WHERE id = $1 LIMIT 1;

-- name: ListProductsByWarehouse :many
SELECT * FROM products WHERE warehouse_id = $1 ORDER BY id;

-- name: SetProductQuantity :one
UPDATE products SET quantity = $2 WHERE id = $1 RETURNING id;

-- name: AddProductQuantity :one
UPDATE products SET quantity = quantity + $2 WHERE id = $1 RETURNING id;

-- name: DeleteProduct :one
DELETE FROM products WHERE id = $1 RETURNING id;

-- name: ReserveProduct :one
UPDATE products SET reserved = reserved + $2 WHERE id = $1 AND (quantity - reserved) >= $2 RETURNING id;

-- name: ReleaseProduct :one
UPDATE products SET quantity = quantity - $2, reserved = reserved - $2 WHERE id = $1 AND reserved >= $2 AND quantity >= $2 RETURNING id;

-- name: CancelReservation :one
UPDATE products SET reserved = reserved - $2 WHERE id = $1 AND reserved >= $2 RETURNING id;