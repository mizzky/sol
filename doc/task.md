## フロント先行ミニマムEC連携（TODO）

### 目的
- ミニマムなEC体験を優先して、フロントとAPIの連携を早期に成立させる。
- 対象: 商品一覧 + 管理者の簡易商品追加 + 簡易ログイン。

### 前提・方針
- API仕様は実装に合わせてdocを整合済み。
- フロントは最小画面: 商品一覧、ログイン、商品追加。
- 認可が必要な操作はJWT（Bearer）を付与する。

### TODO（優先度順）
1. [x] API仕様の整合（レスポンス形式/PUT方針/管理者要約/404追加）
2. [x] フロントのAPI利用設計
   - トークン保持: `localStorage`
   - 状態管理: `zustand`
   - API呼び出し: 小さな fetch util を作成し、認可ヘッダを共通化
   - `GET /api/products` のレスポンス形（`{"products": [...]}`）を反映
   - `POST /api/login` の `token` を保存し、`Authorization: Bearer <token>` で利用
   - `POST /api/products` の必須項目（`name`, `price`, `category_id`, `sku`）を反映
3. [x] 画面設計（最小）
    - 商品一覧
       - 表示項目: `name`, `price`, `category_id`, `sku`
       - 初期表示: `GET /api/products`（レスポンス: `{"products":[...]}`）
       - 表示形式: シンプルなリスト（カード/テーブルは任意）
       - 空表示: 「商品がありません」
       - UI動作: 画面表示時に一覧取得。追加/更新後は再取得またはローカルに即時反映
    - ログイン
       - 入力項目: `email`, `password`
       - 成功時: レスポンスに含まれる `token` を `localStorage.setItem("token", token)` で保存
       - 失敗時: エラーメッセージを表示（バリデーション/認証エラーを分離）
    - 商品追加（管理者のみ）
       - 入力項目: `name`(必須), `price`(必須), `category_id`(必須), `sku`(必須)
       - 認可: リクエストに `Authorization: Bearer <token>` を付与
       - 成功時: 商品一覧を再取得または画面内リストに即時追加して反映
       - 失敗時: `401/403/400/404` に応じたエラーメッセージを表示（`404` はカテゴリ未存在）
4. [x] フロント実装（連携）
   - 商品一覧の取得と表示
   - ログインとトークン保存
   - 商品追加（認可ヘッダ付与）
   - **チケット8完了 (2026-02-19)**: 
     - `useAuthStore` に `login`/`register` 関数を追加
     - `fetchWithAuth` で認証ヘッダの自動付与を実装
     - 401エラー時の自動ログアウトを実装
     - 全テスト (17件) がPASS
5. [ ] 動作確認
   - 未ログイン時の一覧表示
   - ログイン後の商品追加

---
作成日: 2026-02-14
最終更新: 2026-02-19

---

## フロントエンド ページ構成整理（新規追加: 2026-02-19）

### 背景
現在のトップページに一般ユーザー向けの「商品一覧」と管理者向けの「商品登録フォーム」が同居しており、ページ構成が不自然。また、ユーザー登録画面が未実装。

### 目的
- トップページと管理機能を適切に分離
- ユーザー登録画面の実装
- ナビゲーションの実装でUX向上

### TODO（優先度順）
0. [x] **チケット1**: DB スキーマ拡張（reset_token カラム追加） (P0)
   - マイグレーション v5 を執行
   - `000005_add_reset_token_to_users.up.sql` を DB に適用
   - 完了日: 2026-02-19
   - 意義: ユーザー登録機能がDBで正常に動作するための前提条件
   
1. [x] **チケット10**: ユーザー登録画面の実装 (P0)
   - `/register` ページ作成
   - 名前、メール、パスワード入力フォーム
   - `useAuthStore.register()` 連携
   - 登録成功後は `/login` へリダイレクト
   - 実装詳細:
     - 入力項目: 名前、メールアドレス、パスワード、パスワード確認
     - バリデーション: メール形式、パスワード8文字以上、確認一致チェック
     - エラー表示: バリデーションエラー、400/500エラー、重複メールアドレス
   - 完了日: 2026-02-19
   - 影響: `frontend/app/register/page.tsx` (新規)
   - コミット例: `feat(frontend): add user registration page`
   
