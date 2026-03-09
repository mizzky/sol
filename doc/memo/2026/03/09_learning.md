# 2026-03-09 学習ログ

## 本日の作業概要
注文機能（orders/order_items/payments）のDB設計ドキュメント完成 → チケット1（マイグレーション実装）へ進行開始

---

## 学んだ事項

### 1. 設計ドキュメント修正と8つのミスマッチ解消
**詰まったところ:**
- 初回設計で既存コードベースのパターンを見落としていた
- 独自の3層エラーレスポンス構造を提案したが、実装は単一フィールド形式（`{"error": "msg"}` のみ）

**解決策と知見:**
- 既存マイグレーション（v1-7）から一貫性を確認：**BIGSERIAL, TIMESTAMP WITH TIME ZONE が全テーブルで統一**
- 実装コードベース（respond/error.go）からエラー形式が既に決まっていることを確認
- 解決方法：**実装を優先 → 設計を合わせる** アプローチが正しい

**記録：**
- `orders` テーブル：BIGSERIAL (pk, user_id), TIMESTAMP WITH TIME ZONE (created_at, updated_at, cancelled_at)
- `order_items` テーブル：product_name_snapshot フィールド追加で将来の商品名変更対応（設計決定の背景を Appendix に記載）
- `payments` テーブル：MVP では使用予定なし（スキーマのみ作成、INSERT なし）

---

### 2. マイグレーション手法（migrate CLI）の確認
**詰まったところ:**
- ファイル生成と SQL 記述をどの順で進めるか不明確だった
- dirty 状態の解消方法が不明

**学んだ内容：**
- **migrate CLI の基本コマンド**
  - `migrate up / down` — applying/reverting
  - `migrate version` — current state + dirty flag
  - `migrate force <version>` — dirty を解除（破壊的、要バックアップ）
  
- **ファイル命名規則**
  - 既存プロジェクト: `000001_create_users_table.up.sql` / `.down.sql` （連番）
  - migrate create: タイムスタンプ自動付与（既存規則との統合判定が必要）
  - 本プロジェクト: 連番 v8, v9, v10 を選択

- **dirty 状態への対応（重要）**
  - マイグレーション失敗 → schema_migrations テーブルに dirty flag が立つ
  - 対処: DB 修好 → `force` で前バージョン指定 → `up` で再実行
  - 本番環境は必ずバックアップ（`pg_dump`）を取ること

**記録：**
  ```bash
  # 依存管理: orders → order_items → payments の順で up、逆順で down
  # DOWN は依存関係の逆： payments DROP → order_items DROP → orders DROP
  ```

---

### 3. CI/CD Troubleshooting（frontend npm upgrade 問題）
**詰まったところ:**
- CI で npm install -g npm@11.11.0 が失敗
- エラー：Node 18 と npm 11 の互換性（npm11 が Node >=20.17 を要求）

**分析プロセス:**
1. frontend/package.json に engines 指定がないか確認
2. 依存パッケージが Node 18 で動作するか検証
3. CI でこのステップが必須か判断

**結論と知見:**
- `Upgrade npm` ステップ削除は安全（現在の依存パッケージで十分）
- Node 18 + npm 10 で `npm ci` / `npm test` の実行に問題なし
- リスク：将来的に npm 11 機能が必須になるパッケージが追加される場合は Node をアップグレード

**記録：**
- セットアップノード v4 で Node 18 → 次は CI 変更（npm ステップ削除）を提案

---

### 4. TDD アプローチの本格開始
**準備状況:**
- チケット1（マイグレーション）が進行中
- チケット2-1（Transaction Handler Pattern Design）を新規追加
- チケット3-5（ハンドラー層 TDD）の実装ガイドが決定待ち

**学んだ効果:**
- 設計から実装へ移行する際の段階的な手法確認
- TDD mentor モードの事前テスト設計の重要性

---

## 次のアクション
1. チケット1: マイグレーションファイル実装完了 + migrate up/down 動作確認
2. CI 修正: frontend npm upgrade ステップ削除（提案予定）
3. チケット2: sqlc クエリ実装へ進行
4. チケット2-1: Transaction Handler Pattern を設計・ドキュメント化

---

## メモ・参考資料
- 設計ドキュメント：[doc/planning/orders-design-2026-03-08.md](../../planning/orders-design-2026-03-08.md)
- task.md での進捗：チケット1, 2-1 開始準備完了、チケット2-5 待機中
- DB マイグレーション参考コマンド：`migrate -path file://backend/db/migrations -database "$DATABASE_URL" version`
