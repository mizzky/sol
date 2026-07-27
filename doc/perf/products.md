# ListProducts
```sql
-- name: ListProducts :many
SELECT
    id, name, price, is_available, category_id, sku, description, image_url, stock_quantity, created_at, updated_at
FROM products
ORDER BY id;
```

## small before計測

- `backend/perf/sql/queries/products_list.sql`を使用する
- 3回計測して中央値を測定する

```bash
#出力ファイル名を変えて3回コマンド実行する
psql "$PERF_DATABASE_URL"   -X   -q   -P pager=off   -v ON_ERROR_STOP=1   -f backend/perf/sql/queries/products_list.sql   2>&1 | tee "$result_dir/run-*.txt"

awk '/Execution Time:/ {print $3}' \
  "$result_dir"/run-*.txt |
sort -n |
sed -n '2p'
2.744
```

### 測定結果
|測定項目|結果|
|---|---|
|計測対象| 商品一覧|
|計測条件| small|
|実行回数| 3回|
|各Execution Time|2.555/2.744/4.329|
|中央値|2.744|
|Scan type|Index Scan using products_pkey|
|推定rows / 実測rows|10000/10000|
|実行Buffers|242