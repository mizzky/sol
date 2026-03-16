//go:build integration

package tests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sol_coffeesys/backend/db"
	"sol_coffeesys/backend/handler"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// テスト全体で共有する DB 接続
var testDB *sql.DB

func TestMain(m *testing.M) {
	ctx := context.Background()

	// 1. PostgreSQL コンテナを起動
	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:17-trixie",
		tcpostgres.WithDatabase("test_db"),
		tcpostgres.WithUsername("user"),
		tcpostgres.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2),
		),
	)
	if err != nil {
		log.Fatalf("コンテナ起動失敗: %v", err)
	}

	// 2. 接続文字列を取得
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("接続文字列取得失敗: %v", err)
	}

	// 3. DB 接続
	testDB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("DB 接続失敗: %v", err)
	}

	// 4. マイグレーション適用
	// go test はパッケージディレクトリ（backend/tests）を作業ディレクトリにする
	wd, _ := os.Getwd()
	migrationsPath := filepath.Join(wd, "..", "db", "migrations")

	migrator, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		connStr,
	)
	if err != nil {
		log.Fatalf("マイグレーション初期化失敗: %v", err)
	}
	if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("マイグレーション適用失敗: %v", err)
	}

	// 5. テスト実行
	code := m.Run()

	// 6. クリーンアップ（defer は os.Exit 前に実行されないため明示的に呼ぶ）
	testDB.Close()
	pgContainer.Terminate(ctx)

	os.Exit(code)
}

func TestIntegration_DBReady(t *testing.T) {
	if testDB == nil {
		t.Fatalf("testDB is nil")
	}
	if err := testDB.Ping(); err != nil {
		t.Fatalf("test DB ping failed: %v", err)
	}

	var currentDB string
	if err := testDB.QueryRow(`SELECT current_database()`).Scan(&currentDB); err != nil {
		t.Fatalf("failed to query current_database(): %v", err)
	}
	if currentDB != "test_db" {
		t.Fatalf("unexpected database name: got=%s want=test_db", currentDB)
	}
}

func seedHappyPath(t *testing.T) (userID int64, productID int64) {
	t.Helper()

	// 1. user
	err := testDB.QueryRow(`
		INSERT INTO users (name, email, password_hash)
		VALUES('テストユーザー', 'test@example.com', 'dummy_hash')
		RETURNING id
	`).Scan(&userID)
	if err != nil {
		t.Fatalf("user insert failed:%v", err)
	}

	// 2. category(productsがFK参照するため)
	var categoryID int64
	err = testDB.QueryRow(`
		INSERT INTO categories (name)
		VALUES('テストカテゴリ')
		RETURNING id
	`).Scan(&categoryID)
	if err != nil {
		t.Fatalf("category insert failed:%v", err)
	}

	// 3. product(qty=10 price=750)
	err = testDB.QueryRow(`
		INSERT INTO products (name, price, category_id, sku, stock_quantity)
		VALUES('テストコーヒー', 750, $1, 'SKU_TEST-001', 10)
		RETURNING id
	`, categoryID).Scan(&productID)
	if err != nil {
		t.Fatalf("product insert failed:%v", err)
	}

	// 4. cart
	var cartID int64
	err = testDB.QueryRow(`
		INSERT INTO carts(user_id) VALUES ($1) RETURNING id
	`, userID).Scan(&cartID)
	if err != nil {
		t.Fatalf("cart insert failed:%v", err)
	}

	// 5. cart_item(qty=2, price=750*2=1500)
	_, err = testDB.Exec(`
		INSERT INTO cart_items(cart_id, product_id, quantity, price)
		VALUES($1, $2, 2, 1500)
	`, cartID, productID)
	if err != nil {
		t.Fatalf("cart_item insert failed:%v", err)
	}

	// テスト後に全データ掃除
	t.Cleanup(func() {
		cleanupOrderRelatedTables(t)
	})

	return userID, productID

}

