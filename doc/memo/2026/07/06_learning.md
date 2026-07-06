# 学習記録 2026-07-06

## セッション 1 (10:55)

### 取り組んだタスク

- Issue #62 Phase 1のうち、1-3「users・categories投入」をTDDで完了した。
- `generate_series`を使ってusers 1,000件、categories 10件を決定的に生成した。
- 明示投入したIDに合わせて、usersとcategoriesのsequenceを最大IDへ同期した。
- reset、seed、verifyを2回繰り返し、件数、固定値、一意性、sequenceが毎回同じになることを確認した。
- 1-4「products投入」をTDDで完了した。
- smallプロファイル向けにproducts 10,000件を生成し、販売可否、在庫、カテゴリ、価格へ決定的な分布を持たせた。
- 商品検証SQLで、ID範囲、固定命名、価格、販売可否、在庫、カテゴリ、SKU重複、外部キー孤児、日時、sequenceを検証した。
- 商品についてもresetからの投入と検証を2回繰り返し、再現性を確認した。

### ユーザーが質問した内容

- sequence同期とは何か。
- `INSERT ... SELECT`では採番が進まず、`INSERT ... VALUES`では進むのか。
- `generate_series(1, 1000)`は1から1,000までの連番を意味するのか。
- 商品の販売可否を`product_id % 20`で分岐するとき、`IF`と`CASE`のどちらを使うのか。
- `((n - 1) / 20) % 10 AS stock_bucket`と`(n - 1) % 100 AS category_bucket`が何を表すのか。
- 在庫範囲を作るCASE式と、`1 + ((n - 1) % 5)`などの剰余式をどう読むのか。
- productsのseed SQLで、INSERT列とSELECT値をどの順番で対応させるのか。
- products用のverify SQLでは何を検証すべきか。

### 躓いたポイントと解決策

- `02_users_categories.sql`を実行してもusersが0件のままだった。原因はファイルが0バイトで、`psql`が何も処理せず正常終了していたことだった。ファイル内容とサイズを確認し、seed SQLを保存してから再実行して解決した。
- IDを明示して投入しても、`BIGSERIAL`のsequenceは自動では進まないことが分かりづらかった。採番の有無は`SELECT`と`VALUES`の違いではなく、ID列を省略して`nextval`を呼ぶかどうかで決まると整理した。明示投入後は`setval`でsequenceを最大IDへ同期した。
- stock bucketの式が読みづらかった。整数除算で20商品ずつまとめ、`% 10`でbucket 0〜9を200商品周期で繰り返す処理に分解した。
- category bucketは`(n - 1) % 100`で0〜99を作り、CASEで40% / 20% / 残り8カテゴリ各5%へ割り当てる処理だと整理した。
- 在庫値の生成式が複雑に見えた。`最小値 + 剰余`という共通形に分解し、1〜5は5種類、10〜100は91種類、500〜1,000は501種類だと確認した。
- products seedの作成時に、CTE末尾の余分なカンマ、INSERT先11列とSELECT値の不足・順序ずれ、`FROM product_source`の位置、誤った`ORDER BY NOT`が混在した。INSERT列とSELECT値を上から1対1で対応させ、`SELECT ... FROM ... ORDER BY ...;`の後にsequence同期と`COMMIT`を置いて解決した。
- products verifyは商品未投入時に件数0、min/maxがNULLとなって失敗した。これは意図したRedであり、seed後に`PASS: products`となってGreenを確認できた。

### 次回課題

- 1-4の2ファイルが未コミットなら、`feat(perf): add deterministic product seed`でコミットする。
- 1-5「carts・cart_items投入」へ進み、cart 500件、cart_items 1,594件の決定的な生成規則と検証条件をTDDで実装する。
- cart 1を空、cart 2を100明細、残りを各3明細とし、`(cart_id, product_id)`の一意性、外部キー、quantity、priceの整合性を確認する。
