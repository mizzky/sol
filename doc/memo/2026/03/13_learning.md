# 学習記録 2026-03-13

## 今日取り組んだこと

- `feat/create-order-handler` ブランチを作成
- 注文作成ハンドラの設計議論と Testcontainers 環境セットアップ（手順1〜2完了）

---

## 設計の整理

### テスト責務の切り分け

| 層 | 対象 | テスト手法 |
|---|---|---|
| `createOrderLogic` | 業務ロジック（在庫確認、カート操作、明細作成） | MockDB を使ったユニットテスト |
| `CreateOrderHandler` | Tx 制御、HTTP ステータス変換、userID 取得 | Testcontainers を使った統合テスト |

- モックでは BeginTx / Commit / Rollback の挙動を検証できないため、トランザクション境界は実 DB で確認する必要がある
- `createOrderLogic` はすでにモック UTがある（TestCreateOrderLogic）

### CreateOrderHandler の責務
1. context から userID を取得
2. `conn.BeginTx` でトランザクション開始
3. `queries.WithTx(tx)` でトランザクション内クエリを作成
4. `createOrderLogic` を呼び出す
5. エラー時 → Rollback + HTTP ステータス変換（400 / 404 / 409 / 500）
6. 成功時 → Commit + 201 返却

### エラー変換の方針
| エラー文言 | HTTP ステータス |
|---|---|
| "カートが空です" | 400 |
| "商品が見つかりません" | 404 |
| "在庫不足です" | 409 |
| その他 | 500 |

---

## 疑問・詰まった点

### Q1. Testcontainers を使うには Docker が使えないといけない。dev container から Docker は使える？
- `docker ps` が exit 127（コマンド見つからず）だった
- 解決: `devcontainer.json` の `features` に `ghcr.io/devcontainers/features/docker-outside-of-docker:1` を追加して Dev Container を Rebuild することで解消

### Q2. Testcontainers を使うために Dockerfile の RUN に go install を追加する必要がある？
- **不要**。`go install` で Dockerfile に書くのは CLI バイナリ（`air`・`sqlc`・`migrate` のようなコマンド）
- testcontainers-go はコードから `import` して使うライブラリなので、`go get` で `go.mod` に追記するだけでよい

### Q3. トランザクションをモックでテストできないのはなぜ？
- `db.Querier` インターフェースには `BeginTx` が含まれていない
- `Queries.WithTx(tx)` は `*sql.DB` が持つ `BeginTx` を前提にしている
- MockDB は `db.Querier` を満たすモックなので、Tx 開始自体をモックに任せられない
- そのため Tx 境界の検証は実 DB を使った統合テストで行う

---

## 手順2で追加した依存パッケージ

```bash
go get github.com/testcontainers/testcontainers-go
go get github.com/testcontainers/testcontainers-go/modules/postgres
go get github.com/testcontainers/testcontainers-go/wait
go get github.com/golang-migrate/migrate/v4
go get github.com/golang-migrate/migrate/v4/database/postgres
go get github.com/golang-migrate/migrate/v4/source/file
go mod tidy
```

---

## 次回の課題

1. Testcontainers を使った TestMain 基盤を作る（コンテナ起動・マイグレーション適用）
2. 正常系 1 ケースの統合テストを書く（Red）
3. `CreateOrderHandler` の最小実装を追加してテストを通す（Green）
4. `createOrderLogic` 内の在庫更新ロジックの修正検討
   - 現状は `item.ProductStock`（カート取得時のスナップショット）を使っている
   - 本来は `GetProductForUpdate` 後の `product.StockQuantity` を使うべき
