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
			} else if tt.expectedStatus == http.StatusInternalServerError {
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
			name:   "invalid quantity",
			userID: int64(43),
			body:   map[string]interface{}{"prodcut_id": 110, "quantity": -10},
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
				m.On("GetOrCreateCartForUser", mock.Anything, int64(43)).Return(
					db.Cart{
						ID:     11,
						UserID: 43,
					}, nil)
				m.On("AddCartItem", mock.Anything, mock.Anything).Return(
					db.CartItem{}, nil)
			},
			expectedStatus: http.StatusUnauthorized,
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

			b, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/cart/items", bytes.NewBuffer(b))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockDB.AssertExpectations(t)
		})
	}
}
