# Git コミット & ブランチ命名ルール

プロジェクトの履歴を分かりやすく保つため、以下のルールを採用します。

---

## 1. コミットメッセージ
形式: `<type>: <description>`

| Type | 内容 | 例 |
| :--- | :--- | :--- |
| **feat** | 新機能の追加 | `feat: ログインAPIの実装` |
| **fix** | バグ修正 | `fix: バリデーションエラーの修正` |
| **docs** | ドキュメントのみの変更 | `docs: READMEにAPI仕様を追加` |
| **style** | コードの意味を変えない修正（整形など） | `style: gofmtによるコード整形` |
| **refactor** | リファクタリング（機能変更なし） | `refactor: 重複したハンドラー処理の共通化` |
| **test** | テストの追加・修正 | `test: ユーザー登録APIのテストを追加` |
| **chore** | ビルド関連やライブラリ更新など | `chore: sqlcのコード生成を実行` |

---

## 2. ブランチ命名規則
形式: `<type>/<issue>-<description>` （すべて小文字、単語間はハイフン `-`）

| Type | 用途 | 例 |
| :--- | :--- | :--- |
| **feature/** | 新機能の開発 | `feature/add-jwt-authentication` |
| **bugfix/** | 既存バグの修正 | `bugfix/fix-duplicate-email-error` |
| **docs/** | ドキュメントの整備 | `docs/update-api-documentation` |
| **hotfix/** | 緊急のバグ修正 | `hotfix/fix-security-vulnerability` |

---

## 3. ハンドラ関数の命名規則
- ハンドラ関数は `<Action><Resource>Handler` の形式で命名する。
  - `<Action>`: 処理内容を表す動詞（例: Create, Update, Delete, Get, List）。
  - `<Resource>`: 対象リソース（例: Category, User, Product）。
  - `Handler`: ハンドラであることを明示。

### 例
- `CreateCategory` → `CreateCategoryHandler`
- `Login` → `LoginUserHandler`
- `DeleteProduct` → `DeleteProductHandler`

---

## 4. 実践ガイド
1. **作業前にブランチを作成する** `git checkout -b feature/user-login`
2. **意味のある単位でコミットする** SQLを追加したら一度コミット、ハンドラーを作ったらまたコミット、という風に分けると後で追いやすくなります。
3. **完了したらメインブランチへマージ（またはPR）