func TestCreateOrderHandler_HappyPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID, productID := seedHappyPath(t)

	router := gin.New()
	queries := db.New(testDB)
	router.POST("/api/orders", func(c *gin.Context) {
		c.Set("userID", userID)
		handler.CreateOrderHandler(testDB, queries)(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// 注文が正しく処理され、在庫が減少(10-2=8),カートがクリアされることを検証
	assertOrderCountByUser(t, userID, 1)
	assertOrderItemCountByUser(t, userID, 1)
	assertProductStockByID(t, productID, 8)
	assertCartItemCountByUser(t, userID, 0)

}

func seedEmptyCart(t *testing.T) int64 {
	t.Helper()

	var userID int64
	err := testDB.QueryRow(`
        INSERT INTO users (name, email, password_hash)
        VALUES ('空カートユーザー', 'empty@example.com', 'dummy_hash')
        RETURNING id
    `).Scan(&userID)
	if err != nil {
		t.Fatalf("user insert failed: %v", err)
	}

	_, err = testDB.Exec(`INSERT INTO carts(user_id) VALUES ($1)`, userID)
	if err != nil {
		t.Fatalf("cart insert failed: %v", err)
	}

	t.Cleanup(func() {
		cleanupOrderRelatedTables(t)
	})

	return userID
}

// 在庫1個の商品[SKU_00S-001]に対して、数量5の注文をリクエストする
func seedOutOfStock(t *testing.T) int64 {
	t.Helper()

	var userID, categoryID, productID, cartID int64

	err := testDB.QueryRow(`
		INSERT INTO users (name, email, password_hash)
		VALUES ('在庫不足ユーザー', 'outofstock@example.com', 'dummy_hash')
		RETURNING id
	`).Scan(&userID)
	if err != nil {
		t.Fatalf("user insert failed: %v", err)
	}

	err = testDB.QueryRow(`
		INSERT INTO categories (name)
		VALUES ('テストカテゴリ')
		RETURNING id
	`).Scan(&categoryID)
	if err != nil {
		t.Fatalf("category insert failed: %v", err)
	}

	// 在庫 1
	err = testDB.QueryRow(`
		INSERT INTO products (name, price, category_id, sku, stock_quantity)
		VALUES ('在庫不足テスト商品', 750, $1, 'SKU_00S-001', 1)
		RETURNING id
	`, categoryID).Scan(&productID)
	if err != nil {
		t.Fatalf("product insert failed: %v", err)
	}

	err = testDB.QueryRow(`
		INSERT INTO carts(user_id) VALUES ($1) RETURNING id
	`, userID).Scan(&cartID)
	if err != nil {
		t.Fatalf("cart insert failed: %v", err)
	}

	// 数量5の注文明細で在庫不足
	_, err = testDB.Exec(`
		INSERT INTO cart_items (cart_id, product_id, quantity, price)
		VALUES ($1, $2, 5, 3750)
	`, cartID, productID)
	if err != nil {
		t.Fatalf("cart_item insert failed: %v", err)
	}

	t.Cleanup(func() {
		cleanupOrderRelatedTables(t)
	})

	return userID
}

func TestCreateOrderHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupDB        func(t *testing.T) int64 //DBを準備してuserIDを返す
		rawUserID      interface{}
		expectedStatus int
		expectedErrMsg string
		assertDB       func(t *testing.T, userID int64)
	}{
		{
			name:           "異常系：未認証",
			setupDB:        nil,
			rawUserID:      nil,
			expectedStatus: http.StatusUnauthorized,
			expectedErrMsg: "認証が必要です",
			assertDB:       nil,
		},
		{
			name:           "異常系：userIDが不正な型",
			setupDB:        nil,
			rawUserID:      "not-a-number",
			expectedStatus: http.StatusUnauthorized,
			expectedErrMsg: "認証が必要です",
			assertDB:       nil,
		},
		{
			name:           "異常系：カートが空",
			setupDB:        seedEmptyCart,
			rawUserID:      int64(0),
			expectedStatus: http.StatusBadRequest,
			expectedErrMsg: "カートが空です",
			assertDB: func(t *testing.T, userID int64) {
				assertOrderCountByUser(t, userID, 0)
			},
		},
		{
			name:           "異常系：在庫不足",
			setupDB:        seedOutOfStock,
			rawUserID:      int64(0),
			expectedStatus: http.StatusConflict,
			expectedErrMsg: "在庫不足です",
			assertDB: func(t *testing.T, userID int64) {
				assertOrderCountByUser(t, userID, 0)
				assertOrderItemCountByUser(t, userID, 0)
				assertProductStockBySKU(t, "SKU_00S-001", 1)
				assertCartItemCountByUser(t, userID, 1)
			},
		},
		{
			name:           "異常系：userIDがintでも通る",
			setupDB:        seedEmptyCart,
			rawUserID:      int(0),
			expectedStatus: http.StatusBadRequest,
			expectedErrMsg: "カートが空です",
			assertDB: func(t *testing.T, userID int64) {
				assertOrderCountByUser(t, userID, 0)
			},
		},
		{
			name:           "異常系：userIDがfloatでも通る",
			setupDB:        seedEmptyCart,
			rawUserID:      float64(0),
			expectedStatus: http.StatusBadRequest,
			expectedErrMsg: "カートが空です",
			assertDB: func(t *testing.T, userID int64) {
				assertOrderCountByUser(t, userID, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var userID int64
			rawUserID := tt.rawUserID

			if tt.setupDB != nil {
				userID = tt.setupDB(t)

				switch tt.rawUserID.(type) {
				case int:
					rawUserID = int(userID)
				case float64:
					rawUserID = float64(userID)
				case int64:
					rawUserID = userID
				}
			}

			router := gin.New()
			queries := db.New(testDB)

			router.POST("/api/orders", func(c *gin.Context) {
				if rawUserID != nil {
					c.Set("userID", rawUserID)
				}
				handler.CreateOrderHandler(testDB, queries)(c)
			})

			req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewBufferString(`{}`))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var resp map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedErrMsg, resp["error"])

			if tt.assertDB != nil {
				tt.assertDB(t, userID)
			}

		})
	}
}
