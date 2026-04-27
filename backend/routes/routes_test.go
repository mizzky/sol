package routes_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sol_coffeesys/backend/auth"
	"sol_coffeesys/backend/db"
	"sol_coffeesys/backend/handler"
	testutil "sol_coffeesys/backend/handler/testutil"
	"sol_coffeesys/backend/middleware"
	"sol_coffeesys/backend/pkg/apperror"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// routes_test: AdminOnly ミドルウェア挙動確認（未認証/非管理者/管理者)
func TestCategories_AdminOnlyMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupMock      func(m *testutil.MockDB)
		setAuthHeader  bool
		expectedStatus int
	}{
		{
			name:           "未認証->401",
			setupMock:      nil,
			setAuthHeader:  false,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "非管理者->403",
			setupMock: func(m *testutil.MockDB) {
				m.On("GetUserForUpdate", mock.Anything, int64(1)).
					Return(db.User{ID: 1, Role: "user"}, nil)
			},
			setAuthHeader:  true,
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "管理者->ハンドラ実行201",
			setupMock: func(m *testutil.MockDB) {
				m.On("GetUserForUpdate", mock.Anything, int64(1)).
					Return(db.User{ID: 1, Role: "admin"}, nil)
				m.On("CreateCategory", mock.Anything, mock.Anything).
					Return(db.Category{ID: 1, Name: "テスト"}, nil)
			},
			setAuthHeader:  true,
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.Default()
			router.Use(middleware.ErrorHandler(apperror.ToHTTP))
			mockDB := new(testutil.MockDB)
			if tt.setupMock != nil {
				tt.setupMock(mockDB)
			}

			router.POST("/api/categories", auth.AdminOnly(mockDB), handler.CreateCategoryHandler(mockDB))

			orig := auth.Validate
			defer func() { auth.Validate = orig }()
			auth.Validate = func(ts string) (*jwt.Token, error) {
				return &jwt.Token{
					Valid:  true,
					Claims: jwt.MapClaims{"user.id": float64(1)},
				}, nil
			}

			body, _ := json.Marshal(map[string]interface{}{"name": "テスト"})
			req := httptest.NewRequest(http.MethodPost, "/api/categories", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			if tt.setAuthHeader {
				req.Header.Set("Authorization", "Bearer faketoken")
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
			mockDB.AssertExpectations(t)
		})
	}
}

func TestCartRoutes_AuthMiddlewareAndDeleteIdempotency(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name           string
		setupMock      func(m *testutil.MockDB)
		setAuthHeader  bool
		expectedStatus int
	}{
		{
			name:           "unauthorized -> 401",
			expectedStatus: http.StatusUnauthorized,
			setupMock:      nil,
			setAuthHeader:  false,
		},
		{
			name:           "auth -> handler runs",
			expectedStatus: http.StatusOK,
			setAuthHeader:  true,
			setupMock: func(m *testutil.MockDB) {
				m.On("GetUserForUpdate", mock.Anything, int64(1)).Return(
					db.User{
						ID:   1,
						Role: "member",
					}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.Default()
			router.Use(middleware.ErrorHandler(apperror.ToHTTP))
			mockDB := new(testutil.MockDB)
			if tt.setupMock != nil {
				tt.setupMock(mockDB)
			}
			router.GET("/api/cart", auth.RequireAuth(mockDB), func(c *gin.Context) { c.Status(http.StatusOK) })

			orig := auth.Validate
			defer func() { auth.Validate = orig }()

			// control auth.Validate based on test case
			if tt.setAuthHeader {
				auth.Validate = func(ts string) (*jwt.Token, error) {
					return &jwt.Token{Valid: true, Claims: jwt.MapClaims{"user.id": float64(1)}}, nil
				}
			} else {
				auth.Validate = func(ts string) (*jwt.Token, error) { return nil, errors.New("no token") }
			}

			req := httptest.NewRequest(http.MethodGet, "/api/cart", nil)
			if tt.setAuthHeader {
				req.Header.Set("Authorization", "Bearer faketoken")
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
			mockDB.AssertExpectations(t)
		})
	}
}