2. [x] **チケット11**: トップページと管理ページの分離 (P0)
   - **11-1: トップページのリファクタリング**
     - 商品登録フォームを削除
     - 商品一覧表示のみに特化
     - ローディング状態の改善
     - 影響: `frontend/app/page.tsx` (修正)
   
   - **11-2: 管理ページの作成**
     - `/admin/products` ページ新規作成
     - 商品追加フォーム（トップページから移行）
     - 商品一覧（編集・削除ボタン付き）
     - ロール確認: 管理者のみアクセス可
     - 影響: `frontend/app/admin/products/page.tsx` (新規)
   
   - **11-3: ナビゲーションヘッダーの実装**
     - ロゴ/タイトル（トップページへのリンク）
     - 状態別ナビゲーション:
       - 未ログイン: [ログイン] [新規登録]
       - ログイン済み（一般）: [ユーザー名] [ログアウト]
       - ログイン済み（管理者）: [ユーザー名（管理者）] [商品管理] [ログアウト]
     - 影響: `frontend/app/components/Header.tsx` (新規)
   
   - **11-4: レイアウトの更新**
     - Header コンポーネントを全ページに適用
     - 影響: `frontend/app/layout.tsx` (修正)
   
   - 完了日: 2026-02-26
   - 関連テストファイル: `frontend/app/__tests__/` (ページコンポーネントテスト、Header テスト)
   - コミット例: 
     - `feat(frontend): refactor top page (remove form)`
     - `feat(frontend): create admin products page`
     - `feat(frontend): add header navigation component`
     - `feat(frontend): integrate header to all pages`
   
3. [x] **チケット12**: 管理者権限チェックミドルウェア (P1)
   - HOC`AdminRoute` の実装
   - 未ログイン → `/login` へリダイレクト
   - ログイン済みだが非管理者 → `/` へリダイレクト + エラーメッセージ表示
   - ローディング状態の適切な表示
   - 実装詳細:
     - `useAuthStore()` から `user` とロール情報を取得
     - 権限チェック後にページレンダリング
     - リダイレクトロジックをテストカバー
   - 完了日: 2026-02-27
   - 影響: `frontend/app/components/AdminRoute.tsx` (新規)
   - テスト: `frontend/app/__tests__/AdminRoute.test.tsx` (新規)
   - コミット例: `feat(frontend): add admin authorization middleware`

### あるべきページ構成（実装完了）
```
/                     → 商品一覧（誰でも閲覧可）✅ 実装済み
/login                → ログイン ✅ 実装済み
/register             → ユーザー登録 ✅ 実装済み (チケット10)
/admin/products       → 商品管理（管理者のみ）✅ 実装済み (チケット11-2)
```

**実装完了日**: 2026-02-27 (チケット12 完了で全タスク完了)

**ナビゲーション**: Header コンポーネントで状態別メニビ表示 ✅ 実装済み (チケット11-3)

**管理者権限チェック**: AdminRoute HOC で保護 ✅ 実装済み (チケット12)

詳細: [doc/planning/frontend-pages-plan-2026-02-19.md](planning/frontend-pages-plan-2026-02-19.md)

---
追加日: 2026-02-19
最終更新: 2026-03-04

---

## カート操作ハンドラ実装（TDD）（新規追加: 2026-02-23）

### 背景
- DBスキーマとsqlcクエリは完成済み（[cart-plan-2026-02-16.md](planning/cart-plan-2026-02-16.md) チケット2, 3）
- カート操作のAPIハンドラとルーティングが未実装
- ログイン必須のカート機能をTDDサイクルで実装

### 目的
- カート操作の5つのエンドポイントを実装（追加、取得、数量更新、削除、全削除）
- 既存パターン（`handler/product.go`, `handler/user.go`）に準拠
- TDD（テスト駆動開発）で実装し、学習効果を最大化

