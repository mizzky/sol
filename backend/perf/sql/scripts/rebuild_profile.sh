#!/bin/sh
set -eu

usage() {
  echo "Usage: $0 <small|medium|large>" >&2
}

if [ "$#" -ne 1 ]; then
  usage
  exit 1
fi

profile=$1

case "$profile" in
  small|medium|large)
    ;;
  *)
    echo "ERROR: invalid profile: $profile" >&2
    usage
    exit 1
    ;;
esac

: "${PERF_DATABASE_URL:?PERF_DATABASE_URL is required}"

script_dir=$(
  CDPATH= cd -- "$(dirname -- "$0")" && pwd
)
sql_dir=$(
  CDPATH= cd -- "$script_dir/.." && pwd
)
seed_dir="$sql_dir/seed"
verify_dir="$sql_dir/verify"
started_at=$(date +%s)

run_profile_sql() {
  sql_file=$1

  psql "$PERF_DATABASE_URL" \
    -X -q -v ON_ERROR_STOP=1 \
    -v profile="$profile" \
    -f "$sql_file"
}

echo "START: rebuild profile=$profile at $(date '+%F %T')"

psql "$PERF_DATABASE_URL" \
  -X -q -v ON_ERROR_STOP=1 \
  -f "$seed_dir/01_reset.sql"

for name in \
  02_users_categories \
  03_products \
  04_carts \
  05_orders \
  06_order_items
do
  echo "SEED: $name"
  run_profile_sql "$seed_dir/$name.sql"
done

for name in \
  00_profile \
  02_users_categories \
  03_products \
  04_carts \
  05_orders \
  06_order_items
do
  echo "VERIFY: $name"
  run_profile_sql "$verify_dir/$name.sql"
done

echo "ANALYZE"
psql "$PERF_DATABASE_URL" \
  -X -q -v ON_ERROR_STOP=1 \
  -c 'ANALYZE;'

finished_at=$(date +%s)
elapsed_seconds=$((finished_at - started_at))

echo "PASS: rebuilt profile=$profile in ${elapsed_seconds}s"
echo "END: $(date '+%F %T')"
