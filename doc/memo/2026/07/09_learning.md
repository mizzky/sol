# 学習記録 2026-07-09

## セッション 1 (11:58)

### 取り組んだタスク
- 1-5 `carts/cart_items` の verify を実行し、`users/categories/products/carts/cart_items` が期待件数・期待分布で投入済みであることを確認した。
- `tdd` スキルを用いて、1-6 `orders` 投入を Red-Green の流れで進めた。
- Red として `backend/perf/sql/verify/05_orders.sql` を作成し、`orders range mismatch` で失敗することを確認した。
- Green として `backend/perf/sql/seed/05_orders.sql` を段階的に実装し、`PASS: orders` まで到達した。
- `orders_id_seq` の同期漏れによる verify 失敗を修正し、最終的に seed 01〜05 と verify 05 の実行が通った。

### ユーザーが質問した内容
- 1-6 に進む前に verify から始めたい。
- `tdd-mentor` / `tdd` を用いて 1-6 の実装を進めたい。
- Red 用の `verify/05_orders.sql` はこれで良いか。
- seed 実装では、最初から完成コードではなく段階的なヒントで自分で考えたい。
- `updated_at` / `cancelled_at` はどう設定すればよいか。
- 同じ `SELECT` 内で作った `status` や `created_at` をそのまま参照できるのか。
- `orders_id_seq` の verify が `value 1, called f` で落ちた理由は何か。

### 躓いたポイントと解決策
- `verify/05_orders.sql` の初回作成時に、`:` と `;` の typo、変数名 `orphan_order_users` 周辺の typo、`$verify$;` のセミコロン漏れがあった。
  - 構文エラーを潰した後、期待どおり `orders range mismatch` で失敗し、Red 成功と判断した。
- `seed/05_orders.sql` の実装中、`INSERT` の指定カラム数に対して `SELECT` 側の列数が足りない状態になった。
  - `id, user_id, status, total, created_at, updated_at, cancelled_at` の7列を同じ順番で返す必要があると整理した。
- `order_source` CTE 内で `n AS id` とした直後に、同じ `SELECT` 内で `id` を参照しようとして混乱した。
  - 同じ `SELECT` 内では alias の `id` ではなく元の `n` を使い、外側の `SELECT` では `id` を使うと整理した。
- `updated_at = created_at` と書いてしまい、代入のつもりが比較式のような形になった。
  - `SELECT` では代入ではなく値を返すため、`created_at AS updated_at` と書くと整理した。
- `cancelled_at` を作るために、同じ `SELECT` 内の `status` と `created_at` alias を参照しようとして詰まりかけた。
  - `order_source` CTE で `status` と `created_at` を先に作り、外側の `SELECT` で `cancelled_at` を作る構成にした。
- `orders_id_seq` が `value 1, called f` のままで verify が失敗した。
  - 明示的に `id` を投入した場合は sequence が自動で進まないため、`setval(pg_get_serial_sequence('public.orders', 'id'), (SELECT MAX(id) FROM public.orders), true)` で同期する必要があると整理した。

### 次回課題
- 1-7 `order_items・注文合計投入` に進む。
  - 明細数 1/2/6 の分布を 20% / 50% / 30% にする。
  - 同一注文内で `product_id` を重複させない。
  - `order_items` の `quantity`, `unit_price`, `product_name_snapshot` を決定的に作る。
  - `orders.total` を注文明細の `quantity * unit_price` 合計と一致させる。
