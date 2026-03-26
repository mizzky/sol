# 学習記録 2026-03-26

## セッション 1

### 取り組んだタスク
- チケット9（ドキュメント更新）の方針変更に対応し、注文APIの記述先を doc/api.md ではなく OpenAPI に統一する方針を確定
- OpenAPI の検証手順として swagger-cli と Spectral の実行方法を確認
- OpenAPI バリデーションエラーの原因切り分けを実施
- OpenAPI のプレビュー方法（VS Code 拡張 / Redoc 生成）を確認
- devcontainer.json への Swagger Viewer 追加方法を確認

### ユーザーが質問した内容
- swagger / spectral の実行方法
- Spectral 実行時の command not found への対処
- lint 成功時の結果解釈（warnings は問題か）
- openapi.yaml のプレビュー方法
- devcontainer.json に Swagger Viewer を追加する書き方

### 躓いたポイントと解決策
- 躓き: swagger-cli validate で OpenAPI 構文エラーが発生
- 原因:
  - summary のキー typo（summarY）
  - パス typo（/api/orderes/{id}/cancel）
  - $ref の書式ミス（$ref:'#/...' になっておりキー解釈されない）
- 解決策:
  - summary に修正
  - /api/orders/{id}/cancel に修正
  - $ref: '#/components/schemas/Order' に修正

- 躓き: npx spectral lint 実行時に spectral: not found
- 原因:
  - npx が spectral という別パッケージ名を解決しようとした
- 解決策:
  - npx @stoplight/spectral-cli lint ../doc/openapi.yaml -r ../.spectral.yaml を使用
  - もしくは npm i -D @stoplight/spectral-cli 後に npx spectral lint を使用

### 次回課題
- OpenAPI の残りエンドポイント（products の PUT/DELETE、cart 系など）を routes.go と完全同期
- operationId 命名規則を統一（動詞 + 対象）
- OpenAPI lint を CI に組み込み（GitHub Actions）
- チケット9の完了反映を task.md に最終反映

---

## セッション 2

### 取り組んだタスク
- studylog-writer カスタムエージェントの追記対応修正
- SKILL.md のステップ3を「新規 / 既存ファイル」分岐フォーマットに改訂
- studylog-writer.md の実行手順3に既存ファイル確認と追記/新規分岐を追加
- 既存の 26_learning.md を旧フォーマット（h2見出し）から新フォーマット（セッション番号 + h3見出し）へ移行

### ユーザーが質問した内容
- 既存の learning.md がある場合に上書きではなく追記する形にしてほしい（午前・午後と複数回ログ記録するユースケース）
- agents や skills を調査して修正してほしい

### 躓いたポイントと解決策
- 特になし（設計変更のみ）

### 次回課題
- 実際に同日2回目のログ記録が正しく追記されるか動作確認
