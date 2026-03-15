# 学習ログ 2026-03-15

## テーマ: CreateOrderHandler の統合テスト実装

---

## やったこと

- 正常系の統合テスト（`TestCreateOrderHandler_HappyPath`）に `t.Cleanup` を追加してテスト分離を実現
- 異常系のテーブル駆動テスト（`TestCreateOrderHandler`）を実装
  - 未認証
  - userID が不正な型
  - カートが空
  - 在庫不足
  - userID が int / float64 でも通ることの確認
- `assertDB` 関数をケースに持たせ、HTTP ステータスだけでなく DB 状態も検証する形にした
- GitHub Issue にアーキテクチャ課題（TxBeginner インターフェース化）を起票

---

## 疑問と解決

### `wantStatus` vs `expectedStatus`

- `want` / `got` は Go 標準ライブラリやOSSで慣用的なパターン
- ただし自プロジェクトの既存テストが `expectedStatus` で統一されていれば、**一貫性を優先する方が正しい**
- → このプロジェクトでは `expectedStatus` を採用

### `isAuth` vs `setUserID`

- `isAuth` は「状態を表すbool」として自然
- `setUserID` は「動作」のように読める → 関数名向き
- 既存 UT の `userID interface{}` パターンに合わせるか、`bool` にするかはプロジェクト方針次第
- 統合テストでは `rawUserID interface{}` を採用（型分岐も同時に検証できる形にした）

### 統合テストでの正常系・異常系の分け方

- cart の UT → 1ケースあたりの検証コストが均一 → 正常系・異常系まとめてテーブル駆動が読みやすい
- order の統合テスト → 正常系だけ副作用確認（orders / order_items / stock / cart_items）が多い → 単独テストにした方が意図が明確
- **判断基準: テストの責務と密度の違いに応じて形式を変える。テスト駆動で全部を同じ形に揃えることが目的ではない**

### ロールバックのテストについて

- 現在の `createOrderLogic` はバリデーション失敗でエラーを返す（書き込みがそもそも起きていない）
- 本当のロールバックテストには「途中まで書き込んだ後に失敗する」状態が必要
- 現状のアーキテクチャでは `conn` が `*sql.DB`（具体型）のためUTでモック化できない
- → アーキテクチャ課題として Issue に起票し、今は統合テストの assertDB で代替

### テストのアーキテクチャ課題

- `CreateOrderHandler` は `*sql.DB` を直接受け取って内部で `BeginTx` を呼んでいる
- `tx.Rollback()` が呼ばれることを UT で検証する手段がない
- 対応案: `TxBeginner` インターフェースを定義して依存を抽象化する

```go
type TxBeginner interface {
    BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}
```

---

## バグ指摘・気づき

- `t.Cleanup` 内の `TRUNCATE` に **スペース抜け（`RESTARTIDENTITY` → `RESTART IDENTITY`）** があり、クリーンアップが無音で失敗していた  
  → エラーを `_` で捨てているため実行時に気づけない。統合テストでは Cleanup 内のエラーも `t.Errorf` で出すべき

- `assertDB` 内で `Scan` の引数を誤った変数 `&orderCount` に渡していた  
  → `orderItemCount` が常にゼロ値のままアサートが通ってしまう。**型が合えばコンパイルエラーにならないので見落としやすい**

- `rawUserID` が `"not-a-number"` のケースで `c.Set("userID", userID)` と書いており、実際には `int64` が渡っていた  
  → `c.Set("userID", rawUserID)` が正しい。型分岐を検証するケースでは **Set に渡す値が rawUserID であることを必ず確認する**

---

## 次のアクション

- `TxBeginner` インターフェース化（Issue 起票済み）
- 現在の統合テストでカバーできていない「途中書き込み後のロールバック」の検証方針を決める
