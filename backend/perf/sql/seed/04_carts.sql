\set ON_ERROR_STOP ON

BEGIN;

\ir 00_guard.sql

-- 1. carts 500件作成
INSERT INTO public.carts(
    id,
    user_id,
    created_at,
    updated_at
)
SELECT
    n as id,
    n AS user_id,
    TIMESTAMPTZ '2025-01-01 00:00:00+00' AS created_at,
    TIMESTAMPTZ '2025-01-01 00:00:00+00' AS updated_at
FROM generate_series(1, 500) AS series(n)
ORDER BY n;

WITH cart_item_plan AS (
    -- cart2: 100明細
    SELECT
        2 AS cart_id,
        item_ordinal
    FROM generate_series(1, 100) AS item_series(item_ordinal)

    UNION ALL

    -- cart3-500: 各3明細
    SELECT
        cart_id,
        item_ordinal
    FROM generate_series(3, 500) AS cart_series(cart_id)
    CROSS JOIN generate_series(1, 3) AS item_series(item_ordinal)
),
product_candidates AS (
    SELECT
        id AS product_id,
        price,
        row_number() OVER (ORDER BY id) AS candidate_ordinal,
        count(*) OVER () AS candidate_count
    FROM public.products
    WHERE is_available = TRUE
    AND stock_quantity > 0
)
INSERT INTO public.cart_items(
    id,
    cart_id,
    product_id,
    quantity,
    price,
    created_at,
    updated_at
)
SELECT
    row_number() OVER (ORDER BY plan.cart_id, plan.item_ordinal) AS id,
    plan.cart_id,
    products.product_id,
    1 + ((plan.cart_id + plan.item_ordinal - 2) % 3) AS quantity,
    products.price,
    TIMESTAMPTZ '2025-01-01 00:00:00+00' AS created_at,
    TIMESTAMPTZ '2025-01-01 00:00:00+00' AS updated_at
FROM cart_item_plan AS PLAN
JOIN product_candidates AS products
-- 最大容量cart2の幅100に合わせ、各カートを100飛ばしで採番することで視認性を上げる
-- cart2:101-200 cart3:201-203 cart4:301-303...
    ON products.candidate_ordinal = (((plan.cart_id - 1) * 100 + plan.item_ordinal - 1) % products.candidate_count) + 1
ORDER BY plan.cart_id, plan.item_ordinal;

SELECT setval(
    pg_get_serial_sequence('public.carts', 'id'),
    (SELECT MAX(id) FROM public.carts),
    true
);

SELECT setval(
    pg_get_serial_sequence('public.cart_items', 'id'),
    (SELECT MAX(id) FROM public.cart_items),
    true
);


COMMIT;

