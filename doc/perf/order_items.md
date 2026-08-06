# ListOrderItemsByOrderID

```sql
SELECT
    id, order_id, product_id, quantity, unit_price, product_name_snapshot, created_at, updated_at
FROM
    public.order_items
WHERE
    order_id = :order_id
ORDER BY id;
```

## before 計測
- `backend/perf/sql/queries/order_items_by_order.sql`を使用する
- large profileの`order_items`3,000,000件を対象とする
- `order_id=1 / 3 / 8`を使い、明細数`1 / 2 / 6`の場合を比較する
- ウォームアップ後、それぞれ3回計測して中央値を測定する
  

### order_id 1 測定結果
|測定項目|結果|
|---|---|
|計測対象| 注文明細一覧|
|計測条件| order_id 1|
|実行回数| 3回|
|各Execution Time(ms)|0.155/0.073/0.135|
|中央値|0.135|
|Scan type|Index Scan using idx_order_items_order_id on order_items |
|推定rows / 実測rows|5/1|
|実行Buffers|shared hit=10|

### order_id 3 測定結果
|測定項目|結果|
|---|---|
|計測対象| 注文明細一覧|
|計測条件| order_id 3|
|実行回数| 3回|
|各Execution Time(ms)|0.077/0.078/0.069|
|中央値|0.077|
|Scan type|Index Scan using idx_order_items_order_id on order_items |
|推定rows / 実測rows|5/2|
|実行Buffers|shared hit=10|

### order_id 8 測定結果
|測定項目|結果|
|---|---|
|計測対象| 注文明細一覧|
|計測条件| order_id 8|
|実行回数| 3回|
|各Execution Time(ms)|0.130/0.184/0.101|
|中央値|0.130|
|Scan type|Index Scan using idx_order_items_order_id on order_items |
|推定rows / 実測rows|5/6|
|実行Buffers|shared hit=10|


### 結果
- 1・2・6明細のすべてで`idx_order_items_order_id`が使用された
- `ORDER BY id`によるSortは発生したが、quicksortの使用メモリはすべて`25kB`だった
- 実行Buffersはすべて`shared hit=10`で、明細数による実行時間の明確な差は見られなかった
- 単発の明細取得は高速だが、注文一覧では注文ごとに実行されるため、N+1による影響はPhase 5で別途検証する　