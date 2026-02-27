package handler_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sol_coffeesys/backend/db"
	"sol_coffeesys/backend/handler"
	testutil "sol_coffeesys/backend/handler/testutil"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetCartHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name            string
		userID          interface{}
		setupMock       func(*testutil.MockDB)
		expectedStatus  int
		expectedItemLen int
		expectedErr     bool
	}{
		{
			name: "cart with items",
			setupMock: func(m *testutil.MockDB) {
				now := time.Now()
				m.On("ListCartItemsByUser", mock.Anything, int64(42)).Return(
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
						{
							ID:           2,
							CartID:       10,
							ProductID:    101,
							Quantity:     1,
							Price:        2000,
							CreatedAt:    now,
							UpdatedAt:    now,
							ProductName:  "Tea",
							ProductPrice: 2000,
							ProductStock: 30,
						},
					}, nil)
			},
			userID:          int64(42),
			expectedStatus:  http.StatusOK,
			expectedItemLen: 2,
		},
		{
			name:           "missing userID",
			setupMock:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:            "invalid type userID",
			setupMock:       nil,
			userID:          "not-an-id",
			expectedStatus:  http.StatusUnauthorized,
			expectedItemLen: 0,
		},
		{
			name: "userID as Int",
			setupMock: func(m *testutil.MockDB) {
				now := time.Now()
				m.On("ListCartItemsByUser", mock.Anything, int64(45)).Return(
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
						{
							ID:           2,
							CartID:       10,
							ProductID:    101,
							Quantity:     1,
							Price:        2000,
							CreatedAt:    now,
							UpdatedAt:    now,
							ProductName:  "Tea",
							ProductPrice: 2000,
							ProductStock: 30,
						},
					}, nil)
			},
			userID:          int(45),
			expectedStatus:  http.StatusOK,
			expectedItemLen: 2,
		},
		{
			name: "userID as float64",
			setupMock: func(m *testutil.MockDB) {
				now := time.Now()
				m.On("ListCartItemsByUser", mock.Anything, int64(46)).Return(
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
						{
							ID:           2,
							CartID:       10,
							ProductID:    101,
							Quantity:     1,
							Price:        2000,
							CreatedAt:    now,
							UpdatedAt:    now,
							ProductName:  "Tea",
							ProductPrice: 2000,
							ProductStock: 30,
						},
					}, nil)
			},
			userID:          float64(46),
			expectedStatus:  http.StatusOK,
			expectedItemLen: 2,
		},
		{
			name: "empty cart",
			setupMock: func(m *testutil.MockDB) {
				m.On("ListCartItemsByUser", mock.Anything, int64(99)).Return(
					[]db.ListCartItemsByUserRow{},
					nil,
				)
			},
			userID:          99,
			expectedStatus:  http.StatusOK,
			expectedItemLen: 0,
		},
		{
			name: "db error",
			setupMock: func(m *testutil.MockDB) {
				m.On("ListCartItemsByUser", mock.Anything, int64(100)).Return(
					nil,
					errors.New("db connection failed"),
				)
			},
			userID:          100,
			expectedStatus:  http.StatusInternalServerError,
			expectedItemLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			mockDB := new(testutil.MockDB)

			if tt.setupMock != nil {
				tt.setupMock(mockDB)
			}
			if tt.name == "missing userID" {
				router.GET("/api/cart", func(c *gin.Context) {
					handler.GetCartHandler(mockDB)(c)
				})
			} else {
				router.GET("/api/cart", func(c *gin.Context) {
					c.Set("userID", tt.userID)
					handler.GetCartHandler(mockDB)(c)
				})
			}
			req := httptest.NewRequest(http.MethodGet, "/api/cart", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var body map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &body)
				assert.NoError(t, err)
				items, ok := body["items"].([]interface{})
				assert.True(t, ok, "items should be an array")
				assert.Equal(t, tt.expectedItemLen, len(items))
			}
			if tt.expectedStatus == http.StatusInternalServerError {
				var errBody map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &errBody)
				assert.Equal(t, "予期せぬエラーが発生しました", errBody["error"])
			}
			mockDB.AssertExpectations(t)
		})
	}
}

func TestAddToCartHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         interface{}
		body           map[string]interface{}
		setupMock      func(*testutil.MockDB)
		expectedStatus int
	}{
		{
			name:   "success add new item",
			userID: int64(42),
			body:   map[string]interface{}{"product_id": 100, "quantity": 2},
			setupMock: func(m *testutil.MockDB) {
				now := time.Now()
				m.On("GetProduct", mock.Anything, int64(100)).Return(
					db.Product{
						ID:            100,
						Name:          "Coffee",
						Price:         750,
						StockQuantity: 50,
						CreatedAt:     now,
						UpdatedAt:     now,
					}, nil)
				m.On("GetOrCreateCartForUser", mock.Anything, int64(42)).Return(
					db.Cart{
						ID:     10,
						UserID: 42,
					}, nil)
				m.On("AddCartItem", mock.Anything, mock.Anything).Return(
					db.CartItem{
						ID:        1,
						CartID:    10,
						ProductID: 100,
						Quantity:  2,
						Price:     1500,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid quantity (zero)",
			userID:         int64(42),
			body:           map[string]interface{}{"product_id": 100, "quantity": 0},
			setupMock:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "product not found",
			userID: int64(42),
			body:   map[string]interface{}{"product_id": 999, "quantity": 2},
			setupMock: func(m *testutil.MockDB) {
				m.On("GetProduct", mock.Anything, int64(999)).Return(
					db.Product{}, sql.ErrNoRows)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid quantity",
			userID:         int64(43),
			body:           map[string]interface{}{"prodcut_id": 110, "quantity": -10},
			setupMock:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthorized",
			userID:         nil,
			body:           map[string]interface{}{"product_id": 100, "quantity": 1},
			setupMock:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:   "db error on add",
			userID: int64(44),
			body:   map[string]interface{}{"product_id": 100, "quantity": 1},
			setupMock: func(m *testutil.MockDB) {
				now := time.Now()
				m.On("GetProduct", mock.Anything, int64(100)).Return(
					db.Product{
						ID:            100,
						Name:          "Coffee",
						Price:         750,
						StockQuantity: 50,
						CreatedAt:     now,
						UpdatedAt:     now,
					}, nil)
				m.On("GetOrCreateCartForUser", mock.Anything, int64(44)).Return(
					db.Cart{
						ID:     12,
						UserID: 44,
					}, nil)
				m.On("AddCartItem", mock.Anything, mock.Anything).Return(db.CartItem{}, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:   "db error on getproduct",
			userID: int64(44),
			body:   map[string]interface{}{"product_id": 100, "quantity": 1},
			setupMock: func(m *testutil.MockDB) {
				m.On("GetProduct", mock.Anything, int64(100)).Return(
					db.Product{}, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:   "db error on get or create cart for user",
			userID: int64(44),
			body:   map[string]interface{}{"product_id": 100, "quantity": 1},
			setupMock: func(m *testutil.MockDB) {
				now := time.Now()
				m.On("GetProduct", mock.Anything, int64(100)).Return(
					db.Product{
						ID:            100,
						Name:          "Coffee",
						Price:         750,
						StockQuantity: 50,
						CreatedAt:     now,
						UpdatedAt:     now,
					}, nil)
				m.On("GetOrCreateCartForUser", mock.Anything, int64(44)).Return(
					db.Cart{}, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:   "userID as int",
			userID: int(42),
			body:   map[string]interface{}{"product_id": 100, "quantity": 2},
			setupMock: func(m *testutil.MockDB) {
				now := time.Now()
				m.On("GetProduct", mock.Anything, int64(100)).Return(
					db.Product{
						ID:            100,
						Name:          "Coffee",
						Price:         750,
						StockQuantity: 50,
						CreatedAt:     now,
						UpdatedAt:     now,
					}, nil)
				m.On("GetOrCreateCartForUser", mock.Anything, int64(42)).Return(
					db.Cart{
						ID:     10,
						UserID: 42,
					}, nil)
				m.On("AddCartItem", mock.Anything, mock.Anything).Return(
					db.CartItem{
						ID:        1,
						CartID:    10,
						ProductID: 100,
						Quantity:  2,
						Price:     1500,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:   "userID as float",
			userID: float64(42),
			body:   map[string]interface{}{"product_id": 100, "quantity": 2},
			setupMock: func(m *testutil.MockDB) {
				now := time.Now()
				m.On("GetProduct", mock.Anything, int64(100)).Return(
					db.Product{
						ID:            100,
						Name:          "Coffee",
						Price:         750,
						StockQuantity: 50,
						CreatedAt:     now,
						UpdatedAt:     now,
					}, nil)
				m.On("GetOrCreateCartForUser", mock.Anything, int64(42)).Return(
					db.Cart{
						ID:     10,
						UserID: 42,
					}, nil)
				m.On("AddCartItem", mock.Anything, mock.Anything).Return(
					db.CartItem{
						ID:        1,
						CartID:    10,
						ProductID: 100,
						Quantity:  2,
						Price:     1500,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "missing userID",
			body:           map[string]interface{}{"product_id": 100, "quantity": 2},
			setupMock:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "unauthorized",
			userID:         nil,
			body:           map[string]interface{}{"product_id": 100, "quantity": 2},
			setupMock:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid type userID",
			userID:         "not-an-id",
			body:           map[string]interface{}{"product_id": 100, "quantity": 2},
			setupMock:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid JSON type",
			userID:         int64(50),
			body:           nil,
			setupMock:      nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			mockDB := new(testutil.MockDB)
			if tt.setupMock != nil {
				tt.setupMock(mockDB)
			}

			router.POST("/api/cart/items", func(c *gin.Context) {
				if tt.userID != nil {
					c.Set("userID", tt.userID)
				}
				handler.AddToCartHandler(mockDB)(c)
			})

			var b []byte
			if tt.name == "invalid JSON type" {
				b = []byte(`{broken json`)
			} else {
				b, _ = json.Marshal(tt.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/cart/items", bytes.NewBuffer(b))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockDB.AssertExpectations(t)
		})
	}
}

func TestUpdateCartItemHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name           string
		itemID         string
		userID         interface{}
		body           map[string]interface{}
		setupMock      func(*testutil.MockDB)
		expectedStatus int
	}{
		{
			name:           "success",
			itemID:         "1",
			userID:         int64(42),
			expectedStatus: http.StatusOK,
			body:           map[string]interface{}{"quantity": 5},
			setupMock: func(m *testutil.MockDB) {
				now := time.Now()
				m.On("UpdateCartItemQtyByUser", mock.Anything, db.UpdateCartItemQtyByUserParams{
					ID:       1,
					Quantity: 5,
					UserID:   42,
				}).Return(
					db.CartItem{
						ID:        1,
						CartID:    10,
						ProductID: 100,
						Quantity:  5,
						Price:     1500,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
			},
		},
		{
			name:           "invalid quantity(zero)",
			expectedStatus: http.StatusBadRequest,
			userID:         int64(42),
			itemID:         "1",
			body:           map[string]interface{}{"quantyty": 0},
			setupMock:      nil,
		},
		{
			name:           "invalid quantity(negative)",
			expectedStatus: http.StatusBadRequest,
			userID:         int64(42),
			itemID:         "1",
			body:           map[string]interface{}{"quantity": -10},
			setupMock:      nil,
		},
		{
			name:           "unauthorized",
			userID:         nil,
			itemID:         "1",
			expectedStatus: http.StatusUnauthorized,
			setupMock:      nil,
		},
		{
			name:           "invalid id param",
			userID:         int64(42),
			itemID:         "abc",
			expectedStatus: http.StatusBadRequest,
			setupMock:      nil,
		},
		{
			name:           "item not found or not owned",
			userID:         int64(42),
			itemID:         "999",
			body:           map[string]interface{}{"quantity": 10},
			expectedStatus: http.StatusNotFound,
			setupMock: func(m *testutil.MockDB) {
				m.On("UpdateCartItemQtyByUser", mock.Anything, db.UpdateCartItemQtyByUserParams{
					ID:       999,
					Quantity: 10,
					UserID:   42,
				}).Return(db.CartItem{}, sql.ErrNoRows)
			},
		},
		{
			name:           "db error",
			userID:         int64(42),
			itemID:         "1",
			body:           map[string]interface{}{"quantity": 3},
			expectedStatus: http.StatusInternalServerError,
			setupMock: func(m *testutil.MockDB) {
				m.On("UpdateCartItemQtyByUser", mock.Anything, db.UpdateCartItemQtyByUserParams{
					ID:       1,
					Quantity: 3,
					UserID:   42,
				}).Return(db.CartItem{}, errors.New("db connection failed"))
			},
		},
		{
			name:           "userID as int",
			itemID:         "1",
			userID:         int(42),
			expectedStatus: http.StatusOK,
			body:           map[string]interface{}{"quantity": 5},
			setupMock: func(m *testutil.MockDB) {
				now := time.Now()
				m.On("UpdateCartItemQtyByUser", mock.Anything, db.UpdateCartItemQtyByUserParams{
					ID:       1,
					Quantity: 5,
					UserID:   42,
				}).Return(
					db.CartItem{
						ID:        1,
						CartID:    10,
						ProductID: 100,
						Quantity:  5,
						Price:     1500,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
			},
		},
		{
			name:           "userID as float",
			itemID:         "1",
			userID:         float64(42),
			expectedStatus: http.StatusOK,
			body:           map[string]interface{}{"quantity": 5},
			setupMock: func(m *testutil.MockDB) {
				now := time.Now()
				m.On("UpdateCartItemQtyByUser", mock.Anything, db.UpdateCartItemQtyByUserParams{
					ID:       1,
					Quantity: 5,
					UserID:   42,
				}).Return(
					db.CartItem{
						ID:        1,
						CartID:    10,
						ProductID: 100,
						Quantity:  5,
						Price:     1500,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
			},
		},
		{
			name:           "missing userID",
			itemID:         "1",
			expectedStatus: http.StatusUnauthorized,
			setupMock:      nil,
		},
		{
			name:           "invalid type userID",
			userID:         "non-an-id",
			itemID:         "1",
			body:           map[string]interface{}{"quantity": 1},
			expectedStatus: http.StatusUnauthorized,
			setupMock:      nil,
		},
		{
			name:           "invalid JSON type",
			userID:         int64(50),
			itemID:         "1",
			body:           nil,
			expectedStatus: http.StatusBadRequest,
			setupMock:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			mockDB := new(testutil.MockDB)
			if tt.setupMock != nil {
				tt.setupMock(mockDB)
			}

			router.PUT("/api/cart/items/:id", func(c *gin.Context) {
				if tt.userID != nil {
					c.Set("userID", tt.userID)
				}
				handler.UpdateCartItemHandler(mockDB)(c)
			})

			var b []byte
			if tt.name == "invalid JSON type" {
				b = []byte(`{broken json`)
			} else {
				b, _ = json.Marshal(tt.body)
			}
			req := httptest.NewRequest(http.MethodPut, "/api/cart/items/"+tt.itemID, bytes.NewBuffer(b))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
			mockDB.AssertExpectations(t)
		})
	}
}

func TestRemoveCartItemHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name           string
		itemID         string
		userID         interface{}
		setupmock      func(*testutil.MockDB)
		expectedStatus int
	}{
		{
			name:           "success",
			expectedStatus: http.StatusNoContent,
			userID:         int64(42),
			itemID:         "1",
			setupmock: func(m *testutil.MockDB) {
				now := time.Now()
				m.On("GetCartItemByID", mock.Anything, int64(42)).Return(
					db.CartItem{
						ID:        1,
						CartID:    10,
						ProductID: 100,
						Quantity:  5,
						Price:     1500,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
				m.On("GetCartByUser", mock.Anything, int64(42)).Return(
					db.Cart{
						ID:        10,
						UserID:    42,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
				m.On("RemoveCartItemByUser", mock.Anything, db.RemoveCartItemByUserParams{ID: 1, UserID: 42}).Return(nil)
			},
		},
		{
			name:           "item not found",
			expectedStatus: http.StatusNotFound,
			userID:         int64(42),
			itemID:         "999",
			setupmock: func(m *testutil.MockDB) {
				m.On("GetCartItemByID", mock.Anything, int64(999)).Return(
					db.CartItem{}, sql.ErrNoRows,
				)
			},
		},
		{
			name:           "item belongs to another user",
			expectedStatus: http.StatusNotFound,
			userID:         int64(42),
			itemID:         "5",
			setupmock: func(m *testutil.MockDB) {
				now := time.Now()
				m.On("GetCartItemByID", mock.Anything, int64(5)).Return(
					db.CartItem{
						ID:        1,
						CartID:    99,
						ProductID: 100,
						Quantity:  5,
						Price:     1500,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
				m.On("GetCartByUser", mock.Anything, int64(42)).Return(
					db.Cart{
						ID:        10,
						UserID:    42,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
			},
		},
		{
			name:           "user has no cart",
			expectedStatus: http.StatusNotFound,
			userID:         int64(42),
			itemID:         "5",
			setupmock: func(m *testutil.MockDB) {
				now := time.Now()
				m.On("GetCartItemByID", mock.Anything, int64(5)).Return(
					db.CartItem{
						ID:        1,
						CartID:    99,
						ProductID: 100,
						Quantity:  5,
						Price:     1500,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
				m.On("GetCartByUser", mock.Anything, int64(42)).Return(
					db.Cart{}, sql.ErrNoRows,
				)
			},
		},
		{
			name:           "unauthorized(userID nil)",
			expectedStatus: http.StatusUnauthorized,
			userID:         nil,
			itemID:         "1",
			setupmock:      nil,
		},
		{
			name:           "invalid id param",
			expectedStatus: http.StatusBadRequest,
			userID:         int64(42),
			itemID:         "abc",
			setupmock:      nil,
		},
		{
			name:           "db error on GetCartItemByID",
			expectedStatus: http.StatusInternalServerError,
			userID:         int64(42),
			itemID:         "1",
			setupmock: func(m *testutil.MockDB) {
				m.On("GetCartItemByID", mock.Anything, int64(42)).Return(
					db.CartItem{}, errors.New("db access failed"),
				)
			},
		},
		{
			name:           "db error on RemoveCartItemByUser",
			expectedStatus: http.StatusInternalServerError,
			userID:         int64(42),
			itemID:         "1",
			setupmock: func(m *testutil.MockDB) {
				now := time.Now()
				m.On("GetCartItemByID", mock.Anything, int64(42)).Return(
					db.CartItem{
						ID:        1,
						CartID:    10,
						ProductID: 100,
						Quantity:  5,
						Price:     1500,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
				m.On("GetCartByUser", mock.Anything, int64(42)).Return(
					db.Cart{
						ID:        10,
						UserID:    42,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
				m.On("RemoveCartItemByUser", mock.Anything, db.RemoveCartItemByUserParams{ID: 1, UserID: 42}).Return(errors.New("db access failed"))
			},
		},
		{
			name:           "userID as int",
			expectedStatus: http.StatusNoContent,
			userID:         int(42),
			itemID:         "1",
			setupmock: func(m *testutil.MockDB) {
				now := time.Now()
				m.On("GetCartItemByID", mock.Anything, int64(42)).Return(
					db.CartItem{
						ID:        1,
						CartID:    10,
						ProductID: 100,
						Quantity:  5,
						Price:     1500,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
				m.On("GetCartByUser", mock.Anything, int64(42)).Return(
					db.Cart{
						ID:        10,
						UserID:    42,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
				m.On("RemoveCartItemByUser", mock.Anything, db.RemoveCartItemByUserParams{ID: 1, UserID: 42}).Return(nil)
			},
		},
		{
			name:           "userID as float64",
			expectedStatus: http.StatusNoContent,
			userID:         float64(42),
			itemID:         "1",
			setupmock: func(m *testutil.MockDB) {
				now := time.Now()
				m.On("GetCartItemByID", mock.Anything, int64(42)).Return(
					db.CartItem{
						ID:        1,
						CartID:    10,
						ProductID: 100,
						Quantity:  5,
						Price:     1500,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
				m.On("GetCartByUser", mock.Anything, int64(42)).Return(
					db.Cart{
						ID:        10,
						UserID:    42,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
				m.On("RemoveCartItemByUser", mock.Anything, db.RemoveCartItemByUserParams{ID: 1, UserID: 42}).Return(nil)
			},
		},
		{
			name:           "invalid type userID",
			expectedStatus: http.StatusUnauthorized,
			userID:         "not-an-id",
			itemID:         "1",
			setupmock:      nil,
		},
		{
			name:           "missing userID",
			expectedStatus: http.StatusUnauthorized,
			itemID:         "1",
			setupmock:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			mockDB := new(testutil.MockDB)

			if tt.setupmock != nil {
				tt.setupmock(mockDB)
			}

			router.DELETE("/api/cart/items/:id", func(c *gin.Context) {
				if tt.userID != nil {
					c.Set("userID", tt.userID)
				}
				handler.RemoveCartItemHandler(mockDB)(c)
			})

			req := httptest.NewRequest(http.MethodDelete, "/api/cart/items/"+tt.itemID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
			mockDB.AssertExpectations(t)

		})
	}
}
