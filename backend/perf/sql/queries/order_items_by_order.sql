\set ON_ERROR_STOP ON

\ir ../seed/00_guard.sql

\if :{?order_id}
\else
\set order_id 1
\endif

EXPLAIN(ANALYZE, BUFFERS, FORMAT TEXT)
SELECT
    id, order_id, product_id, quantity, unit_price, product_name_snapshot, created_at, updated_at
FROM
    public.order_items
WHERE
    order_id = :order_id
ORDER BY id;