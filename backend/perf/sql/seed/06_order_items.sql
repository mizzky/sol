\set ON_ERROR_STOP on

BEGIN;

\ir 00_guard.sql
\ir 00_profile.sql


-- INSERT
WITH order_item_plan AS (
    SELECT
        orders.id AS order_id,
        orders.created_at,
        CASE 
            WHEN orders.id % 10 IN (1, 2) THEN 1
            WHEN orders.id % 10 IN (3, 4, 5, 6, 7) THEN 2
            ELSE  6
        END AS item_count
    FROM public.orders AS orders
)
INSERT INTO public.order_items(
    id,
    order_id,
    product_id,
    quantity,
    unit_price,
    product_name_snapshot,
    created_at
)
SELECT 
    row_number() OVER (
        ORDER BY plan.order_id, item_series.item_ordinal
    ) AS id,
    plan.order_id,
    products.id AS product_id,
    1 + ((plan.order_id + item_series.item_ordinal - 2) % 3) AS quantity,
    products.price AS unit_price,
    products.name AS product_name_snapshot,
    plan.created_at
FROM order_item_plan AS plan
CROSS JOIN pg_temp.perf_profile AS config
CROSS JOIN generate_series(1, plan.item_count) AS item_series(item_ordinal)
JOIN public.products AS products
    ON products.id = (
        ((plan.order_id - 1) * 6 + item_series.item_ordinal - 1) % config.products_count
    ) + 1
ORDER BY plan.order_id, item_series.item_ordinal;

UPDATE public.orders AS orders
SET total = item_totals.total
FROM (
    SELECT
        order_id,
        sum(quantity * unit_price) AS total
    FROM public.order_items
    GROUP BY order_id
) AS item_totals
WHERE item_totals.order_id = orders.id;


-- sync sequence
SELECT setval(
    pg_get_serial_sequence('public.order_items', 'id'),
    (SELECT max(id) FROM public.order_items),
    true
);


COMMIT;

