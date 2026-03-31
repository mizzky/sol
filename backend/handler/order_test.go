package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sol_coffeesys/backend/db"
	"sol_coffeesys/backend/handler/testutil"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateOrderLogic(t *testing.T) {
	tests := []struct {
		name        string
		userID      int64
		setupMock   func(*testutil.MockDB)
		expectedErr string
	}{
		{
			name:   "U1: 単一商品の注文作成",
			userID: int64(1),
			setupMock: func(m *testutil.MockDB) {
				now := time.Now()
				m.On("GetOrCreateCartForUser", mock.Anything, int64(1)).Return(
					db.Cart{
						ID:        10,
						UserID:    1,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)

				m.On("ListCartItemsByUser", mock.Anything, int64(1)).Return(
					[]db.ListCartItemsByUserRow{
						{
							ID:           1,
							CartID:       10,
							ProductID:    100,
							Quantity:     2,
							Price:        1500,
							CreatedAt:    now,
							UpdatedAt:    now,
							ProductName:  "Coffee",
							ProductPrice: 750,
							ProductStock: 50,
						},
					}, nil)

				m.On("GetProductForUpdate", mock.Anything, int64(100)).Return(
					db.Product{
						ID:            100,
						Name:          "Coffee",
						Price:         750,
						IsAvailable:   true,
						CategoryID:    1,
						Sku:           "COF-100",
						StockQuantity: 50,
						CreatedAt:     now,
						UpdatedAt:     now,
					}, nil)

				m.On("CreateOrder", mock.Anything, mock.Anything).Return(
					db.CreateOrderRow{
						ID:        1,
						UserID:    1,
						Status:    "pending",
						Total:     1500,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)

				m.On("CreateOrderItem", mock.Anything, mock.Anything).Return(
					db.OrderItem{
						ID:                  11,
						OrderID:             1,
						ProductID:           100,
						Quantity:            2,
						UnitPrice:           750,
						ProductNameSnapshot: "Coffee",
						CreatedAt:           now,
						UpdatedAt:           now,
					}, nil)

				m.On("UpdateProductStock", mock.Anything, mock.Anything).Return(
					db.UpdateProductStockRow{
						ID:            100,
						StockQuantity: 48,
					}, nil)

				m.On("ClearCartByUser", mock.Anything, int64(1)).Return(nil)
			},
			expectedErr: "",
		},
		{
			name:   "U2:複数商品の注文作成",
			userID: int64(1),
			setupMock: func(m *testutil.MockDB) {
				now := time.Now()
				m.On("GetOrCreateCartForUser", mock.Anything, int64(1)).Return(
					db.Cart{
						ID:        10,
						UserID:    1,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)

				m.On("ListCartItemsByUser", mock.Anything, int64(1)).Return(
					[]db.ListCartItemsByUserRow{
						{
							ID:           1,
							CartID:       10,
							ProductID:    100,
							Quantity:     2,
							Price:        1500,
							CreatedAt:    now,
							UpdatedAt:    now,
							ProductName:  "Coffee A",
							ProductPrice: 750,
							ProductStock: 50,
						},
						{
							ID:           2,
							CartID:       10,
							ProductID:    101,
							Quantity:     3,
							Price:        2850,
							CreatedAt:    now,
							UpdatedAt:    now,
							ProductName:  "Coffee B",
							ProductPrice: 950,
							ProductStock: 50,
						},
					}, nil)

				m.On("GetProductForUpdate", mock.Anything, int64(100)).Return(
					db.Product{
						ID:            100,
						Name:          "Coffee A",
						Price:         750,
						IsAvailable:   true,
						CategoryID:    1,
						Sku:           "COF-100",
						StockQuantity: 50,
						CreatedAt:     now,
						UpdatedAt:     now,
					}, nil)

				m.On("GetProductForUpdate", mock.Anything, int64(101)).Return(
					db.Product{
						ID:            101,
						Name:          "Coffee B",
						Price:         950,
						IsAvailable:   true,
						CategoryID:    1,
						Sku:           "COF-101",
						StockQuantity: 50,
						CreatedAt:     now,
						UpdatedAt:     now,
					}, nil)

				m.On("CreateOrder", mock.Anything, mock.Anything).Return(
					db.CreateOrderRow{
						ID:        1,
						UserID:    1,
						Status:    "pending",
						Total:     4350,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)

				m.On("CreateOrderItem", mock.Anything, mock.Anything).Return(
					db.OrderItem{
						ID:                  21,
						OrderID:             1,
						ProductID:           100,
						Quantity:            2,
						UnitPrice:           750,
						ProductNameSnapshot: "Coffee A",
						CreatedAt:           now,
						UpdatedAt:           now,
					}, nil).Once()

				m.On("CreateOrderItem", mock.Anything, mock.Anything).Return(
					db.OrderItem{
						ID:                  22,
						OrderID:             1,
						ProductID:           101,
						Quantity:            3,
						UnitPrice:           950,
						ProductNameSnapshot: "Coffee B",
						CreatedAt:           now,
						UpdatedAt:           now,
					}, nil).Once()

				m.On("UpdateProductStock", mock.Anything, mock.Anything).Return(
					db.UpdateProductStockRow{
						ID:            100,
						StockQuantity: 48,
					}, nil).Once()

				m.On("UpdateProductStock", mock.Anything, mock.Anything).Return(
					db.UpdateProductStockRow{
						ID:            101,
						StockQuantity: 47,
					}, nil).Once()

				m.On("ClearCartByUser", mock.Anything, int64(1)).Return(nil)
			},
			expectedErr: "",
		},
		{
			name:   "U4:カートが空",
			userID: int64(1),
			setupMock: func(m *testutil.MockDB) {
				now := time.Now()
				m.On("GetOrCreateCartForUser", mock.Anything, int64(1)).Return(
					db.Cart{
						ID:        10,
						UserID:    1,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
				m.On("ListCartItemsByUser", mock.Anything, int64(1)).Return(
					[]db.ListCartItemsByUserRow{}, nil)
			},
			expectedErr: "カートが空です",
		},
		{
			name:   "U5：カート内の商品が削除されている",
			userID: int64(1),
			setupMock: func(m *testutil.MockDB) {
				now := time.Now()

				m.On("GetOrCreateCartForUser", mock.Anything, int64(1)).Return(
					db.Cart{
						ID:        10,
						UserID:    1,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
				m.On("ListCartItemsByUser", mock.Anything, int64(1)).Return(
					[]db.ListCartItemsByUserRow{
						{
							ID:           1,
							CartID:       10,
							ProductID:    999,
							Quantity:     2,
							Price:        1500,
							CreatedAt:    now,
							UpdatedAt:    now,
							ProductName:  "DeletedProduct",
							ProductPrice: 750,
							ProductStock: 0,
						},
					}, nil)
				m.On("GetProductForUpdate", mock.Anything, int64(999)).Return(
					db.Product{}, sql.ErrNoRows)
			},
			expectedErr: "商品が見つかりません",
		},
		{
			name:   "U6：在庫不足",
			userID: int64(1),
			setupMock: func(m *testutil.MockDB) {
				now := time.Now()

				m.On("GetOrCreateCartForUser", mock.Anything, int64(1)).Return(
					db.Cart{
						ID:        10,
						UserID:    1,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)

				m.On("ListCartItemsByUser", mock.Anything, int64(1)).Return(
					[]db.ListCartItemsByUserRow{
						{
							ID:           1,
							CartID:       10,
							ProductID:    100,
							Quantity:     5,
							Price:        3750,
							CreatedAt:    now,
							UpdatedAt:    now,
							ProductName:  "Coffee",
							ProductPrice: 750,
							ProductStock: 3,
						},
					}, nil)

				m.On("GetProductForUpdate", mock.Anything, int64(100)).Return(
					db.Product{
						ID:            100,
						Name:          "Coffee",
						Price:         750,
						IsAvailable:   true,
						CategoryID:    1,
						Sku:           "COF-100",
						StockQuantity: 3,
						CreatedAt:     now,
						UpdatedAt:     now,
					}, nil)
			},
			expectedErr: "在庫不足です",
		},
		{
			name:   "U7：DB Error CreateOrder",
			userID: int64(1),
			setupMock: func(m *testutil.MockDB) {
				now := time.Now()
				m.On("GetOrCreateCartForUser", mock.Anything, int64(1)).Return(
					db.Cart{
						ID:        10,
						UserID:    1,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)

				m.On("ListCartItemsByUser", mock.Anything, int64(1)).Return(
					[]db.ListCartItemsByUserRow{
						{
							ID:           1,
							CartID:       10,
							ProductID:    100,
							Quantity:     2,
							Price:        1500,
							CreatedAt:    now,
							UpdatedAt:    now,
							ProductName:  "Coffee",
							ProductPrice: 750,
							ProductStock: 50,
						},
					}, nil)

				m.On("GetProductForUpdate", mock.Anything, int64(100)).Return(
					db.Product{
						ID:            100,
						Name:          "Coffee",
						Price:         750,
						IsAvailable:   true,
						CategoryID:    1,
						Sku:           "COF-100",
						StockQuantity: 50,
						CreatedAt:     now,
						UpdatedAt:     now,
					}, nil)
				m.On("CreateOrder", mock.Anything, mock.Anything).Return(
					db.CreateOrderRow{}, errors.New("db access failed"))
			},
			expectedErr: "db access failed",
		},
		{
			name:   "U8：DB Error CreateOrderItem",
			userID: int64(1),
			setupMock: func(m *testutil.MockDB) {
				now := time.Now()
				m.On("GetOrCreateCartForUser", mock.Anything, int64(1)).Return(
					db.Cart{
						ID:        10,
						UserID:    1,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)

				m.On("ListCartItemsByUser", mock.Anything, int64(1)).Return(
					[]db.ListCartItemsByUserRow{
						{
							ID:           1,
							CartID:       10,
							ProductID:    100,
							Quantity:     2,
							Price:        1500,
							CreatedAt:    now,
							UpdatedAt:    now,
							ProductName:  "Coffee",
							ProductPrice: 750,
							ProductStock: 50,
						},
					}, nil)

				m.On("GetProductForUpdate", mock.Anything, int64(100)).Return(
					db.Product{
						ID:            100,
						Name:          "Coffee",
						Price:         750,
						IsAvailable:   true,
						CategoryID:    1,
						Sku:           "COF-100",
						StockQuantity: 50,
						CreatedAt:     now,
						UpdatedAt:     now,
					}, nil)

				m.On("CreateOrder", mock.Anything, mock.Anything).Return(
					db.CreateOrderRow{
						ID:        1,
						UserID:    1,
						Status:    "pending",
						Total:     1500,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)

				m.On("CreateOrderItem", mock.Anything, mock.Anything).Return(
					db.OrderItem{}, errors.New("db access failed"))
			},
			expectedErr: "db access failed",
		},
		{
			name:   "U9：DB Error UpdateProductStock",
			userID: int64(1),
			setupMock: func(m *testutil.MockDB) {
				now := time.Now()
				m.On("GetOrCreateCartForUser", mock.Anything, int64(1)).Return(
					db.Cart{
						ID:        10,
						UserID:    1,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)

				m.On("ListCartItemsByUser", mock.Anything, int64(1)).Return(
					[]db.ListCartItemsByUserRow{
						{
							ID:           1,
							CartID:       10,
							ProductID:    100,
							Quantity:     2,
							Price:        1500,
							CreatedAt:    now,
							UpdatedAt:    now,
							ProductName:  "Coffee",
							ProductPrice: 750,
							ProductStock: 50,
						},
					}, nil)

				m.On("GetProductForUpdate", mock.Anything, int64(100)).Return(
					db.Product{
						ID:            100,
						Name:          "Coffee",
						Price:         750,
						IsAvailable:   true,
						CategoryID:    1,
						Sku:           "COF-100",
						StockQuantity: 50,
						CreatedAt:     now,
						UpdatedAt:     now,
					}, nil)

				m.On("CreateOrder", mock.Anything, mock.Anything).Return(
					db.CreateOrderRow{
						ID:        1,
						UserID:    1,
						Status:    "pending",
						Total:     1500,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)

				m.On("CreateOrderItem", mock.Anything, mock.Anything).Return(
					db.OrderItem{
						ID:                  11,
						OrderID:             1,
						ProductID:           100,
						Quantity:            2,
						UnitPrice:           750,
						ProductNameSnapshot: "Coffee",
						CreatedAt:           now,
						UpdatedAt:           now,
					}, nil)

				m.On("UpdateProductStock", mock.Anything, mock.Anything).Return(
					db.UpdateProductStockRow{}, errors.New("db access failed"))
			},
			expectedErr: "db access failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(testutil.MockDB)

			if tt.setupMock != nil {
				tt.setupMock(mockDB)
			}

			ctx := context.Background()
			order, err := createOrderLogic(ctx, mockDB, tt.userID)

			if tt.expectedErr != "" {
				assert.Error(t, err, tt.name)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err, tt.name)
				assert.NotNil(t, order, tt.name)
				assert.Equal(t, int64(1), order.ID, tt.name)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestCancelOrderLogic(t *testing.T) {
	tests := []struct {
		name        string
		orderID     int64
		userID      int64
		setupMock   func(*testutil.MockDB)
		expectedErr string
	}{
		{
			name:    "U1:単一商品のキャンセル",
			orderID: 1,
			userID:  1,
			setupMock: func(m *testutil.MockDB) {
				now := time.Now()
				m.On("GetOrderByIDForUpdate", mock.Anything, int64(1)).Return(
					db.GetOrderByIDForUpdateRow{
						ID: 1, UserID: 1, Total: 1500, Status: "pending", CreatedAt: now, UpdatedAt: now,
					}, nil)
				m.On("ListOrderItemsByOrderID", mock.Anything, int64(1)).Return(
					[]db.OrderItem{
						{
							ID: 1, OrderID: 1, ProductID: 100, Quantity: 2, UnitPrice: 750, CreatedAt: now, UpdatedAt: now,
						},
					}, nil)
				m.On("UpdateProductStock", mock.Anything, db.UpdateProductStockParams{ID: 100, StockQuantity: 2}).Return(
					db.UpdateProductStockRow{ID: 100, StockQuantity: 52}, nil)
				m.On("UpdateOrderStatus", mock.Anything, db.UpdateOrderStatusParams{ID: 1, Status: "cancelled"}).Return(
					db.UpdateOrderStatusRow{ID: 1, UserID: 1, Total: 1500, Status: "cancelled", CreatedAt: now, UpdatedAt: now}, nil)
			},
			expectedErr: "",
		},
		{
			name:    "U2:複数商品のキャンセル",
			orderID: 2,
			userID:  2,
			setupMock: func(m *testutil.MockDB) {
				now := time.Now()
				m.On("GetOrderByIDForUpdate", mock.Anything, int64(2)).Return(
					db.GetOrderByIDForUpdateRow{ID: 2, UserID: 2, Total: 3000, Status: "pending", CreatedAt: now, UpdatedAt: now}, nil)
				m.On("ListOrderItemsByOrderID", mock.Anything, int64(2)).Return(
					[]db.OrderItem{
						{ID: 1, OrderID: 2, ProductID: 101, Quantity: 1, UnitPrice: 1000, CreatedAt: now, UpdatedAt: now},
						{ID: 2, OrderID: 2, ProductID: 102, Quantity: 2, UnitPrice: 1000, CreatedAt: now, UpdatedAt: now},
					}, nil)
				m.On("UpdateProductStock", mock.Anything, db.UpdateProductStockParams{ID: 101, StockQuantity: 1}).Return(
					db.UpdateProductStockRow{ID: 101, StockQuantity: 11}, nil)
				m.On("UpdateProductStock", mock.Anything, db.UpdateProductStockParams{ID: 102, StockQuantity: 2}).Return(
					db.UpdateProductStockRow{ID: 102, StockQuantity: 22}, nil)
				m.On("UpdateOrderStatus", mock.Anything, db.UpdateOrderStatusParams{ID: 2, Status: "cancelled"}).Return(
					db.UpdateOrderStatusRow{ID: 2, UserID: 2, Total: 3000, Status: "cancelled", CreatedAt: now, UpdatedAt: now}, nil)
			},
			expectedErr: "",
		},
		{
			name:    "U3:注文なし",
			orderID: 10,
			userID:  1,
			setupMock: func(m *testutil.MockDB) {
				m.On("GetOrderByIDForUpdate", mock.Anything, int64(10)).Return(
					db.GetOrderByIDForUpdateRow{}, sql.ErrNoRows)
			},
			expectedErr: "注文が見つかりません",
		},
		{
			name:    "U4:他ユーザーの注文",
			orderID: 11,
			userID:  2,
			setupMock: func(m *testutil.MockDB) {
				now := time.Now()
				m.On("GetOrderByIDForUpdate", mock.Anything, int64(11)).Return(
					db.GetOrderByIDForUpdateRow{ID: 11, UserID: 1, Total: 1000, Status: "pending", CreatedAt: now, UpdatedAt: now}, nil)
			},
			expectedErr: "注文が見つかりません",
		},
		{
			name:    "U5: キャンセル済み",
			orderID: 12,
			userID:  3,
			setupMock: func(m *testutil.MockDB) {
				now := time.Now()
				m.On("GetOrderByIDForUpdate", mock.Anything, int64(12)).Return(
					db.GetOrderByIDForUpdateRow{ID: 12, UserID: 3, Total: 500, Status: "cancelled", CreatedAt: now, UpdatedAt: now}, nil)
			},
			expectedErr: "この注文はキャンセルできません",
		},
		{
			name:    "U6: DBエラー UpdateProductStock",
			orderID: 20,
			userID:  5,
			setupMock: func(m *testutil.MockDB) {
				now := time.Now()
				m.On("GetOrderByIDForUpdate", mock.Anything, int64(20)).Return(
					db.GetOrderByIDForUpdateRow{ID: 20, UserID: 5, Total: 800, Status: "pending", CreatedAt: now, UpdatedAt: now}, nil)
				m.On("ListOrderItemsByOrderID", mock.Anything, int64(20)).Return(
					[]db.OrderItem{
						{ID: 1, OrderID: 20, ProductID: 200, Quantity: 1, UnitPrice: 800, CreatedAt: now, UpdatedAt: now},
					}, nil)
				m.On("UpdateProductStock", mock.Anything, db.UpdateProductStockParams{ID: 200, StockQuantity: 1}).Return(
					db.UpdateProductStockRow{}, errors.New("db error"))
			},
			expectedErr: "db error",
		},
		{
			name:    "U7: DBエラー UpdateOrderStatus",
			orderID: 21,
			userID:  6,
			setupMock: func(m *testutil.MockDB) {
				now := time.Now()
				m.On("GetOrderByIDForUpdate", mock.Anything, int64(21)).Return(
					db.GetOrderByIDForUpdateRow{ID: 21, UserID: 6, Total: 1200, Status: "pending", CreatedAt: now, UpdatedAt: now}, nil)
				m.On("ListOrderItemsByOrderID", mock.Anything, int64(21)).Return(
					[]db.OrderItem{
						{ID: 1, OrderID: 21, ProductID: 201, Quantity: 1, UnitPrice: 1200, CreatedAt: now, UpdatedAt: now},
					}, nil)
				m.On("UpdateProductStock", mock.Anything, db.UpdateProductStockParams{ID: 201, StockQuantity: 1}).Return(
					db.UpdateProductStockRow{ID: 201, StockQuantity: 101}, nil)
				m.On("UpdateOrderStatus", mock.Anything, db.UpdateOrderStatusParams{ID: 21, Status: "cancelled"}).Return(
					db.UpdateOrderStatusRow{}, errors.New("update status error"), nil)
			},
			expectedErr: "update status error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(testutil.MockDB)
			if tt.setupMock != nil {
				tt.setupMock(mockDB)
			}
			result, err := cancelOrderLogic(context.Background(), mockDB, tt.orderID, tt.userID)
			if tt.expectedErr == "" {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			}
			mockDB.AssertExpectations(t)

		})
	}
}

func TestGetOrderLogic(t *testing.T) {
	tests := []struct {
		name           string
		userID         int64
		setupMock      func(*testutil.MockDB)
		expectedErr    string
		expectedOrders int
		expectedItems  int
	}{
		{
			name:           "U1：注文一覧取得成功",
			expectedErr:    "",
			userID:         1,
			expectedOrders: 1,
			expectedItems:  1,
			setupMock: func(m *testutil.MockDB) {
				now := time.Now()
				m.On("ListOrdersByUser", mock.Anything, int64(1)).Return(
					[]db.ListOrdersByUserRow{
						{
							ID:        1,
							UserID:    1,
							Total:     1,
							Status:    "pending",
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil)
				m.On("ListOrderItemsByOrderID", mock.Anything, int64(1)).Return(
					[]db.OrderItem{
						{
							ID:        1,
							OrderID:   1,
							ProductID: 100,
							Quantity:  50,
							UnitPrice: 750,
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil)
			},
		},
		{
			name:           "U2: 注文無し",
			expectedErr:    "",
			userID:         1,
			expectedOrders: 0,
			expectedItems:  0,
			setupMock: func(m *testutil.MockDB) {
				m.On("ListOrdersByUser", mock.Anything, int64(1)).Return(
					[]db.ListOrdersByUserRow{}, nil)
			},
		},
		{
			name:           "U3: DBエラー 注文情報",
			expectedErr:    "db error",
			userID:         1,
			expectedOrders: 0,
			expectedItems:  0,
			setupMock: func(m *testutil.MockDB) {
				m.On("ListOrdersByUser", mock.Anything, int64(1)).Return(
					[]db.ListOrdersByUserRow{}, errors.New("db error"))
			},
		},
		{
			name:           "U4: DBエラー 明細取得",
			expectedErr:    "db error",
			userID:         1,
			expectedOrders: 0,
			expectedItems:  0,
			setupMock: func(m *testutil.MockDB) {
				now := time.Now()
				m.On("ListOrdersByUser", mock.Anything, int64(1)).Return(
					[]db.ListOrdersByUserRow{
						{
							ID:        1,
							UserID:    1,
							Total:     1,
							Status:    "pending",
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil)
				m.On("ListOrderItemsByOrderID", mock.Anything, int64(1)).Return(
					[]db.OrderItem{}, errors.New("db error"))
			},
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(testutil.MockDB)

			if tt.setupMock != nil {
				tt.setupMock(mockDB)
			}

			ctx := context.Background()
			owi, err := getOrderLogic(ctx, mockDB, tt.userID)

			if tt.expectedErr != "" {
				assert.Error(t, err, tt.name)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err, tt.name)
				assert.Len(t, owi, tt.expectedOrders)
				if tt.expectedOrders > 0 {
					assert.Equal(t, int64(1), owi[0].Order.ID)
					assert.Len(t, owi[0].Items, tt.expectedItems)
					assert.Equal(t, int64(1), owi[0].Items[0].ID)
				}
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestGetOrdersHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Now()

	tests := []struct {
		name                        string
		query                       string
		userID                      any
		setupMock                   func(*testutil.MockDB)
		expectedStatus              int
		expectedCount               int
		expectedErrMsg              string
		expectedProductNameSnapshot string
	}{
		{
			name:           "U1: 注文確認成功 フィルタなし",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
			query:          "",
			userID:         int64(1),
			setupMock: func(m *testutil.MockDB) {
				m.On("ListOrdersByUser", mock.Anything, int64(1)).Return(
					[]db.ListOrdersByUserRow{
						{
							ID:        1,
							UserID:    1,
							Total:     1500,
							Status:    "pending",
							CreatedAt: now,
							UpdatedAt: now,
						},
						{
							ID:        2,
							UserID:    1,
							Total:     3000,
							Status:    "pending",
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil)
				m.On("ListOrderItemsByOrderID", mock.Anything, int64(1)).Return(
					[]db.OrderItem{
						{
							ID:        11,
							OrderID:   1,
							ProductID: 100,
							Quantity:  2,
							UnitPrice: 750,
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil)
				m.On("ListOrderItemsByOrderID", mock.Anything, int64(2)).Return(
					[]db.OrderItem{
						{
							ID:        21,
							OrderID:   2,
							ProductID: 200,
							Quantity:  3,
							UnitPrice: 1000,
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil)
			},
		},
		{
			name:           "U2: 注文確認成功 フィルタ=pending",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			query:          "?status=pending",
			userID:         int64(1),
			setupMock: func(m *testutil.MockDB) {
				m.On("ListOrdersByUser", mock.Anything, int64(1)).Return(
					[]db.ListOrdersByUserRow{
						{
							ID:        1,
							UserID:    1,
							Total:     1500,
							Status:    "pending",
							CreatedAt: now,
							UpdatedAt: now,
						},
						{
							ID:        2,
							UserID:    1,
							Total:     3000,
							Status:    "cancelled",
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil)

				m.On("ListOrderItemsByOrderID", mock.Anything, int64(1)).Return(
					[]db.OrderItem{
						{
							ID:        11,
							OrderID:   1,
							ProductID: 100,
							Quantity:  2,
							UnitPrice: 750,
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil)

				m.On("ListOrderItemsByOrderID", mock.Anything, int64(2)).Return(
					[]db.OrderItem{
						{
							ID:        21,
							OrderID:   2,
							ProductID: 200,
							Quantity:  3,
							UnitPrice: 1000,
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil)
			},
		},
		{
			name:           "U3: 注文確認成功 フィルタ=cancelled",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			query:          "?status=cancelled",
			userID:         int64(1),
			setupMock: func(m *testutil.MockDB) {
				m.On("ListOrdersByUser", mock.Anything, int64(1)).Return(
					[]db.ListOrdersByUserRow{
						{
							ID:        1,
							UserID:    1,
							Total:     1500,
							Status:    "pending",
							CreatedAt: now,
							UpdatedAt: now,
						},
						{
							ID:        2,
							UserID:    1,
							Total:     3000,
							Status:    "cancelled",
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil)

				m.On("ListOrderItemsByOrderID", mock.Anything, int64(1)).Return(
					[]db.OrderItem{
						{
							ID:        11,
							OrderID:   1,
							ProductID: 100,
							Quantity:  2,
							UnitPrice: 750,
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil)

				m.On("ListOrderItemsByOrderID", mock.Anything, int64(2)).Return(
					[]db.OrderItem{
						{
							ID:        21,
							OrderID:   2,
							ProductID: 200,
							Quantity:  3,
							UnitPrice: 1000,
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil)
			},
		},
		{
			name:           "U4: 注文確認成功(フィルタで0件表示) フィルタ=cancelled",
			expectedStatus: http.StatusOK,
			expectedCount:  0,
			query:          "?status=cancelled",
			userID:         int64(1),
			setupMock: func(m *testutil.MockDB) {
				m.On("ListOrdersByUser", mock.Anything, int64(1)).Return(
					[]db.ListOrdersByUserRow{
						{
							ID:        1,
							UserID:    1,
							Total:     1500,
							Status:    "pending",
							CreatedAt: now,
							UpdatedAt: now,
						},
						{
							ID:        2,
							UserID:    1,
							Total:     3000,
							Status:    "pending",
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil)

				m.On("ListOrderItemsByOrderID", mock.Anything, int64(1)).Return(
					[]db.OrderItem{
						{
							ID:        11,
							OrderID:   1,
							ProductID: 100,
							Quantity:  2,
							UnitPrice: 750,
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil)

				m.On("ListOrderItemsByOrderID", mock.Anything, int64(2)).Return(
					[]db.OrderItem{
						{
							ID:        21,
							OrderID:   2,
							ProductID: 200,
							Quantity:  3,
							UnitPrice: 1000,
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil)
			},
		},
		{
			name:           "U5: 注文失敗 フィルタ=invalid",
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
			expectedErrMsg: "無効なステータスです",
			query:          "?status=invalid",
			userID:         int64(1),
			setupMock:      nil,
		},
		{
			name:           "U6: userIDが存在しない",
			expectedStatus: http.StatusUnauthorized,
			expectedErrMsg: "認証が必要です",
			query:          "",
			userID:         nil,
			setupMock:      nil,
		},
		{
			name:           "U7: userIDの型がintでも通る",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
			query:          "",
			userID:         int(1),
			setupMock: func(m *testutil.MockDB) {
				m.On("ListOrdersByUser", mock.Anything, int64(1)).Return(
					[]db.ListOrdersByUserRow{
						{
							ID:        1,
							UserID:    1,
							Total:     1500,
							Status:    "pending",
							CreatedAt: now,
							UpdatedAt: now,
						},
						{
							ID:        2,
							UserID:    1,
							Total:     3000,
							Status:    "pending",
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil)
				m.On("ListOrderItemsByOrderID", mock.Anything, int64(1)).Return(
					[]db.OrderItem{
						{
							ID:        11,
							OrderID:   1,
							ProductID: 100,
							Quantity:  2,
							UnitPrice: 750,
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil)
				m.On("ListOrderItemsByOrderID", mock.Anything, int64(2)).Return(
					[]db.OrderItem{
						{
							ID:        21,
							OrderID:   2,
							ProductID: 200,
							Quantity:  3,
							UnitPrice: 1000,
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil)
			},
		},
		{
			name:           "U8: userIDの型がfloatでも通る",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
			query:          "",
			userID:         float64(1),
			setupMock: func(m *testutil.MockDB) {
				m.On("ListOrdersByUser", mock.Anything, int64(1)).Return(
					[]db.ListOrdersByUserRow{
						{
							ID:        1,
							UserID:    1,
							Total:     1500,
							Status:    "pending",
							CreatedAt: now,
							UpdatedAt: now,
						},
						{
							ID:        2,
							UserID:    1,
							Total:     3000,
							Status:    "pending",
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil)
				m.On("ListOrderItemsByOrderID", mock.Anything, int64(1)).Return(
					[]db.OrderItem{
						{
							ID:        11,
							OrderID:   1,
							ProductID: 100,
							Quantity:  2,
							UnitPrice: 750,
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil)
				m.On("ListOrderItemsByOrderID", mock.Anything, int64(2)).Return(
					[]db.OrderItem{
						{
							ID:        21,
							OrderID:   2,
							ProductID: 200,
							Quantity:  3,
							UnitPrice: 1000,
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil)
			},
		},
		{
			name:           "U9: userIDの型が不正",
			expectedStatus: http.StatusUnauthorized,
			query:          "",
			userID:         "not-a-number",
			expectedErrMsg: "認証が必要です",
			setupMock:      nil,
		},
		{
			name:                        "U10: 明細にproduct_name_snapshotが含まれる",
			expectedStatus:              http.StatusOK,
			expectedCount:               1,
			userID:                      int64(1),
			expectedProductNameSnapshot: "House Blend",
			setupMock: func(m *testutil.MockDB) {
				m.On("ListOrdersByUser", mock.Anything, int64(1)).Return(
					[]db.ListOrdersByUserRow{
						{
							ID:        1,
							UserID:    1,
							Total:     1500,
							Status:    "pending",
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil)
				m.On("ListOrderItemsByOrderID", mock.Anything, int64(1)).Return(
					[]db.OrderItem{
						{
							ID:                  11,
							OrderID:             1,
							ProductID:           100,
							Quantity:            2,
							UnitPrice:           750,
							ProductNameSnapshot: "House Blend",
							CreatedAt:           now,
							UpdatedAt:           now,
						},
					}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(testutil.MockDB)
			if tt.setupMock != nil {
				tt.setupMock(mockDB)
			}

			router := gin.New()
			router.GET("/api/orders", func(c *gin.Context) {
				if tt.userID != nil {
					c.Set("userID", tt.userID)
				}
				GetOrdersHandler(mockDB)(c)
			})

			req := httptest.NewRequest(http.MethodGet, "/api/orders"+tt.query, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var body struct {
					Orders []OrderWithItems `json:"orders"`
				}
				err := json.Unmarshal(w.Body.Bytes(), &body)
				assert.NoError(t, err)
				assert.Len(t, body.Orders, tt.expectedCount)

				if tt.query == "?status=pending" && len(body.Orders) > 0 {
					for _, order := range body.Orders {
						assert.Equal(t, "pending", order.Order.Status)
					}
				}
				if tt.query == "?status=cancelled" && len(body.Orders) > 0 {
					for _, order := range body.Orders {
						assert.Equal(t, "cancelled", order.Order.Status)
					}
				}
				if tt.expectedProductNameSnapshot != "" && len(body.Orders) > 0 && len(body.Orders[0].Items) > 0 {
					assert.Equal(t, tt.expectedProductNameSnapshot, body.Orders[0].Items[0].ProductNameSnapshot)
				}
			} else {
				var body struct {
					Error string `json:"error"`
				}
				err := json.Unmarshal(w.Body.Bytes(), &body)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedErrMsg, body.Error)
			}
			mockDB.AssertExpectations(t)

		})
	}
}
