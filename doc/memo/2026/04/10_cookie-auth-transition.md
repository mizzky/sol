# Cookie認証移行（ADR補足）

このファイルは ADR の補足として作成します。過去の設計は [doc/task.md](doc/task.md) を ADR として残す運用とします。

## 現在の方針
- Cookie-only: `access_token` と `refresh_token` を HttpOnly（かつ Secure）Cookie で管理する。
- API の JSON レスポンスにトークンを含めず、ブラウザ側はクッキーで認証情報を保持する。

## 現状の進捗
- ミドルウェア: 更新済み（Cookie ベースの検証/注入に対応）。
- OpenAPI: 更新済み（認証仕様を Cookie 前提に変更）。
- Login / Refresh ハンドラ: 未処理（レスポンスを Cookie に書き換える対応が残る）。

## 注意点
- `refresh` Cookie の `Path` を適切に限定する（例: `/api/auth/refresh`）。
- CSRF 対策を検討する（SameSite のみで不十分な場合は double-submit や CSRF トークンを併用）。
- フロントは全ての認証 API 呼び出しで `credentials: 'include'` を必須にする。

## 次にやること
1. backend: Login/Refresh ハンドラを Cookie 書込へ変更する。
2. テスト: unit/integration の期待値・セットアップを更新する。
3. frontend: API 呼び出しへ `credentials: 'include'` を適用し、必要に応じて CSRF 実装を追加する。
