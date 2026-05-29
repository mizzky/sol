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

---

## セッション 2 (03:30 - 05:30)

### 取り組んだタスク
- admin register endpoint の設計・方針検討
- 初期化フロー用 test.http ファイルの分離方針確認
- README のセットアップ導線整備計画の立案
- chore/add-initialize-local-testdata ブランチでの作業進行

### ユーザーが質問した内容
- admin 登録エンドポイントを公開すべきか、非公開にすべきか？
- 初期データ投入用と一般テスト用の HTTP リクエストファイルをどう分けるべき？
- セットアップフローを README にどのような形で記載すべき？

### 躓いたポイントと解決策

#### 1. Admin ユーザー登録エンドポイントの設計方針
**問題**: 公開の admin 登録エンドポイント（/api/admin/register など）を設けるべきか検討が必要だった
**結論**:
- **公開エンドポイントは不要** - admin 権限が強力すぎるため、セキュリティリスクが高い
- `register` エンドポイントは **member 権限固定** とする
- 初期 admin の作成は、ローカル開発時は DB 直接操作 or シード SQL、本番は管理画面または安全な初期化フローで対応

#### 2. 初期化フロー用 HTTP ファイルの分離戦略
**問題**: test.http.example（テスト用テンプレート）と初期化処理（setup フロー）が混在する
**解決策**:
- `test.http.example` → 一般的なテストシナリオ用（CI/CD にも使用可能）
- `test.http.init` / `test.init.http` → **初期データ投入用に分離**
  - コマンド実行順序の明確化
  - セットアップ後のエンドポイント疎通確認を含む
  - 新しく参加する開発者が簡単に環境構築できる狙い

#### 3. README セットアップ導線の整備
**方針**:
```
Setup (新規環境構築)
  ├─ 1. git clone / docker setup
  ├─ 2. .env ファイル設定
  ├─ 3. docker up with PostgreSQL
  └─ 4. migrate + backend/frontend 起動
  
Initialize (初期データ投入)
  ├─ 1. backend/test.init.http で初期カテゴリ・商品投入
  ├─ 2. 初回 admin ユーザー作成（DB 直接操作）
  └─ 3. backend/test.http で エンドポイント疎通確認
```
- Setup → Migrate → 起動 → 初期化 → 確認の一本道フローを明示

### 次回課題
- [ ] register エンドポイントを member 権限固定に変更するコード確認
- [ ] test.http.init ファイルのテンプレート作成（カテゴリ・商品投入、admin 昇格 SQL の記載）
- [ ] README.md の Setup/Initialize セクションを整形・追記
- [ ] ブランチ chore/add-initialize-local-testdata での全変更を確認
- [ ] PR 作成前に新規環境での一連フロー動作確認

### 設計決定メモ
- **Admin ユーザー方針**: 登録エンドポイント非公開、初期化は管理側の責任
- **HTTP ファイル分離**: テンプレート用 vs 初期化用を明確に分離して、新規参加者の負担軽減
- **ドキュメント整備**: README に「最初の5分」で環境構築できる導線を提供

---

## セッション 3 (11:44 - 12:00+)

### 取り組んだタスク
- README.md の Setup/Initialize セクションを整形し、migrate→起動→初期データ投入の導線を明確化
- backend/test.init.http ファイルを整備し、空 DB から最小限の運用（ユーザー登録、admin 昇格、ログイン、カテゴリ/商品作成、カート、注文）まで一貫して動作確認
- 初期化フロー実行時に遭遇したエラーの原因調査・切り分け手法の習得
- REST Client 実行時の連鎖エラー挙動の理解と運用ルール整備

### ユーザーが質問した内容
- 初期データ投入フロー全体をどのように整理・文書化すべき？
- test.init.http を実行する際、エラーが発生したらどう対応すべき？
- /api/products で 404 エラーが返った場合、何が原因か？
- createCategory 応答から category_id を抽出して後続リクエストに渡す際、400 エラーが返るケースは？
- Gin フレームワークで大量の import undefined エラーが出た場合の対応？
- JWT_SECRET はどのように管理・設定すべき？
- Cookie 認証と Bearer 認証のエンドポイント実装の使い分けは？

### 躓いたポイントと解決策

#### 1. REST Client での連鎖エラー（前段失敗 → response.body 参照失敗 → 後段全滅）
**問題**: test.init.http を順次実行する際、途中のリクエストでエラー（例：400、404）が発生すると、その応答の `response.body` を参照する後続リクエストが全部失敗する
**原因**:
- REST Client の変数展開（`@categoryId = {{response.body.data.id}}`など）は、前段リクエスト失敗時に null / undefined になる
- その変数を含む後続リクエストが実行されると、自動的に 4xx エラーになる
- UI には個別エラーが見えても、根本原因は「前段の失敗」であることに気づきにくい

**解決策**:
- **実行ルール**: 各リクエスト実行後、HTTP ステータスコード（200/201 など成功範囲）を **逐一確認** してから次へ進む
- **失敗時の切り分け手順**:
  1. 当該リクエストが 2xx/3xx を返すか確認
  2. 応答 body を確認（JSON 構造、エラーメッセージ）
  3. 変数展開が成功したか確認（VS Code debug panel で `@変数名` を参照）
  4. 問題解決後、当該リクエストだけを再実行
  5. 初めてそこから次のリクエストへ進む

