# 2026-02-21 学習記録

## セッション概要
- カート機能の開発準備（TDDメンター使用）
- DB設計（マイグレーション + sqlc クエリ）の作成とPR作成

## 学んだこと・解決した課題

### 1. sqlc のメタコメント書式
**詰まった点**:
- `sqlc generate` で "invalid metadata: -- name:UpdateCartItemQty :one" エラーが発生

**原因**:
- sqlc のメタコメントは `-- name:` と名前の間に空白が必須
- 誤: `-- name:UpdateCartItemQty :one`
- 正: `-- name: UpdateCartItemQty :one`

**学び**:
- sqlc の書式は厳密。コメント構文の空白も重要。

### 2. Postgres の ON CONFLICT を使った安全な upsert
**学んだこと**:
- `GetOrCreateCartForUser` クエリで ON CONFLICT を使うことで、トランザクションやロックなしで「存在すればそのまま、なければ作成」を実現できる
- 前提条件: UNIQUE 制約が必要（例: `UNIQUE(user_id)` on carts）
- メリット: ラウンドトリップ削減、競合の自動解決

```sql
INSERT INTO carts (user_id, created_at, updated_at)
VALUES ($1, now(), now())
ON CONFLICT (user_id) DO UPDATE
  SET updated_at = carts.updated_at
RETURNING id, user_id, created_at, updated_at;
```

### 3. カート所有権チェックの設計方針
**設計上の懸念**:
- RemoveCartItem と ClearCart で参照する ID が混同しやすい（cart_item_id vs cart_id）
- クライアントから直接 cart_id を受け取ると所有権チェックが複雑になる

**採用した方針**:
- クライアントには cart_id を渡さない
- サーバ側で認証済みユーザの userID → cart を解決
- DELETE 操作は所有権検証を SQL に組み込む

```sql
-- RemoveCartItem（所有権検証付き）
DELETE FROM cart_items ci
USING carts c
WHERE ci.id = $1         -- cart_item_id
  AND ci.cart_id = c.id
  AND c.user_id = $2;    -- user_id (サーバ側で渡す)
```

**学び**:
- セキュリティを考慮した API 設計では「クライアントに何を渡さないか」も重要
- SQL レベルで所有権チェックを行うことで、アプリ層のバグを防げる

### 4. dev container 環境での migrate コマンド
**環境理解**:
- docker-compose.yml で app と db が同じネットワークを共有（network_mode: service:db）
- app コンテナから localhost:5432 で DB に接続可能
- Postgres の認証情報: user/password/coffeesys_db

**実行手順**:
```bash
# app コンテナ内で実行
export DATABASE_URL='postgres://user:password@localhost:5432/coffeesys_db?sslmode=disable'
migrate -path backend/db/migrations -database "$DATABASE_URL" up
```

**学び**:
- dev container の設定ファイル（docker-compose.yml）を確認することで、正しい接続情報を把握できる
- ローカル開発では sslmode=disable が一般的だが、本番では適切な SSL 設定が必要

### 5. sqlc で Querier インタフェース変更時のテスト対応
**問題**:
- sqlc でクエリを追加すると db.Querier インタフェースにメソッドが追加される
- 既存のテストモック（FakeQuerier/BadQuerier）がそのメソッドを実装していないとコンパイルエラー

**対処**:
- テストモックに新メソッドのスタブを追加
- FakeQuerier: 空実装（nil/空配列を返す）
- BadQuerier: エラーを返す（sql.ErrConnDone）

**学び**:
- sqlc を使う場合、クエリ追加 → インタフェース変更 → 既存モック更新の流れを理解しておく必要がある
- テストインフラの保守も開発の一部

### 6. 在庫確保タイミングの設計判断
**選択肢**:
- AddToCart 時に在庫を確保 → 一貫性は高いがキャンセル/期限管理が必要
- Checkout 時に在庫を検証・確保 → 一般的、在庫の長期ロックを避けられる

**採用した方針**:
- Checkout 時に在庫確保（ユーザからの選択）

**学び**:
- 設計判断は「何を優先するか」で変わる
- 今回は在庫利用効率とシンプルさを優先

## 次のアクション
- PR マージ後、次のステップ:
  1. AuthRequired ミドルウェアの実装
  2. カートハンドラのテストコード作成（TDD）
  3. カートハンドラの実装
  4. フロントエンド（useCartStore, API クライアント）
  5. 統合テスト（在庫競合ケース）

## 参考資料
- doc/planning/cart-plan-2026-02-16.md
- doc/planning/cart-ownership-policy.md
- golang-migrate/migrate: https://github.com/golang-migrate/migrate
- sqlc documentation: https://docs.sqlc.dev/
