# 2026-03-21 Learning Log

## 今日の学習テーマ
- `GET /api/orders`（注文履歴一覧）と `GET /api/orders/:id`（注文詳細）の責務分離
- 所有権チェックをどの層で担保するか（アプリ層 vs DB層）
- 既存 sqlc クエリで実現できる設計範囲の整理

## 実施内容
- `GET /api/orders` の設計フローを確認
  - `RequireAuth` で `userID` を取得
  - `ListOrdersByUser(userID)` で自分の注文一覧を取得
  - 各 `order_id` に対して `ListOrderItemsByOrderID(orderID)` で明細を取得
  - `200` で `orders` を返す
- `GetOrderByID(id int64)` の `id` は `orderID` であり、戻り値に `user_id` が含まれる点を確認
- 一覧APIでは `ListOrdersByUser(userID)` の時点で所有者制約がかかるため、
  `GetOrderByID` の再チェックは通常不要（N+1 増加の懸念）と整理

## 学び・気づき
- `GET /api/orders` は「注文履歴」APIとして扱うのが自然
- 所有権チェックの考え方
  - 一覧API: `WHERE user_id = $1` でDB側フィルタ済みなら追加チェック不要
  - 詳細API: `orderID` 指定のため所有権チェック必須
- 現在のクエリ構成でもアプリ層で所有権チェックは可能
  - `GetOrderByID(orderID)` で取得した `user_id` を認証ユーザーと比較
- ただし DB層で強制したい場合は専用クエリ（例: `WHERE id = $1 AND user_id = $2`）が必要

## 詰まったポイント
- 一覧APIでも `orderID` 単位の所有権チェックが必要かどうかで混乱しやすい
- 「アプリ層で十分」なのか「DBで強制するべき」なのかは要件（防御の深さ）で決まる

## 次の作業（優先順）
1. チケット5（`GetOrdersHandler`）のテスト設計を先に確定する（正常系/空一覧/認証なし/フィルタ）
2. 一覧APIのレスポンスで `items` を含めるかを仕様として固定する
3. 詳細API（`GET /api/orders/:id`）を後続で作る場合の所有権チェック方針を明文化する

## 参考ファイル
- `backend/query.sql`
- `backend/db/querier.go`
- `backend/handler/order.go`
- `doc/task.md`

---
記録者: 開発作業ペア（対話ログ）
作成日時: 2026-03-21
