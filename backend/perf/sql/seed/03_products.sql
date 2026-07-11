\set ON_ERROR_STOP ON

BEGIN;

\ir 00_guard.sql

WITH product_source AS (
    SELECT
        n,
        100 + ((n - 1) % 1000) * 10 AS price,
        ((n - 1) / 20) % 10 AS stock_bucket,
        (n - 1) % 100 AS category_bucket
    FROM generate_series(1, 10000) AS series(n)
)
INSERT INTO public.products (
    id,
    name,
    price,
    is_available,
    category_id,
    sku,
    description,
    image_url,
    stock_quantity,
    created_at ,
    updated_at
)
SELECT
    n AS id,
    format('product-%s', lpad(n::text, 7, '0')) AS name,
    price,
    (n %20 <> 0) AS is_available,
    CASE 
        WHEN category_bucket < 40 THEN 1
        WHEN category_bucket < 60 THEN 2 
        ELSE  3 + ((category_bucket - 60) / 5)
    END AS category_id,

    format('PERF-%s', lpad(n::text, 7, '0')) AS sku,
    format (
        'Performance product %s',
        lpad(n::text, 7, '0')
    ) AS description,

    format(
        'https://example.test/products/%s.jpg',
        lpad(n::text, 7, '0')
    ) AS image_url,

    CASE 
        WHEN stock_bucket = 0 THEN 0
        WHEN stock_bucket IN (1, 2) THEN 1 + ((n - 1) % 5)
        WHEN stock_bucket BETWEEN 3 AND 8 THEN 10 + ((n - 1) %91)
        WHEN stock_bucket = 9 THEN 500 + ((n - 1) % 501)
    END AS stock_quantity,

    TIMESTAMPTZ '2025-01-01 00:00:00+00' AS created_at,
    TIMESTAMPTZ '2025-01-01 00:00:00+00' AS updated_at
FROM product_source
ORDER BY n;

SELECT setval(
    pg_get_serial_sequence('public.products', 'id'),
    (SELECT max(id) FROM public.products),
    true
);

COMMIT;


 
