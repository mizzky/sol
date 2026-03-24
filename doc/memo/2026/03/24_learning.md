# 学習記録 2026-03-24

## 取り組んだタスク
- 注文作成の同時性テスト（チケット6）を実装した。
- 少量並列（5ユーザー）と大量並列（50ユーザー）の2ケースをテーブル駆動で検証した。
- goroutineごとにrouterを作成し、userIDの注入を分離した。
- readyチャネルで一斉開始、MutexでstatusCodesへの同時書き込みを保護した。
- 期待値を固定値ではなく導出式（成功数、競合数、残在庫）に整理し、integration testでGreenを確認した。

## ユーザーが質問した内容
- FOR UPDATEで行ロックできる理解で正しいか。
- handlerはユーザー数分呼ぶ必要があるか。
- goroutineごとにrouterを作るべきか。
- 大量テストケースで失敗する原因は何か。
- 期待値の書き方（success/conflict/remaining stock）はどれがよいか。
- ループ変数キャプチャでハマる理由は何か。
- <- 演算子の意味は何か。

## 躓いたポイントと解決策
- 躓き: seed関数でScan引数やappend手順に不備があり、ID管理が崩れた。
  - 解決策: newUserIDへScanしてからappendし、参照順を修正した。
- 躓き: 大量ケースの在庫アサートが固定値で失敗した。
  - 解決策: 残在庫を productQty % cartQty で導出する形に変更した。
- 躓き: 並列時のuserID混線懸念。
  - 解決策: goroutine引数で値を渡し、goroutineごとにrouterを生成してクロージャの意図を明確化した。
- 躓き: statusCodesへの同時appendでデータ競合リスク。
  - 解決策: Mutexで臨界区間を保護した。

## 次回課題
- チケット6完了としてtaskの進捗を更新する。
- チケット7（エラーハンドリング・エッジケース）に着手する。
- goroutine学習の継続として、channel集約版（Mutexを使わない結果収集）も試す。
