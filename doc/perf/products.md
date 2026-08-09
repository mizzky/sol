# ListProducts
```sql
-- name: ListProducts :many
SELECT
    id, name, price, is_available, category_id, sku, description, image_url, stock_quantity, created_at, updated_at
FROM products
ORDER BY id;
```

## before計測

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

### small 測定結果
|測定項目|結果|
|---|---|
|計測対象| 商品一覧|
|計測条件| small|
|実行回数| 3回|
|各Execution Time(ms)|2.555/2.744/4.329|
|中央値|2.744|
|Scan type|Index Scan using products_pkey|
|推定rows / 実測rows|10,000/10,000|
|実行Buffers|shared hit=242|

### medium 測定結果
|測定項目|結果|
|---|---|
|計測対象| 商品一覧|
|計測条件| medium|
|実行回数| 3回|
|各Execution Time(ms)|27.167/24.910/25.769|
|中央値|25.769|
|Scan type|Index Scan using products_pkey|
|推定rows / 実測rows|100,000/100,000|
|実行Buffers|shared hit=2403


### large 測定結果
|測定項目|結果|
|---|---|
|計測対象| 商品一覧|
|計測条件| large|
|実行回数| 3回|
|各Execution Time(ms)|337.262/359.847/365.246|
|中央値|359.847|
|Scan type|Index Scan using products_pkey|
|推定rows / 実測rows|1,000,000/1,000,000|
|実行Buffers|shared read=24,012 written=55(*run-1のみ)


- `shared hit`：すでにPostgreSQLの共有バッファ上にある8KBブロックを参照
- `shared read`：共有バッファになかったので、OSキャッシュまたはストレージから読み込んで共有バッファへ載せる|

- `shared hit`ではなく`shared read=24,012`になった理由
  - shared_buffersが128MBであったため、容量が足りず次の実行時に再度読み込んでいる
```bash
# postgreSQLの共有バッファ容量を確認
⬢ [Docker] ❯ psql "$PERF_DATABASE_URL" -X -Atc 'SHOW shared_buffers;'
128MB
```

### 結果
- rows約10倍、Buffers約10倍、largeの時間約13.96倍とおおむねデータ量通りの実行時間、実行内容であった
- largeの場合はデータ量が多くバッファキャッシュに収まらないため速度が少し遅い`shared read`でブロックの読み込みを行っていることが分かった


## keyset pagination計測
### 概要
issue #103 の`3-3. 注文一覧before計測`で、OFFSETによるページネーションでは読み取り位置が深くなるにつれて実行時間が増加することがわかったので、Keyset Paginationの方が読み取り位置が深くても読み取り量と実行時間が増加しにくいことを理解するためのタスク。

`products`を読み取る際に、`WHERE id > :cursor`で読み取り位置を変えて`LIMIT 50`で50件表示するSQLで計測を行う。

### 計測SQL・条件
- large profileを使用する

計測に使うSQLは以下

```sql
EXPLAIN(ANALYZE, BUFFERS, FORMAT TEXT)
SELECT
    id,
    name,
    price,
    is_available,
category_id,
    sku,
    description,
    image_url,
    stock_quantity,
    created_at,
    updated_at
FROM public.products
WHERE id > :cursor
ORDER BY id
LIMIT 50;
```

### cursor 0

### cursor 500,000

### cursor 999,950

### 全件取得との比較

### 結果