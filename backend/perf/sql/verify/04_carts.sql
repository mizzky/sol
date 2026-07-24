\set ON_ERROR_STOP on

\ir ../seed/00_guard.sql
\ir ../seed/00_profile.sql

DO $verify$
DECLARE
    actual_carts BIGINT;
    min_cart_id BIGINT;
    max_cart_id BIGINT;

    actual_cart_items BIGINT;
    min_cart_item_id BIGINT;
    max_cart_item_id BIGINT;

    invalid_carts BIGINT;
    invalid_cart_distribution BIGINT;
    duplicate_cart_users BIGINT;
    duplicate_cart_products BIGINT;

    orphan_cart_users BIGINT;
    orphan_cart_items_carts BIGINT;
    orphan_cart_items_products BIGINT;

    invalid_cart_items BIGINT;
    invalid_cart_item_products BIGINT;
    invalid_cart_item_prices BIGINT;

    carts_sequence_value BIGINT;
    carts_sequence_called BOOLEAN;
    cart_items_sequence_value BIGINT;
    cart_items_sequence_called BOOLEAN;

    expected_carts BIGINT;
    expected_cart_items BIGINT;

BEGIN
    SELECT carts_count
    INTO expected_carts
    FROM pg_temp.perf_profile;

    expected_cart_items := 100 + 3 * (expected_carts - 2);

    SELECT count(*), min(id), max(id)
    INTO actual_carts, min_cart_id, max_cart_id
    FROM public.carts;

    IF actual_carts <> expected_carts
       OR min_cart_id IS DISTINCT FROM 1
       OR max_cart_id IS DISTINCT FROM expected_carts THEN
        RAISE EXCEPTION
            'carts range mismatch: expected count/min/max %/1/%, actual %/%/%',
            expected_carts,
            expected_carts,
            actual_carts,
            min_cart_id,
            max_cart_id;
    END IF;

    SELECT count(*), min(id), max(id)
    INTO actual_cart_items, min_cart_item_id, max_cart_item_id
    FROM public.cart_items;

    IF actual_cart_items <> expected_cart_items
       OR min_cart_item_id IS DISTINCT FROM 1
       OR max_cart_item_id IS DISTINCT FROM expected_cart_items THEN
        RAISE EXCEPTION
            'cart_items range mismatch: expected count/min/max %/1/%, actual %/%/%',
            expected_cart_items,
            expected_cart_items,
            actual_cart_items,
            min_cart_item_id,
            max_cart_item_id;
    END IF;

    SELECT count(*)
    INTO invalid_carts
    FROM public.carts
    WHERE user_id IS DISTINCT FROM id
       OR user_id NOT BETWEEN 1 AND expected_carts
       OR created_at IS DISTINCT FROM TIMESTAMPTZ '2025-01-01 00:00:00+00'
       OR updated_at IS DISTINCT FROM TIMESTAMPTZ '2025-01-01 00:00:00+00';

    IF invalid_carts <> 0 THEN
        RAISE EXCEPTION
            'invalid carts found: %',
            invalid_carts;
    END IF;

    SELECT count(*)
    INTO invalid_cart_distribution
    FROM (
        SELECT
            carts.id AS cart_id,
            count(cart_items.id) AS item_count
        FROM public.carts AS carts
        LEFT JOIN public.cart_items AS cart_items
            ON cart_items.cart_id = carts.id
        GROUP BY carts.id
    ) AS cart_counts
    WHERE item_count IS DISTINCT FROM
          CASE
              WHEN cart_id = 1 THEN 0
              WHEN cart_id = 2 THEN 100
              ELSE 3
          END;

    IF invalid_cart_distribution <> 0 THEN
        RAISE EXCEPTION
            'cart item distribution mismatch: mismatched carts %',
            invalid_cart_distribution;
    END IF;

    SELECT count(*) - count(DISTINCT user_id)
    INTO duplicate_cart_users
    FROM public.carts;

    IF duplicate_cart_users <> 0 THEN
        RAISE EXCEPTION
            'duplicate cart users found: %',
            duplicate_cart_users;
    END IF;

    SELECT count(*)
    INTO duplicate_cart_products
    FROM (
        SELECT cart_id, product_id
        FROM public.cart_items
        GROUP BY cart_id, product_id
        HAVING count(*) > 1
    ) AS duplicates;

    IF duplicate_cart_products <> 0 THEN
        RAISE EXCEPTION
            'duplicate cart products found: %',
            duplicate_cart_products;
    END IF;

    SELECT count(*)
    INTO orphan_cart_users
    FROM public.carts AS carts
    LEFT JOIN public.users AS users
        ON users.id = carts.user_id
    WHERE users.id IS NULL;

    IF orphan_cart_users <> 0 THEN
        RAISE EXCEPTION
            'orphan cart users found: %',
            orphan_cart_users;
    END IF;

    SELECT count(*)
    INTO orphan_cart_items_carts
    FROM public.cart_items AS cart_items
    LEFT JOIN public.carts AS carts
        ON carts.id = cart_items.cart_id
    WHERE carts.id IS NULL;

    IF orphan_cart_items_carts <> 0 THEN
        RAISE EXCEPTION
            'orphan cart_items carts found: %',
            orphan_cart_items_carts;
    END IF;

    SELECT count(*)
    INTO orphan_cart_items_products
    FROM public.cart_items AS cart_items
    LEFT JOIN public.products AS products
        ON products.id = cart_items.product_id
    WHERE products.id IS NULL;

    IF orphan_cart_items_products <> 0 THEN
        RAISE EXCEPTION
            'orphan cart_items products found: %',
            orphan_cart_items_products;
    END IF;

    SELECT count(*)
    INTO invalid_cart_items
    FROM public.cart_items
    WHERE quantity <= 0
       OR created_at IS DISTINCT FROM TIMESTAMPTZ '2025-01-01 00:00:00+00'
       OR updated_at IS DISTINCT FROM TIMESTAMPTZ '2025-01-01 00:00:00+00';

    IF invalid_cart_items <> 0 THEN
        RAISE EXCEPTION
            'invalid cart_items found: %',
            invalid_cart_items;
    END IF;

    SELECT count(*)
    INTO invalid_cart_item_products
    FROM public.cart_items AS cart_items
    JOIN public.products AS products
        ON products.id = cart_items.product_id
    WHERE NOT products.is_available
       OR products.stock_quantity <= 0;

    IF invalid_cart_item_products <> 0 THEN
        RAISE EXCEPTION
            'invalid cart item products found: %',
            invalid_cart_item_products;
    END IF;

    SELECT count(*)
    INTO invalid_cart_item_prices
    FROM public.cart_items AS cart_items
    JOIN public.products AS products
        ON products.id = cart_items.product_id
    WHERE cart_items.price IS DISTINCT FROM products.price;

    IF invalid_cart_item_prices <> 0 THEN
        RAISE EXCEPTION
            'cart item price mismatch found: %',
            invalid_cart_item_prices;
    END IF;

    SELECT last_value, is_called
    INTO carts_sequence_value, carts_sequence_called
    FROM public.carts_id_seq;

    IF carts_sequence_value <> expected_carts
       OR NOT carts_sequence_called THEN
        RAISE EXCEPTION
            'carts sequence mismatch: value %, called %',
            carts_sequence_value,
            carts_sequence_called;
    END IF;

    SELECT last_value, is_called
    INTO cart_items_sequence_value, cart_items_sequence_called
    FROM public.cart_items_id_seq;

    IF cart_items_sequence_value <> expected_cart_items
       OR NOT cart_items_sequence_called THEN
        RAISE EXCEPTION
            'cart_items sequence mismatch: value %, called %',
            cart_items_sequence_value,
            cart_items_sequence_called;
    END IF;
END
$verify$;

\echo 'PASS: carts and cart_items'