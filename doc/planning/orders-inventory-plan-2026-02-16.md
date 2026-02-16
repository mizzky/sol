# 取引・在庫処理 — チケット詳細（作成日: 2026-02-16）

目的
- 注文作成時の在庫整合性を担保し、キャンセル・決済連携を考慮した堅牢なトランザクション処理を実装する。

受け入れ基準
- `orders` と `order_items` のマイグレーションが存在する。  
- CreateOrder API が DB トランザクション内で在庫を正しくデクリメント/ロールバックする。  
- 同時実行でオーバーソールドが発生しないことを統合テストで確認。

チケット一覧

1) DBマイグレーション: `orders`, `order_items`, `payments` — Effort: High, Priority: P0
   - 内容: 注文ヘッダ（id,user_id,total, status, created_at）と明細（order_item: order_id, product_id, qty, unit_price）を定義。  
   - 影響ファイル: backend/db/migrations, backend/query.sql
   - 受け入れ条件: マイグレーションファイルとリバート手順があること。

2) sqlc クエリ: GetProductForUpdate / UpdateProductStock / CreateOrder* — Effort: High, Priority: P0
   - 内容: 在庫更新を安全に行うための `SELECT ... FOR UPDATE` を含むクエリと、order 作成系クエリを追加。  
   - 影響ファイル: backend/query.sql, backend/db/querier.go
   - 受け入れ条件: sqlc 再生成手順とサンプルTx呼び出しコードがあること。

3) トランザクションハンドラ: `CreateOrderHandler` — Effort: High, Priority: P0
   - 内容: カート or リクエストを受け取り、DB Tx 内で 1) 各 product を `FOR UPDATE` で取得、2) 在庫チェック、3) 在庫デクリメント、4) order+items 作成、の流れを実装。  
   - 影響ファイル: backend/handler/order.go (新規), backend/routes/routes.go
   - 受け入れ条件: 単体テストと統合テスト（Tx 成功/失敗ケース）があること。

4) キャンセル / 在庫巻き戻しロジック — Effort: Med, Priority: P1
   - 内容: `CancelOrderHandler` を実装し、ステータス遷移と在庫の戻しをTxで行う。  
   - 影響ファイル: backend/handler/order.go, backend/db/querier.go
   - 受け入れ条件: キャンセル後に在庫が元に戻ることを統合テストで確認。

5) 決済抽象化インタフェース — Effort: Med, Priority: P2
   - 内容: `pkg/payment` に `PaymentProvider` インタフェースを用意、モック実装でテスト可能にする。  
   - 影響ファイル: backend/pkg/payment (新規), backend/handler/order.go
   - 受け入れ条件: モックで決済成功/失敗の振る舞いを切替えてテスト可能。

6) 冪等性（idempotency-key）サポート — Effort: Med, Priority: P1
   - 内容: 同一注文再送を匿う idempotency-key ヘッダを受け取り、重複注文を防止する。  
   - 影響ファイル: backend/handler/order.go, backend/db/migrations
   - 受け入れ条件: 同一 key の2回目リクエストで既存注文を返すテストがあること。

7) 同時性テスト（E2E） — Effort: High, Priority: P0
   - 内容: 並列で複数の CreateOrder リクエストをぶつけ、在庫が負にならないことを確認する統合テストを作成。  
   - 影響ファイル: backend/tests, backend/db/migrations
   - 受け入れ条件: CI で実行可能なテストスクリプトがあること。

8) エラーハンドリングとリトライ設計 — Effort: Low, Priority: P1
   - 内容: DB トランザクションの失敗時のエラーマッピング（409/500）とクライアントへの分かりやすいレスポンス設計。  
   - 影響ファイル: backend/pkg/respond, backend/handler/order.go
   - 受け入れ条件: エラー種別に応じたHTTPステータスのユニットテスト。

9) メトリクス/監視案 — Effort: Low, Priority: P2
   - 内容: 注文失敗数、在庫不足回数、注文レイテンシ等のメトリクス設計（Prometheus などを想定）。  
   - 影響ファイル: doc/planning, backend/logging 設計（新規）

実施順の推奨
1→2→3→7→4→6→5→8→9
