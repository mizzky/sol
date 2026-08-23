## このタスクでやること

- UNIQUE制約由来のインデックスと既存インデックスについて、調査・比較を行う

### 現在存在するインデックス一覧

以下のクエリを実行して、現在存在する制約由来・明示作成のインデックスの洗い出しを行った

`backend/perf/sql/queries/index_inventory.sql`

ユニーク由来でpostgresによって自動生成されたインデックスと完全に重複するインデックスを削除することで軽量化を狙うのが目的
`index_inventory.sql`ではPostgreSQL内部に分散しているインデックス情報をシステムカタログから取得し、比較しやすい一覧表へまとめるSQL

OIDで関連付けられているテーブルとインデックスの情報を比較材料として使用し、重複判断にユニーク定義や利用クエリなどを出力することでユーザーの削除対象の判断材料とする

削除対象は以下の通り

|削除候補|代替する制約由来インデックス|サイズ|利用クエリ
|---|---|---|---|
|`idx_users_email`|`users_email_key`|約4.8MB|`GetUserByEmail`|
|`idx_categories_name`|`categories_name_key`|約16KB|`ListCategories`のORDER BY name|
|`idx_products_sku`|`products_sku_key`|約30MB|検索なし、UNIQUE検査|
|`idx_carts_user_id`|`carts_user_id_key`|約1.1MB|`GetCartByUser`|


### 完全重複ではない削除候補

|削除候補|代替候補|サイズ|利用クエリ|
|---|---|---|---|
|`idx_cart_items_cart_id`|`cart_items_cart_id_product_id_key`|約2.4MB|`ListCartItems`、カートJOINなど|

`idx_cart_items_cart_id`は単一列`(cart_id)`であり、複合UNIQUEインデックス`(cart_id, product_id)`とは完全重複ではない。

ただし、B-treeの左端が`cart_id`で一致しているため、`cart_id`のみを条件にした検索でも代替できる可能性があり、実行計画で追加確認する対象とした。

### 削除候補を一時DROPした実行計画

`backend/perf/sql/queries/index_redundancy_check.sql`を使用し、トランザクション内で削除候補5本を一時的にDROPして実行計画を確認した。

|削除したインデックス|削除後の実行計画|
|---|---|
|`idx_users_email`|`users_email_key`によるIndex Scan|
|`idx_categories_name`|`Seq Scan + Sort`|
|`idx_products_sku`|`products_sku_key`によるIndex Scan|
|`idx_carts_user_id`|`carts_user_id_key`によるIndex Scan|
|`idx_cart_items_cart_id`|`cart_items_cart_id_product_id_key`によるIndex Scan|

カートJOINでも`carts_user_id_key`と`cart_items_cart_id_product_id_key`が使用され、削除候補を外した状態でも既存クエリを処理できることを確認した。

検証後は`ROLLBACK`し、削除候補5本がすべて復元されたことを確認した。