\set ON_ERROR_STOP ON

\ir ../seed/00_guard.sql

SELECT
    table_rel.relname AS table_name,
    index_rel.relname AS index_name,
    pg_get_indexdef(index_rel.oid) AS index_definition,
    index_info.indisunique AS is_unique,
    constraint_info.contype AS constraint_type,
    pg_size_pretty(pg_relation_size(index_rel.oid)) AS index_size,
    index_stats.idx_scan
FROM pg_class AS table_rel
JOIN pg_index AS index_info
    ON index_info.indrelid = table_rel.oid
JOIN pg_class AS index_rel
    ON index_rel.oid = index_info.indexrelid
JOIN pg_namespace AS namespace
    ON namespace.oid = table_rel.relnamespace
LEFT JOIN pg_constraint AS constraint_info
    ON constraint_info.conindid = index_rel.oid
    AND constraint_info.conrelid = table_rel.oid
LEFT JOIN pg_stat_user_indexes AS index_stats
    ON index_stats.indexrelid = index_rel.oid
WHERE namespace.nspname = 'public'
ORDER BY table_rel.relname, index_rel.relname;