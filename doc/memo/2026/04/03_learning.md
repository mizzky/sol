# 学習記録 2026-04-03

## セッション 1 (本日)

### 取り組んだタスク

- **テーマ**: JWTのtoken保持改善（localStorage → HttpOnly Cookie）と refresh token導入設計
- **実施内容**:
  1. 現状調査：backend/frontendの認証実装確認
  2. 設計の明確化：refresh tokenはハッシュ保存、同時ログイン端末数は無制限、login responseのtoken即時廃止
  3. セキュリティ方針の整理：再利用方式をまず採用し、ローテーション/再利用検知は将来課題へ
  4. doc/planning/migrate-auth-to-cookie.md へ方針追記
  5. DB migration設計レビュー（timestamp with time zone整合など）
  6. query.sql の refresh tokenクエリレビューと改善
  7. sqlcコード生成からコミット完了

### ユーザーが質問した内容

- Cookie Secure属性の意味：HTTPSのみへの送信制限について
- クライアント種別ごとのトークン保持戦略の違い：ブラウザ vs モバイルアプリ
- ローテーション導入を設計段階では採用せず将来課題にする理由：実装複雑性とセキュリティリスク評価のバランス

### 躓いたポイントと解決策

| 躓きポイント | 原因 | 解決策 |
|----------|------|------|
| query.sql revoke/revoked カラム名ミス | SQL記述時のシンタックスエラー | 正しいカラム名を確認し修正 |
| query.sql VALUES列数ミスマッチ | INSERT時のバインド値と列数の不一致 | 列定義と値の数を一致させる |
| query.sql 検索条件の不一致 | WHERE句の条件ロジックが要件と異なる | 設計要件に基づいて条件を修正 |
| Revoke系の冪等性と意図の混乱 | 単端末失効（refresh token失効）と全端末失効（user全体ブロック）の区別が不明確 | 設計ドキュメントで両パターンを整理し、実装対象を明確化 |
| migration の revoked_at カラムの型と NULL許容性 | timestamp型の整合性チェック漏れ | timestamp with time zone に統一し NULL許容を確認 |

### 次回課題

1. **即時開始**:
   - backendのTDD Red開始：refresh token DB層テスト（query_refresh_token_test.go）
   - login/logout/refresh/middlewareの実装をTDDで段階的に進行

2. **将来課題**（チケット化予定）:
   - refresh token ローテーション方式の設計・実装
   - refresh token 再利用検知メカニズム
   - logout時の全端末トークン失効オプション実装

3. **設計確認**:
   - frontend側の cookie 読み取り挙動確認（HttpOnly Cookie読取不可仕様の再確認）
   - refresh token エンドポイント実装時の CSRF対策設計
