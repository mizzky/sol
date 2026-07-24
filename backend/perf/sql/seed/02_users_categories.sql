\set ON_ERROR_STOP on

BEGIN;

\ir 00_guard.sql
\ir 00_profile.sql

INSERT INTO public.users (
    id,
    name,
    email,
    password_hash,
    role,
    status,
    reset_token,
    created_at,
    updated_at
)
SELECT
    n,
    format('user-%s', lpad(n::text, 6, '0')),
    format('user-%s@example.test', lpad(n::text, 6, '0')),
    'perf-not-a-real-password-hash',
    'member',
    'active',
    NULL,
    TIMESTAMPTZ '2025-01-01 00:00:00+00',
    TIMESTAMPTZ '2025-01-01 00:00:00+00'
FROM pg_temp.perf_profile AS config
CROSS JOIN generate_series(1, config.users_count) AS series(n)
ORDER BY n;

INSERT INTO public.categories (
    id,
    name,
    description,
    created_at,
    updated_at
)
SELECT
    n,
    format('category-%s', lpad(n::text, 2, '0')),
    format('Performance category %s', lpad(n::text, 2, '0')),
    TIMESTAMPTZ '2025-01-01 00:00:00+00',
    TIMESTAMPTZ '2025-01-01 00:00:00+00'
FROM generate_series(1, 10) AS series(n)
ORDER BY n;

SELECT setval(
    pg_get_serial_sequence('public.users', 'id'),
    (SELECT max(id) FROM public.users),
    true
);

SELECT setval(
    pg_get_serial_sequence('public.categories', 'id'),
    (SELECT max(id) FROM public.categories),
    true
);

COMMIT;
