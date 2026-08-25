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


## keyset pagination 計測

### 概要
上記の`3-3. 注文一覧before計測`で計測の応用として、読み取り位置を固定したOFFSET計測と、同位置をカーソル指定したkeyset paginationによる計測の比較を行う。

`backend/perf/sql/queries/orders_by_user_keyset_compare.sql`でOFFSET計測と同値性確認を行い、同じ50件を取得していることを確認した上でkeyset pagination計測を行うことで、OFFSETとkeyset paginationの比較が出来る。

### 取得結果の同値性確認

|比較位置|offset_only|keyset_only|
|---:|---:|---:|
|OFFSET 50（同一時刻のページ境界）|0|0|
|OFFSET 1,000|0|0|
|OFFSET 10,000|0|0|
|OFFSET 100,000|0|0|

### 計測SQL・条件
- large profileを使用する

計測に使うSQLは以下

```sql
EXPLAIN(ANALYZE, BUFFERS, FORMAT TEXT)
SELECT
    id, user_id, total, status, created_at, updated_at
FROM orders
WHERE user_id = :user_id
AND (created_at, id) < (:'cursor_created_at'::timestamptz, :cursor_id::bigint)
ORDER BY created_at DESC, id DESC
LIMIT 50;
```



### OFFSET 0相当 測定結果

|測定項目|結果|
|---|---|
|計測対象|注文一覧|
|計測条件|先頭ページ|
|実行回数|3回|
|各Execution Time(ms)|0.155/0.249/0.266|
|中央値|0.249|
|Scan type|Index Scan using idx_orders_user_id_created_at on orders|
|追加処理|Incremental Sort|
|Index Scan 推定rows / 実測rows|198,813/61|
|Limit 推定rows / 実測rows|50/50|
|実行Buffers|shared hit=17|
|Sort Method|quicksort|
|Sort Memory|29kB|

### OFFSET 1,000相当 測定結果

|測定項目|結果|
|---|---|
|計測対象|注文一覧|
|計測条件|OFFSET 1,000と同じ開始位置|
|カーソル|`2025-01-03 07:16:41+00` / `199001`|
|実行回数|3回|
|各Execution Time(ms)|0.137/0.127/0.144|
|中央値|0.137|
|Scan type|Index Scan using idx_orders_user_id_created_at on orders|
|追加処理|Incremental Sort|
|Index Scan 推定rows / 実測rows|39,354/51|
|Rows Removed by Filter|1|
|Limit 推定rows / 実測rows|50/50|
|実行Buffers|shared hit=17|
|Sort Method|quicksort|
|Sort Memory|27kB|

### OFFSET 10,000相当 測定結果

|測定項目|結果|
|---|---|
|計測対象|注文一覧|
|計測条件|OFFSET 10,000と同じ開始位置|
|カーソル|`2025-01-03 04:46:41+00` / `190001`|
|実行回数|3回|
|各Execution Time(ms)|0.167/0.127/0.264|
|中央値|0.167|
|Scan type|Index Scan using idx_orders_user_id_created_at on orders|
|追加処理|Incremental Sort|
|Index Scan 推定rows / 実測rows|37,492/51|
|Rows Removed by Filter|1|
|Limit 推定rows / 実測rows|50/50|
|実行Buffers|shared hit=18|
|Sort Method|quicksort|
|Sort Memory|27kB|

### OFFSET 100,000相当 測定結果

|測定項目|結果|
|---|---|
|計測対象|注文一覧|
|計測条件|OFFSET 100,000と同じ開始位置|
|カーソル|`2025-01-02 03:46:41+00` / `100001`|
|実行回数|3回|
|各Execution Time(ms)|0.133/0.168/0.131|
|中央値|0.133|
|Scan type|Index Scan using idx_orders_user_id_created_at on orders|
|追加処理|Incremental Sort|
|Index Scan 推定rows / 実測rows|19,690/51|
|Rows Removed by Filter|1|
|Limit 推定rows / 実測rows|50/50|
|実行Buffers|shared hit=17|
|Sort Method|quicksort|
|Sort Memory|27kB|

### OFFSET paginationとの比較

|開始位置|OFFSET中央値(ms)|keyset中央値(ms)|OFFSET実測rows|keyset実測rows|OFFSET Buffers|keyset Buffers|
|---:|---:|---:|---:|---:|---:|---:|
|0|0.061|0.249|50|61|7|17|
|1,000|0.276|0.137|1,050|51|24|17|
|10,000|2.624|0.167|10,050|51|172|18|
|100,000|32.250|0.133|100,050|51|1,645|17|

### 結果

OFFSET paginationでは、開始位置が深くなるほど読み飛ばす行数が増加し、実測rows、Buffers、Execution Timeも増加した。

keyset paginationでは、開始位置が深くなってもIndex Scanの実測rowsは約51件、Buffersは`shared hit=17〜18`、Execution Timeの中央値は`0.133〜0.249 ms`であり、大きな変化はなかった。

OFFSET 100,000では100,050件を読み取って50件を返したのに対し、keyset paginationではカーソル位置から51件を読み取り、カーソル行を除外して50件を返した。

