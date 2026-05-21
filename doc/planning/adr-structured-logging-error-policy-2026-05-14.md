# ADR: 構造化ロギング実装と エラーハンドリング統合設計

**ADR ID**: ADR-2026-001  
**作成日**: 2026-05-14  
**関連 Issue**: #60（構造化ロギング実装）、#65（ログレベル運用定義）、#74（ログ設計書）、#67（カスタムエラー構造体）  
**ステータス**: 提案中（Issue A 実装済み、B～F は検証待ち）  
**優先度**: P0  

---

## 1. 背景

### 問題点
- 従来の文字列ベースログでは、エラー発生時の原因特定に 10 分以上を要していた
- ログが複数箇所で出力され、同一エラーが二重・三重にログされる（ノイズ増加）
- 認証トークンやパスワードなど機微情報がログに含まれるリスク
- ユーザー行動追跡（ユーザーID・リクエスト ID との紐付け）ができない

### 既存成果
- `backend/pkg/apperror` でカスタムエラー型を定義済み
- `backend/middleware/request_id.go` でリクエスト ID 採番・伝搬完了
- `backend/auth/middleware.go` で認証ユーザー ID を Context にセット済み

### 設計基準（Issue #65・#74 から）
- **JSON 構造化ログ**: Go 標準ライブラリ `log/slog` で JSON 出力
- **ログレベル**: DEBUG / INFO / WARN / ERROR を明確に分離
- **必須フィールド**: request_id / user_id / trace_id / event / message / error_type / status / duration_ms / method / route
- **マスキング対象**: password / token / email などの機微情報を `[REDACTED]` に置換
- **責務分離**: ErrorHandler のみでエラーログを出力（二重ログ禁止）

---

## 2. 決定事項

### 2.1 ログ出力原則
1. **ErrorHandler が唯一のエラーログ出力地点**
   - ハンドラ層・サービス層では `return err` のみ
   - ErrorHandler で error_type・status 判定 → ログレベル決定 → slog 出力

2. **正常系イベントログはハンドラ層で出力**
   - INFO/DEBUG レベルのイベント（ログイン成功、商品検索、注文作成等）
   - ハンドラ層で `slog.Info()`/`slog.Debug()` で直接出力
   - ErrorHandler では出力しない（異常系のみ）

3. **ログ出力 ＋ エラー返却の同時実行禁止**
   - ハンドラ層で「ログ出力してからエラーを返す」はしない
   - 責務：ハンドラ層 → エラー返却、ErrorHandler → ログ出力

### 2.2 slog 初期化と設定
- **出力形式**: JSON （運用環境でのログ検索・集計を想定）
- **出力先**: 標準出力（OS/Docker ログシステムに委譲）
- **ReplaceAttr**: password / token / email フィールドを `[REDACTED]` に置換
- **タイムスタンプ**: slog デフォルト（キー名 `time`、RFC3339 形式）
- **グローバル設定**: `slog.SetDefault()` で アプリ全体に適用

### 2.3 ログレベル決定ロジック

| 内部エラー型 | HTTP ステータス | ログレベル | 判定根拠 |
|---|---|---|---|
| `ValidationError` | 400 | **INFO** | ユーザー入力エラー、正常な異常系 |
| `NotFoundError` | 404 | **INFO** | リソース存在しない、正常な異常系 |
| `ConflictError` | 409 | **INFO** | 重複・競合、ビジネスロジック上の正常な異常系 |
| `BusinessLogicError` | 400/422 | **INFO** | ビジネスルール違反、正常な異常系 |
| `UnauthorizedError` | 401 | **WARN** | 認証なし・失敗、セキュリティ関連 |
| `ForbiddenError` | 403 | **WARN** | 権限不足、セキュリティ関連 |
| `InternalError` | 500 | **ERROR** | DB エラー・予期しないエラー、要対応 |

