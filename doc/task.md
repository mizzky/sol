# タスク一覧（議事録: 2026/02/01 参照）

## 概要
2026/02/01 の議事録に基づき、当面の実装優先度と具体タスクを整理する。

## 優先タスク（Phase 1 — 商品・カテゴリ管理）
- **マイグレーション**: `categories` / `products` テーブル作成（not-started）
- **sqlc 設定**: query 定義と `sqlc.yaml` の更新（not-started）
- **CRUD API 実装**: カテゴリと商品の登録・更新・削除・一覧（not-started）

## 次フェーズ（Phase 2 — カート機能）
- **テーブル作成**: `carts` テーブル定義（not-started）
- **API 実装**: カート追加・更新・削除（not-started）

## 取引系（Phase 3 — 注文・取引）
- **テーブル作成**: `orders`, `order_items`（not-started）
- **注文API**: 注文作成（トランザクション）、確認、キャンセル（not-started）

## 認可・ミドルウェア
- **管理者チェックミドルウェア実装**: `AdminOnly()` 相当（in-progress）

## 当面の短期タスク（今週）
- カテゴリ・商品マイグレーション作成とマイグレーション適用手順の確認（owner: backend）
- `sqlc` 用の query ファイルに基本CRUDクエリを追加（owner: backend）
- 管理者ミドルウェアのスケルトン実装とハンドラーへの組み込み（owner: backend）

## 担当と優先度（提案）
- 高: マイグレーション、sqlc 設定、CRUD API（Phase 1 完了が最優先）
- 中: ミドルウェア（認可）、カートAPI（Phase 2）
- 低: 売上レポート等管理者向け機能（Phase 3 後半）

## 次のアクション
1. `backend/migrations` にカテゴリ・商品用 migration を追加する（提案コミット: "Add migrations for categories and products"）
2. `query.sql` に sqlc 用の CRUD クエリを追加する
3. ミドルウェアの雛形を `auth` パッケージに追加する

---
ファイル生成日時: 2026-02-01
