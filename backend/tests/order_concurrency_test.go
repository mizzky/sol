//go:build integration

package tests

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sol_coffeesys/backend/db"
	"sol_coffeesys/backend/handler"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func seedConcurrentOrders(t *testing.T, userCnt int64, cartQty int64, productQty int64) (int64, []int64) {
	t.Helper()

	var userIDs []int64
	var productID int64

	var categoryID int64
	err := testDB.QueryRow(`
		INSERT INTO categories (name)
		VALUES('テストカテゴリ')
		RETURNING id
	`).Scan(&categoryID)
	if err != nil {
		t.Fatalf("category insert failed:%v", err)
	}

	err = testDB.QueryRow(`
		INSERT INTO products (name, price, category_id, sku, stock_quantity)
		VALUES('テストコーヒー', 750, $1, 'SKU_TEST-001', $2)
		RETURNING id
	`, categoryID, productQty).Scan(&productID)
	if err != nil {
		t.Fatalf("product insert failed:%v", err)
	}

	for i := range userCnt {
		var newUserID int64
		err := testDB.QueryRow(`
			INSERT INTO users (name, email, password_hash)
			VALUES ($1, $2, 'dummy_hash')
			RETURNING id
		`, fmt.Sprintf("同期テストユーザー%d", i), fmt.Sprintf("user%d@example.com", i)).Scan(&newUserID)
		if err != nil {
			t.Fatalf("users insert failed: %v", err)
		}
		userIDs = append(userIDs, newUserID)

		var cartID int64
		err = testDB.QueryRow(`
		INSERT INTO carts(user_id) VALUES ($1) RETURNING id
	`, userIDs[i]).Scan(&cartID)
		if err != nil {
			t.Fatalf("cart insert failed:%v", err)
		}

		// 5. cart_item
		_, err = testDB.Exec(`
		INSERT INTO cart_items(cart_id, product_id, quantity, price)
		VALUES($1, $2, $3, 2250)
	`, cartID, productID, cartQty)
		if err != nil {
			t.Fatalf("cart_item insert failed:%v", err)
		}
	}

	t.Cleanup(func() {
		cleanupOrderRelatedTables(t)
	})

	return productID, userIDs
}

func TestCreateOrderHandler_Concurrency(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		userCnt    int64
		cartQty    int64
		productQty int64
	}{
		{
			name:       "ユーザー5 成功3 競合2",
			userCnt:    5,
			cartQty:    3,
			productQty: 10,
		},
		{
			name:       "ユーザー50 成功10 競合40",
			userCnt:    50,
			cartQty:    3,
			productQty: 30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queries := db.New(testDB)
			productID, userIDs := seedConcurrentOrders(t, tt.userCnt, tt.cartQty, tt.productQty)

			var (
				wg          sync.WaitGroup
				mu          sync.Mutex
				statusCodes []int
			)
			ready := make(chan struct{})
			for _, uid := range userIDs {
				wg.Add(1)
				go func(userID int64) {
					defer wg.Done()
					<-ready

					r := gin.New()
					r.POST("/api/orders", func(c *gin.Context) {
						c.Set("userID", userID)
						handler.CreateOrderHandler(testDB, queries)(c)
					})

					req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewBufferString(`{}`))
					req.Header.Set("Content-Type", "application/json")
					w := httptest.NewRecorder()
					r.ServeHTTP(w, req)

					mu.Lock()
					statusCodes = append(statusCodes, w.Code)
					mu.Unlock()
				}(uid)
			}

			close(ready)
			wg.Wait()

			successCnt := int64(0)
			conflictCnt := int64(0)

			for _, code := range statusCodes {
				switch code {
				case http.StatusCreated:
					successCnt++
				case http.StatusConflict:
					conflictCnt++
				default:
					t.Fatalf("unexpected status code: %d", code)
				}

			}

			expectedSuccessCnt := tt.productQty / tt.cartQty
			expectedConflictCnt := tt.userCnt - expectedSuccessCnt
			expectedRemainingStock := int32(tt.productQty % tt.cartQty)

			assert.Len(t, statusCodes, int(tt.userCnt))
			assert.Equal(t, expectedSuccessCnt, successCnt)
			assert.Equal(t, expectedConflictCnt, conflictCnt)
			assertProductStockByID(t, productID, expectedRemainingStock)

		})
	}

}