### 実装方針
- **認証**: ログイン必須（`RequireAuth`ミドルウェアを新規作成）
- **権限**: ユーザーは自分のカートのみ操作可（`ByUser`系クエリを使用）
- **テスト**: テーブル駆動テスト + `testutil.MockDB`
- **開発サイクル**: TDD Mentor スキルに従う（Red → Green → Refactor）

### TODO（優先度順）

#### ステップ1: ブランチ作成
- [x] 作業ブランチ作成: `feature/cart-handlers`

-#### ステップ2: 認証ミドルウェア実装（チケット5対応）
- [x] **チケット13**: `RequireAuth` ミドルウェア実装 (P0)
  - 2-1: テスト設計（正常系、未認証、トークン不正、ユーザー不在）
  - 2-2: テストコード作成（`auth/middleware_test.go`）
  - 2-3: プロダクトコード実装（`auth/middleware.go`）
  - 2-4: テスト実行・確認
  - 影響: `backend/auth/middleware.go`, `backend/auth/middleware_test.go`
  - 受け入れ条件: `AdminOnly`と同様のロジックだがロール確認を除外、`c.Set("userID", user.ID)`を設定
  - コミット: `feat(auth): add RequireAuth middleware for general users`

#### ステップ3: カートハンドラ実装（チケット4対応）

- [x] **チケット14**: GetCartHandler - カート内容取得 (P0)
  - 3-1: 仕様設計（エンドポイント、レスポンス形式）
  - 3-2: テスト設計（正常系、空カート、DBエラー）
  - 3-3: テストコード作成（`handler/cart_test.go`）
  - 3-4: プロダクトコード実装（`handler/cart.go`）
  - 3-5: テスト実行・確認
  - エンドポイント: `GET /api/cart`
  - レスポンス: `{"items": [...]}`
  - コミット: `feat(handler): add GetCartHandler with tests`

- [x] **チケット15**: AddToCartHandler - 商品追加 (P0) (完了: 2026-02-25)
  - 3-6: 仕様設計（在庫確認はCheckout時に実施する方針で合意）
  - 3-7: テスト設計（正常系、商品不在、数量不正、未認証、DBエラー）
  - 3-8: テストコード作成（`TestAddToCartHandler` を追加、表駆動）
  - 3-9: プロダクトコード実装（`backend/handler/cart.go` に最小実装を追加）
  - 3-10: テスト実行・確認（ユニットテスト全件通過、カバレッジ100%）
  - エンドポイント: `POST /api/cart/items`
  - リクエスト: `{"product_id": 1, "quantity": 2}`
  - ブランチ/コミット: `feat/handler/add-to-cart` で実装・テストを追加

- [x] **チケット16**: UpdateCartItemHandler - 数量更新 (P0)
  - 3-11: 仕様設計
  - 3-12: テスト設計（正常系、アイテム不在、他ユーザー、数量不正）
  - 3-13: テストコード作成
  - 3-14: プロダクトコード実装
  - 3-15: テスト実行・確認
  - エンドポイント: `PUT /api/cart/items/:id`
  - リクエスト: `{"quantity": 5}`
  - コミット: `feat(handler): add UpdateCartItemHandler with authorization`

- [x] **チケット17**: RemoveCartItemHandler - アイテム削除 (P0)
  - 3-16: 仕様設計
  - 3-17: テスト設計（正常系、アイテム不在）
  - 3-18: テストコード作成
  - 3-19: プロダクトコード実装
  - 3-20: テスト実行・確認
  - エンドポイント: `DELETE /api/cart/items/:id`
  - APIステータス方針メモ（2026-02-27）:
    - `UpdateCartItemHandler`: `sql.ErrNoRows` は `404 Not Found`
    - `RemoveCartItemHandler`: 本プロジェクトでは `404 Not Found` に統一（不存在/非所有を同一扱い）
  - コミット: `feat(handler): add RemoveCartItemHandler`

- [x] **チケット18**: ClearCartHandler - カート全削除 (P0)
  - 3-21: 仕様設計
  - 3-22: テスト設計（正常系）
  - 3-23: テストコード作成
  - 3-24: プロダクトコード実装
  - 3-25: テスト実行・確認
  - エンドポイント: `DELETE /api/cart`
  - APIステータス方針メモ(2026-03-01):
    - `ClearCartHandler`:Delete処理はカートの有無にかかわらず`204 No Content`を返す（冪等性の保持）
  - コミット: `feat(handler): add ClearCartHandler`

