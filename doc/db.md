# データベース設計 (Database Schema)

## 1. 概要
ユーザー管理およびコーヒー注文管理（予定）のためのデータベース。
- **DBエンジン**: PostgreSQL
- **管理ツール**: sqlc / golang-migrate

## 2. テーブル定義

### users テーブル
ユーザーの基本情報を管理。

| カラム名 | 型 | 制約 | 説明 |
| :--- | :--- | :--- | :--- |
| `id` | serial | PRIMARY KEY | ユーザー固有ID |
| `name` | varchar(255) | NOT NULL | 表示名 |
| `email` | varchar(255) | NOT NULL, UNIQUE | ログイン用メールアドレス |
| `password_hash` | text | NOT NULL | bcryptでハッシュ化したパスワード |
| `role` | varchar(50) | NOT NULL, DEFAULT 'member' | 権限（DB既存） |
| `status` | varchar(50) | NOT NULL, DEFAULT 'active' | アカウント状態（DB既存） |
| `reset_token` | varchar(255) | NULL | パスワードリセット用トークン（ハッシュ保存想定） |
| `created_at` | timestamp | DEFAULT NOW() | 登録日時 |
| `updated_at` | timestamp | DEFAULT NOW() | 更新日時 |

## 3. 関連図（ER図）
現在、単一テーブルですが、今後の拡張イメージ：
- `users` (1) --- (N) `orders`
- `orders` (N) --- (N) `coffee_beans`

## 4. セキュリティ上の配慮
- `password_hash`: 生パスワードは一切保存せず、必ず `bcrypt` でハッシュ化して保存する。
- `email`: 検索を高速化し、重複を防ぐために UNIQUE インデックスを貼る。