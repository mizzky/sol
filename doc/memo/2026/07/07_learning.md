# 学習記録 2026-07-07

## セッション 1 (12:09)

### 取り組んだタスク

- Issue 62 Phase 1 の `1-5. carts・cart_items投入` をTDDのRed-Green-Refactorの流れで進めた。
- `backend/perf/sql/verify/04_carts.sql` を作成し、seed未投入状態で `carts range mismatch: expected count/min/max 500/1/500, actual 0/<NULL>/<NULL>` になるRedを確認した。
- `backend/perf/sql/seed/04_carts.sql` を段階的に作成した。
  - `carts` 500件を `generate_series(1, 500)` で生成した。
  - `cart_items` の設計表 `cart_item_plan` を作り、cart 2を100明細、cart 3〜500を各3明細にした。
  - 商品候補を `is_available = true AND stock_quantity > 0` に限定し、`row_number()` で `candidate_ordinal`、`count(*) OVER ()` で `candidate_count` を付与した。
  - `cart_id` と `item_ordinal` から候補商品を決定的に割り当て、同一cart内のproduct重複を0件にした。
  - `cart_items.price` を `products.price` と一致させ、`quantity` は1〜3で循環させた。
  - `carts_id_seq` と `cart_items_id_seq` を `setval` で同期した。
- `verify/04_carts.sql` で `PASS: carts and cart_items` を確認した。
- reset → users/categories → products → carts/cart_items → verify 02/03/04 の再実行性確認を2周行い、どちらもPASSした。
- 1-5完了後、コミットも完了した。

### ユーザーが質問した内容

- `generate_series(3, 500)` はintの3,4,5...を入れるのと何が違うのか。
- `CROSS JOIN` は何をしているのか。
- `CROSS JOIN LATERAL` は通常の `CROSS JOIN` と何が違うのか。
- `UNION ALL` と `CROSS JOIN LATERAL` は、読みやすさ・保守性・パフォーマンスの観点でどちらが良いのか。
- largeでデータ量が100倍になった場合、`UNION ALL` と `CROSS JOIN LATERAL` のパフォーマンス差が問題になるのか。
- `% products.candidate_count` はどういう意味か。
- `candidate_ordinal` は8550までしかないのに、なぜ剰余が必要なのか。
- `* 100` でcartごとの商品候補開始位置をずらす意図は何か。
- cart 86以降で候補商品リストが循環すると、商品重複が起きるのではないか。
- 商品を50,000件用意して `cart_id + 001〜099` のように割り当てる方が読み解きやすいのではないか。

### 躓いたポイントと解決策

- `CROSS JOIN LATERAL` が難しく感じた。
  - 解決策: 今回はシナリオが「cart 2だけ100明細、cart 3〜500は3明細」と明確なので、`UNION ALL` で分ける方針にした。`LATERAL` は「右側の表生成で左側の行を参照できるJOIN」と整理した。
- `generate_series` の理解が曖昧だった。
  - 解決策: `generate_series(3, 5)` は単なる値リストではなく、3,4,5が1行ずつ入った仮テーブルを作るものとして理解した。
- `CROSS JOIN` の結果イメージが曖昧だった。
  - 解決策: 左5行と右5行を `CROSS JOIN` すると `1-1, 1-2, ..., 5-5` の25パターンができる「全組み合わせ」として理解した。
- `UNION ALL` と `CROSS JOIN LATERAL` の選択に迷った。
  - 解決策: 今回はseed SQLであり、パフォーマンス差よりも仕様の読みやすさを優先した。固定シナリオの表現は `UNION ALL`、行ごとに複雑なルールで件数を変える場合は `CROSS JOIN LATERAL` が向くと整理した。
- `candidate_count` に対する剰余の必要性が分かりにくかった。
  - 解決策: `candidate_ordinal` は1〜8550だが、`((cart_id - 1) * 100 + item_ordinal - 1)` は8550を超えるため、`% candidate_count + 1` で1〜8550の範囲に戻すと理解した。
- `* 100` の意味が曖昧だった。
  - 解決策: 最大100明細のcartがあるため、cartごとに100個幅の候補商品枠をずらすルールだと整理した。完全な視認性のための設計ではなく、限られた8550候補の中で規則性・視認性・cart内重複回避を両立する工夫と位置づけた。
- SQL実装中に構文エラーが出た。
  - `SELECT`句で `max(candidate_count) AS max_candidate_count` の後にカンマがなく、次の `count(*)` で syntax error になった。列区切りのカンマを追加して解決した。
  - `TIMESTAMPTZ '2025-01-01 00:00::00+00'` のように時刻リテラルに `::` が混ざり、timestamp変換エラーになった。`2025-01-01 00:00:00+00` に直して解決した。
  - `ON ... + 1; ORDER BY ...` のようにJOIN条件の後でセミコロンを置いてしまい、`ORDER BY` がSQLの外に出た。セミコロンを `ORDER BY` の後へ移動して解決した。

### 次回課題

- 次は `1-6. orders投入` に進む。
- 1-6では、orders 10,000件、order_items 30,000件、集中ユーザー20%、status 90% pending / 10% cancelled、最新60件の同一時刻境界、注文合計と明細合計の整合性を段階的に確認する。
- 1-5と同様に、まず verify SQL でRedを作り、その後 seed SQL を小さいSELECT確認からGreenへ進める。
