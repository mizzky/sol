-- name: GetProduct :one
SELECT
    id, name, price, is_available, category_id, sku, description, image_url, stock_quantity, created_at, updated_at
FROM products
WHERE id = $1;

-- name: ListProducts :many
SELECT
    id, name, price, is_available, category_id, sku, description, image_url, stock_quantity, created_at, updated_at
FROM products
ORDER BY id;

-- name: CreateProduct :one
INSERT INTO products (
    name, price, is_available, category_id, sku, description, image_url, stock_quantity
) VALUES (
    @name, @price, @is_available, @category_id, @sku, @description, @image_url, @stock_quantity
)
RETURNING id, name, price, is_available, category_id, sku, description, image_url, stock_quantity, created_at, updated_at;

-- name: UpdateProduct :one
UPDATE products
SET
    name = COALESCE(@name, name),
    price = COALESCE(@price, price),
    is_available = COALESCE(@is_available, is_available),
    category_id = COALESCE(@category_id, category_id),
    sku = COALESCE(@sku, sku),
    description = COALESCE(@description, description),
    image_url = COALESCE(@image_url, image_url),
    stock_quantity = COALESCE(@stock_quantity, stock_quantity),
    updated_at = NOW()
WHERE id = @id
RETURNING id, name, price, is_available, category_id, sku, description, image_url, stock_quantity, created_at, updated_at;

-- name: DeleteProduct :exec
DELETE FROM products
WHERE id = $1;

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