現在のインデックスは`(user_id, created_at DESC)`であり、`ORDER BY created_at DESC, id DESC`のうち`id DESC`を満たしていないため、すべての条件で`Incremental Sort`が発生した。

以上から、keyset paginationは読み取り位置が深くなっても読み取り量が増加しにくいことを確認した。


## 複合インデックス変更後のafter計測

keyset paginationでは`ORDER BY created_at DESC, id DESC`で並び順を指定している。

変更前のインデックスは`(user_id, created_at DESC)`であり、`id DESC`を満たしていないため、`Incremental Sort`が発生していた。

そこでクエリの検索条件と並び順をインデックスで満たすことで`Incremental Sort`を不要とすることを期待して、phase4-4にてインデックスを修正した。

クエリの検索条件と並び順に合わせて、インデックスを以下のように変更した。

|変更前|変更後|
|---|---|
|`(user_id, created_at DESC)`|`(user_id, created_at DESC, id DESC)`|

インデックス改善の効果測定としてbeforeにあたるフェーズ4.2で計測したkeyset paginationでのorders計測と、afterにあたる複合インデックスを改良後のkeyset paginationのorders計測を比較する


### OFFSET 0相当 after測定結果

|測定項目|結果|
|---|---|
|計測対象|注文一覧|
|計測条件|先頭ページ|
|実行回数|3回|
|各Execution Time(ms)|0.068 / 0.135 / 0.070|
|中央値|0.070|
|Scan type|Index Scan using idx_orders_user_id_created_at on orders|
|Index Scan 推定rows / 実測rows|198,813 / 50|
|Limit 推定rows / 実測rows|50 / 50|
|実行Buffers|shared hit=7|
|追加処理|なし|

### OFFSET 1,000相当 after測定結果

|測定項目|結果|
|---|---|
|計測対象|注文一覧|
|計測条件|OFFSET 1,000と同じ開始位置|
|カーソル|`2025-01-03 07:16:41+00` / `199001`|
|実行回数|3回|
|各Execution Time(ms)|0.076 / 0.144 / 0.086|
|中央値|0.086|
|Scan type|Index Scan using idx_orders_user_id_created_at on orders|
|Index Scan 推定rows / 実測rows|39,354 / 50|
|Limit 推定rows / 実測rows|50 / 50|
|実行Buffers|shared hit=9|
|追加処理|なし|

### OFFSET 10,000相当 after測定結果

|測定項目|結果|
|---|---|
|計測対象|注文一覧|
|計測条件|OFFSET 10,000と同じ開始位置|
|カーソル|`2025-01-03 04:46:41+00` / `190001`|
|実行回数|3回|
|各Execution Time(ms)|0.075 / 0.079 / 0.079|
|中央値|0.079|
|Scan type|Index Scan using idx_orders_user_id_created_at on orders|
|Index Scan 推定rows / 実測rows|37,492 / 50|
|Limit 推定rows / 実測rows|50 / 50|
|実行Buffers|shared hit=8|
|追加処理|なし|

### OFFSET 100,000相当 after測定結果

|測定項目|結果|
|---|---|
|計測対象|注文一覧|
|計測条件|OFFSET 100,000と同じ開始位置|
|カーソル|`2025-01-02 03:46:41+00` / `100001`|
|実行回数|3回|
|各Execution Time(ms)|0.136 / 0.070 / 0.072|
|中央値|0.072|
|Scan type|Index Scan using idx_orders_user_id_created_at on orders|
|Index Scan 推定rows / 実測rows|19,690 / 50|
|Limit 推定rows / 実測rows|50 / 50|
|実行Buffers|shared hit=8|
|追加処理|なし|

### 複合インデックス変更前後の比較

|開始位置|変更前中央値(ms)|変更後中央値(ms)|変更前実測rows|変更後実測rows|変更前Buffers|変更後Buffers|
|---:|---:|---:|---:|---:|---:|---:|
|0|0.249|0.070|61|50|17|7|
|1,000|0.137|0.086|51|50|17|9|
|10,000|0.167|0.079|51|50|18|8|
|100,000|0.133|0.072|51|50|17|8|

### 結果

変更前のインデックスは`(user_id, created_at DESC)`であり、`ORDER BY created_at DESC, id DESC`のうち`id DESC`を満たしていなかったため、`Incremental Sort`が発生していた。

変更後はインデックスを`(user_id, created_at DESC, id DESC)`としたことで、クエリの検索条件と並び順をインデックスで満たせるようになり、`Incremental Sort`が発生しなくなった。

また、変更前のOFFSET 1,000以降では、カーソル境界の行を含む51件を読み取り、Filterで1件を除外していた。変更後は`(created_at, id)`のカーソル条件全体が`Index Cond`に入り、すべての条件で50件だけを読み取った。

実行Buffersは変更前の`shared hit=17〜18`から、変更後は`shared hit=7〜9`に減少した。Execution Timeの中央値もすべての条件で短縮され、変更後は`0.070〜0.086 ms`だった。

実行時間は非常に短く測定値がぶれやすいものの、実行計画から追加SortとFilterがなくなり、実測rowsとBuffersも減少したことを確認できた。