//go:build integration

package perf

import (
	"context"
	"database/sql"
	"os"
	"sol_coffeesys/backend/db"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
		userID int64 = 1
		limit        = 10
	)
	ctx := t.Context()
	sqlDB := openPerfDB(t, ctx)

	// テストデータが前提件数を満たすことを確認。このSQLはN+1のクエリに含めない
	var availableOrders int
	err := sqlDB.QueryRowContext(
		ctx,
		"SELECT COUNT(*) FROM orders WHERE user_id = $1",
		userID,
	).Scan(&availableOrders)
	require.NoError(t, err)
	require.GreaterOrEqual(t, availableOrders, limit)

	countedDB := newCountingDB(sqlDB)

	got, err := loadOrdersNPlusOne(
		ctx,
		countedDB,
		userID,
		limit,
	)
	require.NoError(t, err)
	require.Len(t, got, limit)

	// 注文一覧1回＋注文明細10回。
	assert.Equal(t, int64(1+limit), countedDB.Count())

	for orderIndex, result := range got {
		// 注文順はcreated_at DESC, id DESC。
		if orderIndex > 0 {
			previous := got[orderIndex-1].Order
			current := result.Order

			correctOrder := previous.CreatedAt.After(current.CreatedAt) ||
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
			// 明細が正しい注文に所属している。
			assert.Equal(t, result.Order.ID, item.OrderID)

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
}
