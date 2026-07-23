# 学習記録 2026-07-23

## セッション 1 (09:10)

### 取り組んだタスク
- Phase 2 の `2-5` に着手し、large profileで未投入だった carts / cart_items の投入を進めた。
- `verify/04_carts.sql` を先に実行してRedを確認した。
- large用の既存 `04_carts.sql` が遅い原因を `EXPLAIN` で調査した。
- `product_candidates` の候補順位と候補総数の役割を分離し、`Merge Join` を選べるCTE構造へ書き換えた。
- `BEGIN ... ROLLBACK` 内の `EXPLAIN ANALYZE` で、実データを残さず cart_items INSERTの実測を取得した。

### ユーザーが質問した内容
- candidatesを使う `04_carts.sql` でもNested Loopが起きるか。
- `cart_item_plan` の推定264万行は `item_ordinal` が大きいためか。
- CTEとTEMP TABLE + インデックスのどちらを選ぶべきか。
- 実行計画から元SQLの問題箇所へどう戻るか。
- `EXPLAIN`、`EXPLAIN ANALYZE`、verifyの役割分担。
- Merge Joinとインデックスの関係。

### 躓いたポイントと解決策
- 旧 `04_carts.sql` は、候補商品約855,000件とcart item計画を `candidate_ordinal` で結合していた。
  - `candidate_ordinal` は `row_number()` で作った値で、結合に使える索引がなかった。
  - `EXPLAIN` で `Nested Loop + Join Filter` を確認し、実測では約150,094件のcart itemに対して候補全体を繰り返し比較し得る構造だと分かった。
- `rows=2640100` は `item_ordinal` の値ではなかった。
  - temp tableを880行、可変上限の `generate_series` を1000行と見積もる統計誤差によるcart item計画の行数推定だった。
  - 実測のcart item計画は150,094行だった。
- `candidate_count` を候補行ごとに持たせていた。
  - `product_candidates` は「候補順位から商品を引く表」、`candidate_config` は「候補総数を持つ1行」に役割を分けた。
  - CTEを一度だけ使う今回の用途では、TEMP TABLE + インデックスより先に CTE + Merge Join を試す方針とした。

### 結果
- 旧計画では `Nested Loop + Join Filter` が出ていた。
- 改善後は `Merge Join + Merge Cond` になった。
- `EXPLAIN ANALYZE` の実測:
  - 候補商品: 855,000行。
  - Merge Join出力: 150,094行。
  - Merge Joinは約4.1秒で出力を完了。
  - INSERT全体の `Execution Time`: 21.312秒。
  - 全体時間の主な部分は、cart_itemsの外部キーtrigger検査だった。
- `EXPLAIN` は計画・仮説の確認、`EXPLAIN ANALYZE` は実測、verifyはデータ整合性の確認に使い分けると整理できた。

### 次回課題
- 改善後の `04_carts.sql` をlarge profileで通常投入し、`verify/04_carts.sql` をPASSさせる。
- `06_order_items.sql` のlarge投入3,000,000件とverifyを行う。
- `ANALYZE`、投入時間、DBサイズを記録して2-5を完了する。
