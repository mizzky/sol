# 学習記録 2026-07-24

## セッション 1 (11:54)

### 取り組んだタスク
- Phase 2 の `2-5` を完了した。
- 改善済みの `04_carts.sql` をlarge profileで通常投入し、`verify/04_carts.sql` をPASSさせた。
- `06_order_items.sql` をlarge profileで投入し、order_items 3,000,000件を作成して `verify/06_order_items.sql` をPASSさせた。
- `ANALYZE` を実行し、性能検証DBのサイズを記録した。

### ユーザーが質問した内容
- 2-5まで完了しているか。
- Nested Loop改善済みのlarge order_items投入は、50分以内ではなく数分で終わる見込みか。

### 躓いたポイントと解決策
- `04_carts.sql` を通常実行すると、既存のcartsが残っていたため主キー重複で失敗した。
  - 直後のverifyがPASSしたのは、今回のseedが成功したためではなく、前回投入済みデータが残っていたため。
  - `00_guard.sql` を通した上で `cart_items` と `carts` だけを `TRUNCATE ... RESTART IDENTITY` し、改善後のseedを改めて実行した。
- `06_order_items.sql` のlarge投入時間を予測した。
  - mediumの300,000件が11.586秒だったため、単純な10倍で約1分56秒と見積もった。
  - largeの実測は2分17.104秒で、索引規模の増加を含めても数分以内に収まった。

### 結果
- large carts / cart_items:
  - carts 50,000件、cart_items 150,094件。
  - 通常seedの所要時間は12.794秒。
  - `verify/04_carts.sql` はPASS。
- large order_items:
  - 3,000,000件。
  - seedの所要時間は2分17.104秒。
  - `verify/06_order_items.sql` はPASS。
- `ANALYZE` を実行済み。
- `coffeesys_perf` のDBサイズは1,031 MB。
- Phase 2のsmall / medium / large展開を完了した。

### 次回課題
- Phase 3の `3-1: 計測条件テンプレート作成` に進む。
- largeデータセットを使い、既存クエリのbefore計測を開始する。
