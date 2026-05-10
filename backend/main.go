package main

import (
	"database/sql"
	"log"
	"log/slog"
	"os"
	"sol_coffeesys/backend/db"
	"sol_coffeesys/backend/middleware"
	"sol_coffeesys/backend/pkg/apperror"
	"sol_coffeesys/backend/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	_ "github.com/lib/pq"
)

func main() {
	slog.SetDefault(middleware.NewJSONLogger(os.Stdout, slog.LevelInfo))

	//1. DB接続
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("DATABASE_URL is not set")
	}
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer conn.Close()

	//2. sqlクエリ初期化
	queries := db.New(conn)

	//3. Ginルーター初期化
	r := gin.Default()

	// 4. ミドルウェア設定
	// request_id生成
	r.Use(middleware.RequestIDMiddleware())

	// CORS設定:Next.jsだけに絞る
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	// エラーハンドラ
	r.Use(middleware.ErrorHandler(apperror.ToHTTP))

	//5. ルーティング設定
	routes.SetupRoutes(r, conn, queries)

	//6. サーバー起動
	slog.Info("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		slog.Error("failed to run server", "error", err)
		os.Exit(1)
	}
}