#### ステップ4: ルーティング設定
- [x] **チケット19**: カートエンドポイント登録 (P0)
  - 4-1: `routes/routes.go` に5つのエンドポイント追加
  - 4-2: ルーティングテスト実行
  - 影響: `backend/routes/routes.go`
  - コミット: `feat(routes): register cart endpoints with RequireAuth` (完了日: 2026-03-02)

#### ステップ5: 検証
- [x] **チケット20**: 統合テスト・手動テスト (P1)
  - 5-1: 統合テスト作成（`tests/cart_integration_test.go`）
  - 5-2: `test.http` にカート操作のリクエスト例追加
  - 5-3: 動作確認（ログイン → カート操作の一連フロー）
  - コミット: `test(integration): add cart flow tests and HTTP examples`

### 参考資料
- 詳細計画: [doc/planning/cart-plan-2026-02-16.md](planning/cart-plan-2026-02-16.md)
- 既存ハンドラパターン: [backend/handler/product.go](../backend/handler/product.go), [backend/handler/user.go](../backend/handler/user.go)
- 既存ミドルウェア: [backend/auth/middleware.go](../backend/auth/middleware.go)
- 既存テストパターン: [backend/handler/product_test.go](../backend/handler/product_test.go)

### 進捗メモ
- 開始日: 2026-02-23
- 現在のステップ: ステップ3（GetCartHandler）完了
- 学習ポイント: TDDサイクル、テーブル駆動テスト、認証ミドルウェア

---
追加日: 2026-02-23

---

## フロント: カート機能実装タスク（追加: 2026-03-04）

### 目的
- トップページにカート機能を追加して、商品を選んでカートへ入れる一連のUXを提供する。

### 実装方針（決定済み）
- カートAPIはバックエンドで実装済みのため、フロントはAPI統合とUI実装に集中する。
- 状態管理は `zustand` を採用し、ヘッダーで数量バッジを常時表示する。
- 詳細なカート操作は専用ページ `/cart` で行えるようにする。

### 実行タスク（優先度順）

- [x] **チケット21**: API層を実装する (P0)
  - ファイル: `frontend/lib/api.ts`
  - 内容: `getCart()`, `addToCart(productId, quantity)`, `updateCartItem(itemId, quantity)`, `removeFromCart(itemId)`, `clearCart()` を実装
  - 影響: `frontend/lib/api.ts` (修正)
  - コミット例: `feat(frontend): add cart API functions`

- [x] **チケット22**: カート状態管理を作成 (P0)
  - ファイル: `frontend/store/useCartStore.ts`
  - 内容: `items`, `totalPrice`, `totalQuantity` と、`setCart`, `addItem`, `updateItem`, `removeItem`, `clearCart` を実装
  - 初期同期で `getCart()` を呼ぶ
  - 影響: `frontend/store/useCartStore.ts` (新規)
  - コミット例: `feat(frontend): add cart state management with Zustand`

- [x] **チケット23**: ヘッダーにカート表示追加 (P0)
  - ファイル: `frontend/app/components/Header.tsx` (既存)
  - 内容: カートアイコン + 数量バッジ。クリックで `/cart` へ遷移
  - `useCartStore` から `totalQuantity` を購読してバッジに表示
  - 影響: `frontend/app/components/Header.tsx` (修正)
  - コミット例: `feat(frontend): add cart icon with badge to header`

- [x] **チケット24**: 商品カードに追加ボタン実装 (P0)
  - ファイル: `frontend/app/page.tsx`（既存の商品の表示箇所）
  - 内容: 各商品に数量入力と「カートに追加」ボタンを追加。押下で `addToCart()` を呼ぶ
  - トースト通知で追加完了を表示
  - 影響: `frontend/app/page.tsx` (修正)
  - コミット例: `feat(frontend): add "Add to Cart" button to product cards`

