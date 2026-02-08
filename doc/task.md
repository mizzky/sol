# タスク一覧（議事録: 2026/02/02 参照）

## 概要
2026/02/02 の議事録に基づき、当面の実装優先度と具体タスクを整理する。

## 優先タスク（Phase 1 — 商品・カテゴリ管理）
- **マイグレーション**: `categories` / `products` テーブル作成（completed）
- **sqlc 設定**: query 定義と `sqlc.yaml` の更新（completed）
- **CRUD API 実装**: カテゴリと商品の登録・更新・削除・一覧（in-progress）
	- カテゴリ作成（`POST /api/categories`）+ テスト（completed）
	- カテゴリ一覧/更新/削除（in-progress）
		- カテゴリ削除（`DELETE /api/categories/:id`）+ テスト（completed: 2026/02/07）
	- 商品CRUD（not-started）

## 次フェーズ（Phase 2 — カート機能）
- **テーブル作成**: `carts` テーブル定義（not-started）
- **API 実装**: カート追加・更新・削除（not-started）

## 取引系（Phase 3 — 注文・取引）
- **テーブル作成**: `orders`, `order_items`（not-started）
- **注文API**: 注文作成（トランザクション）、確認、キャンセル（not-started）

## 認可・ミドルウェア
-- **管理者チェックミドルウェア実装**: `AdminOnly()` 相当（in-progress）

### カテゴリCRUDへの認証導入計画 (概要)

- **目的**: 管理操作（カテゴリの作成・更新・削除）を管理者のみが実行できるようにし、APIの誤用・不正利用を防止する。
- **認証方式**: 既存の JWT 実装 (`auth.ValidateToken`) を利用し、トークン内の `user.id` クレームでユーザーを特定、DB の `users` テーブルの `role` を参照して `admin` 権限を確認する。

### 実施手順（高レベル）

1. `auth` パッケージにミドルウェアを追加
  - 追加ファイル: `backend/auth/middleware.go`
  - エンドポイントから `Authorization: Bearer <token>` を受け取り、`auth.ValidateToken` で検証。
  - 検証済みユーザーID を Gin コンテキストにセット（例: `c.Set("userID", id)`)。
  - ロール確認はミドルウェア内で DB クエリ（`GetUserByID` または `GetUserByEmail`）を実行して `role == "admin"` を確認。管理者でなければ `403 Forbidden` を返す。

2. ルートにミドルウェアを組み込む
  - 対象: `POST /api/categories`, `PUT /api/categories/:id`, `DELETE /api/categories/:id` など管理系のカテゴリルート
  - ルート定義のある `backend/routes/routes.go` にミドルウェアを適用する。

3. テストの更新
  - 既存のハンドラーテストを修正して、認証ヘッダを付与するケースを追加（正常系: 管理者トークン付与、異常系: トークンなし/非管理者トークン/無効トークン）。
  - テスト用に `auth` にテストトークン生成ヘルパーを追加（`auth/test_helper.go` またはモックトークンの埋め込み）。
  - 既存のモックDBパターンを継続利用し、ロール確認の DB 呼び出しをモックする。

4. ドキュメント更新
  - `doc/api.md` のカテゴリエンドポイントに認証要件（管理者）を明記。
  - `doc/task.md` に実施内容とテスト方針を記載（このファイルの更新を含む）。

5. 検証・リリース準備
  - `go test ./...` を実行し、ユニットテストを通す。
  - 手動または簡易統合でエンドポイントに対して動作確認。

### 影響範囲（ファイル）

- `backend/auth/jwt.go`（既存利用）
- `backend/auth/middleware.go`（新規）
- `backend/routes/routes.go`（ミドルウェア組み込み）
- `backend/handler/*.go`（必要に応じてコンテキストから userID を参照）
- `backend/handler/*_test.go`（認証ヘッダを付与するための修正）
- `doc/api.md`, `doc/task.md`（ドキュメント更新）

### テスト方針（詳細）

- 単体テスト: 既存ハンドラーテストに管理者トークンを付与する正常確認ケースと、トークンが無い/無効/非管理者の403を検証するケースを追加する。モックDBでユーザーの role を返すようにする。
- 統合テスト（任意）: DB にテストユーザーを作成して実際にトークン発行・利用してエンドツーエンド確認。

### 目安工数

- 実装: 1〜2 日（ミドルウェア + ルート組込）
- テスト更新: 0.5〜1 日
- ドキュメント更新と検証: 0.5 日

## 当面の短期タスク（今週）
- ✅ user テストケースの不足分を追加（completed: 2026/02/02）
  - `LoginUserHandler` の正常系・異常系テストケースを追加
  - モックの共通化（`mockdb_test.go` を作成）
  - 依存性注入の導入
- RegisterUserHandler のテストケース作成（owner: backend）
- カテゴリ一覧/更新/削除のハンドラー実装（owner: backend）
	- ✅ カテゴリ削除のハンドラーとテスト（completed: 2026/02/07）
- 商品CRUDのハンドラー実装（owner: backend）
- 管理者ミドルウェアのスケルトン実装とハンドラーへの組み込み（owner: backend）
- API エラー共通化の適用範囲整理（owner: backend）

## 担当と優先度（提案）
- 高: マイグレーション、sqlc 設定、CRUD API（Phase 1 完了が最優先）
- 中: ミドルウェア（認可）、カートAPI（Phase 2）
- 低: 売上レポート等管理者向け機能（Phase 3 後半）

## 次のアクション
1. 実装: `auth.AdminOnly` ミドルウェア作成（owner: backend, status: in-progress）
2. ルート統合: カテゴリの管理系ルートへミドルウェアを適用（owner: backend, status: not-started）
3. テスト: 既存ハンドラーテストへ認証ケースを追加（owner: backend, status: not-started）
4. ドキュメント: `doc/api.md` に認証要件を明記（owner: backend, status: not-started）
5. 検証: `go test ./...` 実行して修正（owner: backend, status: not-started）

## 進行中のタスク
1. RegisterUserHandler のテストケース作成に着手する
2. カテゴリ一覧/更新/削除のハンドラー実装に着手する
3. 商品CRUDのハンドラー実装に着手する
4. ミドルウェアの雛形を `auth` パッケージに追加する
5. カテゴリ一覧取得関数のプロダクトコード実装に着手する
6. カテゴリー削除機能のテストケース設計と実装（owner: backend, status: in-progress）

## 完了したタスク
- `RegisterUserHandler`のテストケース作成と改良。
  - 重複メールエラーやデータベース接続エラーのシナリオをカバー。
  - モック設定の調整。
- `HashPassword`のエラーハンドリングテストの問題特定。
  - `%w`を使用したエラーラッピングが原因のエラーメッセージ不一致問題を確認。
- `HashPassword`のエラーメッセージ確認のテストをパス。
- カテゴリ一覧取得関数のテストコードを完成。
