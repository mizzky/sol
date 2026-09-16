//go:build integration

package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"sol_coffeesys/backend/db"

	"github.com/stretchr/testify/require"
)

func TestListOrderItemsByOrderIDs(t *testing.T) {
	unique := time.Now().UnixNano()

	t.Cleanup(func() {
		cleanupOrderRelatedTables(t)
	})

	var userID int64
	err := testDB.QueryRow(`
                INSERT INTO users (
                        name,
                        email,
                        password_hash
                )
                VALUES ($1, $2, $3)
                RETURNING id
        `,
		"明細一括取得テストユーザー",
		fmt.Sprintf("order-items-batch-%d@example.com", unique),
		"dummy_hash",
	).Scan(&userID)
	require.NoError(t, err)

	var categoryID int64
	err = testDB.QueryRow(`
                INSERT INTO categories (name)
                VALUES ($1)
                RETURNING id
        `,
		fmt.Sprintf("明細一括取得カテゴリ-%d", unique),
	).Scan(&categoryID)
	require.NoError(t, err)

	var productID int64
	err = testDB.QueryRow(`
                INSERT INTO products (
                        name,
                        price,
                        category_id,
                        sku,
                        stock_quantity
                )
                VALUES ($1, $2, $3, $4, $5)
                RETURNING id
        `,
		"明細一括取得テスト商品",
		500,
		categoryID,
		fmt.Sprintf("SKU_ORDER_ITEMS_BATCH_%d", unique),
		100,
	).Scan(&productID)
	require.NoError(t, err)

	// 取得対象2注文と、取得対象外1注文を作る。
	orderIDs := make([]int64, 3)

	for i := range orderIDs {
		err = testDB.QueryRow(`
                        INSERT INTO orders (
                                user_id,
                                total,
                                status
                        )
                        VALUES ($1, $2, $3)
                        RETURNING id
                `,
			userID,
			1000,
			"pending",
		).Scan(&orderIDs[i])
		require.NoError(t, err)
	}

	// 最初の2注文は2明細、対象外注文は1明細。
	itemCounts := []int{2, 2, 1}
	expectedItemIDs := make([]int64, 0, 4)

	for orderIndex, itemCount := range itemCounts {
		for itemIndex := 0; itemIndex < itemCount; itemIndex++ {
			var itemID int64

			err = testDB.QueryRow(`
                                INSERT INTO order_items (
                                        order_id,
                                        product_id,
                                        quantity,
                                        unit_price,
                                        product_name_snapshot
                                )
                                VALUES ($1, $2, $3, $4, $5)
                                RETURNING id
                        `,
				orderIDs[orderIndex],
				productID,
				itemIndex+1,
				500,
				"明細一括取得テスト商品",
			).Scan(&itemID)
			require.NoError(t, err)

			if orderIndex < 2 {
				expectedItemIDs = append(expectedItemIDs, itemID)
			}
		}
	}

	queries := db.New(testDB)

	// 入力順ではなく、order_id ASC、id ASCで返ることを確認する。
	got, err := queries.ListOrderItemsByOrderIDs(
		context.Background(),
		[]int64{
			orderIDs[1],
			orderIDs[0],
		},
	)
	require.NoError(t, err)
	require.Len(t, got, 4)

	gotItemIDs := make([]int64, 0, len(got))
	gotOrderIDs := make([]int64, 0, len(got))

	for _, item := range got {
		gotItemIDs = append(gotItemIDs, item.ID)
		gotOrderIDs = append(gotOrderIDs, item.OrderID)
	}

	require.Equal(t, expectedItemIDs, gotItemIDs)
	require.Equal(
		t,
		[]int64{
			orderIDs[0],
			orderIDs[0],
			orderIDs[1],
			orderIDs[1],
		},
		gotOrderIDs,
	)
}
