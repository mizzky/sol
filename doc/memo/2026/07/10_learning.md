# 学習記録 2026-07-10

## セッション 1 (12:02)

### 取り組んだタスク

- 性能検証用seedの1-7「order_items・注文合計投入」にTDDで着手した。
- `backend/perf/sql/verify/06_order_items.sql` を作成し、以下を検証するRedを完成させた。
  - order_itemsの件数、ID範囲、シーケンス
  - orders/productsへの孤児参照、quantityの範囲
  - 商品価格・商品名のスナップショット、注文内の商品重複
  - 注文ごとの明細数分布（1件: 2,000、2件: 5,000、6件: 3,000）
  - `orders.total` と明細合計の一致
- verifyを実行し、seed未実装のため想定どおり `order items range mismatch` で失敗することを確認した。

### ユーザーが質問した内容

- `GROUP BY order_id, product_id` と `HAVING count(*) > 1` を使った重複検出サブクエリの意味。
- 注文ごとの明細数を検証するための二段階集計の組み立て方。
- `orders.total` と明細合計を比較するCTEの書き方と、`orders.id` / `order_items.order_id` の使い分け。
- verify SQLの実行コマンド。

### 躓いたポイントと解決策

- 明細数の期待値を、注文数ではなく明細行数の合計（6000/15000/9000）として捉えかけた。
  - 1件・2件・6件の明細を持つ注文数として、2000/5000/3000を検証する形に修正した。
- `orders` と `order_items` を使う2種類の集計で、GROUP BYの対象を混同した。
  - ordersを起点にLEFT JOINする明細数分布は `GROUP BY orders.id`、order_items単体で合計を作るCTEは `GROUP BY order_id` と整理した。
- `SUM()` はWHERE句に直接置けない。
  - 内側のCTEで注文ごとの `calculated_total` を集計し、外側で `orders.total IS DISTINCT FROM calculated_total` を比較する形にした。

### 次回課題

- Red用のverifyをコミットする。
- `backend/perf/sql/seed/06_order_items.sql` を作成し、10,000注文に対する1/2/6件の明細分布、商品スナップショット、注文合計更新を実装する（Green）。
- seed実行後に `verify/06_order_items.sql` を実行し、PASSまで確認する。
