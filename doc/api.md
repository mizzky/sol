## APIのnull値の取り扱いに関する仕様

APIでは、リクエストおよびレスポンスにおけるnull値の取り扱いを以下のように統一しています。

### リクエストにおけるnull値
- クライアントから送信されるJSONリクエストボディ内で、任意のフィールドがnullである場合、そのフィールドは省略可能です。
- 必須フィールドがnullで送信された場合、サーバーは`400 Bad Request`を返します。

### レスポンスにおけるnull値
- サーバーから返されるJSONレスポンスボディ内で、値が存在しない場合は、該当フィールドをnullとして明示的に返します。
- 例:
  ```json
  {
    "id": 1,
    "name": "サンプル",
    "description": null
  }
  ```

### エラーレスポンスの例
| ステータスコード | 判定条件 | レスポンスボディ (JSON) |
| :--- | :--- | :--- |
| **400 Bad Request** | 必須フィールドがnull | `{"error": "必須フィールドがnullです"}` |
| **500 Internal Server Error** | サーバー内部エラー | `{"error": "予期せぬエラーが発生しました"}` |

この仕様により、クライアントとサーバー間のデータの一貫性を確保します。


## API エンドポイント詳細

## 認証・認可要件

本APIでは以下の方針で認証・認可を扱います。

- **認証方式**: JWT を用いた Bearer トークン（`Authorization: Bearer <token>`）を前提とします。
- **認可方式**: トークンのクレーム内 `user.id` を基に DB を参照し、`role` が `admin` の場合に管理操作を許可します。
- **ステータスマッピング**:
  - トークン無し/不正/クレーム欠落: `401 Unauthorized`（`{"error":"認証が必要です"}`）
  - ログイン済だが管理者でない: `403 Forbidden`（`{"error":"管理者権限が必要です"}`）
  - DB の取得エラー等のサーバー側障害: `500 Internal Server Error`（`{"error":"予期せぬエラーが発生しました"}`）

**管理者権限が必要なエンドポイント（抜粋）**

- `POST /api/products` — 商品登録（管理者のみ）
- `PUT /api/products/:id` — 商品更新（管理者のみ）
- `DELETE /api/products/:id` — 商品削除（管理者のみ）
- `POST /api/categories` — カテゴリ作成（管理者のみ）
- `PUT /api/categories/:id` — カテゴリ更新（管理者のみ）
- `DELETE /api/categories/:id` — カテゴリ削除（管理者のみ）

上記以外の読み取り系エンドポイント（例: `GET /api/products`, `GET /api/categories`）は認証不要です。


## 会員登録 (ユーザー作成)
リクエストに含まれるパスワードをハッシュ化し、安全にDBへ保存します。

- **メソッド/パス:** `POST /api/register`
- **リクエスト形式:** `application/json`

#### リクエストボディ
```json
{
    "name": "田中 太郎",
    "email": "tanaka@example.com",
    "password": "password123"
}
```

#### 成功レスポンス
```json
{
    "id": 1,
    "name": "田中 太郎",
    "email": "tanaka@example.com",
    "role": "member",
    "created_at": "2026-01-31T16:00:00Z"
}
```
#### エラーハンドリグ
会員登録時、以下の条件に応じて適切なステータスコードとメッセージを返却します。

| ステータスコード | 判定条件 | レスポンスボディ (JSON) |
| :--- | :--- | :--- |
| **400 Bad Request** | JSONの構文エラー、または必須項目漏れ | `{"error": "リクエスト形式が正しくありません"}` |
| **400 Bad Request** | メールアドレスが既にDBに存在する場合 (PQ Error `23505`) | `{"error": "このメールアドレスは既に登録されています"}` |
| **500 Internal Server Error** | パスワードのハッシュ化失敗、DB接続断など | `{"error": "予期せぬエラーが発生しました"}` |


## ログイン
登録済みのメールアドレスとパスワードで認証を行い、以降のリクエストに必要な JWT（アクセストークン）を発行します。

- **メソッド/パス:** `POST /api/login`
- **リクエスト形式:** `application/json`
#### リクエストボディ
```json
{
    "email":"tanaka@example.com",
    "password":"password"
}
```

#### 成功レスポンス
`200 STATUS OK`:

