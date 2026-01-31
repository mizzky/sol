# CoffeeSys Backend API 仕様書 (会員基盤)

## 1. システム構成
- **言語:** Go (Golang)
- **Webフレームワーク:** [Gin](https://github.com/gin-gonic/gin)
- **データベース:** PostgreSQL
- **DB操作:** [sqlc](https://sqlc.dev/) (型安全なSQL実行)
- **マイグレーション:** [golang-migrate](https://github.com/golang-migrate/migrate)
- **認証補助:** [bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt) (パスワードハッシュ化)

## 2. ディレクトリ構成
```text
backend/
├── db/              # データベース関連
│   ├── migrations/  # SQLマイグレーションファイル (.up.sql, .down.sql)
│   ├── query.sql    # sqlcの元となるSQLクエリ
│   ├── models.go    # sqlc生成：テーブル構造体
│   ├── db.go        # sqlc生成：共通DB処理
│   └── query.sql.go # sqlc生成：Goの関数群
├── handler/         # HTTPハンドラー (コントローラー)
│   └── user.go      # 会員登録APIのロジック（Ginバージョン）
├── main.go          # エントリーポイント (DB接続、Ginルーター設定)
├── sqlc.yaml        # sqlcの設定ファイル
└── test.http        # REST Client用テストファイル（疎通確認用）
```

## 3. データベース設計

### users テーブル
会員情報を管理します。メールアドレスには一意制約（UNIQUE）を設定し、重複登録を防止しています。

| カラム名 | 型 | 制約 | 説明 |
| :--- | :--- | :--- | :--- |
| `id` | SERIAL | PRIMARY KEY | 自動採番されるユーザーID |
| `name` | TEXT | NOT NULL | 表示名 |
| `email` | TEXT | NOT NULL, UNIQUE | ログイン用メールアドレス（重複不可） |
| `password_hash` | TEXT | NOT NULL | bcryptによりハッシュ化されたパスワード |
| `role` | TEXT | NOT NULL | ユーザー権限（デフォルト: member） |
| `created_at` | TIMESTAMP | DEFAULT NOW() | 登録日時 |

---

## 4. API エンドポイント詳細

### 会員登録 (ユーザー作成)
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

#### 実装上の注意点
- **一意制約違反の識別**: PostgreSQLのエラーコード `23505` (`unique_violation`) を `github.com/lib/pq` ライブラリを用いてキャッチし、重複エラーとしてユーザーに通知します。
- **セキュリティ**: 500エラーの際、DBの生のエラーメッセージをそのままクライアントに返さないよう、汎用的なメッセージにマスクしています。


## 5. 開発フロー (逆引きガイド)
テーブル構造を変更したいとき

    migrate create -ext sql -dir db/migrations -seq [名前] でファイル作成。

    .up.sql に CREATE TABLE 等を記述。

    migrate up でDBに反映。

    sqlc generate でGoの構造体を更新。

新しいSQLクエリを追加したいとき

    db/query.sql にSQLを追記（-- name: [関数名] :one 等のコメントを忘れずに）。

    sqlc generate を実行し、db/query.sql.go に関数が生成されたことを確認。

    handler/ 内のロジックからその関数を呼び出す。

動作確認をしたいとき

    main.go を実行 (go run main.go)。

    VS Codeで test.http を開き、各リクエストの Send Request をクリック。