**理由**:
- INFO：予期された異常系、ビジネスロジックの正常な分岐
- WARN：セキュリティリスク（認証失敗の集計でブルートフォース検知）
- ERROR：システム障害、緊急対応必要

---

## 3. 必須・任意ログフィールド仕様（Issue #74 ベース）

### 3.1 必須フィールド（全ログに必ず含む）

| キー | 型 | 説明 | 例 | 注記 |
|---|---|---|---|---|
| `timestamp` | ISO8601 | ログ出力時刻（slog デフォルト） | `2026-05-14T10:30:45.123Z` | slog の `time` キーに対応 |
| `level` | string | ログレベル（slog デフォルト） | `INFO`, `WARN`, `ERROR` | slog の `level` キー |
| `message` | string | 人間向けのメッセージ | `user login failed: invalid email format` | 詳細な説明 |
| `request_id` | string (UUID) | リクエスト ID（全リクエスト採番） | `550e8400-e29b-41d4-a716-446655440000` | 追跡の中心軸 |
| `method` | string | HTTP メソッド | `GET`, `POST`, `PUT`, `DELETE` | - |
| `route` | string | ルートパターン | `/api/users/:id`, `/api/products` | path ではなく route（パラメータ含む） |
| `status` | integer | HTTP ステータスコード | `200`, `400`, `401`, `404`, `500` | - |

### 3.2 条件付き必須フィールド

| キー | 型 | 条件 | 説明 | 例 | 注記 |
|---|---|---|---|---|---|
| `user_id` | int64 \| null | 認証済みの場合 | リクエストユーザー ID | `12345` | 未認証の場合は null または 0 |
| `event` | string | 異常系以外 | イベント種別（snake_case） | `user_login_succeeded` | ErrorHandler では出力不要 |
| `error_type` | string | エラー系のみ | 内部エラー型分類 | `ValidationError`, `InternalError` | ErrorHandler で判定・出力 |
| `duration_ms` | float64 | INFO/WARN/ERROR | リクエスト処理時間 | `125.45` | ミリ秒単位 |

### 3.3 任意フィールド（拡張時に追加可）

| キー | 型 | 説明 | 例 | 実装時期 |
|---|---|---|---|---|
| `trace_id` | string (UUID) \| null | 分散トレーシング ID（将来） | `550e8400-e29b-41d4-a716-446655440001` | Phase 2（マイクロサービス化時） |
| `client_ip` | string | クライアント IP アドレス | `192.168.1.1` | Phase 2（セキュリティ監視強化時） |
| `operation` | string | 操作内容（ビジネス層） | `create_user`, `update_product` | Phase 2 |
| `response_size` | integer | レスポンスサイズ（バイト） | `1024` | Phase 2 |
| `path` | string | リクエストパス（query を含む） | `/api/products?category=1` | 必要に応じて（オプション） |

**null 許容ルール**:
- `user_id` / `trace_id` は未認証・未実装時に `null` で OK
- その他の必須フィールドは常に値を持つ（存在しない場合はエラー）

---

## 4. ログレベル運用定義と使用例（Issue #65 ベース）

### 4.1 ログレベルの意味と用途

| レベル | 用途 | アラート対象 | 保管期間 | 例 |
|---|---|---|---|---|
| **DEBUG** | 開発時の詳細追跡 | ✗ なし | 1 日 | 関数内の変数値、条件分岐ポイント |
| **INFO** | 正常な異常系（予期された例外） | ✗ なし | 1 ヶ月 | バリデーション失敗、404 Not Found、ログイン成功 |
| **WARN** | セキュリティ関連・要注意 | ○ 集計後（ブルートフォース検知） | 3 ヶ月 | 認証失敗（401）、権限不足（403）、レート制限 |
| **ERROR** | システム障害・緊急対応必要 | ○ 即座 | 1 年 | DB エラー、予期しないクラッシュ、リソース不足 |

### 4.2 実装上の指針