`token`: JWT文字列

`user`: 基本ユーザー情報
```json
{
  "message": "ログイン成功！",
  "token": "eyJhbGciOiJIUzI1Ni...",
  "user": {
    "id": 1,
    "name": "田中太郎",
    "email": "tanaka@example.com"
  }
}
```
#### エラーレスポンス
`400 Bad Request`: バリデーションエラー

`401 Unauthorized`: 認証失敗 *パスワードまたはメールアドレス相違

| ステータスコード | 判定条件 | レスポンスボディ (JSON) |
| :--- | :--- | :--- |
| **400 Bad Request** | バリデーションエラー | `{"error": "リクエストが正しくありません"}` |
| **401 Unauthorized Error** | 認証失敗 | `{"error": "メールアドレスまたはパスワードが正しくありません"}` |
| **500 Internal Server Error** | トークン生成失敗 | `{"error": "トークンの生成に失敗しました"}` |


## 商品一覧取得
登録されているすべての商品一覧を取得します。

- **メソッド/パス:** `GET /api/products`
- **認証:** 不要
- **レスポンス形式:** `application/json`

#### 成功レスポンス
`200 OK`:
```json
{
  "products": [
    {
      "id": 1,
      "name": "アラビカ豆",
      "price": 1500,
      "is_available": true,
      "category_id": 1,
      "sku": "COFFEE-001",
      "description": "高品質なアラビカ豆",
      "image_url": "https://example.com/image1.jpg",
      "stock_quantity": 100,
      "created_at": "2026-02-01T10:00:00Z",
      "updated_at": "2026-02-01T10:00:00Z"
    }
  ]
}
```

#### エラーレスポンス
| ステータスコード | 判定条件 | レスポンスボディ (JSON) |
| :--- | :--- | :--- |
| **500 Internal Server Error** | DB接続失敗など | `{"error": "予期せぬエラーが発生しました"}` |


## 商品（Products）API

商品関連の詳細なエンドポイントをまとめます。以下は `sku` が必須で、更新は `PUT` を採用、管理者のみが `POST` / `PUT` / `DELETE` を実行可能とする仕様です。

- GET /api/products/:id
  - 説明: 指定IDの単一商品を取得します。
  - 認証: 不要
  - レスポンス例 (200 OK):
    ```json
    {
      "id": 1,
      "name": "アラビカ豆",
      "price": 1500,
      "is_available": true,
      "category_id": 1,
      "sku": "COFFEE-001",
      "description": "高品質なアラビカ豆",
      "image_url": "https://example.com/image1.jpg",
      "stock_quantity": 100,
      "created_at": "2026-02-01T10:00:00Z",
      "updated_at": "2026-02-01T10:00:00Z"
    }
    ```
  - エラーレスポンス:
    - 404 Not Found: `{"error": "商品が見つかりません"}`
    - 400 Bad Request: `{"error": "IDが正しくありません"}`
    - 500 Internal Server Error: `{"error": "予期せぬエラーが発生しました"}`

- POST /api/products
  - 説明: 新しい商品を作成します（管理者のみ）。
  - 認証: JWT（管理者ロール）🔒
  - バリデーション:
    - `name` : 必須、非空、最大255文字
    - `price` : 必須、正の整数（> 0）
    - `category_id` : 必須、存在するカテゴリID
    - `sku` : 必須、一意（DBでユニーク制約）
    - `stock_quantity` : 任意、整数（省略時は0）
    - `image_url` : 任意、有効なURL形式
  - リクエスト例:
    ```json
    {
      "name": "アラビカ豆",
      "price": 1500,
      "category_id": 1,
      "sku": "COFFEE-002",
      "description": "新商品説明",
      "image_url": "https://example.com/image2.jpg",
      "stock_quantity": 50
    }
    ```
  - 成功レスポンス: `201 Created`（作成した商品オブジェクト）
  - エラーレスポンス:
    - 400 Bad Request: バリデーションエラー `{"error":"リクエスト形式が正しくありません"}`
    - 401 Unauthorized: `{"error":"認証が必要です"}`
    - 403 Forbidden: `{"error":"管理者権限が必要です"}`
    - 404 Not Found: `{"error":"カテゴリが見つかりません"}`
    - 409 Conflict: `{"error":"SKUが既に存在します"}`
    - 500 Internal Server Error: `{"error":"予期せぬエラーが発生しました"}`

