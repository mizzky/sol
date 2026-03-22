# 学習記録 2026-03-22

## 今日のテーマ
`GetOrdersHandler` 実装に向けたTDD設計と実装（Ticket 5）

---

## 学んだこと・気づき

### 1. Go 構造体タグの文法
**詰まったポイント：** `json:"order` のように閉じダブルクォートが抜けてコンパイルエラー

```
struct field tag `json:"order` not compatible with reflect.StructTag.Get: bad syntax for struct tag value
```

**理解したこと：**
- バッククォート（`` ` ``）はタグ全体を囲む → フィールドの末尾に付ける
- ダブルクォート（`""`）は各値を囲む → キーと値のセット分必要

```go
// 正しい形
type OrderWithItems struct {
    Order db.ListOrdersByUserRow          `json:"order"`
    Items []db.ListOrderItemsByOrderIDRow `json:"items"`
}
```

---

### 2. sqlc 自動生成コードの JSON タグ
sqlc が生成した Row 構造体にはすでに `json:"..."` タグが付いている。
自分でラッパー構造体 `OrderWithItems` を定義したときだけタグを明示する必要がある。

```go
type ListOrdersByUserRow struct {
    ID        int64     `json:"id"`
    UserID    int64     `json:"user_id"`
    // ... sqlc が自動生成
}
```

---

### 3. `[]OrderWithItems` のテスト検証パターン
**詰まったポイント：** `getOrderLogic` が `[]OrderWithItems` を返すのに `order.ID` や `item.ID` で直接アクセスしようとした（スコープエラー）

**正しいアクセス方法：**
```go
assert.Len(t, owi, 1)                          // スライスの長さで注文件数を検証
assert.Equal(t, int64(1), owi[0].Order.ID)     // スライスのインデックスでアクセス
assert.Len(t, owi[0].Items, 1)                 // 明細件数
assert.Equal(t, int64(1), owi[0].Items[0].ID)  // 明細のフィールド
```

**ポイント：** スライスを返す関数のテストは「スライスの長さ」から検証する習慣をつける

---

### 4. レイヤー責務の整理（再確認）

| 層 | 責務 |
|---|---|
| `getOrderLogic` | DB クエリの実行、`OrderWithItems` 構造体の構築 |
| `GetOrdersHandler` | 認証チェック、クエリパラメータ解析、ステータスフィルタ、レスポンス返却 |

ビジネスロジック層は「何を取得するか」、ハンドラ層は「どう返すか」という責務の分離が重要。
ステータスフィルタはハンドラ層でメモリフィルタとして実装する。

---

### 5. `context.CancelCauseFunc` vs `context.Context` の誤用
**詰まったポイント：** 関数シグネチャを書くときに `context.CancelCauseFunc` と書いてしまった

```go
// ❌ 間違い
func getOrderLogic(ctx context.CancelCauseFunc, ...) 

// ✅ 正しい
func getOrderLogic(ctx context.Context, ...)
```

`context.CancelCauseFunc` はキャンセル関数の型であり、コンテキスト渡しには `context.Context` インターフェースを使う。

---

## 次のステップ
- テストコードの検証部分（else ブロック）を `[]OrderWithItems` アクセスに修正
- テストが Red になることを確認
- `getOrderLogic` の実装（Green フェーズ）
- `GetOrdersHandler` の実装とハンドラテスト
