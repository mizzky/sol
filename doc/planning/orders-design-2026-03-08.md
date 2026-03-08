# 注文・在庫システム設計ドキュメント（2026-03-08）

## 1. 要件定義

### 1.1 目的
EC サイト（sol_coffeesys）における注文作成と在庫管理の堅牢なシステムを構築する。

### 1.2 MVPスコープ
以下の機能をスコープとします（ドラフト plan より抽出）：

| チケット | 機能 | 優先度 | スコープ内 |
|---------|------|--------|---------|
| 1 | DBマイグレーション（orders, order_items, payments） | P0 | ✅ |
| 2 | sqlcクエリ（FOR UPDATE含む） | P0 | ✅ |
| 3 | CreateOrderHandler | P0 | ✅ |
| 4 | CancelOrderHandler | P0 | ✅ |
| 5 | 決済抽象化インタフェース | P2 | ⛔（モック実装のみ） |
| 6 | 冪等性（idempotency-key） | P1 | ⛔（将来タスク） |
| 7 | 同時性テスト | P0 | ✅ |
| 8 | エラーハンドリング | P1 | ✅ |
| 9 | メトリクス/監視 | P2 | ⛔（将来タスク） |

**除外理由:**
- チケット 5: 決済処理はモック実装で十分。本番連携は別フェーズ。
- チケット 6: 注文重複防止は、冪等性キーより「クライアント側の重複送信防止ロジック」を優先。
- チケット 9: メトリクスはシステム成熟後に追加。

### 1.3 受け入れ基準

#### 機能面
- [ ] DBマイグレーション（v8, v9, v10）により orders, order_items, payments テーブルが作成される
- [ ] `GET /api/orders` で認証済みユーザーの注文一覧が取得できる
- [ ] `POST /api/orders` で注文が DB トランザクション内で安全に作成される
  - 各商品の在庫が正しくデクリメントされる
  - 在庫不足時は注文が作成されず、409 Conflict が返される
- [ ] `PUT /api/orders/:id/cancel` でキャンセル時に在庫が巻き戻される
- [ ] 複数の同時リクエストでもオーバーソールドが発生しない（同時性テスト）

#### テスト面
- [ ] ユニットテスト（ハンドラー層）全件 PASS、カバレッジ > 80%
- [ ] 統合テスト（トランザクション成功/失敗ケース）が実行可能
- [ ] 同時性テスト（数十の並列リクエスト）で在庫整合性を確認

---

## 2. ユースケース

### 2.1 注文作成フロー（カート Checkout）

```
ユーザー: ログイン済み、カートに商品が存在

1. フロント: POST /api/orders（リクエストボディ空。暗黙的にカート参照）を送信
2. バック: リクエストを受け取り、ユーザーのカートを取得
         → カートが空なら 400 "EmptyCart" を返す
         → 各商品を SELECT FOR UPDATE で取得＆ロック（デッドロック防止に ID 昇順）
         → 各商品の在庫チェック実施
         → 不足時: 409 Conflict + ロールバック
         → 十分なら: 在庫デクリメント
         → orders レコード作成（status='pending'）
         → order_items レコード作成（product_name_snapshot 含む、複数行）
         → cart_items をクリア（カート内容をコピー元から削除）
         → Tx コミット
3. 成功時: Status 201 Created、注文ID + 合計金額を返却
4. 失敗時: Status 400/404/409 に応じたエラー（既述）
```

### 2.2 注文キャンセルフロー

```
ユーザー: 注文が pending ステータス

1. フロント: POST /api/orders/:id/cancel を送信（リクエストボディ空）
2. バック: authorization チェック（ユーザー自身の注文のみ）
         → 注文を SELECT FOR UPDATE で取得（FOR UPDATE で排他ロック）
         → ステータスチェック
           ├─ pending のみキャンセル可能
           └─ already cancelled なら 400 "InvalidStatusTransition"で返す
         → すべての order_items を取得
         → 各商品の在庫を INCREMENT（quantity 分）
         → orders.status = 'cancelled'
         → orders.cancelled_at = NOW()
         → Tx コミット
3. 成功時: Status 200 OK、キャンセル完了を返却
4. 失敗時: Status 400/404（既述）
```

### 2.3 注文一覧表示

