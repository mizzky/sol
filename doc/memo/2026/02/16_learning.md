## 2026-02-16 学習ログ

### 概要
- フロントエンドのログイン周りの TDD サイクルを進めた。
- Jest を導入・設定し、ユニットテストを追加して全件パスさせた。
- 小さな `AuthLoader` コンポーネントを実装してレイアウトに組み込み、認証復元を行うようにした。
- CI (GitHub Actions) にフロント/バックエンド両方のテストジョブを追加した。
- バックエンドの JWT シークレット管理を環境変数 `JWT_SECRET` 読み取りに変更し、テストでの扱い方を整理した。

### 実施したこと（詳細）
1. フロント側テスト追加
   - `frontend/lib/__tests__/api.login.test.ts` を作成（`login()` の成功／失敗ケース）。
   - `frontend/lib/__tests__/useAuthStore.test.ts` を作成（`setToken`/`logout`/`loadFromStorage` の挙動確認）。
   - `frontend/lib/__tests__/page.test.tsx` を作成（`LoginPage` の送信・成功/失敗ハンドリング）。
   - テスト実行のために `jest` 設定と `jest.setup.ts` を修正し、`@testing-library/jest-dom` の読み込みと `localStorage` モックを追加。
   - `console.error` の出力を抑制するため、テスト内で `jest.spyOn(console,'error')` を使用。

2. AuthLoader の追加
   - `frontend/app/components/AuthLoader.tsx` を作成し、`useAuthStore.getState().loadFromStorage()` をマウント時に呼ぶように実装。
   - `frontend/app/layout.tsx` に `<AuthLoader />` を挿入。
   - 対応テスト `frontend/app/components/__tests__/AuthLoader.test.tsx` を追加。

3. CI 追加
   - `.github/workflows/ci.yml` を追加・更新し、backend の `gofmt`/`go test`、frontend の `npm ci && npm test` を実行するジョブを作成。

4. JWT シークレットの env 化
   - `backend/auth/jwt.go` を修正し、`JWT_SECRET` 環境変数からシークレットを読み取るようにした（フォールバックを残す）。
   - `init()` と `getJWTSecret()` の違い、テスト時の注意点（`init()` はテスト実行前に呼ばれるため `os.Setenv` が間に合わない）を確認。
   - テストでの対処法を決定：
     - CI/実行時に環境変数を渡す（`JWT_SECRET=test-secret go test ./...`）
     - またはテスト用に setter / TestMain で `jwtSecret` を上書きする方法を提示。

### 実行コマンド（参考）
- フロントのテスト実行:
```bash
cd frontend
npm test
```
- バックエンドのテスト実行（環境変数指定）:
```bash
JWT_SECRET="local-test-secret" cd backend && go test ./...
```
- devcontainer での永続設定例（`.devcontainer/devcontainer.json`）:
```json
"containerEnv": { "JWT_SECRET": "dev-only-secret" }
```

### 学んだこと / 気づき
- Jest の設定（setup ファイル）で `@testing-library/jest-dom` の import パスはバージョンによって異なるため注意。
- Node のテスト環境で `localStorage` が未定義なのでモックが必要。
- Go の `init()` はパッケージ初期化時に実行されるため、テスト内での `os.Setenv` は `init()` に影響しない（実行時渡しか明示的なオーバーライドが必要）。

### 次のアクション（提案）
1. 本番運用に向けて `JWT_SECRET` を CI/CD の Secret として登録する手順をドキュメント化する。
2. httpOnly cookie 方式への移行設計（CSRF 対策含む）を検討する。
3. E2E テスト（Playwright/Cypress）でログインフローをカバーする。

---
記録者: 作業者
日時: 2026-02-16
2026-02-16 学習ログ
=====================

概要
--
- `GET /api/me` に関する TDD を実践。

今日やったこと
--
- `backend/handler/me_test.go` を作成し、table-driven テストで以下を追加・確認した。
  - 正常系（有効トークン）→ 200 と `{"user": {...}}`
  - ヘッダ無し / フォーマット不正 / 無効トークン → 401
  - クレーム欠如・型不正（文字列で数値に変換できない）→ 401
  - DB の `sql.ErrNoRows`（ユーザー未検出）→ 401
  - DB のその他エラー → 500
  - 追加の境界ケース（`Claims` が `jwt.MapClaims` でないケース、`user.id` が bool 等の不正型）も追加して検証

- `backend/handler/me.go` に `MeHandler` を実装（トークン検証→`claims["user.id"]`→`GetUserForUpdate`→`{"user":{...}}` を返す）。

- テスト実行：`go test ./backend -run TestMeHandler` を実行し、テストはパス。

学び・決定事項
--
- トークンのクレーム `user.id` は `float64` / `string` の両方を扱う実装とし、`strconv.ParseInt(...,10,64)` を使うことで `int64` を直接得る方が安全（`Atoi` は環境依存の `int` を返すため非推奨）。
- レスポンスは `Login` と整合するよう `{"user": {...}}` 形式に統一。
- テストでは `auth.Validate` を差し替えてトークンの振る舞いをモックすると高速に境界条件を検証できる。

次にやること（提案）
--
- `backend/routes/routes.go` に `GET /api/me` を登録してフルルートで動作確認。
- フロント側の `useAuthStore.loadFromStorage` と統合してブラウザで未認証→ログイン遷移を確認。
- （任意）API ドキュメント `doc/api.md` に `/api/me` を追記。

備考
--
- 本日の変更は `doc/` 配下のログに記録済。