- [x] **チケット25**: カート詳細ページを作成 (P0)
  - ファイル: `frontend/app/cart/page.tsx`
  - 内容: カート内アイテム一覧、数量変更、削除、合計金額表示、クリア/チェックアウトボタン
  - 空カート時の表示対応
  - 影響: `frontend/app/cart/page.tsx` (新規)
  - コミット例: `feat(frontend): create cart detail page with full CRUD operations`

- [x] **チケット26**: レイアウトへヘッダー統合 (P1)
  - ファイル: `frontend/app/layout.tsx`
  - 内容: 全ページで `Header` を表示するように調整（既に実装済みの可能性あり、確認が必要）
  - 影響: `frontend/app/layout.tsx` (確認・修正)
  - コミット例: `refactor(frontend): ensure header displays on all pages`

- [x] **チケット27**: テストを作成・実行 (P1)
  - ファイル例:
    - `frontend/store/__tests__/useCartStore.test.ts`
    - `frontend/app/components/__tests__/Header.test.tsx`（カートバッジ部分）
    - `frontend/app/cart/__tests__/page.test.tsx`
  - 内容: 状態更新、バッジ表示、数量変更等のユニット/コンポーネントテストを作成
  - 影響: 複数のテストファイル (新規)
  - コミット例: `test(frontend): add comprehensive tests for cart functionality`

- [x] **チケット28**: ドキュメントにタスク追記 (P0)
  - ファイル: `doc/task.md`（追記済み）
  - 完了日: 2026-03-04

### 受け入れ基準
- 商品一覧から商品を追加するとヘッダーのバッジが即時更新される
- `/cart` ページで数量変更・削除が可能で、合計金額が正しく計算される
- APIエラーや未認証時の挙動が適切にハンドリングされる

---

作成日: 2026-03-04


## 次にやりたいこと

- [ ] APIドキュメント整備
  - [ ] 実装済みのAPIについてドキュメンテーション
  - [ ] 今後実装予定のAPIについてAPI定義の設計

---

## API ドキュメント作成 (OpenAPI) — タスク

- [ ] [doc/openapi.yaml](doc/openapi.yaml) の初版ドラフトを作成（[backend/routes/routes.go](backend/routes/routes.go) を基に paths を埋める）
- [ ] components.schemas をハンドラの構造体に基づき追加（[backend/handler/](backend/handler/) を参照）
- [ ] JWT 認証（bearerAuth）と共通エラーレスポンスを components に定義（backend/auth を参照）
- [ ] swagger-cli と spectral で YAML の検証と lint を実行
- [ ] Redoc によるバンドルと doc/ 配置（例: doc/openapi.html を生成）
- [ ] CI（GitHub Actions）に OpenAPI lint を追加（/.github/workflows/openapi-lint.yml）

担当: あなた（backend を編集） — 私はドラフト作成支援・レビューツールや CI スニペットを提供します。

---

## 注文・在庫システム実装タスク（新規追加: 2026-03-08）

### 目的
- EC サイトの注文作成・キャンセル機能を実装し、同時実行での在庫整合性を保証する
- TDD サイクルで、テストを軸に安全で保守性の高いコードを構築

### 実装スコープ
✅ 対象（チケット 1-4, 7-8）:
- DB マイグレーション（orders, order_items, payments）
- sqlc クエリ（FOR UPDATE 含む）
- CreateOrderHandler（注文作成）
- CancelOrderHandler（注文キャンセル）
- 同時性テスト（オーバーソール防止確認）
- エラーハンドリング（HTTP ステータス・エラーコード）

⛔ 除外（将来タスク）:
- チケット 5: 決済抽象化（テスト用モック実装のみ）
- チケット 6: 冪等性（idempotency-key）
- チケット 9: メトリクス/監視

### API 仕様（確定）
- `GET /api/orders` — 認証済みユーザーの注文一覧取得
- `POST /api/orders` — 注文作成（商品 ID + 数量リストを受け取り）
- `POST /api/orders/:id/cancel` — 注文キャンセル（ステータス pending → cancelled、在庫巻き戻し）

### 実装計画（優先度順）

