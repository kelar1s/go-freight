-- name: CreateWarehouse :one
INSERT INTO warehouses(name, location) VALUES($1, $2) RETURNING *;

-- name: GetWarehouse :one
SELECT * FROM warehouses WHERE id = $1 LIMIT 1;

-- name: ListWarehouses :many
SELECT * FROM warehouses;

-- name: UpdateWarehouse :exec
UPDATE warehouses SET name = $2, location = $3 WHERE id = $1;

-- name: DeleteWarehouse :exec
DELETE FROM warehouses WHERE id = $1;