```
ユーザー: ログイン済み

1. フロント: GET /api/orders を送信（クエリパラメータ: status=pending など）
2. バック: 認証済みユーザーの注文のみ取得し、items と紐付けて返却
3. 成功時: Status 200 OK、注文一覧（最大100件、ページネーション対応予定）
```

---

## 3. API 仕様設計

### 3.1 注文一覧取得

```http
GET /api/orders?status=pending&limit=20&offset=0
Authorization: Bearer <token>
```

**レスポンス (200 OK):**
```json
{
  "orders": [
    {
      "id": 1,
      "user_id": 1,
      "status": "pending",
      "total": 5000,
      "created_at": "2026-03-08T10:00:00Z",
      "items": [
        {
          "id": 10,
          "product_id": 1,
          "product_name": "Coffee Beans",
          "quantity": 2,
          "unit_price": 2000,
          "subtotal": 4000
        }
      ]
    }
  ],
  "total_count": 5
}
```

**エラーレスポンス (401 Unauthorized):**
```json
{
  "error": "Unauthorized",
  "message": "Please provide a valid token"
}
```

---

### 3.2 注文作成（Checkout）

```http
POST /api/orders
Authorization: Bearer <token>
Content-Type: application/json

{}
```

**説明:** リクエストボディは空。バックエンド側で認証済みユーザーの現在のカート（GetOrCreateCartForUser）から自動取得。

**エラーレスポンス (401 Unauthorized):**
```json
{
  "error": "認証が必要です"
}
```

**レスポンス (201 Created):**
```json
{
  "id": 1,
  "user_id": 1,
  "status": "pending",
  "total": 5000,
  "created_at": "2026-03-08T10:00:00Z",
  "items": [
    {
      "product_id": 1,
      "quantity": 2,
      "unit_price": 2000,
      "subtotal": 4000,
      "product_name_snapshot": "Coffee Beans"
    }
  ]
}
```

**エラーレスポンス (400 Bad Request - カート空):**
```json
{
  "error": "カートが空です"
}
```

**エラーレスポンス (409 Conflict - 在庫不足):**
```json
{
  "error": "在庫不足です。商品ID 1 は 2 個しかありません（リクエスト: 5 個）"
}
```

**エラーレスポンス (404 Not Found - 商品不在):**
```json
{
  "error": "商品が見つかりません"
}
```

---

### 3.3 注文キャンセル

```http
POST /api/orders/:id/cancel
Authorization: Bearer <token>
```

**レスポンス (200 OK):**
```json
{
  "id": 1,
  "user_id": 1,
  "status": "cancelled",
  "total": 5000,
  "created_at": "2026-03-08T10:00:00Z",
  "cancelled_at": "2026-03-08T11:30:00Z"
}
```

**エラーレスポンス (404 Not Found - 注文不在または非所有):**
```json
{
  "error": "注文が見つかりません"
}
```

**エラーレスポンス (400 Bad Request - 既にキャンセル済み):**
```json
{
  "error": "注文は既にキャンセルされています"
}
```

---

## 4. データベース設計

### 4.1 products テーブル（既存）

**既存スキーマの実態:**
```sql
-- 当初
CREATE TABLE products (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  price INTEGER NOT NULL,
  is_available BOOLEAN NOT NULL DEFAULT TRUE
);

-- マイグレーション v4 で拡張
ALTER TABLE products
ADD COLUMN category_id BIGINT NOT NULL DEFAULT 1,
ADD COLUMN sku VARCHAR(100) NOT NULL UNIQUE,
ADD COLUMN description TEXT,
ADD COLUMN image_url TEXT,
ADD COLUMN stock_quantity INTEGER NOT NULL DEFAULT 0,
ADD COLUMN created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
ADD COLUMN updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
```

**注文で使用するカラム:**
- `stock_quantity` — 在庫数（Integer）
- `name` — 注文時の商品名スナップショット用
- `price` — 参考用（order_items.unit_price が正式）

### 4.2 orders テーブル（新規）

**スキーマ設計:** 既存スキーマ（users, cart）に準拠（BIGSERIAL, TIMESTAMP WITH TIME ZONE）

