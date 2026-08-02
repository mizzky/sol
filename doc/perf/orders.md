# ListOrdersByUser

```sql
SELECT
    id, user_id, total, status, created_at, updated_at
FROM orders
WHERE user_id = :user_id
ORDER BY created_at DESC
LIMIT 50
OFFSET :offset;
```

## before 計測

- `backend/perf/sql/queries/orders_by_user_offset.sql`を使用する
- `:offset`を`0 / 1,000 / 10,000 / 100,000`と切り替えて計測する
- large profileの集中ユーザー`user_id=1`を対象に、`LIMIT 50`で計測する
- それぞれ3回ずつ計測し、中央値を測定する

```bash
base_dir="doc/local/issue-62/results/phase3/orders/large"

for offset in 0 1000 10000 100000
do
  result_dir="$base_dir/offset-$offset"
  mkdir -p "$result_dir"

  # ウォームアップ
  psql "$PERF_DATABASE_URL" \
    -X -q -P pager=off -v ON_ERROR_STOP=1 \
    -v user_id=1 \
    -v offset="$offset" \
    -f backend/perf/sql/queries/orders_by_user_offset.sql \
    >/dev/null

  # 正式計測3回
  for run in 1 2 3
  do
    psql "$PERF_DATABASE_URL" \
      -X -q -P pager=off -v ON_ERROR_STOP=1 \
      -v user_id=1 \
      -v offset="$offset" \
      -f backend/perf/sql/queries/orders_by_user_offset.sql \
      2>&1 | tee "$result_dir/run-$run.txt"
  done
done

# 中央値抽出
base_dir="doc/local/issue-62/results/phase3/orders/large"

for offset in 0 1000 10000 100000
do
  result_dir="$base_dir/offset-$offset"

  median=$(
    awk '/Execution Time:/ {print $3}' \
      "$result_dir"/run-*.txt |
    sort -n |
    sed -n '2p'
  )

  echo "offset=$offset median=${median}ms"
done
```

### offset 0 測定結果
|測定項目|結果|
|---|---|
|計測対象| 注文一覧|
|計測条件| offset 0|
|実行回数| 3回|
|各Execution Time(ms)|0.061/0.074/0.057|
|中央値|0.061|
|Scan type|Index Scan using idx_orders_user_id_created_at on orders|
|推定rows / 実測rows|198,833/50|
|実行Buffers|shared hit=7|

### offset 1,000 測定結果
|測定項目|結果|
|---|---|
|計測対象| 注文一覧|
|計測条件| offset 1,000|
|実行回数| 3回|
|各Execution Time(ms)|0.257/0.276/0.458|
|中央値|0.276|
|Scan type|Index Scan using idx_orders_user_id_created_at on orders|
|推定rows / 実測rows|198,833/1,050|
|実行Buffers|shared hit=24|

### offset 10,000 測定結果
|測定項目|結果|
|---|---|
|計測対象| 注文一覧|
|計測条件| offset 10,000|
|実行回数| 3回|
|各Execution Time(ms)|2.624/3.188/2.401|
|中央値|2.624|
|Scan type|Index Scan using idx_orders_user_id_created_at on orders|
|推定rows / 実測rows|198,833/10,050|
|実行Buffers|shared hit=172|

### offset 100,000 測定結果
|測定項目|結果|
|---|---|
|計測対象| 注文一覧|
|計測条件| offset 100,000|
|実行回数| 3回|
|各Execution Time(ms)|26.092/32.250/37.061|
|中央値|32.250|
|Scan type|Index Scan using idx_orders_user_id_created_at on orders|
|推定rows / 実測rows|198,833/100,050|
|実行Buffers|shared hit=1645|


### 結果
- 想定通り、offsetに比例して実行時間が大きくなった
- OFFSETの増加に伴い、Buffersと実行時間も増加した
- すべて`shared hit`のため、今回の実行時間増加はディスク読み込みではなく、読み飛ばす行数の増加によるものと考えられる