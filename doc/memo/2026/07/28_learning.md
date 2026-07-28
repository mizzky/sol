# 学習記録 2026-07-28

## セッション 1 (11:31)

### 取り組んだタスク

- `rebuild_profile.sh`でmedium・largeデータセットを再構築し、`ListProducts`のbefore計測を各3回実施した。
- mediumでは100,000行を`products_pkey`の`Index Scan`で取得し、Execution Time中央値25.769 ms、実行Buffers `shared hit=2,403`を記録した。
- largeでは1,000,000行を同じ`Index Scan`で取得し、Execution Time中央値359.847 ms、実行Buffers `shared read=24,012`を記録した。
- small・medium・largeで推定rowsと実測rowsが一致し、Scan typeが変わらないことを確認した。
- mediumはsmallに対してrowsが10倍、Buffersが約9.93倍、Execution Time中央値が約9.39倍になった。
- largeはmediumに対してrowsとBuffersが約10倍、Execution Time中央値が約13.96倍になった。
- `SHOW shared_buffers`で共有バッファが128MBであることを確認し、largeの約187.6MiB相当のブロックアクセスを保持しきれず、繰り返し実行でも`shared read`が発生したと整理した。
- small・medium・largeの計測値と共有バッファに関する結果を`doc/perf/products.md`へまとめた。

### ユーザーが質問した内容

- mediumの商品数が10倍になった場合、Execution TimeやBuffersも単純に10倍になるのか。
- `BUFFERS`は何を計測しているのか。
- 全件取得ではインデックスがあっても仕事量が減らないのか、クエリ改善で10倍の増加を避けられるのか。
- `shared_buffers=128MB`に対し、largeだけ`shared read=24,012`になった理由。
- `shared_buffers`を256MBへ増やした場合、`shared hit`が増えて高速化を見込めるか。
- 共有バッファは複数クエリで発生するすべてのI/Oを扱うのか。
- Oracleで重いクエリを同時実行すると、ほかのクライアントのサービスまで遅くなる事例と共有資源の関係。
- `products.md`の計測記録として必要な粒度とコミットメッセージ。

### 躓いたポイントと解決策

- インデックスが使われていればデータ増加の影響を抑えられるのかが曖昧だった。今回のIndex Scanは`ORDER BY id`のSortを省略するが、全件を返す行数は減らさないため、基本的な仕事量は`O(N)`のままだと整理した。
- mediumの約10倍という結果を単純な時間比例だけで捉えかけた。rows、Buffers、Execution Timeを分けて比較し、読み取るブロック数もデータ量にほぼ比例していることを確認した。
- largeではウォームアップ後も`shared read`が続いた。24,012ブロックを8KB換算した約187.6MiBが`shared_buffers=128MB`を超えており、共有バッファへ保持しきれず再読込が発生したと判断した。
- `shared read`を物理ディスクアクセスと同一視しないよう整理した。共有バッファにはないが、OSのファイルキャッシュから読み込まれた可能性もある。
- `shared_buffers`を増やせば問題が解決するように見えた。反復実行は高速化する可能性がある一方、1,000,000行の走査・転送・変換は残るため、根本的にはpaginationで1リクエストの取得件数を制限する必要がある。
- 共有バッファをDBの全I/Oと捉えかけた。テーブル・インデックスの共有キャッシュであり、一時ファイル、WAL、checkpoint、OSキャッシュなどは別に存在すると整理した。
- 計測記録の説明を厳密にしすぎると読みづらくなった。明確な誤りだけ修正し、ユーザー自身の表現を優先する方針へ戻した。

### 次回課題

- `doc/perf/products.md`のmedium・large計測結果を`docs(perf): record medium and large products measurements`でコミットする。
- Phase 3の3-2「商品一覧before計測」を完了とし、3-3「注文一覧before計測」へ進む。
- 現行のユーザー別注文一覧SQLを特定し、OFFSET `0 / 1,000 / 10,000 / 100,000`で変化する処理を実行前に予想する。
- 注文一覧用の計測SQLを作成し、small・medium・largeでScan type、rows、Buffers、Execution Timeを保存する。