```sql
CREATE TABLE IF NOT EXISTS orders (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  status VARCHAR(50) NOT NULL DEFAULT 'pending',  -- pending, cancelled
  total BIGINT NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  cancelled_at TIMESTAMP WITH TIME ZONE
);

-- インデックス: ユーザーごとの注文取得を高速化
CREATE INDEX IF NOT EXISTS idx_orders_user_id_created_at ON orders(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
```

**ステータス値:**
- `pending` — 作成直後、在庫確保済み
- `cancelled` — ユーザーによるキャンセル、在庫復旧済み
- ※ `completed` は将来タスク（決済連携後）。MVP では未使用。

### 4.3 order_items テーブル（新規）

**スキーマ設計:** 既存スキーマ（cart_items）に準拠（BIGSERIAL, TIMESTAMP WITH TIME ZONE）

```sql
CREATE TABLE IF NOT EXISTS order_items (
  id BIGSERIAL PRIMARY KEY,
  order_id BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  product_id BIGINT NOT NULL REFERENCES products(id),
  quantity INTEGER NOT NULL CHECK (quantity > 0),
  unit_price BIGINT NOT NULL,  -- スナップショット：注文時の価格保持
  product_name_snapshot TEXT NOT NULL,  -- スナップショット：商品名保持（将来の商品変名対応）
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- インデックス: 注文キャンセル時に各商品を素早く取得
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);
CREATE INDEX IF NOT EXISTS idx_order_items_product_id ON order_items(product_id);
```

**スナップショット戦略:**
- `unit_price` — 注文時の価格を保存（将来の値下げ・値上げ表示用）
- `product_name_snapshot` — 注文時の商品名を保存（将来の商品名変更対応）

### 4.4 payments テーブル（新規、現在は使用予定なし）

**スキーマ設計:** 決済連携タスク（P2 フェーズ）の実装時に使用予定。MVP では空テーブル。

```sql
CREATE TABLE IF NOT EXISTS payments (
  id BIGSERIAL PRIMARY KEY,
  order_id BIGINT NOT NULL UNIQUE REFERENCES orders(id) ON DELETE CASCADE,
  amount BIGINT NOT NULL,
  status VARCHAR(50) NOT NULL DEFAULT 'pending',  -- pending, completed, failed
  payment_method VARCHAR(50),
  external_transaction_id VARCHAR(100),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
```

**注記:**
- MVP では INSERT なし。status = 'pending' のままで注文手数料なし
- P2 決済連携タスク開始時に、決済プロバイダ統合（Stripe 等）と併せて本実装
- status 値: `pending` (初期), `completed` (決済成功), `failed` (決済失敗)

---

## 5. トランザクションフロー

### 5.1 注文作成時（CreateOrderHandler）

```
1. リクエストバリデーション
   ├─ items が空でないか
   ├─ product_id, quantity が妥当か
   └─ 認証済みユーザーか

2. トランザクション開始
   ├─ FOR UPDATE で各商品の在庫ロック（デッドロック防止に product_id 順）
   ├─ 各商品が存在するか確認
   ├─ 各商品の在庫が十分か確認
   │  └─ 不足時: エラーレスポンス + ロールバック
   ├─ 各商品の在庫をデクリメント
   ├─ orders レコード作成（status='pending', total計算）
   ├─ order_items レコード作成（複数行）
   └─ トランザクションコミット

3. レスポンス返却（201 Created）
```

### 5.2 注文キャンセル時（CancelOrderHandler）

```
1. 認証チェック
   └─ ユーザーが自分の注文のみ操作可能

2. トランザクション開始
   ├─ FOR UPDATE で orders レコード取得
   ├─ ステータスチェック（pending のみキャンセル可能）
   │  └─ 他のステータス時: エラーレスポンス + ロールバック
   ├─ order_items を全取得
   ├─ 各商品の在庫をインクリメント
   ├─ orders.status = 'cancelled'
   ├─ orders.cancelled_at = NOW()
   └─ トランザクションコミット

3. レスポンス返却（200 OK）
```

---

## 6. エラーハンドリング方針

### 6.1 HTTP ステータスコード

