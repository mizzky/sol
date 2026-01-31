# API エンドポイント詳細

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
