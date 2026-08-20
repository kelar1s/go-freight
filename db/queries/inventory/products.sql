-- name: CreateProduct :one
INSERT INTO products(warehouse_id, name, quantity, reserved) VALUES($1, $2, $3, 0) RETURNING *;

-- name: GetProduct :one
SELECT * FROM products WHERE id = $1 LIMIT 1;

-- name: DeleteProduct :one
DELETE FROM products WHERE id = $1 RETURNING id;

-- name: ListProductsByWarehouse :many
SELECT * FROM products WHERE warehouse_id = $1 ORDER BY id;

-- name: AdjustProductQuantity :one
UPDATE products SET quantity = quantity + $2 WHERE id = $1 AND (quantity + $2) >= reserved RETURNING id;

-- name: ReserveProduct :one
UPDATE products SET reserved = reserved + $2 WHERE id = $1 AND $2 > 0 AND (quantity - reserved) >= $2 RETURNING id;

-- name: ReleaseProduct :one
UPDATE products SET quantity = quantity - $2, reserved = reserved - $2 WHERE id = $1 AND $2 > 0 AND reserved >= $2 AND quantity >= $2 RETURNING id;

-- name: CancelReservation :one
UPDATE products SET reserved = reserved - $2 WHERE id = $1 AND $2 > 0 AND reserved >= $2 RETURNING id;