#### フェーズ 0: 設計（完了日: 2026-03-08）
- [x] **チケット 0-1**: 要件定義・API 設計ドキュメント完成
  - ファイル: `doc/planning/orders-design-2026-03-08.md`（本ドラフト）
  - 内容: ユースケース、API 仕様、DB スキーマ、トランザクションフロー、テスト戦略
  
- [x] **チケット 0-2**: DB マイグレーション番号確定（v8, v9, v10）
  - スキーマ最終確認後にファイル名を決定

#### フェーズ 1: DB マイグレーション実装（予定: 3/10）
- [x] **チケット 1**: DB マイグレーション作成 (P0, Effort: High)
  - 前提: DB スキーマ最終確認
  - 内容:
    - `000008_create_orders_table.up/down.sql`
    - `000009_create_order_items_table.up/down.sql`
    - `000010_create_payments_table.up/down.sql`
  - 受け入れ条件:
    - [x] マイグレーション実行で 3 テーブル作成
    - [x] リバート確認（down スクリプト実行で テーブル削除）
    - [x] sqlc code generation で型生成に支障なし
  - ファイル影響: `backend/db/migrations/`
  - コミット例: `feat(db): create orders, order_items, payments tables`

#### フェーズ 2: sqlc クエリ層実装（予定: 3/12）
- [x] **チケット 2**: sqlc クエリ拡張 (P0, Effort: High)
  - 前提: マイグレーション実行済み
  - 内容:
    - `GetProductForUpdate(ctx, id)` — FOR UPDATE で在庫ロック
    - `UpdateProductStock(ctx, id, decrement)` — 在庫デクリメント/インクリメント
    - `CreateOrder(ctx, userID, total, status)` — 注文ヘッダ作成
    - `CreateOrderItem(ctx, orderID, productID, qty, unitPrice)` — 注文アイテム作成（複数行可）
    - `GetOrdersByUser(ctx, userID)` — ユーザーの注文一覧取得
    - `GetOrderByID(ctx, id)` — 注文取得（FOR UPDATE 版あり）
    - `GetOrderItemsByOrderID(ctx, orderID)` — 注文の商品明細取得
    - `UpdateOrderStatus(ctx, id, status)` — ステータス更新
    - `GetOrderCount(ctx, userID)` — ユーザーの注文件数（ページネーション用）
  - 受け入れ条件:
    - [x] `backend/query.sql` にクエリを追記
    - [x] `sqlc generate` 実行成功
    - [x] `backend/db/querier.go` に新メソッドが型安全に追加
  - ファイル影響: `backend/query.sql`, `backend/db/querier.go`
  - コミット例: `feat(db): add order-related sqlc queries with FOR UPDATE`

- [x] **チケット 2-1**: Transaction Handler Pattern Design (P0, Effort: Medium)
  - 前提: チケット 2 完了後
  - 目的: handler 内での db.BeginTx() 呼び出しパターンと WithTx の使用法を統一化
  - 内容:
    - [x] handler 内での `db.BeginTx(ctx context.Context, opts *sql.TxOptions)` 呼び出しパターンを決定
    - [x] `Queries.WithTx(tx *sql.Tx)` による再バインディング方法を確認・ドキュメント化
    - [x] ユニットテスト時の MockDB 制限（Tx をモックできない）を認識→統合テストで Tx 検証するアプローチを決定
    - [x] Service 層導入か handler 内ローカル処理か、チームで統一パターンを決定
    - [x] 決定内容を `Transaction Pattern.md` にドキュメント（チケット 3-5 の実装ガイドラインとして利用）
  - 受け入れ条件:
    - [x] Transaction Pattern.md が作成され、handler 内 Tx 処理のサンプルコードを掲載
    - [x] チケット 3-5 の実装者が参照できるレベルの詳細度
  - ファイル影響: `doc/planning/Transaction-Pattern.md` (新規)
  - 完了日: 2026-03-11
  - 参照: `doc/planning/Transaction-Pattern.md`
  - コミット例: `docs(design): add transaction handler pattern documentation`

#### フェーズ 3: ハンドラー層実装 - TDD（予定: 3/17）