- PUT /api/products/:id
  - 説明: 商品の全体更新を行います（管理者のみ）。`PUT` を採用し、全フィールドを送信することを前提とします。
  - 認証: JWT（管理者ロール）🔒
  - バリデーション:
    - `name` : 必須、非空、最大255文字
    - `price` : 必須、正の整数（> 0）
    - `category_id` : 必須、存在するカテゴリID
    - `sku` : 必須、一意（DBでユニーク制約）。重複時は `409 Conflict` を返す。
    - その他フィールドは型チェックおよび文字数制限を適用
  - リクエスト例（フル更新）:
    ```json
    {
      "name": "アラビカ豆",
      "price": 1600,
      "is_available": true,
      "category_id": 1,
      "sku": "COFFEE-002",
      "description": "更新後の説明",
      "image_url": "https://example.com/image2.jpg",
      "stock_quantity": 80
    }
    ```
  - 成功レスポンス: `200 OK`（更新後の商品オブジェクト）
  - エラーレスポンス:
    - 400 Bad Request: `{"error":"リクエストが正しくありません"}`
    - 401 Unauthorized / 403 Forbidden
    - 404 Not Found: `{"error":"商品が見つかりません"}`
    - 409 Conflict: `{"error":"SKUが既に存在します"}`
    - 500 Internal Server Error

- DELETE /api/products/:id
  - 説明: 商品を削除します（管理者のみ）。物理削除または論理削除のいずれかを実装可。
  - 認証: JWT（管理者ロール）🔒
  - 成功レスポンス:
    - 204 No Content
  - エラーレスポンス:
    - 401 / 403 / 404 / 500（上記と同様）

### 実装メモ
- ルーティング: `backend/routes/routes.go` にハンドラを追加
- DB クエリ: `backend/query.sql` に CRUD 用クエリを追加（sku 一意制約を確認）
- ハンドラ: `backend/handler/product.go` を作成/編集
- テスト: `backend/handler/product_test.go` を追加/更新（正常系・異常系）
- マイグレーション: 必要に応じて `backend/db/migrations/000004_alter_products_table.up.sql` を更新して `sku` カラムの制約やインデックスを追加

## 商品登録
新しい商品を登録します。（管理者のみ）

- **メソッド/パス:** `POST /api/products`
- **認証:** JWT（管理者ロール）🔒
- **リクエスト形式:** `application/json`

#### リクエストボディ
```json
{
  "name": "アラビカ豆",
  "price": 1500,
  "category_id": 1,
  "sku": "COFFEE-001",
  "description": "高品質なアラビカ豆",
  "image_url": "https://example.com/image1.jpg",
  "stock_quantity": 100
}
```

#### 成功レスポンス
`201 Created`:
```json
{
  "id": 1,
  "name": "アラビカ豆",
  "price": 1500,
  "is_available": true,
  "category_id": 1,
  "sku": "COFFEE-001",
  "description": "高品質なアラビカ豆",
  "image_url": "https://example.com/image1.jpg",
  "stock_quantity": 100,
  "created_at": "2026-02-02T10:00:00Z",
  "updated_at": "2026-02-02T10:00:00Z"
}
```

#### エラーレスポンス
| ステータスコード | 判定条件 | レスポンスボディ (JSON) |
| :--- | :--- | :--- |
| **400 Bad Request** | バリデーションエラー | `{"error": "リクエスト形式が正しくありません"}` |
| **401 Unauthorized** | 認証失敗 | `{"error": "認証が必要です"}` |
| **403 Forbidden** | 管理者権限がない | `{"error": "管理者権限が必要です"}` |
| **500 Internal Server Error** | DB接続失敗など | `{"error": "予期せぬエラーが発生しました"}` |


## カテゴリ一覧取得
登録されているすべてのカテゴリ一覧を取得します。

- **メソッド/パス:** `GET /api/categories`
- **認証:** 不要
- **レスポンス形式:** `application/json`

