# 認証方式移行計画: localStorage トークン → HTTP-only Cookie

## 目的
- フロントが `localStorage` に保存する JWT を廃止し、サーバ発行の HttpOnly Cookie を利用することでトークン盗難（XSS）のリスクを低減する。

## 前提
- 現状フロー: フロントが `auth_token` を `localStorage` に保存し、API 呼び出しで `Authorization: Bearer <token>` を付与している。
- サーバは JWT（HS256）を発行している（`backend/auth/jwt.go`）。
- 本計画はドメインが同一、もしくは API とフロントが分離される場合も想定している（CORS 設定が必要）。

## 設計決定（提案）
- Cookie 名: `access_token`
- Cookie 属性: `HttpOnly; Secure; SameSite=Lax; Path=/; Max-Age=86400`（アクセストークン24時間を想定）
- リフレッシュ戦略: 将来的に `refresh_token` を HttpOnly Cookie にて運用し、`/api/refresh` を設ける（この移行ではオプション）。
- CSRF 対策: `SameSite=Lax` に加え、重要な変更系エンドポイントでは CSRF トークン（Double Submit Cookie または同期トークン）を導入することを推奨。

## バックエンドでの変更点
1. ログイン処理 (`LoginUserHandler`)
   - 現状: レスポンスボディに `{ token, user }` を返している。
   - 変更: レスポンスに加え `Set-Cookie: access_token=<jwt>; HttpOnly; Secure; SameSite=Lax; Path=/; Max-Age=...` を付与。必要に応じ body に user を返す。

2. ミドルウェア/ハンドラの認証
   - 現状: `Authorization` ヘッダからトークンを読み検証している。
   - 変更: `Authorization` が無い場合は Cookie（`access_token`）を優先して読み取るロジックを追加。

3. ログアウト
   - `Set-Cookie` で `access_token` の Max-Age=0 または Expires 過去日を返してクッキーを消去するエンドポイントを用意。

4. 環境変数・秘密管理
   - `JWT_SECRET` をソースにハードコードしない。`os.Getenv("JWT_SECRET")` 等で管理。

5. CORS / クレデンシャル
   - フロントとAPIが別オリジンの場合は `Access-Control-Allow-Credentials: true` を許可し、`Access-Control-Allow-Origin` はワイルドカード不可で明示的オリジンを設定。

## フロントエンドでの変更点
1. API 呼び出し共通化 (`frontend/lib/api.ts`)
   - `Authorization` ヘッダ付与を廃止（またはヘッダ付与はオプション化）。
   - fetch の `credentials` をエンドポイントに応じて `include`（クロスオリジン）または `same-origin` に設定。

2. 認証ストア (`frontend/store/useAuthStore.ts`)
   - `auth_token` の localStorage 保存／読み書きを削除する。
   - アプリ起動時のユーザ復元は `GET /api/me` を `credentials: 'include'` で呼び、成功時に user を store にセットする。

3. ログインページ (`frontend/app/login/page.tsx`)
   - ログイン成功時はサーバがクッキーをセットすることを前提に、返却された user 情報を store にセットしてリダイレクトする。必要に応じ `/api/me` で確認。

4. ログアウト処理
   - フロントは `POST /api/logout` を呼ぶのみでよく、サーバ側がクッキー消去を行う。

5. E2E / 自動テスト
   - テスト実行時にクッキーを正しく扱う設定が必要（`credentials` 設定、テストランナーの cookie サポート）。

## CSRF 対策（推奨手順）
- 短期（速やかに実施）: `SameSite=Lax` を設定し、危険なGET以外のリクエストで確認。POST/PUT/DELETE の API は適切な認可チェックを強化。
- 中長期（推奨）: Double Submit Cookie パターンか、同期トークンパターンを導入。
  - 例: サーバは `csrf_token` を通常の（HttpOnly ではない）Cookie に入れ、クライアントはヘッダ `X-CSRF-Token` に同値を付与して送信しサーバで比較する。

## マイグレーション手順（ステップ）
1. 設計確定: Cookie 名・属性、CSRF 方針、リフレッシュ戦略、環境変数名を決定する。
2. バックエンド実装（ステージングブランチ）: `LoginUserHandler` の Set-Cookie、ミドルウェアの Cookie 読取、logout を実装。`JWT_SECRET` を env 参照に。
3. フロント実装（同時ブランチ）: `lib/api.ts` と `useAuthStore` を修正し、`login` 呼び出し後の store 更新を行う。localStorage トークンは移行時に削除するロジックを追加。
4. 移行用互換レイヤ（暫定）: 既存 localStorage トークンが残るクライアント向けに、起動時 `auth_token` があれば安全な移行エンドポイントへ一時送信してサーバで cookie 発行 → localStorage 削除するスクリプトを用意する（注意: 送信は HTTPS 上で行う）。
5. テスト: ユニット・統合・E2E を作成/更新。
6. ステージングで検証: ログイン/ログアウト/保護ルート/E2E を実行。
7. 本番ロールアウト: フロントとバックエンドを同時にデプロイ（互換期間を短くする）。

## テストケース（抜粋）
- 正常: ログインで `Set-Cookie` が返り、`/api/me` が cookie により認証される。
- 異常: Cookie が無い/無効な場合、`/api/me` が 401 を返す。
- CSRF: Double Submit による検証が期待通りに失敗/成功するか。
- 移行: localStorage トークンが存在するクライアントで移行フローが期待通り cookie を発行し localStorage を削除するか。

## 必要なドキュメント追記
- `doc/api.md` に `Set-Cookie` の仕様（cookie 名・属性）と logout の挙動を明記。

## 開発時の注意点
- 開発ローカル (HTTP) では `Secure` が効かないため、local dev 用の設定やローカル証明書の導入を検討する。
- CORS 設定で `Allow-Credentials` を有効にする場合、`Allow-Origin` は特定オリジンに限定すること。

---
### 付録: 実装例スニペット

Go (Login handler の一部、擬似コード):

```go
cookie := &http.Cookie{
    Name:     "access_token",
    Value:    tokenString,
    Path:     "/",
    HttpOnly: true,
    Secure:   true, // 本番
    SameSite: http.SameSiteLaxMode,
    MaxAge:   86400,
}
http.SetCookie(c.Writer, cookie)
// body に user を返す
c.JSON(http.StatusOK, gin.H{"user": user})
```

TypeScript (fetch 共通関数の例、擬似コード):

```ts
export async function apiFetch(path: string, opts: RequestInit = {}) {
  const res = await fetch(API_URL + path, {
    credentials: 'include',
    headers: { 'Content-Type': 'application/json', ...(opts.headers || {}) },
    ...opts,
  })
  return res.json()
}
```

フロント起動時のユーザ復元（擬似コード）:

```ts
async function loadUser() {
  try {
    const data = await apiFetch('/api/me')
    setUser(data.user)
  } catch (e) {
    setUser(null)
  }
}
```

---
作成日: 2026-02-16