**ログ汚染防止**:
- DEBUG ログはデフォルトで出力しない（開発時のみ環境変数で有効化）
- 頻出する INFO（例：全リクエストの成功ログ）は出力時間帯を制限

**機密情報除外**:
- password / token / email は `[REDACTED]` に置換
- リクエスト Body・Response JSON 全体をログに出力しない（必要なキーのみ抽出）

**コンテキスト付与**:
- 全ログに `request_id` を付与（追跡の起点）
- 認証済みユーザーのログには `user_id` を付与

---

## 5. ErrorHandler 責務と通常イベントログ責務の境界

### 5.1 ErrorHandler（最上位エラーハンドラ）
**責務**:
- エラーハンドリング（HTTP ステータス返却）
- **エラーログ出力のみ**（INFO/WARN/ERROR で分類）
- マスキング（slog の ReplaceAttr）

**出力対象**:
- 異常系：ValidationError, NotFoundError, UnauthorizedError, InternalError 等

**出力しない**:
- 正常系イベント（ログイン成功、商品検索完了等）

### 5.2 ハンドラ・サービス層
**責務**:
- ビジネスロジック実装
- **正常系イベントのみ slog.Info/Debug で出力**（処理開始・完了時点）
- エラーは return で ErrorHandler に委譲

**出力対象**:
- 正常系：`user_login_succeeded`, `product_list_fetched`, `order_created` 等

**出力しない**:
- エラー情報（ErrorHandler に委譲）
- 同じイベントの重複出力（"ログイン開始"→"ログイン成功" で 2 行だが、重複とは異なる）

### 5.3 見かけ上の矛盾の解消

> Issue B：「ErrorHandler のみでログ出力」と、Issue F：「ハンドラ/サービスで通常イベントログ出力」は矛盾していないか？

**回答**：**責務が異なるため矛盾しない**
- ErrorHandler：**エラーログ出力** （INFO/WARN/ERROR で事象分類）
- ハンドラ層：**イベントログ出力** （INFO/DEBUG で正常系フロー追跡）

具体例：

```
【ログイン正常系】
1. ハンドラ層: slog.Info("event", "user_login_started", "email", user.Email)  ← 処理開始
2. DB クエリ成功
3. ハンドラ層: slog.Info("event", "user_login_succeeded", "user_id", user.ID)  ← 処理完了
（ErrorHandler は出力されない）

【ログイン エラー系】
1. ハンドラ層: 不正なメール形式
2. ハンドラ層: return ValidationError("invalid email format")  ← ハンドラは return のみ
3. ErrorHandler: slog.Info("event", ..., "error_type", "ValidationError", "status", 400)  ← ここでログ出力
（ハンドラ層では出力されない）
```

**結論**: 正常系はハンドラ層、異常系は ErrorHandler が責務。**二重ログにはならない**。

---

## 6. マスキング方針

### 6.1 マスキング対象とルール

| 対象情報 | 例 | マスキング方法 | 適用層 |
|---|---|---|---|
| **password** | `password123` | 値全体を `[REDACTED]` に置換 | slog ReplaceAttr（ログ出力時） |
| **token** (JWT・API トークン) | `eyJhbGciOiJIUzI1NiIs...` | 値全体を `[REDACTED]` に置換 | slog ReplaceAttr（ログ出力時） |
| **email** | `user@example.com` | 値全体を `[REDACTED]` に置換 | slog ReplaceAttr + apperror ReplaceAttr（エラー返却時） |

### 6.2 マスキングの2層設計

**層 1：apperror（エラー生成時）**
- 対象：ドメイン・エラーレスポンスに乗る可能性のある値（email）
- 理由：クライアントに返却されるため、PII を含めない
- 実装：`apperror.ValidationError()` で `message` を生成時にマスク

**層 2：slog（ログ出力時）**
- 対象：ログに溜まる機微情報（password, token）
- 理由：ログは運用フェーズで大量に溜まり、アクセス制御外
- 実装：`slog.Handler` の `ReplaceAttr()` で値を置換

