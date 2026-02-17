# ユーザー管理 — チケット詳細（作成日: 2026-02-16）

目的
- ユーザ登録・認証・ロール管理・トークン管理を堅牢にし、フロントと安全に連携する。

受け入れ基準
- 管理者がAPIでユーザロールを変更できる。  
- サーバ側バリデーションが強化され、不正入力で400を返す。  
- RefreshToken は将来のために設計ドキュメントが存在する（実装はP2）。

チケット一覧

1) ユーザテーブル拡張案作成 — Effort: Low, Priority: P0
   - 内容: `users` テーブルに `role VARCHAR` と `reset_token VARCHAR NULL` を追加するマイグレーション草案を作成。  
   - 影響ファイル: backend/db/migrations, backend/db/models.go
   - 受け入れ条件: SQL草案と ER 図（簡易）がdocにあること。

   - 補足（チケット1の設計メモ）
      - 既に users に role/status が存在するため、P0では reset_token のみ追加を前提とする。

      SQL草案（P0・設計）
      ALTER TABLE users
        ADD COLUMN reset_token VARCHAR(255) NULL;

      - reset_token はハッシュ保存を想定（デファクト準拠）。
      - 有効期限は仕様として定義（例: 30分〜1時間）。DBカラム追加はP1で検討。

      簡易ER図（P0時点）
      users (1) --- (N) orders

2) ロール管理API: `SetUserRoleHandler` 実装 — Effort: Low, Priority: P0
   - 内容: 管理者のみが `/admin/users/:id/role` にPATCHできるエンドポイントを実装。リクエスト: { role: "user"|"admin" }。  
   - 影響ファイル: backend/handler/user.go, backend/routes/routes.go, backend/auth/middleware.go
   - 受け入れ条件: 管理者トークンで変更成功、非管理者は403になる単体テストがあること。

3) サーバ側入力バリデーション共通化 — Effort: Low, Priority: P0
   - 内容: Register/Login リクエストのバリデーション（メール形式、パスワード最小長、禁止文字）を共通ユーティリティにする。  
   - 影響ファイル: backend/handler/user.go, backend/pkg/validation (新規)
   - 受け入れ条件: テーブル駆動テストで正常/異常ケースを網羅。

4) ユーザ関連の sqlc クエリとテスト整備 — Effort: Low, Priority: P0
   - 内容: `GetUserByID`, `UpdateUserRole`, `SetResetToken` 等のsqlcクエリ追加。  
   - 影響ファイル: backend/query.sql, backend/db/querier.go
   - 受け入れ条件: sqlc 再生成手順を doc に記載。

5) パスワードリセット（設計→実装候補） — Effort: Med, Priority: P1
   - 内容: リセット要求→メール（トークン）→トークン検証→パスワード更新 のAPIとメールテンプレ設計。  
   - 影響ファイル: backend/handler/user.go, doc/planning, backend/db/migrations
   - 受け入れ条件: 設計ドキュメントとユニットテスト（token検証のモック）。

6) RefreshToken 方針設計（ドキュメントのみ） — Effort: Med, Priority: P2
   - 内容: refresh token を DB に保存するか、cookie にするか、寿命、盗難対策をドキュメント化。  
   - 影響ファイル: backend/auth/jwt.go, doc/planning
   - 受け入れ条件: 選択肢と移行手順が明記されていること。

7) ログイン試行制限 / レートリミット設計 — Effort: Med, Priority: P2
   - 内容: IP/アカウント単位の失敗回数カウントと一時ロックの設計（ミドルウェア案）。  
   - 影響ファイル: backend/auth/middleware.go (新規), doc/planning
   - 受け入れ条件: 設計文書と単体テストのテストケース案。

8) フロント統合タスク — Effort: Low, Priority: P0
   - 内容: `useAuthStore` のトークン取扱い（login/logout/refresh の呼び出し）をAPI定義に合わせる。  
   - 影響ファイル: frontend/store/useAuthStore.ts, frontend/lib/api.ts
   - 受け入れ条件: フロント側の簡易結合テスト（Jest）を追加。

9) 監査ログ（ユーザロール変更） — Effort: Low, Priority: P1
   - 内容: 管理者によるロール変更を `audit_logs` テーブルに記録する設計。  
   - 影響ファイル: backend/db/migrations, backend/handler/user.go
   - 受け入れ条件: ロール変更時にログ作成のユニットテストがあること。

実施順の推奨
1→3→4→2→8→5→9→6→7