**重要な運用ルール**: README と test.init.http に「**順次成功確認しながら実行し、各ステップの成功を確認後に次へ進むこと**」と明記

#### 2. エンドポイント名の誤り（/api/product → /api/products）
**事例**: test.init.http に `/api/product` と記載したところ 404
**原因**: 実装エンドポイントは `/api/products`（複数形）
**解決**: 
- routes.go を確認して正しいパスをリスト化
- test.init.http に明確なコメント付きで全エンドポイント列記

**学習**: API 設計時に「複数形 vs 単数形」を RESTful 設計に合わせて統一し、ドキュメント化することの重要性

#### 3. createCategory 応答から category_id を抽出して POST する際、400 が返る
**事例**:
```
@categoryId = {{response.body.data.id}}
// ↓ 後続リクエストで使用
POST /api/products
{ "category_id": "{{categoryId}}" }
// → 400 Bad Request
```
**原因**:
- 前段の createCategory 自体が失敗していた（例：マイグレーション不完全、カラム型が不正、権限不足）
- 応答ボディに `data.id` が存在しない
- `{{categoryId}}` が null / undefined で、JSON 構造が壊れている

**解決手順**:
1. createCategory のレスポンスをステップバイステップで確認
2. HTTP ステータスが 201 Created か確認
3. 応答 body に `{"data":{"id":"...","name":"..."},...}` の構造があるか確認
4. ID 値をコピーして、次のリクエストで手動テスト
5. 初めてテンプレート変数を使用

**一般則**: **前段エラー時は、変数展開に頼らず、手動で値をコピーして検証** → 変数化

#### 4. Go 側の大量 import / undefined エラー（LSP 不整合）
**症状**: VS Code editor で、backend コードに赤波線が大量に出現（import undefined、型不整合など）
**原因**: VS Code の Go Language Server (gopls) がプロジェクト構造を正しく認識していない状態
**解決策**:
- VS Code の「Go: Restart Language Server」コマンドを実行
- または、VS Code を再起動

**重要**: エラーは「コード不具合」ではなく「LSP 状態不整合」であり、実際のテスト実行（`go test ./...` または `air`）では成功することが多い。混乱を避けるため README に注記

#### 5. 認証周りの理解（Cookie vs Bearer Token、JWT_SECRET 管理）
**学習内容**:

**現状の実装**:
- login エンドポイント → JWT トークン **生成後、Cookie に設定** → レスポンスボディではトークン文字列を返さない
- ブラウザ自動送信される Cookie ベース認証
- test.init.http では Cookie 手動管理が必要

**vs 旧設計**:
- 一般的な API では `Authorization: Bearer <token>` で明示的にトークン送付
- レスポンスボディに token を含める例が多い

**対応**:
- test.init.http には「**Cookie の手動設定手順**」を記載
- README auth セクションに「Cookie ベース認証を採用、初回ログイン後自動送信される」と明記
- ブラウザテスト時は自動、curl / Postman / REST Client テスト時は手動設定

**JWT_SECRET 管理**:
- 現在: ハードコード / env 未使用 の可能性
- 推奨: `.env` に `JWT_SECRET=<値>` として管理
- 初期化フロー時も `.env` 設定の有無を確認する

#### 6. オーバーエンジニアリング回避の具体例
**観察**: セッション前半で「placeholder 化、テンプレート化」を過度に検討していた
**現実**:
- ローカル初期投入は **固定値でシンプル**に（placeholder は本番設定時）
- test.init.http も「最小限のシナリオ」で充分
- 新規参加者が「実際に動く例」を手で実行 → 学習 → カスタマイズの流れが効率的

### 次回課題
- [ ] README.md の Setup/Initialize セクションを完成（実行ディレクトリ、各ステップの確認ポイント明記）
- [ ] backend/test.init.http に全リクエストの「**成功確認ステップ**」をコメントで記載
- [ ] Cookie ベース認証の仕組みを README / doc/api.md に説明
- [ ] JWT_SECRET を env 管理に変更・確認
- [ ] test.init.http の実行手順を doc/dev.md に明記
- [ ] 連鎖エラーの切り分け手法を doc/dev.md（troubleshooting セクション）に固定化
- [ ] admin ユーザー作成の初期 SQL を test.init.http または doc/ に記載
- [ ] 新規参加者が 15分で「ローカル完全起動・初期データ投入・エンドポイント疎通確認」できることを検証

### 学習ポイント
**環境復帰手順の固定化が開発効率の鍵**: セットアップ失敗は、ほぼ常に「初期化フロー未実施 or 段階スキップ」が原因。一度 README に「正規フロー」を明記すれば、チーム全体でズレが減少。

**連鎖エラーの発見法**: HTTP リクエストチェーン内で「前段失敗 → 変数 null → 後段自動失敗」の罠は、各ステップを丁寧に確認する習慣で回避可能。本番 CI/CD でも同じ原理が適用可能。

**コード vs ドキュメント**: 「正しく動く initial setup」さえあれば、新規参加者は迷いなく環境構築できる。テンプレート化・抽象化より、「実行例の明確さ」が勝る。