#### 成功レスポンス
`200 OK`:
```json
{
  "categories": [
    {
      "id": 1,
      "name": "コーヒー豆",
      "description": "各種コーヒー豆を取り扱います",
      "created_at": "2026-02-02T10:00:00Z",
      "updated_at": "2026-02-02T10:00:00Z"
    }
  ]
}
```

#### エラーレスポンス
| ステータスコード | 判定条件 | レスポンスボディ (JSON) |
| :--- | :--- | :--- |
| **500 Internal Server Error** | DB接続失敗など | `{"error": "予期せぬエラーが発生しました"}` |


## カテゴリ作成
新しいカテゴリを登録します。（管理者のみ）

- **メソッド/パス:** `POST /api/categories`
- **認証:** JWT（管理者ロール）🔒
- **リクエスト形式:** `application/json`

#### リクエストボディ
```json
{
  "name": "コーヒー豆",
  "description": "各種コーヒー豆を取り扱います"
}
```

#### 成功レスポンス
`201 Created`:
```json
{
  "id": 1,
  "name": "コーヒー豆",
  "description": "各種コーヒー豆を取り扱います",
  "created_at": "2026-02-02T10:00:00Z",
  "updated_at": "2026-02-02T10:00:00Z"
}
```

#### エラーレスポンス
| ステータスコード | 判定条件 | レスポンスボディ (JSON) |
| :--- | :--- | :--- |
| **400 Bad Request** | バリデーションエラー（name 空など） | `{"error": "カテゴリ名は必須です"}` |
| **400 Bad Request** | name が既に存在 | `{"error": "このカテゴリ名は既に存在します"}` |
| **401 Unauthorized** | 認証失敗 | `{"error": "認証が必要です"}` |
| **403 Forbidden** | 管理者権限がない | `{"error": "管理者権限が必要です"}` |
| **500 Internal Server Error** | DB接続失敗など | `{"error": "予期せぬエラーが発生しました"}` |


## カテゴリ更新
既存のカテゴリを更新します。（管理者のみ）

- **メソッド/パス:** `PUT /api/categories/:id`
- **認証:** JWT（管理者ロール）🔒
- **リクエスト形式:** `application/json`

#### リクエストボディ
```json
{
  "name": "プレミアムコーヒー豆",
  "description": "高級コーヒー豆の取り扱い"
}
```

#### 成功レスポンス
`200 OK`:
```json
{
  "id": 1,
  "name": "プレミアムコーヒー豆",
  "description": "高級コーヒー豆の取り扱い",
  "created_at": "2026-02-02T10:00:00Z",
  "updated_at": "2026-02-02T14:30:00Z"
}
```

#### エラーレスポンス
| ステータスコード | 判定条件 | レスポンスボディ (JSON) |
| :--- | :--- | :--- |
| **400 Bad Request** | JSON構文エラー | `{"error": "リクエスト形式が正しくありません"}` |
| **400 Bad Request** | `id` が数値でない場合 | `{"error": "IDが正しくありません"}` |
| **400 Bad Request** | `name` が `null` または空文字 | `{"error": "カテゴリ名は必須です"}` |
| **401 Unauthorized** | 認証失敗 | `{"error": "認証が必要です"}` |
| **403 Forbidden** | 管理者権限がない | `{"error": "管理者権限が必要です"}` |
| **404 Not Found** | 指定されたカテゴリが存在しない場合 | `{"error": "カテゴリが見つかりません"}` |
| **500 Internal Server Error** | DB接続失敗など | `{"error": "予期せぬエラーが発生しました"}` |


## カテゴリ削除
既存のカテゴリを削除します。（管理者のみ）

- **メソッド/パス:** `DELETE /api/categories/:id`
- **認証:** JWT（管理者ロール）🔒

#### 成功レスポンス
`204 No Content`: ボディなし

#### エラーレスポンス
| ステータスコード | 判定条件 | レスポンスボディ (JSON) |
| :--- | :--- | :--- |
| **401 Unauthorized** | 認証失敗 | `{"error": "認証が必要です"}` |
| **403 Forbidden** | 管理者権限がない | `{"error": "管理者権限が必要です"}` |
| **404 Not Found** | カテゴリが存在しない | `{"error": "カテゴリが見つかりません"}` |
| **500 Internal Server Error** | DB接続失敗など | `{"error": "予期せぬエラーが発生しました"}` |
