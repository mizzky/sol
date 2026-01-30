package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"sol_coffeesys/backend/db" // 生成されたパッケージをインポート

	_ "github.com/lib/pq"
)

func main() {
	// 接続文字列（docker-composeの設定に準拠）
	connStr := "host=db user=user password=password dbname=coffeesys_db sslmode=disable"
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// sqlcが生成したQueryオブジェクトを作成
	queries := db.New(conn)
	ctx := context.Background()

	// 1. 商品を追加してみる
	newProduct, err := queries.CreateProduct(ctx, db.CreateProductParams{
		Name:  "エチオピア・シダモ",
		Price: 650,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created: %s (ID: %d)\n", newProduct.Name, newProduct.ID)

	// 2. 一覧を取得してみる
	products, err := queries.ListProducts(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Current Menu:")
	for _, p := range products {
		fmt.Printf("- %s: ¥%d\n", p.Name, p.Price)
	}
}
