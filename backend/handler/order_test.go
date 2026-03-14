package handler

import (
	"context"
	"database/sql"
	"errors"
	"sol_coffeesys/backend/db"
	"sol_coffeesys/backend/handler/testutil"
	"testing"
	"time"

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
