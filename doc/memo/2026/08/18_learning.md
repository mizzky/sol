# 学習記録 2026-08-18

## セッション 1 (10:03)

### 取り組んだタスク

- Phase 4-1として、商品一覧のkeyset paginationをlarge profile（100万件）で計測した。
  - `WHERE id > :cursor ORDER BY id LIMIT 50`を、cursor `0 / 500,000 / 999,950`で比較した。
  - カーソル位置に応じてIndex Scanの推定rowsは`1,000,000 / 501,699 / 50`と変化したが、実測rowsはすべて50、Buffersはすべて`shared hit=8`だった。
  - Execution Time中央値は`0.200 / 0.096 / 0.188 ms`で、読み取り位置が深くなっても実際の読み取り量が増えないことを確認した。
- Phase 4-2として、注文一覧のOFFSET paginationとkeyset paginationを同じ開始位置で比較した。
  - `(created_at, id)`を複合カーソルにして、OFFSET `50 / 1,000 / 10,000 / 100,000`相当の取得結果を比較した。
  - `offset_only=0`、`keyset_only=0`となり、同じ`created_at`を持つ注文がページ境界をまたいでも、同じ50件を重複・欠落なく取得できることを確認した。
  - keyset版のExecution Time中央値は、OFFSET `0 / 1,000 / 10,000 / 100,000`相当で`0.249 / 0.137 / 0.167 / 0.133 ms`だった。
  - keyset版はIndex Scan実測rowsが`61 / 51 / 51 / 51`、Buffersが`17 / 17 / 18 / 17`でほぼ一定だった。一方、OFFSET 100,000では実測100,050行、Buffers 1,645、中央値`32.250 ms`だった。
- Phase 4-3として、同じ`created_at`を持つ注文が`id DESC`で返ることを要求する統合テストを追加し、Redを確認した。
  - 期待値が新しいID順、実際値が作成順となり、元の`ORDER BY created_at DESC`では同時刻内の順序が保証されないことを再現した。
  - `ListOrdersByUser`を`ORDER BY created_at DESC, id DESC`へ変更し、sqlcを再生成してGreenにした。
- Phase 4-4として、注文一覧の並び順に合わせて複合インデックスを変更した。
  - `（user_id, created_at DESC）`から`（user_id, created_at DESC, id DESC）`へ変更するmigration 13のup/downを作成した。
  - 性能検証DBで`up → down → up`を実行し、バージョンとインデックス定義が期待どおり切り替わることを確認した。
  - 改善後の実行計画では`Incremental Sort`が消え、複合カーソル条件全体が`Index Cond`になった。Index Scan実測rowsは61から50、実行Buffersは17から7へ減少した。
  - 対象統合テスト、バックエンド全UT、全ITがすべてPASSした。

### ユーザーが質問した内容

- 商品keyset計測と注文keyset計測は同じことをしているように見えるが、観点にどのような違いがあるか。
- 注文keyset計測の目的を、OFFSET版との同値性確認と性能比較としてどう表現すればよいか。
- `created_at`だけで注文を並べた場合、同じ作成日時の注文で何が問題になるか。
- `id DESC`がない場合、出力順は一意ではないのか。
- この問題はクエリ設計時点のミス、またはkeyset pagination導入時に露見した考慮漏れなのか。
- 完全に同じ日時を持つ注文は通常運用でどの程度発生するのか。
- Green後は複合インデックスを変更するのか。
- migration変更後、対象ITだけでなくUTや全ITも実行すべきか。

### 躓いたポイントと解決策

- `created_at`だけのソートでは、同じ時刻の行同士の順序がSQL上で定義されない。
  - 一意な主キー`id`をタイブレーカーに加え、`ORDER BY created_at DESC, id DESC`として決定的な順序を定義した。
- `created_at`だけをkeysetカーソルにすると、同一時刻の注文がページ境界をまたいだ際に残りの注文が欠落する可能性がある。
  - `(created_at, id)`の複合カーソルを使用し、OFFSET版との差分が0であることを複数の深さと同一時刻境界で検証した。
- 既存インデックス`(user_id, created_at DESC)`では、クエリが要求する`id DESC`まで満たせず`Incremental Sort`が発生した。
  - インデックスを`(user_id, created_at DESC, id DESC)`へ拡張し、Sortの消滅と複合条件のIndex Cond化を確認した。
- 同一時刻は通常の個別リクエストでは頻発しないように見える。
  - PostgreSQLの`now()`は同一トランザクション内で同じ値になり、一括INSERT、バッチ、データ移行、時刻の丸め、seedなどでも発生するため、頻度が低くてもページネーションでは考慮すべき境界条件と整理した。
- migration変更後の確認範囲を対象ITだけにするか迷った。
  - 性能検証DBでのup/down/up、対象IT、全UT、全ITを分けて実行し、既存DBでの可逆性と新規DB上の回帰の両方を確認した。

### 次回課題

- migration 13のup/downをコミットする。
- Phase 4-5として、注文・カート周辺の重複インデックスを調査し、削除または維持の根拠を整理する。
- Phase 4の正式なafter計測をウォームアップ後に3回ずつ実施し、before/afterの実行計画、rows、Buffers、Execution Time中央値をドキュメントへ記録する。
- 複合インデックス変更後のkeyset paginationについて、各カーソル位置の正式なafter値を取得する。
