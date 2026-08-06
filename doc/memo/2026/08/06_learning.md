# 学習記録 2026-08-06

## セッション 1 (10:06)

### 取り組んだタスク

- Phase 3の3-3「注文一覧before計測」を実施した。
  - large profileの集中ユーザー`user_id=1`を対象に、`LIMIT 50`、OFFSET `0 / 1,000 / 10,000 / 100,000`を各3回計測した。
  - 全ケースで`idx_orders_user_id_created_at`によるIndex Scanが使われ、Sortは発生しなかった。
  - OFFSETごとのIndex Scan実測rowsは`50 / 1,050 / 10,050 / 100,050`、Buffersは`7 / 24 / 172 / 1,645`、Execution Time中央値は`0.061 / 0.276 / 2.624 / 32.250 ms`だった。
  - 返却件数は50件のままでも、`OFFSET + LIMIT`件を読み取って先頭を捨てるため、深いOFFSETほど処理量が増えることを確認した。
- Phase 3の3-4「注文明細before計測」を実施した。
  - large profileの300万明細から、1・2・6明細を持つ`order_id=1 / 3 / 8`を計測した。
  - すべて`idx_order_items_order_id`によるIndex Scanと、`ORDER BY id`による25kBのquicksortになった。
  - Execution Time中央値は`0.135 / 0.077 / 0.130 ms`、Buffersはすべて`shared hit=10`で、最大6明細では有意な性能差が見られなかった。
  - 単発の`ListOrderItemsByOrderID`は高速であり、注文一覧から繰り返し呼ばれるN+1の影響は別Issueで扱うことにした。
- Phase 3の3-5「カートbefore計測」を実施した。
  - large profileで、0・3・100明細を持つ`user_id=1 / 3 / 2`を計測した。
  - `carts`は`idx_carts_user_id`、`cart_items`は`idx_cart_items_cart_id`、`products`は`products_pkey`のIndex Scanになった。
  - JOINはNested Loopで、productsへのloopsは明細数に応じて`never executed / 3 / 100`になった。
  - Execution Time中央値は`0.089 / 0.198 / 0.364 ms`、Buffersは`11 / 24 / 413`だった。
  - 100明細でも推定rowsは平均的な3件のままで、実測100件とのずれが発生することを確認した。
- 計測用SQLを`backend/perf/sql/queries/`、結果を`doc/perf/`、raw出力を`doc/local/issue-62/results/phase3/`へ整理した。
- Phase 3のbefore計測をPR #104としてまとめ、mainへマージした。
- Phase 4とN+1改善の順序を検討し、DB単体のbefore/after比較を完結させるためPhase 4を先に進める方針にした。

### ユーザーが質問した内容

- 注文一覧のOFFSET計測後に、次に何を進めるべきか。
- `orders_by_user_offset.sql`をbefore計測用としてそのまま使ってよいか。
- psqlの`\g`にはどのような意味があり、3回計測では何回実行すればよいか。
- `ListOrderItemsByOrderID`は何を確認する計測なのか、1・2・6明細で結果がほぼ変わらないと判断してよいか。
- JOINクエリのScan typeはproductsだけを記録すればよいか、Sortも記録すべきか。
- カートJOINでproducts 100万件と100明細がどのように比較されるのか。
- Nested Loopが使われた場合、以前のseed投入で問題になったNested Loopと同じ問題なのか。
- Phase 3完了時点でPRに分けてよいか。
- Phase 4と別IssueのN+1改善のどちらを先に進めるべきか。

### 躓いたポイントと解決策

- OFFSETが深くても同じIndex Scanなので高速だと考えかけた。実測rowsが`OFFSET + LIMIT`まで増え、返却前にOFFSET分を捨てるため、インデックスがあっても仕事量は増えると整理した。
- 初回の注文明細計測が`3.221 ms`、再実行が`0.033 ms`となり差が大きかった。初回は`shared read=1`、再実行はすべて`shared hit`であり、ウォームアップと共有バッファの影響を確認した。
- 1・2・6明細のExecution Timeを単純比較しようとしたが、すべて`0.2 ms`未満で順序も明細数に比例しなかった。少数行では処理量の差より計測ノイズが大きく、単発クエリは現状ボトルネックではないと判断した。
- JOINクエリのScan typeをproductsだけで表そうとした。3-5の完了条件に合わせ、carts、cart_items、productsそれぞれのScan typeとJoin typeを分けて記録した。
- products 100万件とカート100明細から1億件比較が発生すると予想した。実際には`products_pkey`を使った1件検索が`loops=100`であり、100万件の全件走査ではないことを実行計画から確認した。
- Nested Loop自体を性能問題と捉えかけた。今回のように外側が最大100件で内側がIndex Scanなら適切であり、巨大な候補同士を`Join Filter`で比較していたseedのNested Loopとは性質が異なると整理した。
- カート100明細だけ推定3件・実測100件になった。Plannerは多くのカートに共通する平均3明細を推定し、`user_id=2`から特殊な100明細カートへ到達するテーブル間の偏りまでは把握できないと理解した。

### 次回課題

- Phase 4の新しい作業ブランチを用意する。
- 4-1「商品keyset比較」に着手する。
- `backend/perf/sql/queries/products_keyset.sql`を作成し、`WHERE id > :cursor ORDER BY id LIMIT 50`を検証する。
- large profileでcursor `0 / 500,000 / 999,950`を計測し、全件取得のbefore結果とrows、Buffers、Execution Timeを比較する。
- keysetではcursorが深くても主キーインデックスから直接50件を取得できるか確認する。
- Phase 4完了後、別IssueでN+1改善に着手する。
