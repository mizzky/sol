# ListCartItemsByUser

```sql
SELECT
    ci.id,
    ci.cart_id,
    ci.product_id,
    ci.quantity,
    ci.price,
    ci.created_at,
    ci.updated_at,
    p.name AS product_name,
    p.price AS product_price,
    p.stock_quantity AS product_stock
FROM public.cart_items AS ci
JOIN public.carts AS c
    ON ci.cart_id = c.id
JOIN public.products AS p
    ON p.id = ci.product_id
WHERE c.user_id = :user_id
ORDER BY ci.id;
```

## before 計測
- `backend/perf/sql/queries/cart_items_by_user.sql`を使用する
- 3回計測して中央値を測定する
- large profileを対象とする
- `user_id=1 / 3 / 2`を使い、カート明細数`0 / 3 / 100`の場合を比較する

### user_id 1 測定結果
|測定項目|結果|
|---|---|
|計測対象| カート明細一覧|
|計測条件| user_id 1|
|実行回数| 3回|
|各Execution Time(ms)|0.087/0.117/0.089|
|中央値|0.089|
|carts Scan|Index Scan using idx_carts_user_id on carts c|
|cart_items Scan|Index Scan using idx_cart_items_cart_id on cart_items ci|
|products Scan|Index Scan using products_pkey on products p（never executed）|
|Join type|Nested Loop|
|Sort|quicksort  Memory: 25kB|
|推定rows / 実測rows|3/0|
|実行Buffers|shared hit=11|

### user_id 2 測定結果
|測定項目|結果|
|---|---|
|計測対象| カート明細一覧|
|計測条件| user_id 2|
|実行回数| 3回|
|各Execution Time(ms)|0.364/0.326/0.408|
|中央値|0.364|
|carts Scan|Index Scan using idx_carts_user_id on carts c|
|cart_items Scan|Index Scan using idx_cart_items_cart_id on cart_items ci|
|products Scan|Index Scan using products_pkey on products p|
|Join type|Nested Loop|
|Sort|quicksort  Memory: 34kB|
|推定rows / 実測rows|3/100|
|実行Buffers|shared hit=413|

### user_id 3 測定結果
|測定項目|結果|
|---|---|
|計測対象| カート明細一覧|
|計測条件| user_id 3|
|実行回数| 3回|
|各Execution Time(ms)|0.249/0.198/0.185|
|中央値|0.198|
|carts Scan|Index Scan using idx_carts_user_id on carts c|
|cart_items Scan|Index Scan using idx_cart_items_cart_id on cart_items ci|
|products Scan|Index Scan using products_pkey on products p|
|Join type|Nested Loop|
|Sort|quicksort  Memory: 25kB|
|推定rows / 実測rows|3/3|
|実行Buffers|shared hit=24|


### 結果
- すべてのケースでcarts、cart_items、productsにIndex Scanが選択された
- JOINにはNested Loopが使われ、productsへのアクセス回数はカート明細数と一致した
- 明細数の増加に伴ってBuffersと実行時間が増加したが、100明細でも中央値は0.364msだった
- 推定rowsは一律3件であり、100明細のカートでは推定3件に対して実測100件となった
- `ORDER BY ci.id`によるSortは発生したが、最大34kBのquicksortで負荷は小さかった