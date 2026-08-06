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

EXPLAIN(ANALYZE, BUFFERS, FORMAT TEXT)
SELECT
    id, user_id, total, status, created_at, updated_at
FROM orders
WHERE user_id = :user_id
ORDER BY created_at DESC
LIMIT 50
OFFSET :offset;