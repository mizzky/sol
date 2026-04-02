# 学習記録 2026-04-02

## セッション 1 (10:42)

### 取り組んだタスク
- 架空EC MVP のREADME作成を一区切りにするための相談と文章作成
- READMEの構成検討、セットアップ手順の精密化、ドキュメント導線の整備

### ユーザーが質問した内容
1. MVPとして形になったか評価してほしい
2. READMEのおすすめ構成に沿って自分の言葉で書くためのアドバイスがほしい
3. Setupはどう書くべきか（Dev Container前提、Docker Desktopが必須か、.envは何を書くか）
4. .vscode/tasks.json が Git管理外の場合、起動手順はコマンドベースで書くべきか
5. 作成したREADME案のレビュー

### 躓いたポイントと解決策

**ポイント1: Setup手順の前提条件の記述範囲**
- 躓き：どこまで詳細に前提条件を記載すべきか不明確
- 解決策：DEV Container・WSL2・Docker環境を前提とし、各OSに応じた最小限の情報（Docker Desktop vs Docker Engine）を明記

**ポイント2: Docker Desktop必須性の表現**
- 躓き：Windows/macOS での Docker Desktop の必須性をどう表現するか
- 解決策：OS別に明示
  - Windows/macOS: Docker Desktop（Docker Desktop Installer を使用）が必須
  - Linux: Docker Engine で対応可能（OS別の公式インストール手順へのリンク）

**ポイント3: .env の必須項目を最小化**
- 躓き：.env に何を書くべきか、どこまで最小化するか
- 解決策：
  - 必須: `JWT_SECRET`（認証トークン用途）
  - 必要に応じて: `DATABASE_URL`（DB接続文字列、デフォルト設定が動作する場合は省略可）

**ポイント4: VS Code tasks 依存の記述**
- 躓き：.vscode/tasks.json が Git管理外の場合、起動手順をタスク名で書くべきか、コマンド名で書くべきか
- 解決策：README はコマンド実行による再現性を最優先とする
  - タスクはあくまで開発者の利便性（`Ctrl+Shift+B` で起動可能）
  - README には `air`（Backend）と `npm run dev`（Frontend）の直接コマンドを記載

**ポイント5: パス誤記と記法の統一**
- 躓き：`/workspace` と `/workspaces` の表記ゆれ、bashコメント記法の誤り
- 解決策：
  - 起動パス: `/workspaces/sol_coffeesys` が正（Dev Container付属）
  - bashコメント: `#` を使用（`//` は JavaScript）
  - Setup表記: "## Setup" から "## セットアップ" への日本語統一提案

### 次回課題
1. READMEにドキュメント導線を追記
   - `doc/api.md` - API 仕様
   - `doc/openapi.yaml` - OpenAPI スキーマ
   - `doc/task.md` - 実装タスク管理
   - `doc/db.md` - DB スキーマ記述

2. 実装済み/未対応の境界を明文化
   - MVP に含まれる機能（ユーザー認証、商品管理、カート、注文）
   - フェーズ2以降の機能（支払い処理、在庫検出、通知など）

3. README 最終レビュー
   - 誤字チェック
   - 手順の再現性検証（Dev Container環境での実行確認）
   - OS別セットアップ手順の実装版作成
