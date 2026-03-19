//go:build integration

package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// order件数
func assertOrderCountByUser(t *testing.T, userID int64, want int) {
	t.Helper()
	var got int
	err := testDB.QueryRow(`SELECT COUNT(*) FROM orders WHERE user_id=$1`, userID).Scan(&got)
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

// orderItem件数
func assertOrderItemCountByUser(t *testing.T, userID int64, want int) {
	t.Helper()
	var got int
	err := testDB.QueryRow(`
		SELECT COUNT(*)
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		WHERE o.user_id = $1
	`, userID).Scan(&got)
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

// cartItem件数
func assertCartItemCountByUser(t *testing.T, userID int64, want int) {
	t.Helper()
	var got int
	err := testDB.QueryRow(`
		SELECT COUNT(*)
		FROM cart_items ci
		JOIN carts c ON c.id = ci.cart_id
		WHERE c.user_id = $1
	`, userID).Scan(&got)
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

// productIDベースの在庫検証
func assertProductStockByID(t *testing.T, productID int64, want int32) {
	t.Helper()
	var got int32
	err := testDB.QueryRow(`
		SELECT stock_quantity FROM products WHERE id = $1
	`, productID).Scan(&got)
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

// skuベースの在庫検証
func assertProductStockBySKU(t *testing.T, sku string, want int32) {
	t.Helper()
	var got int32
	err := testDB.QueryRow(`
		SELECT stock_quantity FROM products WHERE sku = $1
	`, sku).Scan(&got)
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func assertOrderStatus(t *testing.T, orderID int64, want string) {
	t.Helper()
	var status string
	err := testDB.QueryRow(`
		SELECT status FROM orders WHERE id = $1
	`, orderID).Scan(&status)
	assert.NoError(t, err)
	assert.Equal(t, want, status)
}

// DBクリーンアップ
func cleanupOrderRelatedTables(t *testing.T) {
	t.Helper()
	_, err := testDB.Exec(`
		TRUNCATE TABLE order_items, orders, cart_items, carts, products, categories, users
		RESTART IDENTITY CASCADE
	`)
	assert.NoError(t, err)
}
