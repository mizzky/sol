# 学習記録 2026-08-19

## セッション 1 (09:57)

### 取り組んだタスク

- Phase 4-5の重複インデックス調査を開始した。
- `backend/perf/sql/queries/index_inventory.sql`を作成し、`public`スキーマにあるインデックスについて、定義、UNIQUE属性、制約種別、サイズ、`idx_scan`を一覧化した。
- 修正後のSQLで、30本のインデックスを重複行なしで取得できることを確認した。
- UNIQUE制約由来インデックスと明示的インデックスを比較し、次の完全重複候補4本を抽出した。
  - `idx_users_email`：`users_email_key`で代替可能。サイズは4,880kB。利用クエリは`GetUserByEmail`。
  - `idx_categories_name`：`categories_name_key`で代替可能。サイズは16kB。利用クエリは`ListCategories`の`ORDER BY name`。
  - `idx_products_sku`：`products_sku_key`で代替可能。サイズは30MB。検索用途はなく、UNIQUE検査に使われる。
  - `idx_carts_user_id`：`carts_user_id_key`で代替可能。サイズは1,112kB。利用クエリは`GetCartByUser`など。
- 上記4本を削除できた場合、概算で約36MB削減できることを確認した。
- `idx_cart_items_cart_id`は完全重複ではないが、`cart_items_cart_id_product_id_key (cart_id, product_id)`の左端が`cart_id`であるため、追加検証が必要な削除候補として整理した。サイズは2,432kBで、代替候補の複合インデックスは4,632kBだった。
- VS CodeのDatabase Clientでも、SQLエディタやテーブル配下の`Indexes`からクエリ実行・インデックス確認ができることを確認した。
- `PRIMARY KEY`と`UNIQUE`制約は対応するUNIQUEインデックスを自動作成するが、UNIQUEインデックスが必ず主キー由来とは限らないことを整理した。

### ユーザーが質問した内容

- Phase 4-5の重複インデックス調査をどのように始めるか。
- `index_inventory.sql`の内容が適切か。
- VS CodeのDatabase Clientでクエリ実行やインデックス調査ができるか。
- Database ClientのどこでSQLを実行するか。
- `users`テーブルに設定されているインデックスをどう確認するか。
- UNIQUEインデックスは主キーに対して自動作成されるものか。
- 実行計画の確認前に区切ってコミットしてよいか。

### 躓いたポイントと解決策

- 最初の`index_inventory.sql`では、同じ主キーインデックスが複数行表示された。
  - 原因：`pg_constraint`との結合条件が`constraint_info.conindid = index_rel.oid`だけだったため、対象テーブル自身の主キー・UNIQUE制約だけでなく、そのインデックスを参照する外部キー制約も結合された。`constraint_type=f`の行が重複していた。
  - 解決策：結合条件へ`constraint_info.conrelid = table_rel.oid`を追加し、対象テーブル自身に属する制約だけへ限定した。これにより各インデックスを1行ずつ取得できた。
- `idx_scan=0`を未使用インデックスの削除根拠にしてよいか迷った。
  - 原因：`idx_scan`は統計リセットやインデックス再作成以降の利用回数であり、観測期間外の利用を表さない。実際にPhase 3で使用したインデックスにも0が含まれていた。
  - 解決策：`idx_scan`だけでは削除判断をせず、インデックス定義、制約、サイズ、利用クエリ、削除候補を外した場合の実行計画を組み合わせて判断する方針にした。
- `idx_cart_items_cart_id`は複合UNIQUEインデックスと定義が完全には一致しなかった。
  - 原因：単一列`(cart_id)`と複合列`(cart_id, product_id)`のため、完全重複ではない。
  - 解決策：B-treeの左端一致により複合インデックスが`WHERE cart_id = ...`を処理できる可能性があるため、即時削除せず実行計画で代替可否を確認する候補とした。

### 次回課題

- `backend/perf/sql/queries/index_redundancy_check.sql`を作成する。
- トランザクション内で次の削除候補を一時的に`DROP INDEX`する。
  - `idx_users_email`
  - `idx_categories_name`
  - `idx_products_sku`
  - `idx_carts_user_id`
  - `idx_cart_items_cart_id`
- 削除候補を外した状態で、次のクエリの`EXPLAIN (ANALYZE, BUFFERS)`を確認する。
  - `users.email`による検索
  - `categories.name`による並び替え
  - `carts.user_id`による検索
  - `cart_items.cart_id`による検索
  - 既存のカートJOIN
- 次の代替インデックスへ切り替わるか確認する。
  - `idx_users_email`から`users_email_key`
  - `idx_carts_user_id`から`carts_user_id_key`
  - `idx_cart_items_cart_id`から`cart_items_cart_id_product_id_key`
- 検証後は必ず`ROLLBACK`し、実際のインデックスを保持する。
- 実行計画の結果をもとに削除候補と維持対象を確定し、調査結果を文書化する。
