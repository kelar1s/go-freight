-- name: CreateProductMeta :one
INSERT INTO products(warehouse_id, name) 
VALUES($1, $2) RETURNING *;

-- name: CreateProductStock :exec
INSERT INTO product_stocks(product_id, quantity) 
VALUES($1, $2);

-- name: GetProductMeta :one
SELECT id, warehouse_id, name, created_at 
FROM products 
WHERE id = $1 LIMIT 1;

-- name: GetProductStock :one
SELECT product_id, quantity, reserved 
FROM product_stocks 
WHERE product_id = $1 LIMIT 1;

-- name: DeleteProduct :one
DELETE FROM products WHERE id = $1 RETURNING id;

-- name: ListProductsByWarehouse :many
SELECT p.id, p.warehouse_id, p.name, p.created_at, s.quantity, s.reserved 
FROM products p
JOIN product_stocks s ON p.id = s.product_id
WHERE p.warehouse_id = $1 
ORDER BY p.id;

-- name: AdjustProductQuantity :one
UPDATE product_stocks 
SET quantity = quantity + $2 
WHERE product_id = $1 AND (quantity + $2) >= reserved 
RETURNING product_id;

-- name: ReserveProduct :one
UPDATE product_stocks 
SET reserved = reserved + $2 
WHERE product_id = $1 AND $2 > 0 AND (quantity - reserved) >= $2 
RETURNING product_id;

-- name: ReleaseProduct :one
UPDATE product_stocks 
SET quantity = quantity - $2, reserved = reserved - $2 
WHERE product_id = $1 AND $2 > 0 AND reserved >= $2 AND quantity >= $2 
RETURNING product_id;

-- name: CancelReservation :one
UPDATE product_stocks 
SET reserved = reserved - $2 
WHERE product_id = $1 AND $2 > 0 AND reserved >= $2 
RETURNING product_id;