### 6.3 実装上の注意
- キー名は小文字統一（`password`, `token`, `email` — 大文字 `PASSWORD` は対象外）
- ネストされたキー（例：`user.password`）もマスク対象とする可能性 → Phase 2 で検討
- `[REDACTED]` 固定値（長さ可変は不可、検索性低下）

---

## 7. 実装フェーズ（Issue A～E/F）

### 7.1 段階的実装スケジュール

```
Issue A (slog 初期化)
    ↓ [実装完了済み] PR #79
Issue B (ErrorHandler 拡張)
    ↓ [実装予定 P0]
Issue C (request_id 付与)
    ↓ [実装予定 P0]
Issue D (user_id 付与)
    ↓ [実装予定 P0]
Issue F (通常イベントログ INFO/DEBUG)
    ↓ [実装予定 P0]
Issue E (redaction ユーティリティ統合)
    ↓ [実装予定 P1、将来タスク]
```

### 7.2 各フェーズの概要

| Issue | タイトル | 優先度 | 実装箇所 | 依存関係 | 状態 |
|---|---|---|---|---|---|
| **A** | slog ロガー初期化とマスキング設定 | P0 | `middleware/error_handler.go`, `main.go` | なし | ✅ 完了 |
| **B** | ErrorHandler に slog ログ出力を組み込む | P0 | `middleware/error_handler.go` | A | ⏳ 予定中 |
| **C** | request_id をログフィールドへ付与 | P0 | `middleware/error_handler.go` | B | ⏳ 予定中 |
| **D** | user_id をログフィールドへ付与 | P0 | `middleware/error_handler.go` | C | ⏳ 予定中 |
| **F** | 通常イベントログ（INFO/DEBUG）実装 | P0 | `handler/user.go`, `handler/product.go`, `handler/order.go` | D | ⏳ 予定中 |
| **E** | redaction ユーティリティ統合 | P1 | `backend/pkg/redaction/` | A, B | ⏳ 将来（Phase 2） |

---

## 8. 整合性チェック結果

現行の実装計画（doc/task.md #60 セクション）と、Issue #65・#74 設計要件を照らし合わせた結果、以下の矛盾・不足を検出しました。

### 8.1 ✅ 一貫性が確認できた項目

| 項目 | task.md の記載 | #65/#74 の要件 | 判定 |
|---|---|---|---|
| ログ出力地点統一 | ErrorHandler 最上位のみ | 記載なし（暗黙） | ✅ 一貫 |
| ログレベル分類 | INFO/WARN/ERROR | Issue #65 で同じ | ✅ 一貫 |
| request_id 付与 | Issue C で実装 | 必須フィールド | ✅ 一貫 |
| user_id 付与 | Issue D で実装 | 条件付き必須 | ✅ 一貫 |
| マスキング対象 | password, token, email | Issue #65 同じ | ✅ 一貫 |

### 8.2 ⚠️ 明確化が必要な項目

#### 8.2.1 ErrorHandler と通常イベントログの責務分離（**軽度の矛盾**）

**現象**:
- Issue B：「ErrorHandler でログ出力」
- Issue F：「ハンドラ/サービスで通常イベントログ出力」
- → 「どちらで出力する？」という混乱の余地

**実態**:
- 責務が異なる（エラー vs 正常系イベント）ため矛盾しない

**是正案**:
- **Issue B の説明文を修正**: 「ErrorHandler は**エラーログのみ**出力。正常系は Issue F で実装」と明記
- **Issue F の説明文を修正**: 「正常系イベント**のみ**出力。エラーは ErrorHandler に委譲」と明記
- → 責務分離を図 1 で可視化（本 ADR の「5. ErrorHandler 責務と通常イベントログ責務の境界」参照）

#### 8.2.2 必須フィールドの仕様差分（**中程度の不足**）

