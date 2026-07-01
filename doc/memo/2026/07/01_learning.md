# 学習記録 2026-07-01

## セッション 1 (12:56)

### 取り組んだタスク

- 大量データ投入では単に件数を増やすのではなく、検証対象クエリが劣化するデータ分布を再現する方針を整理した。
  - PostgreSQLの `generate_series` と `INSERT ... SELECT` を使って一括投入する。
  - small（users 1,000、products 10,000、orders 10,000、order_items 30,000）から始め、medium、largeへ段階的に拡張する。
  - 注文が集中する検証用ユーザー、日時の重複、statusの偏りを設け、深いOFFSET、keyset pagination、scan種別の変化を検証できるようにする。
  - 外部キー順に投入し、連番と剰余で再現可能なデータを生成する。専用DB確認、再実行可能な初期化、件数・整合性検査、投入後の `ANALYZE` を組み込む計画とした。
- 通常の開発DBと分離した性能検証用DB `coffeesys_perf` を既存PostgreSQL内に作成し、`PERF_DATABASE_URL` で接続できることを確認した。
- `coffeesys_perf` にmigration v1〜v12を適用し、versionが `12` であることを確認した。
- `schema_migrations` を含む10テーブルと、主要5テーブルの既存インデックス17件を確認した。現在のインデックスは変更せず、性能計測のbefore基準として使うことにした。
  - 重複候補として `products_sku_key` と `idx_products_sku`、`carts_user_id_key` と `idx_carts_user_id`、`cart_items_cart_id_product_id_key` と `idx_cart_items_cart_id` を記録した。
- アプリの `main.go` は従来どおり `DATABASE_URL` のみを読み、起動時に `PERF_DATABASE_URL` の値を `DATABASE_URL` として渡す切替方式を整理した。
- VS Code tasksに `Go: Air (Backend / Perf DB)` と `Launch Fullstack (Perf DB)` を追加し、バックエンドタスクの環境変数を `"DATABASE_URL": "${env:PERF_DATABASE_URL}"` とする構成を確認した。通常の `Launch Fullstack (Both)` はデフォルトのまま維持した。

### ユーザーが質問した内容

- 100万件規模のデータを安全かつ検証目的に合う形で投入する計画はどうするか。
- 性能検証にはTestcontainersではなく専用DBを用意すべきか。
- 開発DBと性能検証DBを簡単に切り替えられるか、`main.go` の修正が必要か。
- `PERF_DATABASE_URL` の設定範囲と永続化方法、シェルやVS Code Taskへの引き継がれ方はどうなるか。
- VS Code tasksに性能検証用タスクを追加した場合、`Ctrl + Shift + B` でどちらが起動するか、提示した設定内容で問題ないか。
- Perf用Taskを起動した際の `DATABASE_URL is not set` エラーの原因は何か。

### 躓いたポイントと解決策

- `PERF_DATABASE_URL` を一時的に `export` しただけでは、そのターミナル内と子プロセスにしか反映されず、VS Code Taskには引き継がれないことを整理した。永続利用する値はルートの `.env` に正確な変数名で追加し、Dev ContainerをRebuildしてVS Code側へ読み込ませる。
- Perf用Taskで `DATABASE_URL is not set` となった。原因は `.env` に `PERF_DATABASE_URL` が未設定で、Task内の `${env:PERF_DATABASE_URL}` が空文字へ展開されたため。DB作成やmigrationの失敗ではない。
- `DATABASE_URL="$PERF_DATABASE_URL"` の有効範囲をコンテナ全体と捉えかけたが、実際はコマンド単位または現在のシェル単位であることを確認した。通常起動とPerf起動をVS Code Taskで分けることで、接続先を明示する。
- 通常版とPerf版のバックエンドはどちらも8080番ポートを使うため同時起動できない。切替前に動作中のバックエンドタスクを停止する。

### 次回課題

- ルートの `.env` に `PERF_DATABASE_URL=.../coffeesys_perf?sslmode=disable` を追加する。
- Dev ContainerをRebuildし、新しいターミナルで `PERF_DATABASE_URL` が設定されていることを値を表示せず確認する。
- `psql "$PERF_DATABASE_URL" -Atc "SELECT current_database();"` の結果が `coffeesys_perf` になることを確認する。
- `Launch Fullstack (Perf DB)` を起動し、バックエンドが性能検証DBへ接続できることを確認する。
- small規模の投入仕様と、期待件数・外部キー孤児・一意性・数量・注文合計・検証用ユーザーの注文数を確認する検証SQLを先に設計する。
- small投入を通した後、medium／largeへ件数を変数化して拡張し、投入時間・件数・DBサイズを記録する。
