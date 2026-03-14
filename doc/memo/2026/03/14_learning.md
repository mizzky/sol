# 学習記録 2026-03-14

## 今日取り組んだこと
- Testcontainers を使った統合テスト基盤のセットアップを実施
- `//go:build integration` タグ付きで `order_integration_test.go` を作成
- `TestMain` で PostgreSQL コンテナ起動・マイグレーション適用・クリーンアップの流れを実装
- `TestIntegration_DBReady` を実行して、`test_db` への接続確認を実施

## 実行した内容と結果
- 実行コマンド: `go test -tags=integration ./tests -run TestIntegration_DBReady -v`
- 結果: PASS
- 確認できたこと:
  - Docker へ接続できている
  - `postgres:17-trixie` コンテナが起動する
  - テスト終了後にコンテナが停止・破棄される

## 理解が深まったポイント
- `TestMain` は自分で呼び出す関数ではなく、`go test` が自動的に検出して実行する
- `m.Run()` のタイミングで各 `TestXxx` が実行される
- Testcontainers は「イメージを毎回ビルド」するのではなく、
  既存イメージを再利用して毎回新しいコンテナを作成する
- ログに出る2つのコンテナの役割:
  - `testcontainers/ryuk`: 後片付け用
  - `postgres:17-trixie`: テスト対象DB用

## 疑問・詰まった点と解決
- 疑問: `gopls` で `No packages found ...` が出た
  - 解決: build tag による想定挙動。必要なら `gopls.buildFlags` に `-tags=integration` を追加
- 疑問: 統合テストを常時実行したくない
  - 解決: build tag 方式で UT と統合テストを分離

## 次にやること
1. `TestCreateOrderHandler_HappyPath`（正常系1ケース）を追加する
2. まず Red を確認する
3. `CreateOrderHandler` の最小実装で Green 化する

---

## 追加で進めた内容（午後）

### 実施したこと
- `TestCreateOrderHandler_HappyPath` の Red を仕様ベースへ肉付け
  - ルーター経由で `POST /api/orders` を実行
  - `201 Created` と DB状態（orders/order_items/stock/cart_items）を検証
- `CreateOrderHandler` を実装し、トランザクション制御（BeginTx/Commit/Rollback）を追加
- `CreateOrderItem` クエリを修正し `product_name_snapshot` を保存するように変更
- `sqlc generate` を再実行して生成コードを更新

### 失敗の原因分析と修正
- 症状: 統合テストで HTTP 500 が返る
- 主因:
  - `order_items.product_name_snapshot` が `NOT NULL` なのに INSERT していなかった
  - 在庫更新クエリ `stock_quantity = stock_quantity + $2` に対して、減算用の値を渡せていなかった
- 修正内容:
  - `CreateOrderItem` に `product_name_snapshot` を追加
  - `UpdateProductStock` 呼び出しで `-item.Quantity` を渡すように変更
  - 統合テスト側 SQL の typo (`JOIN order` -> `JOIN orders`) を修正
  - `Scan` エラーを都度 `err` で受けるように修正

### テスト結果
- 実行コマンド: `go test -tags=integration ./tests -run TestCreateOrderHandler_HappyPath -v`
- 結果: PASS（Green）
- 実行コマンド: `go test -tags=integration ./tests -v`
- 結果: PASS

### 学び
- Red は「未実装だから落ちる」ではなく「仕様アサートで落ちる」形にすると次の実装が明確になる
- Testcontainers での統合テストは、DB制約起因の不整合を早期に検出できる