**現象**:
- Issue B（ErrorHandler）の受け入れ条件に記載：`event`, `error_type`, `status`, `method`, `route`
- Issue #74（ログ設計書）の必須フィールド：`timestamp`, `level`, `message`, `request_id`, `method`, `route`, `status` + 条件付き `user_id`, `event`, `error_type`, `duration_ms`
- → **完全に一致していない**（どの項目を ErrorHandler で、どれを通常イベント時に出力するのか不明）

**実態**:
- Issue B は「ErrorHandler のみの必須」
- Issue #74 は「全ログの必須」を定義している
- → **フェーズによって必須項目セットが異なる**

**是正案**:
```
【Phase 1（Issue A〜D 完了時点）】
ErrorHandler で出力する項目:
  - timestamp, level, message, request_id, method, route, status, error_type, duration_ms, user_id（認証済み時）

【Phase 2（Issue F 完了時点）】
ハンドラ層で出力する項目（正常系イベント）:
  - timestamp, level, message, request_id, event, method, route, status, duration_ms, user_id（認証済み時）
  - error_type は出力しない（正常系のため）

【Phase 3（Issue E 完了時点、将来）】
任意フィールド（trace_id, client_ip, operation）を段階的に追加
```

→ **実装ガイドラインに「フェーズごとの必須フィールド表」を追加する**（本 ADR の「3.2 条件付き必須フィールド」参照）

#### 8.2.3 slog デフォルトキー（time）と #74 要件（timestamp）の不一致（**軽度の不足**）

**現象**:
- slog JSONHandler デフォルト：キー名は `time`（RFC3339 形式）
- Issue #74 の仕様：`timestamp`

**実態**:
- 単なるキー名の違い（同一データ）
- 「どちらで統一するか」の設計が曖昧

**是正案**:
- **slog のデフォルトをそのまま使用** → JSON 出力で `"time"` キーで固定
- **Issue #74 の仕様を修正**: `timestamp` ではなく `time`（slog ネイティブ） に合わせる
- **理由**: slog の標準出力に任せることで、後続ログ処理・集計ツール（CloudWatch, Datadog 等）との連携性向上

**修正案（Issue #74 へのコメント）**:
```
出力スキーマの `timestamp` フィールドについて：
実装では slog JSONHandler デフォルト（キー名 `time`, RFC3339 形式）を使用する予定です。
理由：ログ集計ツール（CloudWatch, Datadog等）がslog標準形式を認識するため
修正提案：本仕様の `timestamp` → `time` に読み替えて運用してください。
```

#### 8.2.4 route（必須）と path（任意）の使い分け（**中程度の注意**）

**現象**:
- Issue #74 で `route` は必須、`path` は任意と定義
- Issue B のテスト項目に `route` の記載があるが、**実装者向けの「パラメータ付きルート」定義がない**

**実態**:
- Gin では `c.FullPath()` で完全パス（含パラメータ）が取得可能
- ただし ERR が複数パターン（`/api/users/1`, `/api/users/2` は同じ route `/api/users/:id`）の場合、ルート定義が必要

**実装上の注意**:
- `route` の値：`/api/users/:id`, `/api/products`, `/api/cart` など（パラメータ名含む）
- `path` の値（任意）：`/api/users/123?sort=name` （実際のクエリ含む）
- **ルート情報の抽出方法**: Gin の MatchedRoute 機能を活用

**是正案**:
- **Issue B の受け入れ条件に追記**: 「ログ JSON に `route` フィールドが含まれ、パラメータ名が含まれていることを確認する」
- **実装ガイド**: 以下コード例をドキュメント化
```go
// ErrorHandler 内で
route := c.FullPath()  // ← パラメータ名を含む
// 例: "/api/users/:id" （パラメータは :id のままで出力）
```

#### 8.2.5 null 許容ルール の明文化（**軽度の不足**）

**現象**:
- Issue D で「user_id が存在しない場合もログ出力が正常に完了すること」と記載
- **具体的に「null なのか 0 なのか」が明確でない**

