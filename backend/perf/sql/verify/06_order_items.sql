\set ON_ERROR_STOP ON

\ir ../seed/00_guard.sql

DO $verify$
DECLARE
    actual_order_items BIGINT;
    min_order_items_id BIGINT;
    max_order_items_id BIGINT;

    orphan_order_items_orders BIGINT;
    orphan_order_items_products BIGINT;

    invalid_quantity BIGINT;
    invalid_price_snapshot BIGINT;
    invalid_name_snapshot BIGINT;
    duplicate_order_products BIGINT;

    one_item_orders BIGINT;
    two_item_orders BIGINT;
    six_item_orders BIGINT;

    order_total_mismatches BIGINT;

    order_items_sequence_value BIGINT;
    order_items_sequence_called BOOLEAN;

BEGIN
    SELECT count(*), min(id), max(id)
    INTO actual_order_items, min_order_items_id, max_order_items_id
    FROM public.order_items;

    IF actual_order_items <> 30000
        OR min_order_items_id IS DISTINCT FROM 1
        OR max_order_items_id IS DISTINCT FROM 30000 THEN
            RAISE EXCEPTION
                'order items range mismatch: expected count/min/max 30000/1/30000,  actual %/%/%',
                actual_order_items, min_order_items_id, max_order_items_id;
    END IF;

    SELECT count(*)
    INTO orphan_order_items_orders
    FROM public.order_items AS order_items
    LEFT JOIN public.orders AS orders
    ON orders.id = order_items.order_id
    WHERE orders.id IS NULL;

    IF orphan_order_items_orders <> 0 THEN
        RAISE EXCEPTION
            'orphan order items orders found: %',
            orphan_order_items_orders;
    END IF;

    SELECT count(*)
    INTO orphan_order_items_products
    FROM public.order_items AS order_items
    LEFT JOIN public.products AS products
    ON products.id = order_items.product_id
    WHERE products.id IS NULL;

    IF orphan_order_items_products <> 0 THEN
        RAISE EXCEPTION
            'orphan order items products found: %',
            orphan_order_items_products;
    END IF;

    SELECT count(*)
    INTO invalid_quantity
    FROM public.order_items
    WHERE quantity NOT BETWEEN 1 AND 3;

    IF invalid_quantity <> 0 THEN
        RAISE EXCEPTION
            'invalid order item quantity found: %',
            invalid_quantity;
    END IF;

    SELECT count(*)
    INTO invalid_price_snapshot
    FROM public.order_items AS order_items
    JOIN public.products AS products
    ON products.id = order_items.product_id
    WHERE order_items.unit_price IS DISTINCT FROM products.price;

    IF invalid_price_snapshot <> 0 THEN
        RAISE EXCEPTION
            'invalid order items price snapshot found: %',
            invalid_price_snapshot;
    END IF;

    SELECT count(*)
    INTO invalid_name_snapshot
    FROM public.order_items AS order_items
    JOIN public.products AS products
    ON products.id = order_items.product_id
    WHERE order_items.product_name_snapshot IS DISTINCT FROM products.name;

    IF invalid_name_snapshot <> 0 THEN
        RAISE EXCEPTION
            'invalid order item name snapshot found: %',
            invalid_name_snapshot;
    END IF;

    SELECT count(*)
    INTO duplicate_order_products
    FROM (
        SELECT order_id, product_id
        FROM public.order_items
        GROUP BY order_id, product_id
        HAVING count(*) > 1
    ) AS duplicates;

    IF duplicate_order_products <> 0 THEN
        RAISE EXCEPTION
            'duplicate order products found: %',
            duplicate_order_products;
    END IF;

    SELECT
        count(*) FILTER (WHERE item_count = 1),
        count(*) FILTER (WHERE item_count = 2),
        count(*) FILTER (WHERE item_count = 6)
    INTO one_item_orders, two_item_orders, six_item_orders
    FROM (
        SELECT
            orders.id AS order_id,
            count(order_items.id) AS item_count
        FROM public.orders AS orders
        LEFT JOIN public.order_items AS order_items
        ON order_items.order_id = orders.id
        GROUP BY orders.id
    ) AS order_item_counts;

    IF one_item_orders <> 2000 OR two_item_orders <> 5000 OR six_item_orders <> 3000 THEN
        RAISE EXCEPTION
            'order items by orders mismatch: expected one/two/six 2000/5000/3000,  actual %/%/%',
            one_item_orders, two_item_orders, six_item_orders;
    END IF;

    WITH calculated_order_items AS(
        SELECT
            order_id,
            sum(quantity * unit_price) AS calculated_total
        FROM public.order_items
        GROUP BY order_id
    ) 
    SELECT count(*)
    INTO order_total_mismatches
    FROM calculated_order_items
    JOIN public.orders AS orders
    ON orders.id = calculated_order_items.order_id
    WHERE calculated_total IS DISTINCT FROM orders.total;

    IF order_total_mismatches <> 0 THEN
        RAISE EXCEPTION
            'order total mismatches found: %',
            order_total_mismatches;
    END IF;


    SELECT last_value, is_called
    INTO order_items_sequence_value, order_items_sequence_called
    FROM public.order_items_id_seq;

    IF order_items_sequence_value <> 30000 OR NOT order_items_sequence_called THEN
        RAISE EXCEPTION
            'order items sequence mismatch: value %, called %',
            order_items_sequence_value, order_items_sequence_called;
    END IF;

END ;
$verify$;

\echo 'PASS: order items'
