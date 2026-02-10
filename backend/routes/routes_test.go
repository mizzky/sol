package routes_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sol_coffeesys/backend/auth"
	"sol_coffeesys/backend/db"
	"sol_coffeesys/backend/handler"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockDB struct {
	db.Querier
	mock.Mock
}

func (m *MockDB) GetUserForUpdate(ctx context.Context, id int64) (db.User, error) {
	args := m.Called(ctx, id)
	u := db.User{}
	if v := args.Get(0); v != nil {
		u = v.(db.User)
	}
	return u, args.Error(1)
}

func (m *MockDB) CreateCategory(ctx context.Context, p db.CreateCategoryParams) (db.Category, error) {
	args := m.Called(ctx, p)
	c := db.Category{}
	if v := args.Get(0); v != nil {
		c = v.(db.Category)
	}
	return c, args.Error(1)
}

// routes_test: AdminOnly ミドルウェア挙動確認（未認証/非管理者/管理者)
func TestCategories_AdminOnlyMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupMock      func(m *MockDB)
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
			setupMock: func(m *MockDB) {
				m.On("GetUserForUpdate", mock.Anything, int64(1)).
					Return(db.User{ID: 1, Role: "user"}, nil)
			},
			setAuthHeader:  true,
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "管理者->ハンドラ実行201",
			setupMock: func(m *MockDB) {
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
			mockDB := new(MockDB)
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
