# 学習記録 2026-03-23

## 取り組んだタスク
- **TDD Step 1-6 完了**: `GetOrdersHandler` の Red-Green-Refactor サイクル全実装
  - Step 1: `getOrderLogic()` 実装（4テストケース、すべてGreen）
  - Step 2: GetOrdersHandler Red テスト設計（5ケース → 8ケースに拡張）
  - Step 3: テストコード作成と Red 状態確認
  - Step 4: ハンドラー実装（Green、全8テスト合格）
  - Step 5: エッジケース追加（0件フィルタ検証）
  - Step 6: コード品質レビュー（実装品質確認）

## ユーザーが質問した内容

### 設計・責務分離に関する質問
1. **Status フィルタの責務**: URL パラメータの status フィルタはどの層が担うのか？
   - 判定：Handler 層（HTTP 層）で実装、Logic 層は DB Query 実行のみ
   
2. **テストファイル位置**: `order_test.go` を `package handler` か `package handler_test` か？
   - 判定：`package handler` に統一（Logic UT と Handler UT の両方を管理）
   
3. **userID 型の柔軟性**: Context から取得した userID がどの型でも対応すべき？
   - 判定：int64/int/float64 すべて対応（type assertion で処理）

### コード実装に関する質問
4. **Status バリデーション方式**: ハードコード vs Map+ヘルパー関数？
   - 判定：`validOrderStatuses` Map + `isValidOrderStatus()` 関数を採用（保守性向上）

5. **Handler 引数**: `GetOrdersHandler(queries)` のみで Tx は不要？
   - 判定：読み取り専用なので Tx 不要（CreateOrderHandler と異なる設計）

6. **DB エラーの重複テスト**: Logic UT と Handler UT で DB エラーをテストすべき？
   - 判定：Logic UT で実施、Handler UT は HTTP 変換部分のみテスト（責務分離）

## 躓いたポイントと解決策

### 1. 責務分離の曖昧性
**問題**: Status フィルタ実装をどの層で行うか、テスト構成をどうするか不明確

**原因**: TDD 初実装で層分離の概念が未確立

**解決策**:
- Handler = HTTP Protocol → Go 値への変換責務（Status 検証含む）
- Logic = DB Query 実行と結果集約責務（フィルタは Handler から指示）
- テストは責務に応じて分割（Logic UT: 4ケース、Handler UT: 8ケース）

### 2. テスト設計の重複チェック
**問題**: Handler UT と Logic UT で DB エラーをテストすべきか曖昧

**原因**: テスト層間の責務が未定義

**解決策**:
- Logic UT で DB failures 実装（`ListOrdersByUser` error, `ListOrderItemsByOrderID` error）
- Handler UT は HTTP layer concerns に集中（Status validation, userID auth, JSON serialization）
- 各層が担うテスト責務を明確化

### 3. エッジケース見落とし
**問題**: Status フィルタで全件がフィルタされた場合（0件） のテストが初期設計になかった

**原因**: Happy-path 中心のテスト設計

**解決策**:
- Test U4 追加: pending な注文のみ存在 → `status=cancelled` で検索 → 0件配列返却
- 空配列の JSON シリアライズ確認

## 技術的な学び

### 1. TDD サイクルの実装

**Red Phase**:
```go
// テスト先行（失敗を確認）
func TestGetOrdersHandler(t *testing.T) {
    // Case: U1 filter pending → 1 件取得
    // Case: U2 filter cancelled → 0 件取得
    // ...etc
}
```

**Green Phase**:
```go
// 最小限の実装でテストをパス
func GetOrdersHandler(queries db.Querier) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Auth check (userID)
        // 2. Parse status param
        // 3. Call getOrderLogic
        // 4. Filter by status (in-memory)
        // 5. Return JSON
    }
}
```

**Refactor Phase**:
```go
// 保守性向上（Map + Helper 関数）
var validOrderStatuses = map[string]struct{}{"pending":{}, "cancelled":{}}
func isValidOrderStatus(status string) bool {
    if status == "" { return true }
    _, ok := validOrderStatuses[status]
    return ok
}
```

### 2. Go テストの Table-Driven パターン

```go
type test struct {
    name            string
    query           string
    userID          interface{}
    setupMock       func(*testutil.MockDB)
    expectedStatus  int
    expectedCount   int
    expectedErrMsg  string
}

tests := []test{
    {
        name: "U1: filter=pending で pending 注文が返却される",
        query: "?filter=pending",
        userID: int64(1),
        setupMock: func(m *testutil.MockDB) {
            m.On("ListOrdersByUser", ...).Return(...)
            m.On("ListOrderItemsByOrderID", ...).Return(...)
        },
        expectedStatus: 200,
        expectedCount: 1,
    },
    // ...
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // テスト実装
    })
}
```

### 3. 設計の一貫性チェック

**既存コード参照**:
- `cart_test.go`: userID をテスト内で inline 設定（middleware 不使用）
- `cart.go`: Handler が Tx 管理（create は Tx 必須）
- `order.go`: ReadOnly → Tx 不要

**決定基準**:
- Create/Update/Delete: Tx 必須、Handler 層で Tx 管理
- Read: Tx 不要、Querier インタフェースのみ
- テスト userID 設定: inline style（cart_test.go に統一）

## 次回課題

### 優先度・高
1. **Ticket 8: GetOrdersHandler をルーティング登録**
   - routes.go に `GET /api/orders` エンドポイント追加
   - RequireAuth middleware 適用
   - test.http で動作確認
   - リクエスト例:
     ```
     GET /api/orders?filter=pending
     ```

2. **マイナーリファクタ**: テスト名の重複 U4 → U5, U6... に修正

### スキルの定着確認
- ✔ TDD Red-Green-Refactor サイクルを実装レベルで理解
- ✔ 層間の責務分離（Handler vs Logic）を実装で示している
- ✔ テーブル駆動テスト、MockDB pattern を習得
- 🔄 次フェーズ: Router 統合テスト（エンドツーエンド確認）

### 継続学習
- 注文 Create/Cancel ハンドラーのテスト設計パターンも同様か確認
- POST/PUT メソッドの Tx パターンを CartHandler と比較
- エラーレスポンス形式の統一化を確認

---
**セッション時間**: ~2時間  
**実装行数**: ~150行（Handler + Tests + Mock）  
**テスト結果**: 21/21 GREEN ✅