| 状況 | ステータス | メッセージ例 |
|------|----------|-----------|
| 認証なし | 401 | `認証が必要です` |
| 注文不在または非所有 | 404 | `注文が見つかりません` |
| 商品不在 | 404 | `商品が見つかりません` |
| 在庫不足 | 409 | `在庫が不足しています` |
| ステータス遷移エラー | 400 | `この注文はキャンセルできません` |
| DB エラー | 500 | `サーバーエラーが発生しました` |

**統一方針:** 既存コードベース（cart.go, product.go よりの respond.RespondError 呼び出し） に合わせ、エラーコード廃止して単一の error フィールドのみ使用。

### 6.2 エラーレスポンス形式

**統一形式（既存コード準拠）:**
```json
{
  "error": "<ユーザー向けメッセージ>"
}
```

**例（認証エラー）:**
```json
{
  "error": "認証が必要です"
}
```

**例（商品不在）:**
```json
{
  "error": "商品が見つかりません"
}
    "product_id": 1,
    "product_name": "Coffee Beans",
    "requested": 5,
    "available": 2
  }
}
```

---

## 7. テスト戦略

### 7.1 ユニットテスト

**対象**: `handler/order.go`

**テストケース:**
- CreateOrderHandler
  - ✅ 正常系：注文が正しく作成される、order_items に product_name_snapshot が保存される
  - ✅ 認証なし：401 Unauthorized
  - ✅ 商品不在：404 ProductNotFound
  - ✅ 在庫不足：409 InsufficientStock
  - ✅ 不正な数量：400 Bad Request
  - ✅ DB エラー：500 Internal Server Error
  
- CancelOrderHandler
  - ✅ 正常系：注文がキャンセルされ在庫が巻き戻される
  - ✅ 注文不在：404 Not Found
  - ✅ 既にキャンセル済み：400 InvalidStatusTransition
  - ✅ 他ユーザーの注文：404 Not Found（非所有を見えなくする）
  - ✅ DB エラー：500 Internal Server Error

- GetOrdersHandler
  - ✅ 正常系：自分の注文一覧を取得
  - ✅ ステータスフィルター機能（pending/cancelled）
  - ✅ ページネーション: limit/offset パラメータが正常に機能
  - ✅ total_count フィールドが正確（count query 使用）
  - ✅ 認証なし：401 Unauthorized

**テストパターン**: テーブル駆動テスト + mockdb を使用

### 7.2 統合テスト

**対象**: `tests/order_integration_test.go`

**テストケース:**
- ✅ 注文作成 → キャンセル → 在庫確認のフルフロー
- ✅ product_name_snapshot が注文時のスナップショット値で保存されることを検証
- ✅ トランザクション成功時のコミット確認
- ✅ トランザクション失敗時のロールバック確認（在庫が戻ること）
- ✅ 複数商品の同時発注、複数 order_items レコードの正確性
- ✅ 複数ユーザーの同時注文

### 7.3 同時性テスト

**対象**: `tests/concurrency_test.go`（新規）

**シナリオ:**
- ✅ 同じ商品に対して複数ユーザーが同時注文
  - 期待：在庫が正確にデクリメントされ、合計が元の在庫-合計数量になる
  - 期待：在庫不足時のエラーが正しく発動する
  - テスト方法：goroutine で N 本の CreateOrder リクエストを並列実行、最後に在庫を確認
  
- ✅ 同一ユーザーが重複した注文を送信
  - 期待：複数注文が作成される（重複を許容せず分離）
  - テスト方法：同じカート内容の CreateOrder を 2 回送信、2 つの order_id が異なることを確認



---

## 8. チケット分割案（実装順序）

### フェーズ 1: 設計・準備（0.5 日）

- **チケット 0-1**: 要件定義・API 設計ドキュメント完成（本ファイル）
- **チケット 0-2**: DB スキーマ設計確定、マイグレーション番号決定（v8, v9, v10）

### フェーズ 2: DB 準備（1 日）

- **チケット 1**: DB マイグレーション実装
  - contents:
    - `000008_create_orders_table.up/down.sql`
    - `000009_create_order_items_table.up/down.sql`
    - `000010_create_payments_table.up/down.sql`
  - テスト: マイグレーション実行 OK、リバート OK

### フェーズ 3: クエリ層実装（1.5 日）

