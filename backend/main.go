package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"sol_coffeesys/backend/db"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"golang.org/x/crypto/bcrypt"

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

	ctx := context.Background()
	rawPassword := "mypassword123"
	email := "text@example.com"
	name := "田中 太郎"

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("failed to hash password :%v", err)
	}

	newUser, err := queries.CreateUser(ctx, db.CreateUserParams{
		Name:         name,
		Email:        email,
		PasswordHash: string(hashedPassword),
		Role:         "member",
	})
	if err != nil {
		log.Fatalf("ユーザー登録に失敗:%v", err)
	}
	fmt.Printf("登録成功 ID:%d, Name: %s\n", newUser.ID, newUser.Name)

	//3. Ginルーター初期化
	r := gin.Default()

	// CORS設定:Next.jsだけに絞る
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:3000"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept"},
	}))

	// /apiのグループ作成
	api := r.Group("/api")
	{
		//エンドポイント：商品一覧取得
		api.GET("/products", func(c *gin.Context) {
			products, err := queries.ListProducts(c.Request.Context())
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			// DBから取得したスライスをそのままJSONとして返す
			c.JSON(http.StatusOK, products)
		})

		//エンドポイント：商品登録
		api.POST("/products", func(c *gin.Context) {
			// フロントから送られてくるJSON受け皿
			var input struct {
				Name  string `json: "name"`
				Price int32  `json: "price"`
			}

			// JSON解析
			if err := c.ShouldBindJSON(&input); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			// save DB
			product, err := queries.CreateProduct(c.Request.Context(), db.CreateProductParams{
				Name:        input.Name,
				Price:       input.Price,
				IsAvailable: true,
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, product)
		})
	}

	//5. サーバー起動
	log.Println("Server starting on :8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("failed to run server:", err)
	}
}
