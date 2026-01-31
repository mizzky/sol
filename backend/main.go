package main

import (
	"database/sql"
	"log"
	"net/http"
	"sol_coffeesys/backend/db"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	_ "github.com/lib/pq"
)

func main() {
	//1. DB接続
	connStr := "host=db user=user password=password dbname=coffeesys_db sslmode=disable"
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer conn.Close()

	//2. sqlクエリ初期化
	queries := db.New(conn)

	//3. Ginルーター初期化
	r := gin.Default()

	// CORS設定:Next.jsだけに絞る
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:3000"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept"},
	}))

	//4. エンドポイント：商品一覧取得
	r.GET("/products", func(c *gin.Context) {
		products, err := queries.ListProducts(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		// DBから取得したスライスをそのままJSONとして返す
		c.JSON(http.StatusOK, products)
	})

	//5. サーバー起動
	log.Println("Server starting on :8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("failed to run server:", err)
	}
}