**実態**:
- Go では `int64` は null を表せない（nil ポインタか 0 のどちらか）
- JSON では `null` か数値 `0` か判断必要

**是正案**:
- **未認証時は `user_id` を出力しない（フィールドを省略）** が最適
  - 理由：JSON のサイズ削減、ログ集計ツールの null 値フィルタリング不要
  - 実装：ErrorHandler で `if userID > 0 { attrs = append(attrs, slog.Int64("user_id", userID)) }`
  
- **またはデフォルト値 `0` を出力** する方針もあり
  - 理由：常に同じキー構造で集計ツール側が統一処理可能

**推奨**: **フィールド省略方式** を採用
```json
// 未認証の場合
{
  "timestamp": "2026-05-14T10:30:45.123Z",
  "level": "WARN",
  "message": "invalid token",
  "request_id": "550e8400-...",
  "route": "/api/users/:id",
  "status": 401
  // ← user_id は出力しない
}

// 認証済みの場合
{
  "timestamp": "2026-05-14T10:30:45.123Z",
  "level": "INFO",
  "message": "user registered",
  "request_id": "550e8400-...",
  "user_id": 12345,  // ← 出力
  "route": "/api/users",
  "status": 201
}
```

#### 8.2.6 UnauthorizedError/ForbiddenError の WARN 固定化の根拠（**軽度の説明不足**）

**現象**:
- Issue B で UnauthorizedError/ForbiddenError を WARN に固定
- **「なぜ WARN か」の根拠がテキスト内に明示されていない**

**実態**:
- Issue #65 でログレベル定義があるが、**根拠理由がない**

**是正案**:
```
【WARN 固定化の根拠】
1. セキュリティ監視の軸：ブルートフォース攻撃検知
   - 認証失敗（401）が短時間に多発 → WARN ログの集計で異常検知可能

2. 運用アラート設定の標準化
   - ERROR は DB 障害など即座対応、WARN は傾向監視
   - 401/403 は「ユーザーの利用ミス」か「攻撃」かを数分後に判定できる粒度

3. ノイズ低減
   - INFO だと誤認証が大量にログされてノイズになる
   - WARN はフィルタリング対象（デフォルト表示）となり、分析性向上
```

→ **Issue #65 の仕様に「根拠」セクションを追加**

#### 8.2.7 二重ログ防止規約（**中程度の不足**）

**現象**:
- task.md に「二重ログ禁止」と記載あるが、**「ハンドラ層が return 時にログを出さない」という実装規約が明記されていない**

**実態**:
- PR レビュー時に「このハンドラはなぜログ出力がないのか？」という質問が発生しやすい

**是正案**:
- **実装ガイドラインを追加**:
```
【二重ログ防止の実装規約】

NG 例（禁止）：
  if err != nil {
    slog.Error("database error", "err", err)  // ← ハンドラでログ出力（禁止）
    return err
  }

OK 例（推奨）：
  if err != nil {
    return err  // ← ErrorHandler が拾ってログ出力する
  }
```

→ **PR テンプレート・Lint ルールに「二重ログチェック」を追加**（実装フェーズで）

---

## 9. 是正アクション（優先度付き）

### Priority: P0（Issue A～F の実装に同期）

| No | アクション | 対象 | 実行時期 | 備考 |
|---|---|---|---|---|
| **1** | Issue B・F の説明文修正：責務分離を明記 | task.md | Issue B 着手前 | ErrorHandler（異常系）vs ハンドラ層（正常系）の境界を図示 |
| **2** | フェーズごとの「必須フィールド表」を追加 | 本 ADR + task.md | Issue B 着手前 | Phase 1/2/3 で出力項目を明示（本 ADR 3.2 参照） |
| **3** | route/path の実装ガイドを追加 | 本 ADR + task.md | Issue B 着手前 | `:id` などのパラメータ名を含める例示 |
| **4** | user_id の null 許容ルール（フィールド省略案）を明記 | 本 ADR | Issue D 着手前 | JSON フォーマット例を併記 |

