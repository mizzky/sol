# タスク一覧（議事録: 2026/02/02 参照）

## 概要
2026/02/02 の議事録に基づき、当面の実装優先度と具体タスクを整理する。

## 優先タスク（Phase 1 — 商品・カテゴリ管理）
- **マイグレーション**: `categories` / `products` テーブル作成（completed）
- **sqlc 設定**: query 定義と `sqlc.yaml` の更新（completed）
- **CRUD API 実装**: カテゴリと商品の登録・更新・削除・一覧（in-progress）
	- カテゴリ作成（`POST /api/categories`）+ テスト（completed）
	- カテゴリ一覧/更新/削除（not-started）
	- 商品CRUD（not-started）

## 次フェーズ（Phase 2 — カート機能）
- **テーブル作成**: `carts` テーブル定義（not-started）
- **API 実装**: カート追加・更新・削除（not-started）

## 取引系（Phase 3 — 注文・取引）
- **テーブル作成**: `orders`, `order_items`（not-started）
- **注文API**: 注文作成（トランザクション）、確認、キャンセル（not-started）

## 認可・ミドルウェア
- **管理者チェックミドルウェア実装**: `AdminOnly()` 相当（not-started）

## 当面の短期タスク（今週）
- ✅ user テストケースの不足分を追加（completed: 2026/02/02）
  - `LoginHandler` の正常系・異常系テストケースを追加
  - モックの共通化（`mockdb_test.go` を作成）
  - 依存性注入の導入
- RegisterHandler のテストケース作成（owner: backend）
- カテゴリ一覧/更新/削除のハンドラー実装（owner: backend）
- 商品CRUDのハンドラー実装（owner: backend）
- 管理者ミドルウェアのスケルトン実装とハンドラーへの組み込み（owner: backend）
- API エラー共通化の適用範囲整理（owner: backend）

## 担当と優先度（提案）
- 高: マイグレーション、sqlc 設定、CRUD API（Phase 1 完了が最優先）
- 中: ミドルウェア（認可）、カートAPI（Phase 2）
- 低: 売上レポート等管理者向け機能（Phase 3 後半）

## 次のアクション
1. RegisterHandler のテストケース作成に着手する
2. カテゴリ一覧/更新/削除のハンドラー実装に着手する
3. 商品CRUDのハンドラー実装に着手する
4. ミドルウェアの雛形を `auth` パッケージに追加する

---
ファイル生成日時: 2026-02-03
