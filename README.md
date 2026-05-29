### 概要
- このプロジェクトは学習目的で作成した架空のコーヒーショップECサイトMVPです。
- フロントエンドとAPI連携を通して設計、実装、テストの流れを実践
  
### 機能
- ユーザー
  - 商品一覧の閲覧
  - ログイン
  - カート操作
  - 注文作成
- 管理者
  - 商品管理
  - カテゴリー管理
  - ユーザー権限管理

### 技術スタック
- frontend
  - Next.js
- backend
  - Go(Gin)
- Database
  - PostgreSQL + sqlc

### SetUp
1. 前提
   - Dockerが起動していること(windows/macOSはDocker Desktop, LinuxはDocker engine)
   - VSCodeにDev Containers拡張が入っていること
2. 開発コンテナ起動
3. 環境変数
   - .envに`JWT_SECRET`,`DATABASE_URL`を設定する

4. DBマイグレーション
   ```bash
   cd /workspaces/sol_coffeesys/backend
   migrate -path db/migrations -database "$DATABASE_URL" up
   ```

5. アプリケーション起動
   ```bash
   # terminal1(backend)
   cd /workspaces/sol_coffeesys/backend
   air
   # terminal2(frontend)
   cd /workspaces/sol_coffeesys/frontend
   npm run dev
   ```
6. 初期データ投入
  - `test.init.http`をrest clientで実行する

7. 動作確認
   - ブラウザでログインから商品閲覧、カート、注文作成までの基本フローが動作することを確認   
### ドキュメント
- API設計書
- `doc/openapi.yaml`