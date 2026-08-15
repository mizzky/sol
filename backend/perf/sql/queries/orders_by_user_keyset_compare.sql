\set ON_ERROR_STOP ON

\ir ../seed/00_guard.sql

\if :{?user_id}
\else
\set user_id 1
\endif

\if :{?offset}
\else
\set offset 0
\endif


\if :{?cursor_created_at}
\else
\set cursor_created_at infinity
\endif

\if :{?cursor_id}
\else
\set cursor_id 9223372036854775807
\endif

WITH offset_page AS (
    SELECT id, created_at
    FROM public.orders
    WHERE user_id = :user_id
    ORDER BY created_at DESC, id DESC
    LIMIT 50 OFFSET :offset
),
keyset_page AS (
    SELECT id, created_at
    FROM public.orders
    WHERE user_id = :user_id
      AND (created_at, id) <
          (:'cursor_created_at'::timestamptz, :cursor_id::bigint)
    ORDER BY created_at DESC, id DESC
    LIMIT 50
)
SELECT 'offset_only' AS difference, count(*)
FROM (
    SELECT * FROM offset_page
    EXCEPT
    SELECT * FROM keyset_page
) AS difference
UNION ALL
SELECT 'keyset_only', count(*)
FROM (
    SELECT * FROM keyset_page
    EXCEPT
    SELECT * FROM offset_page
) AS difference;