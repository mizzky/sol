# 2026-03-31 Learning Log

## 今日の学習テーマ
- 注文履歴の明細に `product_name_snapshot` が正しく表示されない不具合対応（mizzky/sol#49）

## 実施内容
- `backend/query.sql` の `ListOrderItemsByOrderID` に `product_name_snapshot` を追加
- `sqlc generate` を実行して `backend/db/query.sql.go` を更新。結果的に `ListOrderItemsByOrderIDRow` が `OrderItem` に統一された
- `backend/handler/order_test.go` に注文履歴APIのJSON検証ケース（U10）を追加し、先に Red（コンパイルエラー）を確認後、Green で修正してテストを通した
- テストやモック内の関連参照を `OrderItem` に置換して `go test ./...` を実行し全テストパスを確認

## 実行したコマンド
```bash
cd backend
sqlc generate
go test ./...
```

## 学び・気づき
- `sqlc` は返すカラムがテーブルの全カラムと一致するとテーブルのモデル型（`OrderItem`）を流用する。部分列だと専用の `...Row` 型が生成される
- TDD（Red→Green）で「まず失敗させる」ことで、必要な修正箇所を最小限に絞りやすかった

## 次の作業
1. 変更をコミットしてブランチ作成 → PR と CI 実行
2. フロント側で商品名が正しく表示されるか手動または E2E で確認
3. 必要なら issue に修正内容を追記（対応済みである旨）

---
記録者: セッション
作成日時: 2026-03-31