### Priority: P1（Phase 2 以降）

| No | アクション | 対象 | 実行時期 | 備考 |
|---|---|---|---|---|
| **5** | slog の `time` キー vs Issue #74 の `timestamp` 統一 | Issue #74 修正 + 実装 | Issue F 後 | ログ集計ツール連携時の標準化 |
| **6** | UnauthorizedError/ForbiddenError を WARN にする根拠をドキュメント化 | Issue #65 | Issue D 後 | セキュリティ監視・ブルートフォース検知の文脈を明記 |
| **7** | 二重ログ防止の実装規約を明記 | 本 ADR + PR テンプレート | Issue E 後 | Lint ルール（slog 出力検知）の検討 |
| **8** | redaction ユーティリティの拡張性設計 | Issue E | Phase 2 計画 | trace_id, client_ip, operation のマスキング検討 |

---

## 10. 非採用案

### 10.1 検討したが採用しなかった案

#### 案 1: ErrorHandler 以外の層でもエラーログ出力を許可
- **理由**: ログの一貫性喪失、二重ログ増加、原因特定時の混乱
- **判定**: ❌ 不採用

#### 案 2: 全ログレベルを INFO に統一（WARN/ERROR なし）
- **理由**: ログの優先度判定が困難、アラート設定が不可能
- **判定**: ❌ 不採用

#### 案 3: パスワード・トークンの値をマスキングせず出力
- **理由**: ログが大量に溜まる運用環境でセキュリティリスク
- **判定**: ❌ 不採用

#### 案 4: request_id を各ハンドラで独立採番（ミドルウェア不要）
- **理由**: ハンドラ重複実装、採番ルール の一貫性喪失
- **判定**: ❌ 不採用（ミドルウェア採番で統一）

#### 案 5: slog 以外の logger ライブラリを採用（例：zap, logrus）
- **理由**: Go 1.21 標準ライブラリで十分、外部依存最小化、メンテナンス簡化
- **判定**: ❌ 不採用（slog 標準採用）

---

## 11. 受け入れ基準（#60 完了クローズ基準）

実装完了時、以下の基準を全て満たすことで #60 を Close とする。

### 11.1 ログ出力の正確性

- [ ] 正常系フロー 3 つ（ログイン成功、商品一覧取得、注文作成）で request_id 起点に `request_id`・`event`・`method`・`route`・`status`・`duration_ms` が欠落なく出力される
- [ ] 異常系シナリオ 3 つ（バリデーション失敗、認証失敗、DB エラー）で `error_type`・`status` が正確に判定・出力される
- [ ] JSON ログから 10 秒以内に原因特定できる

### 11.2 ログレベル決定の正確性

- [ ] ValidationError / NotFoundError で INFO ログが出力される
- [ ] UnauthorizedError / ForbiddenError で WARN ログが出力される
- [ ] InternalError で ERROR ログが出力される
- [ ] 各ログレベルの採用理由について自分の論理的根拠がある（本 ADR「4. ログレベル運用定義」参照可）

### 11.3 エラーログ一貫性

- [ ] エラーが最上位（ErrorHandler）で重複なく構造化ログとして出力される
- [ ] ハンドラ層でのログ出力がなく、ErrorHandler が唯一のエラーログ地点
- [ ] 二重ログがない（同一エラーが複数回出力されない）

### 11.4 機密情報保護

- [ ] JWT 署名・認証トークンが `[REDACTED]` に置換される
- [ ] DB パスワードが `[REDACTED]` に置換される
- [ ] ユーザーメールアドレスが `[REDACTED]` に置換される（エラー返却時含む）
- [ ] password / token / email キー以外の値は変わらない（マスク対象外は出力）

### 11.5 実装完了性

