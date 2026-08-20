\set ON_ERROR_STOP ON

\ir ../seed/00_guard.sql

BEGIN;

DROP INDEX idx_users_email;
DROP INDEX idx_categories_name;
DROP INDEX idx_products_sku;
DROP INDEX idx_carts_user_id;
DROP INDEX idx_cart_items_cart_id;

\echo 'users.email lookup'

EXPLAIN(ANALYZE, BUFFERS, FORMAT TEXT)
SELECT *
FROM public.users
WHERE email = 'user-050000@example.test'
LIMIT 1;

\echo 'categories.name ordering'
EXPLAIN(ANALYZE, BUFFERS, FORMAT TEXT)
SELECT
    id, name, description, created_at, updated_at
FROM public.categories
ORDER BY name;

\echo 'products.sku lookup'
EXPLAIN(ANALYZE, BUFFERS, FORMAT TEXT)
SELECT id
FROM public.products
WHERE sku = 'PERF-0500000';

\echo 'carts.user_id lookup'
EXPLAIN(ANALYZE, BUFFERS, FORMAT TEXT)
SELECT id, user_id, created_at, updated_at
FROM public.carts
WHERE user_id = 2
LIMIT 1;

\echo 'cart_items.cart_id lookup'
EXPLAIN(ANALYZE, BUFFERS, FORMAT TEXT)
SELECT
    id, cart_id, product_id, quantity, price, created_at, updated_at
FROM public.cart_items
WHERE cart_id = 2
ORDER BY id;

\echo 'cart items by user join'
EXPLAIN(ANALYZE, BUFFERS, FORMAT TEXT)
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
FROM public.cart_items AS ci
JOIN public.carts AS c
    ON ci.cart_id = c.id
JOIN public.products AS p
    ON p.id = ci.product_id
WHERE c.user_id = 2
ORDER BY ci.id;

ROLLBACK;