\set ON_ERROR_STOP ON

\ir ../seed/00_guard.sql

\if :{?cursor}
\else
\set cursor 0
\endif

EXPLAIN(ANALYZE, BUFFERS, FORMAT TEXT)
SELECT
    id,
    name,
    price,
    is_available,
category_id,
    sku,
    description,
    image_url,
    stock_quantity,
    created_at,
    updated_at
FROM public.products
WHERE id > :cursor
ORDER BY id
LIMIT 50;