- [ ] Issue A～F の全 GitHub Issue がクローズされている
- [ ] 各 Issue ごとに PR がマージされている
- [ ] 全ユニット・統合テストがPASS
- [ ] コードレビュー指摘が全て対応済み

---

## 12. 参考リンク・関連ドキュメント

### GitHub Issues
- [#60](https://github.com/mizzky/sol/issues/60) - 構造化ロギング実装（メイン Issue）
- [#65](https://github.com/mizzky/sol/issues/65) - ログレベル運用定義書
- [#67](https://github.com/mizzky/sol/issues/67) - カスタムエラー構造体設計
- [#74](https://github.com/mizzky/sol/issues/74) - ログ設計書
- [#75](https://github.com/mizzky/sol/issues/75) - request_id ミドルウェア実装
- [#77](https://github.com/mizzky/sol/issues/77) - slog 初期化と ReplaceAttr（Issue A の実装）

### 内部ドキュメント
- [doc/task.md](../task.md) - 全体的な実装計画（構造化ロギング実装（#60対応）セクション）
- [doc/memo/2026/05/20260501_learning.md](../memo/2026/05/20260501_learning.md) - ログ設計書深掘り（Issue #74 設計過程）
- [doc/memo/2026/05/20260502_learning.md](../memo/2026/05/20260502_learning.md) - ログレベル運用定義（Issue #65 設計過程）

### Go 標準ドキュメント
- [Go Package slog](https://pkg.go.dev/log/slog) - Go 1.21 標準ロギングライブラリ
- [JSONHandler ReplaceAttr](https://pkg.go.dev/log/slog#JSONHandler.ReplaceAttr) - マスキング実装方法

### 実装コード参照
- `backend/pkg/apperror/` - カスタムエラー型定義
- `backend/middleware/error_handler.go` - 拡張対象のメインハンドラ
- `backend/middleware/request_id.go` - request_id ミドルウェア（既実装）
- `backend/auth/middleware.go` - 認証ミドルウェア（user_id Context セット済み）

---

## 13. 附録：実装時チェックリスト

### Issue A（実装完了）
- [x] slog JSONHandler の初期化
- [x] ReplaceAttr で password/token/email をマスキング
- [x] main.go の起動時エラー処理を slog.Error() で統一
- [x] ユニットテスト（ReplaceAttr の動作確認）

### Issue B（次フェーズ）
- [ ] ErrorHandler で error_type を判定
- [ ] ログレベル決定ロジック実装（INFO/WARN/ERROR 分岐）
- [ ] `event`, `error_type`, `status`, `method`, `route` を slog に出力
- [ ] ユニットテスト（各エラー型のログ出力確認）

### Issue C（次々フェーズ）
- [ ] Context から `request_id` を取得
- [ ] slog に `request_id` フィールド追加
- [ ] `request_id` なしでもパニックしない（nil チェック）
- [ ] ユニットテスト（request_id の付与確認）

### Issue D（その後）
- [ ] Context から `userID` キーで user_id を取得
- [ ] slog に `user_id` フィールド追加（存在時のみ）
- [ ] 未認証時もパニックしない
- [ ] ユニットテスト（user_id の付与確認、未認証時の動作確認）

### Issue F（その後）
- [ ] ハンドラ層で `slog.Info()` でイベント出力
- [ ] event 名を snake_case で統一
- [ ] request_id, user_id, duration_ms をイベントログに含める
- [ ] 同一イベントの重複出力がないか確認
- [ ] ユニット・統合テスト（3 つの正常系フロー確認）

### Issue E（将来 Phase 2）
- [ ] 共通 redaction パッケージ作成
- [ ] apperror と slog が同じマスキング結果を返す
- [ ] キー追加時の拡張性をテストで確認

---

**ADR 作成日**: 2026-05-14  
**次回レビュー予定**: Issue B 着手時（2026-05-20 目標）  
**責任者**: ユーザー（実装 Mentor）