- [ ] **チケット 3**: CreateOrderHandler 実装 (P0, Effort: High)
  - タイプ: TDD サイクル（Red → Green → Refactor）
  - ステップ 1: テスト設計
    - [ ] テストケースリスト作成（正常系、異常系、エッジケース）
    - [ ] MockDB の準備確認
  - ステップ 2: テストコード作成（`handler/order_test.go`）
    - [ ] テストケース実装（テーブル駆動）
      - 正常系: 複数商品の注文作成成功
      - 異常系: 認証なし、商品不在、在庫不足、DB エラー
      - 副作用: 在庫正確にデクリメント、合計金額計算
    - [ ] テスト実行でサイクル確認（Red）
  - ステップ 3: プロダクトコード実装（`handler/order.go`）
    - [ ] トランザクション開始
    - [ ] 各商品を FOR UPDATE で取得＆ロック
    - [ ] 在庫チェック → 不足時は 409 Conflict（ロールバック）
    - [ ] デクリメント＆ orders, order_items 作成
    - [ ] コミット後にレスポンス返却（201 Created）
    - [ ] テスト全 PASS（Green）
  - ステップ 4: リファクタリング（Refactor）
    - [ ] エラーハンドリング改善
    - [ ] トランザクション処理の可読性向上
  - 受け入れ条件:
    - [ ] ユニットテスト全 PASS、カバレッジ > 80%
    - [ ] 在庫がトランザクション内で正確にデクリメント
    - [ ] 在庫不足時にロールバック・409 返却
  - ファイル影響: `backend/handler/order.go` (新規), `backend/handler/order_test.go` (新規)
  - コミット例: `feat(handler): add CreateOrderHandler with TDD (tests + implementation)`

- [ ] **チケット 4**: CancelOrderHandler 実装 (P0, Effort: High)
  - タイプ: TDD サイクル（Red → Green → Refactor）
  - ステップ 1: テスト設計
    - [ ] テストケースリスト作成
  - ステップ 2: テストコード作成（`handler/order_test.go` に追加）
    - [ ] テーブル駆動テスト
      - 正常系: ステータス pending → cancelled、在庫巻き戻し
      - 異常系: 注文不在(404)、非所有(404)、既にキャンセル済み(400)
      - 副作用: 各商品の在庫が正確にインクリメント
    - [ ] テスト実行でサイクル確認（Red）
  - ステップ 3: プロダクトコード実装（`handler/order.go` に追加）
    - [ ] authorization チェック（自分の注文のみ）
    - [ ] トランザクション開始 & orders を FOR UPDATE 取得
    - [ ] ステータスチェック（pending のみキャンセル可）
    - [ ] order_items を全取得 → 各商品インクリメント
    - [ ] orders.status = 'cancelled', cancelled_at = NOW()
    - [ ] コミット＆レスポンス（200 OK）
    - [ ] テスト全 PASS（Green）
  - ステップ 4: リファクタリング（Refactor）
  - 受け入れ条件:
    - [ ] ユニットテスト全 PASS
    - [ ] キャンセル後の在庫が統合テストで確認可能
  - コミット例: `feat(handler): add CancelOrderHandler with authorization and rollback`

- [ ] **チケット 5**: GetOrdersHandler 実装 (P0, Effort: Med)
  - タイプ: 標準実装（既存パターン参考）
  - 内容:
    - [ ] ユーザーの注文一覧を取得（自分の注文のみ）
    - [ ] ステータスフィルター機能（?status=pending など）
    - [ ] order_items と紐付けて返却
  - テストケース: 正常系、認証なし、空一覧、フィルター検証
  - ファイル影響: `backend/handler/order.go`
  - コミット例: `feat(handler): add GetOrdersHandler with status filtering`

#### フェーズ 4: 検証テスト（予定: 3/20）

- [ ] **チケット 6**: 同時性テスト実装 (P0, Effort: High)
  - ファイル: `backend/tests/order_concurrency_test.go` (新規)
  - 内容:
    - [ ] 複数 goroutine で同じ商品に対して並列 CreateOrder
    - [ ] 在庫が正確にデクリメント（オーバーソールなし）
    - [ ] テスト シナリオ:
      - 在庫 10 個の商品を、5 ユーザーが各 3 個ずつ同時注文 → 最後の 2 個は 409 Conflict
      - 並列度: N=50 程度の大量リクエスト確認
  - 受け入れ条件:
    - [ ] テスト実行で 100% 成功（キャンセル時には確認ポイント多し）
    - [ ] 結果ログで在庫が正確に管理されていることを確認
  - コミット例: `test(order): add concurrency tests for overbooking prevention`

