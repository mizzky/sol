package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
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

// Table-driven tests for important flow cases:
// 1) Unauthorized across flow
// 2) DB error during Add -> ensure Add fails and Get returns empty
// 3) Ownership violation on Remove
func TestCartFlow_TableTopCases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type step struct {
		method string
		path   string
		body   interface{}
		want   int
	}

	cases := []struct {
		name      string
		setupMock func(m *testutil.MockDB)
		withAuth  bool // whether handlers set userID into context
		steps     []step
	}{
		{
			name:     "happy path (add -> get -> update -> remove -> clear)",
			withAuth: true,
			setupMock: func(m *testutil.MockDB) {
				now := time.Now()
				// GetProduct
				m.On("GetProduct", mock.Anything, int64(100)).Return(
					db.Product{ID: 100, Name: "Coffee", Price: 750, StockQuantity: 50, CreatedAt: now, UpdatedAt: now}, nil)
				// GetOrCreateCartForUser
				m.On("GetOrCreateCartForUser", mock.Anything, int64(42)).Return(db.Cart{ID: 10, UserID: 42}, nil)
				// AddCartItem
				m.On("AddCartItem", mock.Anything, mock.Anything).Return(db.CartItem{ID: 1, CartID: 10, ProductID: 100, Quantity: 2, Price: 1500, CreatedAt: now, UpdatedAt: now}, nil)
				// ListCartItemsByUser
				m.On("ListCartItemsByUser", mock.Anything, int64(42)).Return([]db.ListCartItemsByUserRow{
					{ID: 1, CartID: 10, ProductID: 100, Quantity: 2, Price: 1500, CreatedAt: now, UpdatedAt: now, ProductName: "Coffee", ProductPrice: 750, ProductStock: 50},
				}, nil)
				// UpdateCartItemQtyByUser
				m.On("UpdateCartItemQtyByUser", mock.Anything, db.UpdateCartItemQtyByUserParams{ID: 1, Quantity: 5, UserID: 42}).Return(db.CartItem{ID: 1, CartID: 10, ProductID: 100, Quantity: 5, Price: 1500, CreatedAt: now, UpdatedAt: now}, nil)
				// GetCartItemByID / GetCartByUser / RemoveCartItemByUser / ClearCartByUser
				m.On("GetCartItemByID", mock.Anything, int64(1)).Return(db.CartItem{ID: 1, CartID: 10, ProductID: 100, Quantity: 5, Price: 1500, CreatedAt: now, UpdatedAt: now}, nil)
				m.On("GetCartByUser", mock.Anything, int64(42)).Return(db.Cart{ID: 10, UserID: 42}, nil)
				m.On("RemoveCartItemByUser", mock.Anything, db.RemoveCartItemByUserParams{ID: 1, UserID: 42}).Return(nil)
				m.On("ClearCartByUser", mock.Anything, int64(42)).Return(nil)
			},
			steps: []step{
				{method: http.MethodPost, path: "/api/cart/items", body: map[string]interface{}{"product_id": 100, "quantity": 2}, want: http.StatusCreated},
				{method: http.MethodGet, path: "/api/cart", body: nil, want: http.StatusOK},
				{method: http.MethodPut, path: "/api/cart/items/1", body: map[string]interface{}{"quantity": 5}, want: http.StatusOK},
				{method: http.MethodDelete, path: "/api/cart/items/1", body: nil, want: http.StatusNoContent},
				{method: http.MethodDelete, path: "/api/cart", body: nil, want: http.StatusNoContent},
			},
		},
		{
			name:     "add product not found -> expect 404 then get empty",
			withAuth: true,
			setupMock: func(m *testutil.MockDB) {
				m.On("GetProduct", mock.Anything, int64(999)).Return(db.Product{}, sql.ErrNoRows)
				// ensure ListCartItemsByUser returns empty
				m.On("ListCartItemsByUser", mock.Anything, int64(42)).Return([]db.ListCartItemsByUserRow{}, nil)
			},
			steps: []step{
				{method: http.MethodPost, path: "/api/cart/items", body: map[string]interface{}{"product_id": 999, "quantity": 1}, want: http.StatusNotFound},
				{method: http.MethodGet, path: "/api/cart", body: nil, want: http.StatusOK},
			},
		},
		{
			name:     "unauthorized across flow",
			withAuth: false,
			setupMock: func(m *testutil.MockDB) {
				// no DB expectations needed; requests should short-circuit at auth check
			},
			steps: []step{
				{method: http.MethodPost, path: "/api/cart/items", body: map[string]interface{}{"product_id": 100, "quantity": 1}, want: http.StatusUnauthorized},
				{method: http.MethodGet, path: "/api/cart", body: nil, want: http.StatusUnauthorized},
				{method: http.MethodPut, path: "/api/cart/items/1", body: map[string]interface{}{"quantity": 2}, want: http.StatusUnauthorized},
			},
		},
		{
			name:     "db error during add -> get returns empty",
			withAuth: true,
			setupMock: func(m *testutil.MockDB) {
				now := time.Now()
				// GetProduct ok
				m.On("GetProduct", mock.Anything, int64(100)).Return(db.Product{
					ID: 100, Name: "Coffee", Price: 750, StockQuantity: 50, CreatedAt: now, UpdatedAt: now,
				}, nil)
				// GetOrCreateCartForUser ok
				m.On("GetOrCreateCartForUser", mock.Anything, int64(42)).Return(db.Cart{ID: 10, UserID: 42}, nil)
				// AddCartItem fails (simulate DB error)
				m.On("AddCartItem", mock.Anything, mock.Anything).Return(db.CartItem{}, sql.ErrConnDone)
				// ListCartItemsByUser returns empty after failed add
				m.On("ListCartItemsByUser", mock.Anything, int64(42)).Return([]db.ListCartItemsByUserRow{}, nil)
			},
			steps: []step{
				{method: http.MethodPost, path: "/api/cart/items", body: map[string]interface{}{"product_id": 100, "quantity": 2}, want: http.StatusInternalServerError},
				{method: http.MethodGet, path: "/api/cart", body: nil, want: http.StatusOK},
			},
		},
		{
			name:     "ownership violation on remove",
			withAuth: true,
			setupMock: func(m *testutil.MockDB) {
				now := time.Now()
				// Add flow
				m.On("GetProduct", mock.Anything, int64(100)).Return(db.Product{
					ID: 100, Name: "Coffee", Price: 750, StockQuantity: 50, CreatedAt: now, UpdatedAt: now,
				}, nil)
				m.On("GetOrCreateCartForUser", mock.Anything, int64(42)).Return(db.Cart{ID: 10, UserID: 42}, nil)
				m.On("AddCartItem", mock.Anything, mock.Anything).Return(db.CartItem{
					ID: 1, CartID: 10, ProductID: 100, Quantity: 2, Price: 1500, CreatedAt: now, UpdatedAt: now,
				}, nil)
				// Get returns the item
				m.On("ListCartItemsByUser", mock.Anything, int64(42)).Return([]db.ListCartItemsByUserRow{
					{
						ID: 1, CartID: 10, ProductID: 100, Quantity: 2, Price: 1500, CreatedAt: now, UpdatedAt: now,
						ProductName: "Coffee", ProductPrice: 750, ProductStock: 50,
					},
				}, nil)
				// Remove flow: GetCartItemByID returns item with cartID=10
				m.On("GetCartItemByID", mock.Anything, int64(1)).Return(db.CartItem{
					ID: 1, CartID: 10, ProductID: 100, Quantity: 2, Price: 1500, CreatedAt: now, UpdatedAt: now,
				}, nil)
				// But GetCartByUser returns a different cart (ownership violation)
				m.On("GetCartByUser", mock.Anything, int64(42)).Return(db.Cart{ID: 99, UserID: 42}, nil)
			},
			steps: []step{
				{method: http.MethodPost, path: "/api/cart/items", body: map[string]interface{}{"product_id": 100, "quantity": 2}, want: http.StatusCreated},
				{method: http.MethodGet, path: "/api/cart", body: nil, want: http.StatusOK},
				{method: http.MethodDelete, path: "/api/cart/items/1", body: nil, want: http.StatusNotFound},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			router := gin.New()
			mockDB := new(testutil.MockDB)
			if tc.setupMock != nil {
				tc.setupMock(mockDB)
			}

			// register handlers; optionally set auth into context depending on case
			if tc.withAuth {
				router.POST("/api/cart/items", func(c *gin.Context) { c.Set("userID", int64(42)); handler.AddToCartHandler(mockDB)(c) })
				router.GET("/api/cart", func(c *gin.Context) { c.Set("userID", int64(42)); handler.GetCartHandler(mockDB)(c) })
				router.PUT("/api/cart/items/:id", func(c *gin.Context) { c.Set("userID", int64(42)); handler.UpdateCartItemHandler(mockDB)(c) })
				router.DELETE("/api/cart/items/:id", func(c *gin.Context) { c.Set("userID", int64(42)); handler.RemoveCartItemHandler(mockDB)(c) })
				router.DELETE("/api/cart", func(c *gin.Context) { c.Set("userID", int64(42)); handler.ClearCartHandler(mockDB)(c) })
			} else {
				// register handlers without setting userID to simulate unauthorized
				router.POST("/api/cart/items", func(c *gin.Context) { handler.AddToCartHandler(mockDB)(c) })
				router.GET("/api/cart", func(c *gin.Context) { handler.GetCartHandler(mockDB)(c) })
				router.PUT("/api/cart/items/:id", func(c *gin.Context) { handler.UpdateCartItemHandler(mockDB)(c) })
				router.DELETE("/api/cart/items/:id", func(c *gin.Context) { handler.RemoveCartItemHandler(mockDB)(c) })
				router.DELETE("/api/cart", func(c *gin.Context) { handler.ClearCartHandler(mockDB)(c) })
			}

			for _, s := range tc.steps {
				var b []byte
				if s.body != nil {
					b, _ = json.Marshal(s.body)
				}
				req := httptest.NewRequest(s.method, s.path, bytes.NewBuffer(b))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				assert.Equal(t, s.want, w.Code)
			}

			mockDB.AssertExpectations(t)
		})
	}
}
