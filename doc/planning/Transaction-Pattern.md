# Transaction Handler Pattern

作成日: 2026-03-11
対象タスク: doc/task.md チケット 2-1

## 目的

注文系ハンドラで必要になるトランザクション処理を、実装者ごとにばらつかない形で統一する。

このドキュメントでは以下を決める。

- handler から `BeginTx` を開始する責務
- `Queries.WithTx(tx)` の使い方
- ユニットテストと統合テストの責務分担
- チケット 3-5 で採用する実装パターン

## 結論

現時点では Service 層は新設せず、handler 内でトランザクションを開始する。

理由は以下の通り。

- 既存コードは handler が `db.Querier` を直接受け取る構成で揃っている
- 注文系だけのために大きな層追加を行うと、学習コストに対して差分が大きい
- `sqlc` が生成する `Queries.WithTx(tx)` をそのまま活用できる

ただし、`BeginTx` と `Commit/Rollback` の記述は重複させず、小さなヘルパーで共通化する。

## 採用パターン

### 1. 依存の渡し方

注文系 handler は以下 2 つを受け取る。

- `*sql.DB`: トランザクション開始元
- `*db.Queries`: `sqlc` の通常クエリ実行元

既存の `db.Querier` だけでは `BeginTx` を呼べないため、注文系だけは明示的に `*sql.DB` も注入する。

推奨シグネチャ:

```go
func CreateOrderHandler(conn *sql.DB, queries *db.Queries) gin.HandlerFunc
func CancelOrderHandler(conn *sql.DB, queries *db.Queries) gin.HandlerFunc
func GetOrdersHandler(queries db.Querier) gin.HandlerFunc
```

`GetOrdersHandler` は更新処理を持たないため、既存パターン通り `db.Querier` だけでよい。

### 2. トランザクションヘルパー

handler から直接 `BeginTx` と `Rollback` を毎回書くのではなく、同一ファイル内または小さな補助関数で包む。

```go
func runInTx(ctx context.Context, conn *sql.DB, queries *db.Queries, fn func(*db.Queries) error) error {
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	qtx := queries.WithTx(tx)
	if err := fn(qtx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return errors.Join(err, rbErr)
		}
		return err
	}

	return tx.Commit()
}
```

この形にすると、注文作成・キャンセルの両方で以下を統一できる。

- `BeginTx` の開始位置
- `WithTx` の再バインド
- エラー時の `Rollback`
- 正常時の `Commit`

### 3. handler 内の責務分離

handler 内では責務を次の順に置く。

1. 認証済み `userID` の取得
2. リクエストバリデーション
3. `runInTx(...)` 開始
4. Tx 内での DB 読み書き
5. Tx 成功後にレスポンス整形

この順にすると、HTTP 層と Tx 層が混ざりにくい。

## チケット 3: CreateOrderHandler の実装ガイド

### 処理フロー

```text
1. userID を context から取得
2. runInTx を開始
3. カート取得
4. カート item 一覧取得
5. 商品 ID 昇順で FOR UPDATE 取得
6. 在庫確認
7. 不足があれば 409 用エラーを返して rollback
8. orders 作成
9. order_items 作成
10. products.stock_quantity を減算
11. cart_items を削除
12. commit
13. 201 を返す
```

### 注意点

- `FOR UPDATE` を使う商品取得順は ID 昇順で統一する
- `Commit` 前に HTTP レスポンスを書かない
- 409, 404, 400 を表現する業務エラーは、Tx 内で返し、handler 外側で HTTP に変換する

## チケット 4: CancelOrderHandler の実装ガイド

### 処理フロー

```text
1. userID を context から取得
2. orderID を path param から取得
3. runInTx を開始
4. 対象注文を FOR UPDATE 取得
5. 所有者確認
6. status が pending か確認
7. order_items を取得
8. 各 product の在庫を加算
9. order status を cancelled に更新
10. commit
11. 200 を返す
```

### 注意点

- 非所有と不存在は、既存方針に合わせて `404 Not Found` に寄せる
- `pending` 以外のキャンセルは `400 Bad Request`
- 在庫巻き戻しも同一 Tx に含める

## テスト戦略

## 1. ユニットテストで担保すること

`MockDB` を使うユニットテストでは、以下を中心に確認する。

- 認証なしで 401
- path param / request body のバリデーション
- DB から返った業務エラーを適切な HTTP ステータスへ変換できること

## 2. 統合テストで担保すること

以下は `MockDB` だけでは十分に検証できないため、実 DB の統合テストで確認する。

- `BeginTx` から `Commit/Rollback` までの一連の流れ
- `FOR UPDATE` による排他制御
- 在庫減算/巻き戻しの原子性
- 失敗時に途中状態が残らないこと

## 3. MockDB の制約

現在の `backend/handler/testutil/mockdb.go` は `db.Querier` の差し替えに向いている一方、以下は扱いにくい。

- `*sql.DB.BeginTx(...)`
- `*sql.Tx.Commit()`
- `*sql.Tx.Rollback()`

そのため、チケット 3-4 の本質的な受け入れ条件は統合テストで担保する。

## 実装判断

今回のフェーズでは、以下を採用する。

- 採用する: handler 内 Tx 開始 + `runInTx` ヘルパー
- 採用しない: 新しい Service 層の導入
- 採用しない: Tx 自体を無理にモックするための大きな抽象化

理由は、まず注文機能を最小差分で前進させ、その後に重複が明確になったら抽象化を検討する方が安全だからである。

## ルーティング変更方針

注文系 handler を登録する際は、`routes.SetupRoutes` が `*sql.DB` も受け取れるように変更する。

イメージ:

```go
func SetupRoutes(r *gin.Engine, conn *sql.DB, queries *db.Queries) {
	api := r.Group("/api")
	api.POST("/orders", auth.RequireAuth(queries), handler.CreateOrderHandler(conn, queries))
	api.POST("/orders/:id/cancel", auth.RequireAuth(queries), handler.CancelOrderHandler(conn, queries))
	api.GET("/orders", auth.RequireAuth(queries), handler.GetOrdersHandler(queries))
}
```

## チケット 3 開始前チェックリスト

- [ ] `CreateOrderHandler` の request/response 仕様を最終確認した
- [ ] 統合テストで使う注文用テストデータの投入方法を決めた
- [ ] 409 用の業務エラー表現を決めた
- [ ] 商品ロック順を ID 昇順で統一すると合意した
- [ ] `routes.SetupRoutes` の引数変更影響を把握した
