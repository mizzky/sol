\set ON_ERROR_STOP on

BEGIN;

\ir 00_guard.sql
\ir 00_profile.sql


-- INSERT
WITH order_source AS (
    SELECT
        n AS id,
        CASE 
            WHEN n BETWEEN 1 AND config.orders_count * 20 / 100 THEN 1
            ELSE  2 + ((n - (config.orders_count * 20 / 100) - 1) % (config.users_count - 1))
        END AS user_id,
        CASE 
            WHEN n % 10 =0 THEN  'cancelled'
            ELSE 'pending'
        END AS status,
        CASE 
        WHEN n BETWEEN (config.orders_count * 20 / 100) - 59 AND config.orders_count * 20 / 100 THEN  TIMESTAMPTZ '2025-02-01 00:00:00+00'
        ELSE  TIMESTAMPTZ '2025-01-01 00:00:00+00' + (n * INTERVAL '1 second')
        END AS created_at
    FROM pg_temp.perf_profile AS config
    CROSS JOIN generate_series(1, config.orders_count) AS series(n)
)
INSERT INTO public.orders (
    id,
    user_id,
    status,
    total,
    created_at,
    updated_at,
    cancelled_at
)
SELECT
    id,
    user_id,
    status,
    0 AS total,
    created_at,
    created_at AS updated_at,
    CASE 
        WHEN status = 'pending' THEN NULL 
        ELSE  created_at + INTERVAL '1 day'
    END AS cancelled_at 
FROM order_source
ORDER BY id;

-- sync sequence
SELECT setval(
    pg_get_serial_sequence('public.orders', 'id'),
    (SELECT max(id) FROM public.orders),
    true
);

COMMIT;
