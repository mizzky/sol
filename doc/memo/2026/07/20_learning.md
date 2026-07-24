# 学習記録 2026-07-20

## セッション 1 (08:56)

### 取り組んだタスク
- Phase 2 の `2-1` を最終確認し、`profile=small` で `01_reset.sql` から `06_order_items.sql` まで投入した。
- `verify/00_profile.sql` と `verify/02` から `06` を実行し、small の全検証が PASS することを確認した。これにより `2-1: プロファイル切り替え` を完了とした。
- `2-2: medium投入検証` として、`profile=medium` で seed `01` から `06` をクリーン投入した。
- medium の verify `00_profile`, `02` から `06` を実行し、全件 PASS を確認した。
- `ANALYZE;` を実行し、`coffeesys_perf` のDBサイズを取得した。

### ユーザーが質問した内容
- `2-2` で何を確認すればよいか。
- seed と verify を実行するコマンド。

### 躓いたポイントと解決策
- Phase 2 の `2-2` は新規実装を作る工程ではなく、mediumデータセットの受入検証と計測の工程だった。
  - smallの回帰確認、mediumの全投入、整合性検証、`ANALYZE`、DBサイズ計測の順に実施した。
- `06_order_items.sql` の投入時間は個別に `time` で記録した。
  - mediumの全投入は開始 `09:11:48`、終了 `09:13:08` で約1分20秒。
  - `06_order_items.sql` 単体は `18.200s`。

### 結果
- mediumの期待規模である users 10,000、products 100,000、carts 5,000、orders 100,000、order_items 300,000 のデータセットを投入できた。
- 全verifyがPASSした。
- `ANALYZE` 実行後の `coffeesys_perf` のDBサイズは `111 MB`。

### 次回課題
- `2-3` として、largeプロファイルでまず users / products の投入時間と容量を確認する。
- mediumの測定結果を基に、large投入の時間・ディスク使用量を観測しながら段階的に進める。
