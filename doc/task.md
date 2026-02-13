## 商品管理CRUD（TODO）

### 目的
- 商品の作成・更新・削除・取得（一覧/詳細）APIをTDDで実装し、API仕様と整合すること。

### 前提・方針
- 対象範囲: 商品CRUD（POST/PUT/DELETE/GET一覧/GET詳細）
- 認可: `POST /api/products`, `PUT /api/products/:id`, `DELETE /api/products/:id` は管理者のみ。`GET` は認証不要。
- エラー応答は `doc/api.md` の仕様に合わせる。
- 実装はTDD: まずテストを作成してからプロダクトコードを実装する。

### TODO（優先度順）
1. [x] API仕様の確認・補完
   - `doc/api.md` に商品エンドポイントの仕様（リクエスト/レスポンス/バリデーション/ステータスコード）を追加・整備。
2. [x] DBクエリの追加（sqlc用）
   - `query.sql` に `CreateProduct`, `GetProduct`, `ListProducts`, `UpdateProduct`, `DeleteProduct` を定義。
   - `sqlc generate` の実行手順を記載（実行は手元で行ってください）。
3. [x] ルーティング設計
   - 既存ルートへ `/api/products` グループを追加。管理者ミドルウェアの適用範囲を決定。
4. [x] テスト設計（先行）
   - `backend/handler/product_test.go` を作成し、正常系・異常系（バリデーション・JSON不正・ID不正・DBエラー）を網羅するテストを記述。
5. [x] ハンドラー実装
   - テストが通る最小実装を行い、必要に応じてリファクタリング。
6. [ ] 統合テスト
   - ルーター経由で管理者認可ケースや未認証のGETを検証する統合テストを追加。
7. [ ] 動作確認
   - `go test ./...` を実行して全体のテストを確認。
8. [ ] ドキュメント更新
   - 実装・仕様変更を `doc/api.md` と本 `doc/task.md` に反映。

### バリデーション案
- `name`: 必須、文字数上限あり
- `price`: 必須、数値、`price > 0`（仕様に合わせて調整）
- `description`: 任意

### 注意点
- DBのNULL値（`sql.NullString` 等）はAPIレスポンスに変換する。
- エラーフォーマットは `pkg/respond` の仕様に揃える。

### 実行コマンド例（参照）
```bash
# sqlc で型定義とクエリを生成（環境で実行してください）
sqlc generate

# 全テスト実行
go test ./...
```

---
作成日: 2026-02-12