- [ ] **チケット 7**: エラーハンドリング・エッジケース (P1, Effort: Low)
  - 内容:
    - [ ] HTTP ステータス確認（400/401/404/409/500）
    - [ ] エラーレスポンス形式確認（error, message, details）
    - [ ] トランザクション失敗時の振る舞い
    - [ ] テストケース: 不正リクエスト、DB エラーシミュレーション
  - ファイル影響: `backend/handler/order_test.go`, `backend/pkg/respond/`（必要に応じて）
  - コミット例: `test(handler): add error handling and edge case tests for orders`

#### フェーズ 5: 統合・ルーティング（予定: 3/21）

- [ ] **チケット 8**: ルーティング登録 + 最終確認 (P0)
  - 前提: すべてのハンドラー実装 + テスト全 PASS
  - 内容:
    - [ ] `backend/routes/routes.go` に 3 エンドポイント登録
      - `GET /api/orders` — GetOrdersHandler
      - `POST /api/orders` — CreateOrderHandler
      - `POST /api/orders/:id/cancel` — CancelOrderHandler
    - [ ] RequireAuth ミドルウェア装着（認証必須）
    - [ ] `backend/test.http` に実行例追記
      - ログイン → 注文作成 → 注文一覧 → キャンセル の一連フロー
    - [ ] 手動テスト実行確認（API 全体連携確認）
  - ファイル影響: `backend/routes/routes.go`, `backend/test.http`
  - コミット例: `feat(routes): register order endpoints with RequireAuth`

- [ ] **チケット 9**: ドキュメント更新 (P1)
  - 内容:
    - [ ] `doc/api.md` に注文 API セクション追記（エンドポイント、リクエスト/レスポンス例）
    - [ ] `backend/test.http` の実行例説明を追記
    - [ ] `doc/task.md` に本タスク群の完了マーク
  - ファイル影響: `doc/api.md`, `doc/task.md`
  - コミット例: `docs(api): add order endpoints documentation`

### 受け入れ基準（全体）
- [ ] マイグレーション v8, v9, v10 が実行・リバート可能
- [ ] 3 ハンドラー（Get/Create/Cancel）が GET /api/products と同レベルの品質でテストカバー
- [ ] `handler/order_test.go` で正常系・異常系・エッジケースが全網羅
- [ ] 同時性テストで 50+ 並列リクエストでのオーバーソール防止を確認
- [ ] 手動テスト（test.http）で既存 API との連携確認

### タイムライン
| フェーズ | 期限 | 工数 | 成果物 |
|---------|-----|------|--------|
| 0: 設計 | 3/8（完了） | 1日 | 要件定義・API設計ドキュメント ✅ |
| 1: マイグレーション | 3/10 | 1日 | マイグレーション 3 個 |
| 2: sqlc クエリ | 3/12 | 1.5日 | query.sql 拡張・querier.go code gen |
| 3: ハンドラー層 | 3/17 | 2.5日 | 3 ハンドラー + テスト（TDD） |
| 4: 検証テスト | 3/20 | 1.5日 | 同時性テスト + エラーハンドリング |
| 5: 統合 | 3/21 | 1日 | ルーティング + ドキュメント |
| **合計** | **3/21** | **~9日** | **注文・在庫システム完成** |

### 詳細設計書
- 詳細計画（テストケース、スキーマ詳細、トランザクションフロー）: [doc/planning/orders-design-2026-03-08.md](planning/orders-design-2026-03-08.md)

### 推奨開発スタイル
- **TDD 推奨**: ハンドラー層（チケット 3-5）ではテストファイルを先に作成し、プロダクトコードを実装
- **メンター支援**: TDD メンター agent を活用して、テスト設計 → テストコード → 実装 のサイクルをガイド

---


