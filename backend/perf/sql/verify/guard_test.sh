#!/bin/sh
set -eu

guard_sql="backend/perf/sql/seed/00_guard.sql"

: "${DATABASE_URL:?DATABASE_URL is required}"
: "${PERF_DATABASE_URL:?PERF_DATABASE_URL is required}"

dev_db=$(
  psql "$DATABASE_URL" \
    -X -Atq -v ON_ERROR_STOP=1 \
    -c 'SELECT current_database();'
)

perf_db=$(
  psql "$PERF_DATABASE_URL" \
    -X -Atq -v ON_ERROR_STOP=1 \
    -c 'SELECT current_database();'
)

if [ "$dev_db" = "coffeesys_perf" ]; then
  echo "FAIL: DATABASE_URL must not point to coffeesys_perf"
  exit 1
fi

if [ "$perf_db" != "coffeesys_perf" ]; then
  echo "FAIL: PERF_DATABASE_URL must point to coffeesys_perf"
  exit 1
fi

if psql "$DATABASE_URL" \
  -X -q -v ON_ERROR_STOP=1 \
  -f "$guard_sql"
then
  echo "FAIL: guard accepted unexpected database: $dev_db"
  exit 1
fi

psql "$PERF_DATABASE_URL" \
  -X -q -v ON_ERROR_STOP=1 \
  -f "$guard_sql"

echo "PASS: database guard"