- **チケット 2**: sqlc クエリ実装（query.sql → code generation）
  - contents:
    - GetProductForUpdate（FOR UPDATE）
    - UpdateProductStock（デクリメント用）
    - CreateOrder, CreateOrderItems（複数行挿入）
    - GetOrdersByUser, GetOrderByID, UpdateOrderStatus
    - CancelOrder（ステータス更新）
    - GetOrderItemsByOrderID（キャンセル時の逆引き）
    - GetOrderCount（ページネーション用 COUNT クエリ）
  - テスト: code generation 成功、型安全性確認

### フェーズ 3-補: トランザクション設計確定（0.5 日）

- **チケット 2-1**: Transaction Handler Pattern Design
  - contents:
    - handler 内での db.BeginTx() 呼び出しパターン検討
    - WithTx(tx *sql.Tx) による Queries 再バインディング確認
    - テスト時の MockDB では Tx モック不可であることを確認、ユニットテストの工夫方法を検討
    - Service層又は handler 内での統一的パターン決定・ドキュメント化
  - 成果物: Transaction Pattern ドキュメント、チケット 3-5 の実装ガイドライン

### フェーズ 4: ハンドラー層実装（2.5 日）

- **チケット 3**: CreateOrderHandler - TDD
  - Step 1: テスト設計とテストコード作成
  - Step 2: プロダクトコード実装（トランザクション処理、チケット 2-1 のパターンに従う）
  - Step 3: ユニット・統合テスト全 PASS
  
- **チケット 4**: CancelOrderHandler - TDD
  - Step 1: テスト設計とテストコード作成
  - Step 2: プロダクトコード実装（ステータス遷移・在庫巻き戻し）
  - Step 3: ユニット・統合テスト全 PASS
  
- **チケット 5**: GetOrdersHandler - TDD
  - Step 1: テスト設計とテストコード作成
  - Step 2: プロダクトコード実装
  - Step 3: テスト全 PASS

### フェーズ 5: 検証・テスト（1.5 日）

- **チケット 6**: 同時性テスト実装
  - contents:
    - 複数 goroutine で CreateOrder 並列実行
    - 在庫整合性確認
    - テスト成功基準：オーバーソールなし
  
- **チケット 7**: エラーハンドリング・エッジケース
  - contents:
    - HTTP ステータス確認（400/401/404/409/500）
    - エラーレスポンス形式確認
    - トランザクション失敗時の振る舞い

### フェーズ 6: 統合・ドキュメント（0.5 日）

- **チケット 8**: ルーティング登録 + 最終確認
  - contents:
    - `routes/routes.go` に 3 エンドポイント登録
    - RequireAuth ミドルウェア装着
    - `test.http` に実行例追記
  
- **チケット 9**: ドキュメント更新
  - contents:
    - `doc/api.md` に注文 API 追記
    - `doc/task.md` に進捗記録

---

## 9. 実装ロードマップ

| フェーズ | 期限 | 担当 | 成果物 |
|---------|-----|------|--------|
| 0 | 3/8（今日） | メンター+ユーザー確認 | 本ドキュメント（確定） |
| 1 | 3/10 | ユーザー | マイグレーションファイル 3 個 |
| 2 | 3/12 | ユーザー | query.sql 拡張・code generation OK |
| 3 | 3/17 | ユーザー（TDD メンター支援） | 3 ハンドラー + テスト全 PASS |
| 4 | 3/20 | ユーザー | 同時性テスト OK、エラーハンドリング確認 |
| 5 | 3/21 | ユーザー | ルーティング登録、ドキュメント完成 |

**想定工数**: 約 15 日（実装 + TDD メンター支援）

---

## 10. 設計決定の背景（Appendix）

### 10.1 Status 値の選定（pending + cancelled のみ）

**決定内容**: MVP では status = {pending, cancelled} のみ。completed は不実装。

**理由:**
- `completed` ステータスは決済完了を意味するが、MBA第1版ではモック決済を使用
- 実際の決済プロバイダ連携は P2 フェーズ（後続スプリント）で実装予定
- MVP では order.status = pending のままで手数料課金なし、キャンセルのみ可能
- completed 遷移ロジックは P2 で追加（決済成功をトリガーとして status 更新）

**将来計画 (P2 决済統合タスク):**
- payments テーブルに決済ステータスを記録
- Webhook で決済成功時に order.status → completed に更新
- 配送ステータス遷移（pending → shipped → delivered）は P3 以降

