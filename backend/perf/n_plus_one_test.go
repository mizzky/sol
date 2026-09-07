//go:build integration

package perf

import (
	"context"
	"database/sql"
	"os"
	"sol_coffeesys/backend/db"
	"sort"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type orderWithItems struct {
	Order db.ListOrdersByUserRow
	Items []db.OrderItem
}

// countingDBは検証処理が発行したDBクエリ数を記録する
type countingDB struct {
	delegate db.DBTX
	count    atomic.Int64
}

func newCountingDB(delegate db.DBTX) *countingDB {
	return &countingDB{
		delegate: delegate,
	}
}

func (c *countingDB) ExecContext(
	ctx context.Context,
	query string,
	args ...interface{},
) (sql.Result, error) {
	c.count.Add(1)
	return c.delegate.ExecContext(ctx, query, args...)
}

func (c *countingDB) PrepareContext(
	ctx context.Context,
	query string,
) (*sql.Stmt, error) {
	return c.delegate.PrepareContext(ctx, query)
}

func (c *countingDB) QueryContext(
	ctx context.Context,
	query string,
	args ...interface{},
) (*sql.Rows, error) {
	c.count.Add(1)
	return c.delegate.QueryContext(ctx, query, args...)
}

func (c *countingDB) QueryRowContext(
	ctx context.Context,
	query string,
	args ...interface{},
) *sql.Row {
	c.count.Add(1)
	return c.delegate.QueryRowContext(ctx, query, args...)
}

func (c *countingDB) Count() int64 {
	return c.count.Load()
}

func openPerfDB(t *testing.T, ctx context.Context) *sql.DB {
	t.Helper()

	databaseURL := os.Getenv("PERF_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("PERF_DATABASE_URL is not set")
	}

	sqlDB, err := sql.Open("postgres", databaseURL)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, sqlDB.Close())
	})

	require.NoError(t, sqlDB.PingContext(ctx))

	var databaseName string
	err = sqlDB.QueryRowContext(
		ctx,
		"SELECT current_database()",
	).Scan(&databaseName)
	require.NoError(t, err)
	require.Equal(t, "coffeesys_perf", databaseName)

	return sqlDB
}

func TestLoadOrdersNPlusOne(t *testing.T) {
	const (
		userID   int64 = 1
		maxLimit int   = 500
	)

	tests := []struct {
		name  string
		limit int
	}{
		{
			name:  "N=10",
			limit: 10,
		},
		{
			name:  "N=100",
			limit: 100,
		},
		{
			name:  "N=500",
			limit: 500,
		},
	}

	ctx := t.Context()
	sqlDB := openPerfDB(t, ctx)

	// fixture確認用SQLはクエリ数の計測に含めない。
	var availableOrders int
	err := sqlDB.QueryRowContext(
		ctx,
		"SELECT COUNT(*) FROM orders WHERE user_id = $1",
		userID,
	).Scan(&availableOrders)
	require.NoError(t, err)
	require.GreaterOrEqual(t, availableOrders, maxLimit)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// サブテストごとにカウンターを0から始める。
			countedDB := newCountingDB(sqlDB)

			got, err := loadOrdersNPlusOne(
				ctx,
				countedDB,
				userID,
				tt.limit,
			)
			require.NoError(t, err)
			require.Len(t, got, tt.limit)

			assert.Equal(
				t,
				int64(1+tt.limit),
				countedDB.Count(),
			)

			for orderIndex, result := range got {
				// 注文順はcreated_at DESC, id DESC。
				if orderIndex > 0 {
					previous := got[orderIndex-1].Order
					current := result.Order

					correctOrder :=
						previous.CreatedAt.After(current.CreatedAt) ||
							(previous.CreatedAt.Equal(current.CreatedAt) &&
								previous.ID > current.ID)

					assert.True(
						t,
						correctOrder,
						"orders are not sorted at index %d",
						orderIndex,
					)
				}

				for itemIndex, item := range result.Items {
					assert.Equal(
						t,
						result.Order.ID,
						item.OrderID,
					)

					// 注文内の明細順はid ASC。
					if itemIndex > 0 {
						assert.Less(
							t,
							result.Items[itemIndex-1].ID,
							item.ID,
						)
					}
				}
			}
		})
	}
}

func loadOrdersNPlusOne(
	ctx context.Context,
	dbtx db.DBTX,
	userID int64,
	limit int,
) ([]orderWithItems, error) {
	const listLimitedOrders = `
  SELECT
      id,
      user_id,
      total,
      status,
      created_at,
      updated_at
  FROM orders
  WHERE user_id = $1
  ORDER BY created_at DESC, id DESC
  LIMIT $2
  `

	rows, err := dbtx.QueryContext(
		ctx,
		listLimitedOrders,
		userID,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]db.ListOrdersByUserRow, 0, limit)

	for rows.Next() {
		var order db.ListOrdersByUserRow

		if err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Total,
			&order.Status,
			&order.CreatedAt,
			&order.UpdatedAt,
		); err != nil {
			return nil, err
		}

		orders = append(orders, order)
	}

	if err := rows.Close(); err != nil {
		return nil, err
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	queries := db.New(dbtx)
	result := make([]orderWithItems, 0, len(orders))

	for _, order := range orders {
		items, err := queries.ListOrderItemsByOrderID(ctx, order.ID)
		if err != nil {
			return nil, err
		}

		result = append(result, orderWithItems{
			Order: order,
			Items: items,
		})
	}

	return result, nil
}

