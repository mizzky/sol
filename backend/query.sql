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
    updated_at = NOW()
WHERE id = $1
RETURNING id, name, description, created_at, updated_at;

-- name: DeleteCategory :exec
DELETE FROM categories
WHERE id = $1;


-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: UpdateUserRole :one
UPDATE users
SET role = @role,
    updated_at = NOW()
WHERE id = @id
RETURNING *;

-- name: SetResetToken :one
UPDATE users
SET reset_token = @reset_token,
    updated_at = NOW()
WHERE id = @id
RETURNING *;

-- name: CreateCart :one
 INSERT INTO carts (user_id, created_at, updated_at)
 VALUES($1, NOW(), NOW())
 RETURNING id, user_id, created_at, updated_at;

-- name: GetCartByUser :one
 SELECT id, user_id, created_at, updated_at
 FROM carts
 WHERE user_id = $1
 LIMIT 1;

-- name: GetOrCreateCartForUser :one
-- Requires UNIQUE(user_id) on carts
 INSERT INTO carts(user_id, created_at, updated_at)
 VALUES($1, NOW(), NOW())
 ON CONFLICT (user_id) DO UPDATE SET updated_at = carts.updated_at
 RETURNING id, user_id, created_at, updated_at;

-- name: ListCartItems :many
 SELECT
    ci.id,
    ci.cart_id,
    ci.product_id,
    ci.quantity,
    ci.price,
    ci.created_at,
    ci.updated_at,
    p.name AS product_name,
    p.price AS product_price,
    p.stock_quantity AS product_stock
FROM cart_items ci
JOIN products p ON p.id = ci.product_id
WHERE ci.cart_id = $1
ORDER BY ci.id;

-- name: ListCartItemsByUser :many
SELECT
    ci.id,
    ci.cart_id,
    ci.product_id,
    ci.quantity,
    ci.price,
    ci.created_at,
    ci.updated_at,
    p.name AS product_name,
    p.price AS product_price,
    p.stock_quantity AS product_stock
FROM cart_items ci
JOIN carts c ON ci.cart_id = c.id
JOIN products p ON p.id = ci.product_id
WHERE c.user_id = $1
ORDER BY ci.id;

-- name: AddCartItem :one
-- Requires UNIQUE(cart_id, product_id) on cart_items
INSERT INTO cart_items (cart_id, product_id, quantity, price, created_at, updated_at)
VALUES ($1, $2, $3, $4, NOW(), NOW())
ON CONFLICT(cart_id, product_id) DO UPDATE
SET quantity = cart_items.quantity + EXCLUDED.quantity,
    price = EXCLUDED.price,
    updated_at = NOW()
RETURNING id, cart_id, product_id, quantity, price, created_at, updated_at;

-- name: GetCartItemByID :one
SELECT id, cart_id, product_id, quantity, price, created_at, updated_at
FROM cart_items
WHERE id = $1
LIMIT 1;

-- name: UpdateCartItemQty :one
UPDATE cart_items
SET quantity = $2, updated_at = NOW()
WHERE id = $1
RETURNING id, cart_id, product_id, quantity, price, created_at, updated_at;

-- name: UpdateCartItemQtyByUser :one
UPDATE cart_items ci
SET quantity = $2, updated_at = NOW()
FROM carts c
WHERE ci.id = $1
and ci.cart_id = c.id
AND c.user_id = $3
RETURNING ci.id, ci.cart_id, ci.product_id, ci.quantity, ci.price, ci.created_at, ci.updated_at;

-- name: RemoveCartItem :exec
DELETE FROM cart_items
WHERE id = $1;

-- name: RemoveCartItemByUser :exec
DELETE FROM cart_items ci
USING carts c
WHERE ci.id = $1
AND ci.cart_id = c.id
AND c.user_id = $2;

-- name: ClearCart :exec
DELETE FROM cart_items
WHERE cart_id = $1;

-- name: ClearCartByUser :exec
DELETE FROM cart_items
WHERE cart_id = (
    SELECT id FROM carts WHERE user_id = $1
);