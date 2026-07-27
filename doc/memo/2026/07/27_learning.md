# 学習記録 2026-07-27

## セッション 1 (11:38)

### 取り組んだタスク

- Phase 3の商品一覧before計測に使用する`backend/perf/sql/queries/products_list.sql`を段階的に作成した。
- 性能DBガードを読み込んだ後、現行の`ListProducts`を`EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)`で計測できるようにした。
- `psql`から計測SQLを実行し、`products_pkey`を使った`Index Scan`、推定・実測10,000行、実行時の`shared hit=242`を確認した。
- `tee`を使い、smallのraw実行計画3回分を`doc/local/issue-62/results/phase3/products/small/`へ保存した。
- 3回のExecution Time（2.744 / 2.555 / 4.329 ms）を抽出し、中央値2.744 msを算出した。
- Pages向けの学習記事`measurements-products.md`と、計測結果を管理する`products.md`の責務を分けた。
- `feat(perf): add products list before measurement`として計測SQLとsmallの計測結果をコミットした。

### ユーザーが質問した内容

- `products_list.sql`へ何を記述し、profile切り替えをどこで担当させるか。
- `psql`の`-X`、`-q`、`-P pager=off`、`-v ON_ERROR_STOP=1`、`-f`が何をするか。
- raw出力をファイルへ保存する処理を、すぐスクリプト化する必要があるか。
- `2>&1`、パイプ、`tee`、`set -o pipefail`がどのように出力を扱うか。
- 同じprofileで3回計測するとき、毎回`rebuild_profile.sh`を実行する必要があるか。
- 3回分のExecution Timeから中央値を算出する方法。
- `rg`が使えない環境でExecution Timeを抽出する方法。
- 計測結果とPages向け学習記事を同じMarkdownへまとめるべきか。

### 躓いたポイントと解決策

- 完成済みの計測コマンドを実行するだけでは理解が薄くなる懸念があった。安全ガード、通常の`EXPLAIN`、`ANALYZE`・`BUFFERS`追加、raw保存の順に一段ずつ進めた。
- `rebuild_profile.sh`と計測反復の役割が混ざりかけた。rebuildはprofileごとに1回だけ行い、ウォームアップ後の3回は同じデータセットとSQLで実行すると整理した。
- raw出力の保存方法が不明だった。`tee`でターミナル表示とファイル保存を同時に行い、`doc/local/`配下へGit管理外のまま保存した。
- Execution Timeの中央値算出に`rg`を使えなかった。`awk`で数値を抽出し、`sort -n`と`sed -n '2p'`で3件の中央を取得した。
- `Buffers: shared hit=242`と`Planning Buffers: shared hit=145`を一つの値として記録しかけた。実行時とPlanning時のBuffersを分離し、今回の要約には実行時の`shared hit=242`だけを記録した。
- 計測ごとの「観察結果」「まだ分からないこと」が過剰に感じられた。各profileでは事実だけを記録し、small・medium・largeが揃った後に比較考察を書く方針へ簡略化した。

### 次回課題

- `doc/perf/products.md`の実行Buffersを`shared hit=242`へ直した未コミット差分をコミットする。
- `rebuild_profile.sh medium`でmediumデータセットを準備する。
- 同じ`products_list.sql`をウォームアップ後に3回実行し、raw出力、Execution Time中央値、Scan type、rows、実行Buffersを記録する。
- smallと同じ手順でlargeを計測した後、3 profile間の時間・Buffers・実行計画の変化を比較する。
