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
[
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
```

#### エラーレスポンス
| ステータスコード | 判定条件 | レスポンスボディ (JSON) |
| :--- | :--- | :--- |
| **500 Internal Server Error** | DB接続失敗など | `{"error": "予期せぬエラーが発生しました"}` |


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
[
  {
    "id": 1,
    "name": "コーヒー豆",
    "description": "各種コーヒー豆を取り扱います",
    "created_at": "2026-02-02T10:00:00Z",
    "updated_at": "2026-02-02T10:00:00Z"
  }
]
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
