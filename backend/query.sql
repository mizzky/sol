-- name: GetProduct :one
SELECT * FROM products
WHERE id = $1 LIMIT 1;

-- name: ListProducts :many
SELECT * FROM products
ORDER BY name;

-- name: CreateProduct :one
INSERT INTO products (
    name, price, is_available
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: CreateUser :one
INSERT INTO users (
    name, email, password_hash, role
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users 
WHERE email = $1 LIMIT 1;

-- name: GetUserForUpdate :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: CreateCategory :one
INSERT INTO categories (
    name, description
) VALUES (
    $1, $2
)
RETURNING id, name, description, created_at, updated_at;

-- name: GetCategory :one
SELECT id, name, description, created_at, updated_at
FROM categories
WHERE id = $1;

-- name: ListCategories :many
SELECT id, name, description, created_at, updated_at
FROM categories
ORDER BY name;

-- name: UpdateCategory :one
UPDATE categories
SET
    name = $2,
    description = $3,
    updated_at = NO()
WHERE id = $1
RETURNING id, name, description, created_at, updated_at;

-- name: DeleteCategory :exec
DELETE FROM categories
WHERE id = $1;

