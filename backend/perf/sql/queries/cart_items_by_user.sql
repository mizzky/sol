\set ON_ERROR_STOP ON

\ir ../seed/00_guard.sql

\if :{?user_id}
\else
\set user_id 2
\endif

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
WHERE c.user_id = :user_id
ORDER BY ci.id;