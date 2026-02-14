# 2026-02-14 学習ログ

本日行ったこと（要点）

- `product` 関連の TDD 作業を実施。
  - テストを先行で追加：SKU 重複時の `409 Conflict`、管理者認可（`AdminOnly`）の統合テスト、`name` の文字数オーバー検証など。
  - モックで Postgres の一意制約エラーを再現するために `*pq.Error{Code: "23505"}` を返すようにした。
- ハンドラ実装の修正：
  - `backend/handler/product.go` にバリデーション（`name` 最大255文字、`price > 0` 等）を追加。
  - `CreateProduct` / `UpdateProduct` の DB エラー処理で `pq.Error` を `errors.As` で判定し、`23505` の場合は `409 Conflict` を返すようにした。
- ドキュメント更新：
  - `doc/api.md` を更新して `PATCH` から `PUT` に方針を統一（PUTはフル更新）、`sku` 重複時の `409`、`name` 上限255、`price > 0` を明記。
  - `doc/task.md` に作業完了の追記を行った。
- テスト実行：
  - 単体テストを実行し、追加したテストを含むハンドラ周りのテストが通ったことを確認（`go test ./...` 実行済）。

学び・注意点

- 単体テストで DB エラーを検出する場合、単に `errors.New("pq: ...")` を返すとハンドラ側の `errors.As(err, &pq.Error)` にマッチしない。実装とテストでエラー型を一致させることが重要。
- JWT を使ったミドルウェアでは、`Authorization` ヘッダのフォーマット（`Bearer <token>` の空白）やトークンの `user.id` クレームの型（float64/string）に配慮する必要がある。
- API 仕様と実装の不一致（PATCH vs PUT、部分更新 vs フル更新）は混乱を招くので、ドキュメントに明示して一貫させるべき。

次のアクション

- CI 向けに統合テストの DB 初期化手順（マイグレーションの適用）を整備する。
- `doc/api.md` の変更をリリースノートに反映するか検討する。

記録者: 作業ペア
作成日: 2026-02-14


# 2026-02-14 学習ログ（追記）

追記: フロントエンド認証・連携作業 (2026-02-14)

- フロント側での準備:
  - `frontend/lib/api.ts` に `API_URL` のエクスポートと API ユーティリティ（`getProducts` / `login` / `createProduct`）を整備。
  - `frontend/store/useAuthStore.ts` を追加し、`token`/`user` の zustand ストア、localStorage 永続化、`loadFromStorage` に `/api/me` を試す復元ロジックを組み込んだ。
  - `frontend/app/page.tsx` の fetch ハンドリングを型安全に修正（`unknown` の利用、`useCallback` 化）。

- バックエンド対応・デバッグ:
  - CORS エラー（ブラウザの `Failed to fetch`）を調査。原因は Gin のミドルウェア登録順がルート定義の後になっていて CORS ヘッダが付与されていなかったこと。
  - `backend/main.go` で CORS ミドルウェアをルート設定の前に移動し、開発時の origin（http://localhost:3000, http://localhost:3001）を許可するよう修正。

- 作業の結果:
  - フロントからバックエンドへの API 呼び出しが正常化し、商品一覧取得のエラーが解消された。

今後の注意点:
- `loadFromStorage` はトークンでのサーバ復元(`/api/me`)を優先し、なければ `localStorage` の `auth_user` をフォールバックで利用する方針。
- localStorage にトークンを保存する実装は学習用としては有用だが、本番では HttpOnly Cookie 等を検討する。

記録者: 作業アシスタント
作成日: 2026-02-14
