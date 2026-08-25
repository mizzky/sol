# 学習記録 2026-08-24

## セッション 1 (21:56)

### 取り組んだタスク

- Phase 4-7を開始し、正式な作業範囲が商品・注文・注文明細のafter実行計画取得とbefore/after比較であることを確認した。
- 注文keyset paginationについて、複合インデックス変更後のafter計測を開始位置0、OFFSET 1,000相当、10,000相当、100,000相当の4条件で実施した。
- 各条件で3回計測し、Execution Timeの中央値、実測rows、Buffers、実行計画を変更前と比較した。
- 複合インデックスを`(user_id, created_at DESC)`から`(user_id, created_at DESC, id DESC)`へ変更した効果を確認した。
  - `Incremental Sort`が発生しなくなった。
  - `(created_at, id)`のカーソル条件全体が`Index Cond`に入った。
  - OFFSET 1,000相当以降で発生していたFilterと`Rows Removed by Filter: 1`がなくなった。
  - すべての条件で必要な50件だけを読み取った。
- afterの中央値とBuffersは以下の結果になった。

|開始位置|変更前中央値(ms)|変更後中央値(ms)|変更前実測rows|変更後実測rows|変更前Buffers|変更後Buffers|
|---:|---:|---:|---:|---:|---:|---:|
|0|0.249|0.070|61|50|17|7|
|1,000|0.137|0.086|51|50|17|9|
|10,000|0.167|0.079|51|50|18|8|
|100,000|0.133|0.072|51|50|17|8|

- 4-4で行った注文クエリ用インデックス改善について、目的と具体的な変更内容の記録が漏れていたことを確認し、`doc/perf/orders.md`へ補完した。
- `orders.md`に4条件のafter測定結果、変更前後の比較表、`Incremental Sort`とFilterがなくなった理由を記録した。
- `ca285df docs(perf): record order index improvement results`としてコミットした。

### ユーザーが質問した内容

- Phase 4-6の次は4-7か、4-7をどのような順番で進めればよいか。
- 商品keysetの前に`orders.md`を更新してよいか。
- 実行時間が短縮した理由は、`idx_orders_user_id_created_at`を`id`まで含む複合インデックスへ変更したためか。
- インデックス改良の内容を既にperfドキュメントへ記録していたか。
- 4-4時点で目的と変更内容を記録すべきだったものが漏れており、注文の改善なので`orders.md`へ記録する認識でよいか。
- 4-4の説明とbefore/after結果をどのようにまとめればよいか。
- `git diff --check`が指摘した末尾の空白とは何か。

### 躓いたポイントと解決策

- 当初、4-7をPhase 4全体のまとめと捉えたが、計画を確認すると商品・注文・注文明細のafter実行計画取得が正式なタスクだった。注文keysetの再計測から始め、その後に商品、注文明細、before/after整理へ進む順番に修正した。
- 変更前のインデックスは`id DESC`を含まないため、`ORDER BY created_at DESC, id DESC`を完全には満たせず、`Incremental Sort`と一部Filterが必要だった。`id DESC`を含む複合インデックスにより、検索条件と並び順をインデックスで満たせるようになったことを実行計画で確認した。
- Execution Timeが0.1ms前後と短く測定値がぶれやすかった。時間だけで判断せず、`Incremental Sort`とFilterの消滅、実測rows、Buffers、`Index Cond`も合わせて改善を評価した。
- 4-4のインデックス改善は`doc/perf`内に具体的な目的・変更内容が残っていなかった。重複インデックス調査用の`indexes.md`ではなく、注文クエリの改善を扱う`orders.md`へ変更前後のインデックス構成と狙いを補完した。
- before結果のFilterに関する説明は、先頭ページではなくOFFSET 1,000相当以降に該当するため、対象条件を限定した表現へ修正した。
- Markdownの空行に半角スペースが残り`git diff --check`で検出された。動作には影響せず、今回は内容を優先してコミットした。

### 次回課題

- 商品keyset paginationのafter実行計画をカーソル位置ごとに計測する。
- 注文明細クエリのafter実行計画を計測する。
- Phase 3のbefore結果とPhase 4のafter結果を整理し、改善内容を比較する。
- 必要に応じて、4-6で実施した重複インデックス削除後の容量差、migrationのup/down、UT・IT結果を`doc/perf/indexes.md`へ記録する。
