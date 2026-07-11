\set ON_ERROR_STOP ON

\ir ../seed/00_guard.sql

DO $verify$
DECLARE
    actual_orders BIGINT;
    min_order_id BIGINT;
    max_order_id BIGINT;

    user1_orders BIGINT;
    pending_orders BIGINT;
    cancelled_orders BIGINT;

    invalid_status_time BIGINT;
    invalid_created_at BIGINT;

    latest_same_time_orders BIGINT;
    invalid_latest_page_orders BIGINT;

    orphan_order_users BIGINT;
    orders_sequence_value BIGINT;
    orders_sequence_called BOOLEAN;

BEGIN
    SELECT count(*), min(id), max(id)
    INTO actual_orders, min_order_id, max_order_id
    FROM public.orders;

    IF actual_orders <> 10000
        OR min_order_id IS DISTINCT FROM 1
        OR max_order_id IS DISTINCT FROM 10000 THEN
            RAISE EXCEPTION
                'orders range mismatch: expected count/min/max 10000/1/10000,  actual %/%/%',
                actual_orders, min_order_id, max_order_id;
    END IF;

    SELECT count(*)
    INTO user1_orders
    FROM public.orders
    WHERE user_id = 1;

    IF user1_orders <> 2000 THEN
        RAISE EXCEPTION
            'user_id=1 orders mismatch: expected 2000, actual %',
            user1_orders;
    END IF;

    SELECT
        count(*) FILTER (WHERE status = 'pending'),
        count(*) FILTER (WHERE status = 'cancelled')
    INTO pending_orders, cancelled_orders
    FROM public.orders;

    IF pending_orders <> 9000 OR cancelled_orders <> 1000 THEN
        RAISE EXCEPTION
            'status distribution mismatch: expected pending/cancelled 9000/1000, actual %/%',
            pending_orders, cancelled_orders;
    END IF;

    SELECT count(*)
    INTO invalid_status_time
    FROM public.orders
    WHERE (status = 'pending' AND cancelled_at IS NOT NULL)
    OR (status = 'cancelled' AND cancelled_at IS DISTINCT FROM created_at + INTERVAL '1 day');

    IF invalid_status_time <> 0 THEN
        RAISE EXCEPTION
            'invalid status/cancelled_at rows found: %',
            invalid_status_time;
    END IF;

    SELECT count(*)
    INTO invalid_created_at
    FROM public.orders
    WHERE updated_at IS DISTINCT FROM created_at
    OR created_at IS DISTINCT FROM
        CASE 
            WHEN id BETWEEN 1941 AND 2000 THEN  TIMESTAMPTZ '2025-02-01 00:00:00+00'
            ELSE  TIMESTAMPTZ '2025-01-01 00:00:00+00' + (id * INTERVAL '1 second')
        END;
    
    IF invalid_created_at <> 0 THEN
        RAISE EXCEPTION
            'invalid created_at/updated_at rows found: %',
            invalid_created_at;
    END IF;

    SELECT count(*)
    INTO latest_same_time_orders
    FROM public.orders
    WHERE created_at = TIMESTAMPTZ '2025-02-01 00:00:00+00';

    IF latest_same_time_orders <> 60 THEN
        RAISE EXCEPTION
            'latest same-time orders mismatch: expected 60, actual %',
            latest_same_time_orders;
    END IF;

    SELECT count(*)
    INTO invalid_latest_page_orders
    FROM (
        SELECT id, user_id, created_at
        FROM public.orders
        ORDER BY created_at DESC, id DESC
        LIMIT 60
    ) AS latest_orders
    WHERE user_id <> 1
    OR id NOT BETWEEN 1941 AND 2000
    OR created_at IS DISTINCT FROM TIMESTAMPTZ '2025-02-01 00:00:00+00';

    IF invalid_latest_page_orders <> 0 THEN
        RAISE EXCEPTION
            'latest page orders mismatch: invalid rows %',
            invalid_latest_page_orders;
    END IF;

    SELECT count(*)
    INTO orphan_order_users
    FROM public.orders AS orders
    LEFT JOIN public.users AS users
    ON users.id = orders.user_id
    WHERE users.id IS NULL;

    IF orphan_order_users <> 0 THEN
        RAISE EXCEPTION
            'orphan order users found: %',
            orphan_order_users;
    END IF;

    SELECT last_value, is_called
    INTO orders_sequence_value, orders_sequence_called
    FROM public.orders_id_seq;

    IF orders_sequence_value <> 10000 OR NOT orders_sequence_called THEN
        RAISE EXCEPTION
            'orders sequence mismatch: value %, called %',
            orders_sequence_value, orders_sequence_called;
    END IF;

END 
$verify$;

\echo 'PASS: orders'
