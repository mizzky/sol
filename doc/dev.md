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
├── auth             # 認証関連
│   └── jwt.go
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

### sqlc 再生成の詳細手順

新しいクエリやカラムを追加した場合（例: reset_token カラム、GetUserByID クエリ）に実行します。

1. **query.sql にクエリを追加**
   ```sql
   -- name: GetUserByID :one
   SELECT * FROM users WHERE id = $1 LIMIT 1;
   ```

2. **sqlc generate を実行して Go コードを自動生成**
   ```bash
   cd backend
   sqlc generate
   ```

3. **生成物を確認**
   - `db/models.go` に新フィールド（例: `ResetToken *string`）が追加されたか確認
   - `db/query.sql.go` に新関数（例: `GetUserByID()`）が生成されたか確認
   - `db/querier.go` のインターフェースに新メソッドが追加されたか確認

4. **テストで動作確認**
   ```bash
   cd backend
   go test ./db -v
   go test ./handler -v
   ```

5. **全テストが通ることを確認**
   ```bash
   go test ./...
   ```

6. **生成物をコミット** — models.go, query.sql.go, querier.go は自動生成だが git 管理対象です

動作確認をしたいとき

    main.go を実行 (go run main.go)。

    VS Codeで test.http を開き、各リクエストの Send Request をクリック。