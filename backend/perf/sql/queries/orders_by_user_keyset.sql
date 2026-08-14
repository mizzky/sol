\set ON_ERROR_STOP ON

\ir ../seed/00_guard.sql

\if :{?user_id}
\else
\set user_id 1
\endif

\if :{?cursor_created_at}
\else
\set cursor_created_at infinity
\endif

\if :{?cursor_id}
\else
\set cursor_id 9223372036854775807
\endif

EXPLAIN(ANALYZE, BUFFERS, FORMAT TEXT)
SELECT
    id, user_id, total, status, created_at, updated_at
FROM orders
WHERE user_id = :user_id
AND (created_at, id) < (:'cursor_created_at'::timestamptz, :cursor_id::bigint)
ORDER BY created_at DESC, id DESC
LIMIT 50;