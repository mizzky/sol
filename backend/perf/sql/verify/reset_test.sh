#!/bin/sh
set -eu

reset_sql="backend/perf/sql/seed/01_reset.sql"
sentinel="__perf_reset_test__"

: "${PERF_DATABASE_URL:?PERF_DATABASE_URL is required}"

perf_db=$(
  psql "$PERF_DATABASE_URL" \
    -X -Atq -v ON_ERROR_STOP=1 \
    -c 'SELECT current_database();'
)

if [ "$perf_db" != "coffeesys_perf" ]; then
  echo "FAIL: unexpected database: $perf_db"
  exit 1
fi

cleanup() {
  psql "$PERF_DATABASE_URL" \
    -X -q -v ON_ERROR_STOP=1 \
    -c "DELETE FROM categories WHERE name = '$sentinel';" \
    >/dev/null 2>&1 || true
}

trap cleanup EXIT
cleanup

migration_before=$(
  psql "$PERF_DATABASE_URL" \
    -X -Atq -v ON_ERROR_STOP=1 \
    -c "SELECT version || ':' || dirty FROM schema_migrations;"
)

psql "$PERF_DATABASE_URL" \
  -X -q -v ON_ERROR_STOP=1 \
  -c "INSERT INTO categories (name, description)
      VALUES ('$sentinel', 'reset test');"

psql "$PERF_DATABASE_URL" \
  -X -q -v ON_ERROR_STOP=1 \
  -f "$reset_sql"

remaining=$(
  psql "$PERF_DATABASE_URL" \
    -X -Atq -v ON_ERROR_STOP=1 \
    -c "SELECT count(*) FROM categories WHERE name = '$sentinel';"
)

migration_after=$(
  psql "$PERF_DATABASE_URL" \
    -X -Atq -v ON_ERROR_STOP=1 \
    -c "SELECT version || ':' || dirty FROM schema_migrations;"
)

if [ "$remaining" -ne 0 ]; then
  echo "FAIL: business data was not reset"
  exit 1
fi

if [ "$migration_before" != "$migration_after" ]; then
  echo "FAIL: schema_migrations was changed"
  exit 1
fi

echo "PASS: performance database reset"