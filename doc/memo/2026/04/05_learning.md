# 学習記録 2026-04-05

## セッション 1 (14:18)

### 取り組んだタスク
- `LoginUserHandler` にリフレッシュトークンを追加。
  - 32バイトのランダム値を生成し hex エンコードした refresh token を作成。
  - refresh token を sha256 でハッシュ化して `q.CreateRefreshToken` で DB に保存（expires_at を 14日後設定）。
  - `access_token` と `refresh_token` を `HttpOnly` Cookie として発行。`access_token` は Path=/, SameSite=Lax, TTL=15分。`refresh_token` は Path=/api/refresh, SameSite=Strict, TTL=14日（MaxAge 秒指定）。
- `auth/jwt.go` の JWT `exp` を 15分に変更して access cookie と整合させた。
- テストの追加/修正:
  - `TestLoginUserHandler_SetsCookies` を作成・拡張。Cookie の存在、`HttpOnly`、`Path`、`MaxAge`、`SameSite`、refresh 値長（64）を検証。
  - `CreateRefreshToken` の呼び出し引数をキャプチャして、DB に保存された `TokenHash` が `sha256(refresh_cookie.Value)` と一致することを検証するようにした。
  - モックの不足（`CreateRefreshToken` モックが未設定で panic）に遭遇し、`mock.On("CreateRefreshToken", mock.Anything, mock.Anything).Return(...)` を追加して解決。
  - `testutil.MockDB` に `CreateRefreshToken` のメソッド実装を追加（テストで使用できるようにした）。
- `user.go` の `accessCookie` に `MaxAge` を追加、`refreshCookie` にも `MaxAge` を追加（秒で 1209600）して互換性を確保。

### ユーザーが質問した内容
- なし

### 躓いたポイントと解決策
- テスト実行時に `assert: mock: I don't know what to return because the method call was unexpected.` による panic（`CreateRefreshToken` がモックされていなかった）。
  - 対応: テスト内で `CreateRefreshToken` の expectation を追加し、呼び出しをスタブ化して戻り値を設定した。さらに引数のキャプチャと追加検証を実装した。
- `access_token` の TTL と JWT `exp` の不整合。
  - 対応: `auth/jwt.go` の `exp` を 15分に変更して整合させた。

### 次回課題
1. `/api/refresh` の Red テストを作成（リフレッシュ時のローテーション、DB更新、Cookie 更新、拒否ケース）。
2. `/api/refresh` を最小実装してテストを Green にする（新トークン発行＋旧トークン無効化）。
3. `/api/logout` の実装とテスト（Cookie を無効化し DB の該当 refresh token を revoke）。
4. CSRF 対策（double-submit cookie + `X-CSRF-Token` 検証）を導入し、ミドルウェアと関連テストを追加。

### 変更ファイル（編集・確認）
- backend/handler/user.go
- backend/auth/jwt.go
- backend/handler/testutil/mockdb.go
- backend/handler/user_test.go
