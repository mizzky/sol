# 学習記録 2026-06-18

## セッション 1 (12:10)

### 取り組んだタスク
- Issue62 の DB パフォーマンス学習を、`tdd-mentor` の方針で読み解き中心に進めた。
- PostgreSQL 公式ドキュメント `14.1. Using EXPLAIN` を題材に、`EXPLAIN` / `EXPLAIN ANALYZE` の読み方を学習した。
- `Seq Scan` / `Index Scan` / `Bitmap Index Scan` / `Bitmap Heap Scan` / `Sort` / `Nested Loop` の違いを整理した。
- EXPLAIN に関する Sub issue 文面を作成し、表現の修正フィードバックを行った。
- PostgreSQL 公式ドキュメント `11.2.1 B-Tree` と `11.3 Multicolumn Indexes` を読み、B-Tree インデックスと複合インデックスの要点を整理した。
- B-Tree / 複合インデックスに関する Sub issue メモについて、誤り・タイポ・欠落観点に絞ってフィードバックした。

### ユーザーが質問した内容
- `Nested Loop` は結局どういう処理なのか。
- `WHERE` 句ではカーディナリティが高い条件を先に書いた方がよいのか。
- カーディナリティが低いと `Seq Scan` になりやすいのか、またカーディナリティは無関係なのか。
- `member` / `admin` のように値の種類は少なくても、データ分布によってインデックスの効き方が変わるのか。
- `users.email` のようなインデックス付き完全一致検索は `Index Scan` と考えてよいのか。
- `Bitmap Index Scan` / `Bitmap Heap Scan` の例や、heap ページ・index のイメージはどう捉えるとよいのか。
- `Bitmap Heap Scan` だけが出るケースはあるのか。
- スキャンノードは絶対ルールではなく、レコード数や統計情報に応じてプランナが選ぶものなのか。
- B-Tree はどのような検索に効き、B-Tree 以外の索引をジュニア段階で深掘りすべきなのか。
- 適切なインデックスを張って `EXPLAIN` のコストを見ながら改善することは、DB チューニング入門に当たるのか。
- 複合インデックス `(user_id, created_at DESC)` は、単一列インデックスが2個あるという意味なのか。
- `ORDER BY created_at DESC` や `LIMIT 20` と複合インデックスの相性はどう理解すればよいのか。
- Sub issue の B-Tree / 複合インデックス説明が、この粒度で十分か。

### 躓いたポイントと解決策
- `EXPLAIN` と `EXPLAIN ANALYZE` の違いが曖昧だった。`EXPLAIN` は SQL を実行せず、統計情報から `cost` / `rows` / `width` を推定するもの、`EXPLAIN ANALYZE` は実際にクエリを実行して `actual time` / `rows` / `loops` などの実測値を見るものと整理した。
- `Nested Loop` の動きが抽象的だった。外側の結果1行ごとに内側を検索し、外側が少なければ強いが、多いと内側検索が繰り返されて重くなりやすい JOIN として理解した。
- カーディナリティだけで `Seq Scan` / `Index Scan` を判断しそうになった。重要なのは値の種類数だけではなく、条件でどれだけ絞れるか、つまり選択度とデータ分布であると整理した。
- heap ページと index の関係が掴みにくかった。heap はテーブル本体の行データが保存される領域、index は検索キーから heap 上の行位置を探す索引、という役割で分けて理解した。
- `Bitmap Heap Scan` の位置づけが曖昧だった。`Bitmap Index Scan` などで作った行位置情報をもとに、必要な heap ページをまとめて読む処理と整理した。
- B-Tree 以外の索引まで掘りすぎそうになった。ジュニア段階では、まず B-Tree が等価検索・範囲検索・並び順に強いことを押さえ、B-Tree が苦手な検索には別方式があると知る程度でよいと整理した。
- 複合インデックスを単一列インデックスの集合として捉えかけた。`(user_id, created_at DESC)` は `user_id` で並び、同じ `user_id` 内で `created_at DESC` に並ぶ1つの索引だと整理した。
- `WHERE user_id = 1 ORDER BY created_at DESC LIMIT 20` でなぜ Sort を避けられる可能性があるのかが引っかかった。対象ユーザーの範囲に移動し、その範囲をインデックス順に読めるため、実行計画上の明示的な `Sort` を避け、必要件数を読んだ時点で止まれる可能性があると理解した。

### 次回課題
- 複合インデックスの左端 prefix を整理する。
- 範囲条件以降の列がどこまで効くのかを整理する。
- 既存の orders 系クエリに対して、どのインデックスが効くか / 効きにくいかを当てはめる。
- `EXPLAIN ANALYZE` を使って、インデックス有無やクエリ形状による実測差を比較する。
