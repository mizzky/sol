# 学習記録 2026-05-29

## セッション 1 (00:00 - 03:30)

### 取り組んだタスク
- 開発環境の起動確認（backend air + frontend next）
- フロント起動エラーの調査・解決
- API /api/products の 500 エラー調査・解決
- データベース初期化とマイグレーション実行
- 初期データ投入フローの設計

### ユーザーが質問した内容
- VSCode tasks の Ctrl+Shift+B が起動しない理由は？
- npm run dev で「next: not found」が発生した場合の対応は？
- API が 500 エラーを返す原因は何か？
- migrate コマンドのバージョン確認方法は？
- 初期データをどのように投入すべきか？

### 躓いたポイントと解決策

#### 1. VSCode tasks の Ctrl+Shift+B が動作しない
**原因**: build task がデフォルトに設定されていなかった
**解決**: 
- tasks.json の `dev:all` タスクに `"group": { "kind": "build", "isDefault": true }` を設定
- これにより Ctrl+Shift+B で dev:all が起動

#### 2. frontend の npm run dev で「next: not found」エラー
**原因**: frontend/node_modules が存在しない（依存パッケージ未インストール）
**解決**:
- `cd frontend && npm ci` を実行
- npm ci は package-lock.json の内容を厳密に再現する（CI環境推奨）

#### 3. /api/products が 500 InternalError を返す
**原因**: PostgreSQL に products テーブルが存在しない（DB マイグレーション未実行）
**解決**:
- `migrate -path db/migrations -database "$DATABASE_URL" up` を実行
- DB テーブルが作成され、API は 200 {"products":[]} を返すように
- **根本原因**: 環境差分（Windows→Mac 端末切替時に DB の初期化状態がズレた）

#### 4. migrate コマンドの使用方法
**問題**: `migrate version` を実行すると「URL cannot be empty」で失敗
**学習内容**:
- `migrate -version` で CLI バージョン確認
- `migrate -path db/migrations -database "$DATABASE_URL" version` で DB 適用済みバージョン確認
- source/database の指定が必須

#### 5. 端末切替による状態差分
**観察**: Windows と Mac で接続先 DB が異なる場合、マイグレーション適用状況とデータ有無がズレる
**対応**: 現在の DB は全テーブル（users, categories, products, carts, orders）が 0 件状態

### 次回課題
- [ ] README.md の Setup 章を整形（不要行削除、実行ディレクトリを明記）
- [ ] backend/test.http.init を完成（商品/カテゴリ投入の HTTP リクエスト整備）
- [ ] 初回 admin 昇格の手順を明示（DB 直接更新 SQL を test.http.init に記載するか doc で示す）
- [ ] 初期データ投入後、全エンドポイントの動作確認
- [ ] 「migrate→起動→初期データ投入→確認」の一連フローを README に明記
- [ ] branch `chore/add-initialize-local-testdata` で PR 作成前に確認作業完了

### 学習ポイント
**重要**: 問題が連鎖しても、症状→原因→検証の順で分解すれば必ず収束できる。本日の根本原因はコード不具合ではなく環境差分（依存未導入・DB未初期化）であり、再発防止は「初回セットアップ手順の明文化」と「初期データ投入フローの固定化」が鍵となる。