### 10.2 Checkout モデル（空の POST /api/orders）

**決定内容**: POST /api/orders のリクエストボディは空。バックエンド側でセッション内のカートを自動参照。

**理由:**
- 既存 cart/cart_items テーブルがあり、カート状態が DB に永続化されている
- frontend で user_id を特定済み（RequireAuth 認証済み）
- RequestBody に商品 ID とか数量を送信する二度手間を避ける
- RESTful では「カートを注文に変換する」という操作なので POST /api/orders で表現

**代替案 (検討済みだが採用しなかった):**
1. POST /api/orders { items: [{ product_id, quantity }] } — cart 無視、直接送信
   - 理由: 既存 cart UX を破壊するため非採用
2. POST /api/orders?cart_id=X で明示指定
   - 理由: ユーザーごとカート 1 個制約なので不要

### 10.3 product_name_snapshot の必要性

**決定内容**: order_items に product_name_snapshot TEXT フィールドを追加。

**理由:**
- 将来、管理画面で商品名を変更される可能性がある
- 過去の注文履歴で「今は存在しない商品名」を表示できないと UX 破壊（"商品 ID: 1234" という表示になる）
- unit_price と同様スナップショット戦略で解決

**実装例:**
```go
// query.sql で
INSERT INTO order_items (order_id, product_id, quantity, unit_price, product_name_snapshot)
VALUES ($1, $2, $3, $4, (SELECT name FROM products WHERE id = $2))
```

### 10.4 payments テーブル（MVP は未使用）

**決定内容**: payments テーブルを作成するもスキーマのみ。MVP ではレコード挿入なし。

**理由:**
- DB マイグレーション管理上、決済連携時の追加修正を失くしたい
- チケット 9 で ER 図更新時に既に payments テーブル存在の想定が可能
- P2 フェーズで決済プロバイダ（Stripe 等）と連携時、カラム調整のみで済む

**MVP の支払い方法:** 全注文で status = pending のまま。手数料・決済なし。

**P2 フェーズでの拡張:**
- payments.status = completed に更新するロジック追加
- order.status → completed (Webhook で自動遷移)
- テスト環境で Stripe テストモード使用

### 10.5 Transaction Pattern（db.BeginTx の使用）

**決定内容**: CreateOrderHandler, CancelOrderHandler でトランザクション開始。

**パターン:**
```go
// handler/order.go
func CreateOrderHandler(c *gin.Context) {
  // ... 認証・検証
  
  tx, _ := db.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelSerializable})
  q := db.queries.WithTx(tx)
  
  // 商品ロック、在庫確認、注文作成...
  
  err := tx.Commit()
}
```

**テスト時の工夫:**
- ユニットテストでは tx をモックできない（MockDB の制限）
- 代替案: handler を unit test せず、統合テスト（実DB）で Tx 検証
- Unit test は各 query 関数のクエリ正確性に絞る

**実装ガイド (チケット 2-1 の成果物):**
- BeginTx 呼び出し箇所を統一化（Service層 vs handler 比較検討）
- MockDB の工夫（tx 関数呼び出しを Skip して handler mock を単純化）

---

## 付録: 質問と検討事項

### Q1: 重複注文防止（冪等性）
**現在の決定**: MVP では不実装。クライアント側で重複送信を防止するロジックを実装。
**将来タスク**: idempotency-key ヘッダを受け取り、同一キーなら既存注文を返す

### Q2: ページネーション
**現在の決定**: GET /api/orders?limit=20&offset=0 をサポート
**実装レベル**: シンプルなオフセット式（後に cursor-based への移行検討）

### Q3: 部分キャンセル（一部商品のキャンセル）
**現在の決定**: 注文全体のキャンセルのみ。部分キャンセルは将来タスク

### Q4: クーポン・割引
**現在の決定**: MVP では不実装。order_items.unit_price にスナップショット保持するのみ

### Q5: 配送・ステータス遷移
**現在の決定**: pending → cancelled のみ。配送状態は将来タスク

---

**ドキュメント作成日**: 2026-03-08  
**ステータス**: 要ユーザー確認  
**次アクション**: ユーザーが仕様を確認・修正依頼後、チケット分割にて task.md を更新
