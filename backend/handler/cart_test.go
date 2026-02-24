package handler_test

import (
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
