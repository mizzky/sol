# 2026-03-19 Learning Log

## 今日の学習テーマ
- CancelOrderHandler の異常系テストをテーブル駆動で設計・実装する
- テストケースの分類（正常系・異常系・準正常系）の考え方を整理する
- integration test における seed 関数の設計方針を学ぶ

## 実施内容
- CancelOrderHandler 異常系の統合テストを `TestCancelOrderHandler` としてテーブル駆動で実装
  - 対象ケース: 他人の注文 / 注文が存在しない / 既にキャンセル済み / 未認証 / 無効な注文ID / userID が int/float64
  - HTTP ステータス・エラーメッセージ・DB 副作用（orders.status, products.stock_quantity）を検証
- 異常系 seed として `seedOthersOrderForCancel` / `seedCancelledOrderForCancel` を個別に作成
- ループ本体で `expectedErrMsg` の JSON 検証と `assertDB` の呼び出しを追加
- CreateOrderHandler 側のケース名も誤分類を修正
  - `異常系：userIDがintでも通る` → `正常系：userIDがintでも通る` （2件）

## 実行したコマンド
- `go test -tags=integration ./tests -run TestCancelOrderHandler -v` → PASS
- `go test ./...` → PASS

## 学び・気づき

### seed 設計方針
- 汎用 seed（引数でパターンを切り替える）は再利用性が高いが、テストの可読性が落ちる
- 個別 seed は「何を準備するか」が一目でわかり、保守がしやすい
- 今回は可読性優先で個別 seed を選択した（正解は状況次第で変わる）

### テストケースの分類
- **正常系**: 仕様として受け入れる入力で成功するケース
- **異常系**: 仕様として拒否する入力・状態のケース
- userID が int / float64 のケースはハンドラが明示的に `switch case` で受け入れているため「正常系」に分類するのが正しい
- 判断基準: 「その入力を仕様として受け入れるようコードに書いてあるか」

### nil 関数呼び出しによる panic
- テーブル駆動テストで `pathBuilder` を設定せずにループを回すと `nil pointer dereference` で panic になる
- 未完成のケースを一時的に置くときは `t.Fatalf("pathBuilder is nil: %s", tt.name)` などのガードを入れると安全

### HTTP ステータスの使い分け
- 400 Bad Request: リクエスト自体が不正（例: orderID が数値でない）
- 404 Not Found: 形式は正しいがリソースが存在しない（例: 存在しない orderID）
- 注文が存在しない / 他人の注文を「どちらも 404」にする理由は、存在有無を外部に教えないための設計
- 既に cancelled は 400。将来ステータスが増えた場合は 409 Conflict も候補になりうる

### go mod tidy の permission denied
- `/go/pkg/mod` の所有者が root になっていると `permission denied` が発生する
- 対処: `sudo chown -R $(id -u):$(id -g) /go/pkg/mod` か、GOPATH/GOMODCACHE を HOME 配下に移す

## 詰まったポイント

### seedOthersOrderForCancel の `RETURNING id` 抜け
- `INSERT INTO products ... VALUES ...` に `RETURNING id` を書き忘れたため `sql: no rows in result set` が発生
- Scan に渡す変数があるのに RETURNING がないとこのエラーになる

### 異常系：注文が存在しない の pathBuilder が nil
- テーブル定義だけ追加して中身を埋めずにループを回したため panic になった
- 空のケースで panic せずに失敗メッセージを出すには `pathBuilder == nil` ガードが有効

## 次の作業（優先順）
1. userID が不正型（string 等）のケースを cancel 側にも 1 件追加する（401 確認）
2. `doc/task.md` の CancelOrderHandler 進捗を更新する
3. コミット: `test(integration): add error case tests for CancelOrderHandler`

## 参考ファイル
- `backend/tests/order_integration_test.go`
- `backend/handler/order.go`
- `backend/tests/sql_assert_helper_test.go`

---
記録者: 開発作業ペア（対話ログ）
作成日時: 2026-03-19
