# その他機能・インフラ・CI — チケット詳細（作成日: 2026-02-16）

目的
- プロジェクト運用に必要な周辺機能、CI、ドキュメント、セキュリティチェックを整備する。

チケット一覧

1) sqlc ワークフロー文書化 — Effort: Low, Priority: P0
   - 内容: `backend/query.sql` を更新→`sqlc generate` を行う手順、必要なバイナリ/バージョンを `doc/` に記載。  
   - 影響ファイル: doc/planning, backend/query.sql
   - 受け入れ条件: 手順書があること。

2) テスト基盤: 統合テスト用 DB スクリプト — Effort: Med, Priority: P0
   - 内容: ローカルでマイグレーション適用→テスト実行→ロールバック のスクリプト（Makefile かシェル）を作成。  
   - 影響ファイル: backend/tests, backend/db/migrations
   - 受け入れ条件: ローカルで `./scripts/test-e2e.sh` のように実行できること。

3) CI ジョブ提案（unit / integration 分離） — Effort: Low, Priority: P1
   - 内容: unit tests は mock ベース、integration は DB マイグレーション適用で実行する2ジョブ構成の YAML 素案を作成。  
   - 影響ファイル: .github/workflows (提案), doc/planning
   - 受け入れ条件: CI提案ドキュメントがあること。

4) セキュリティチェックリスト — Effort: Low, Priority: P0
   - 内容: XSS/CSRF/SQLi/認証トークン扱い/Secrets管理 のチェックリストを作成。  
   - 影響ファイル: doc/planning/security.md (新規)
   - 受け入れ条件: チェックリストが存在すること。

5) API ドキュメント（OpenAPI 草案） — Effort: Med, Priority: P1
   - 内容: 主要エンドポイント（ユーザ/カート/注文）の OpenAPI スキーマ草案を作成。  
   - 影響ファイル: doc/planning/openapi.yaml (新規)
   - 受け入れ条件: 主要なPOST/GET/PATCH の request/response が記述されていること。

6) ロギング/監視設計（シンプル版） — Effort: Low, Priority: P2
   - 内容: どのイベントを INFO/WARN/ERROR で出すか、重要メトリクス一覧を作成。  
   - 影響ファイル: doc/planning, backend/logging (提案)

7) ドキュメントテンプレート整備 — Effort: Low, Priority: P0
   - 内容: 設計書・運用ドキュメントのテンプレート（設計/テスト/リリース）を `doc/planning/templates/` に追加。  
   - 影響ファイル: doc/planning/templates/

実施順の推奨
1→2→3→5→4→7→6