func TestMeasureOrdersNPlusOne(t *testing.T) {
	const userID int64 = 1

	const (
		maxLimit = 500
		runCount = 5
	)

	tests := []struct {
		name  string
		limit int
	}{
		{name: "N=10", limit: 10},
		{name: "N=100", limit: 100},
		{name: "N=500", limit: 500},
	}

	ctx := t.Context()
	sqlDB := openPerfDB(t, ctx)

	// fixture確認は計測対象外。
	var availableOrders int
	err := sqlDB.QueryRowContext(
		ctx,
		"SELECT COUNT(*) FROM orders WHERE user_id = $1",
		userID,
	).Scan(&availableOrders)
	require.NoError(t, err)
	require.GreaterOrEqual(t, availableOrders, maxLimit)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ウォームアップは本計測に含めない。
			warmupDB := newCountingDB(sqlDB)
			warmupResult, err := loadOrdersNPlusOne(
				ctx,
				warmupDB,
				userID,
				tt.limit,
			)
			require.NoError(t, err)
			require.Len(t, warmupResult, tt.limit)
			require.Equal(
				t,
				int64(1+tt.limit),
				warmupDB.Count(),
			)

			durations := make([]time.Duration, 0, runCount)
			itemCounts := make([]int, 0, runCount)

			for run := 0; run < runCount; run++ {
				// runごとにクエリ数を0から数える。
				countedDB := newCountingDB(sqlDB)

				startedAt := time.Now()
				got, err := loadOrdersNPlusOne(
					ctx,
					countedDB,
					userID,
					tt.limit,
				)
				elapsed := time.Since(startedAt)

				// 検証処理は計測時間に含めない。
				require.NoError(t, err)
				require.Len(t, got, tt.limit)
				require.Equal(
					t,
					int64(1+tt.limit),
					countedDB.Count(),
				)

				itemCount := 0
				for _, result := range got {
					itemCount += len(result.Items)
				}

				durations = append(durations, elapsed)
				itemCounts = append(itemCounts, itemCount)
			}

			for run := 1; run < len(itemCounts); run++ {
				require.Equal(t, itemCounts[0], itemCounts[run])
			}

			sort.Slice(durations, func(i, j int) bool {
				return durations[i] < durations[j]
			})
			median := durations[len(durations)/2]

			t.Logf(
				"N=%d median=%s queries=%d orders=%d items=%d runs=%d",
				tt.limit,
				median,
				1+tt.limit,
				tt.limit,
				itemCounts[0],
				runCount,
			)
		})
	}
}

func TestLoadOrdersBatch(t *testing.T) {
	const userID int64 = 1

	tests := []struct {
		name  string
		limit int
	}{
		{name: "N=10", limit: 10},
	}

	ctx := t.Context()
	sqlDB := openPerfDB(t, ctx)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want, err := loadOrdersNPlusOne(
				ctx,
				sqlDB,
				userID,
				tt.limit,
			)
			require.NoError(t, err)
			require.Len(t, want, tt.limit)

			countedDB := newCountingDB(sqlDB)

			got, err := loadOrdersBatch(
				ctx,
				countedDB,
				userID,
				tt.limit,
			)
			require.NoError(t, err)
			require.Equal(t, want, got)
			require.Equal(t, int64(2), countedDB.Count())
		})
	}
}

func loadOrdersBatch(
	ctx context.Context,
	dbtx db.DBTX,
	userID int64,
	limit int,
) ([]orderWithItems, error) {
	const listLimitedOrders = `
        SELECT
                id,
                user_id,
                total,
                status,
                created_at,
                updated_at
        FROM orders
        WHERE user_id = $1
        ORDER BY created_at DESC, id DESC
        LIMIT $2
        `

	const listOrderItemsByOrderIDs = `
        SELECT
                id,
                order_id,
                product_id,
                quantity,
                unit_price,
                product_name_snapshot,
                created_at,
                updated_at
        FROM order_items
        WHERE order_id = ANY($1::bigint[])
        ORDER BY order_id, id
        `

	// 1クエリ目: 注文一覧を取得する。
	orderRows, err := dbtx.QueryContext(
		ctx,
		listLimitedOrders,
		userID,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer orderRows.Close()

	orders := make([]db.ListOrdersByUserRow, 0, limit)

	for orderRows.Next() {
		var order db.ListOrdersByUserRow

		if err := orderRows.Scan(
			&order.ID,
			&order.UserID,
			&order.Total,
			&order.Status,
			&order.CreatedAt,
			&order.UpdatedAt,
		); err != nil {
			return nil, err
		}

		orders = append(orders, order)
	}

	if err := orderRows.Close(); err != nil {
		return nil, err
	}

	if err := orderRows.Err(); err != nil {
		return nil, err
	}

	// 一括取得する注文IDを作る。
	orderIDs := make([]int64, 0, len(orders))
	for _, order := range orders {
		orderIDs = append(orderIDs, order.ID)
	}

	// 2クエリ目: 全注文の明細を一括取得する。
	itemRows, err := dbtx.QueryContext(
		ctx,
		listOrderItemsByOrderIDs,
		pq.Array(orderIDs),
	)
	if err != nil {
		return nil, err
	}
	defer itemRows.Close()

	itemsByOrderID := make(
		map[int64][]db.OrderItem,
		len(orders),
	)
	for itemRows.Next() {
		var item db.OrderItem

		if err := itemRows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.Quantity,
			&item.UnitPrice,
			&item.ProductNameSnapshot,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}

		itemsByOrderID[item.OrderID] = append(
			itemsByOrderID[item.OrderID],
			item,
		)
	}

	if err := itemRows.Close(); err != nil {
		return nil, err
	}

	if err := itemRows.Err(); err != nil {
		return nil, err
	}

	// 注文一覧の順序を保ちながら明細を結合する。
	result := make([]orderWithItems, 0, len(orders))

	for _, order := range orders {
		result = append(result, orderWithItems{
			Order: order,
			Items: itemsByOrderID[order.ID],
		})
	}

	return result, nil
}
