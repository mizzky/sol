package handler_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sol_coffeesys/backend/auth"
	"sol_coffeesys/backend/db"
	"sol_coffeesys/backend/handler"
	testutil "sol_coffeesys/backend/handler/testutil"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMeHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	origValidate := auth.Validate
	t.Cleanup(func() {
		auth.Validate = origValidate
	})

	tests := []struct {
		name         string
		authHeader   string
		validateFunc func(string) (*jwt.Token, error)
		setupMock    func(m *testutil.MockDB)
		expectStatus int
		checkBody    func(t *testing.T, body []byte)
	}{
		{
			name:       "正常系：有効トークン",
			authHeader: "Bearer valid",
			validateFunc: func(ts string) (*jwt.Token, error) {
				return &jwt.Token{Valid: true, Claims: jwt.MapClaims{"user.id": float64(1)}}, nil
			},
			setupMock: func(m *testutil.MockDB) {
				m.On("GetUserForUpdate", mock.Anything, int64(1)).Return(db.User{ID: 1, Name: "Taro", Email: "taro@example.com"}, nil)
			},
			expectStatus: http.StatusOK,
			checkBody: func(t *testing.T, body []byte) {
				var resp map[string]interface{}
				assert.NoError(t, json.Unmarshal(body, &resp))
				user, ok := resp["user"].(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, "taro@example.com", user["email"])
			},
		},
		{
			name:       "クレームが数字文字列->200",
			authHeader: "Bearer valid",
			validateFunc: func(ts string) (*jwt.Token, error) {
				return &jwt.Token{Valid: true, Claims: jwt.MapClaims{"user.id": "1"}}, nil
			},
			setupMock: func(m *testutil.MockDB) {
				m.On("GetUserForUpdate", mock.Anything, int64(1)).Return(db.User{ID: 1, Name: "Taro", Email: "taro@example.com"}, nil)
			},
			expectStatus: http.StatusOK,
			checkBody: func(t *testing.T, body []byte) {
				var resp map[string]interface{}
				assert.NoError(t, json.Unmarshal(body, &resp))
				user, ok := resp["user"].(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, "taro@example.com", user["email"])
			},
		},
		{
			name:         "ヘッダ無し->401",
			authHeader:   "",
			validateFunc: nil,
			setupMock:    nil,
			expectStatus: http.StatusUnauthorized,
			checkBody: func(t *testing.T, body []byte) {
				var resp map[string]interface{}
				assert.NoError(t, json.Unmarshal(body, &resp))
				_, ok := resp["error"]
				assert.True(t, ok)
			},
		},
		{
			name:         "ヘッダフォーマット不正->401",
			authHeader:   "invalidHeader",
			validateFunc: nil,
			setupMock:    nil,
			expectStatus: http.StatusUnauthorized,
			checkBody: func(t *testing.T, body []byte) {
				var resp map[string]interface{}
				assert.NoError(t, json.Unmarshal(body, &resp))
				_, ok := resp["error"]
				assert.True(t, ok)
			},
		},
		{
			name:       "無効なトークン（Validateエラー）->401",
			authHeader: "Bearer invalid",
			validateFunc: func(ts string) (*jwt.Token, error) {
				return nil, fmt.Errorf("invalid")
			},
			setupMock:    nil,
			expectStatus: http.StatusUnauthorized,
			checkBody: func(t *testing.T, body []byte) {
				var resp map[string]interface{}
				assert.NoError(t, json.Unmarshal(body, &resp))
				_, ok := resp["error"]
				assert.True(t, ok)
			},
		}, {
			name:       "トークン欠落->401",
			authHeader: "Bearer",
			validateFunc: func(ts string) (*jwt.Token, error) {
				return nil, fmt.Errorf("invalid")
			},
			setupMock:    nil,
			expectStatus: http.StatusUnauthorized,
			checkBody: func(t *testing.T, body []byte) {
				var resp map[string]interface{}
				assert.NoError(t, json.Unmarshal(body, &resp))
				_, ok := resp["error"]
				assert.True(t, ok)
			},
		},
		{
			name:         "クレームなし->401",
			authHeader:   "Bearer noclaim",
			validateFunc: func(ts string) (*jwt.Token, error) { return &jwt.Token{Valid: true, Claims: jwt.MapClaims{}}, nil },
			setupMock:    nil,
			expectStatus: http.StatusUnauthorized,
			checkBody: func(t *testing.T, body []byte) {
				var resp map[string]interface{}
				assert.NoError(t, json.Unmarshal(body, &resp))
				_, ok := resp["error"]
				assert.True(t, ok)
			},
		},
		{
			name:       "クレームの型が不正（bool）->401",
			authHeader: "Bearer noclaim",
			validateFunc: func(ts string) (*jwt.Token, error) {
				return &jwt.Token{Valid: true, Claims: jwt.MapClaims{"user.id": true}}, nil
			},
			setupMock:    nil,
			expectStatus: http.StatusUnauthorized,
			checkBody: func(t *testing.T, body []byte) {
				var resp map[string]interface{}
				assert.NoError(t, json.Unmarshal(body, &resp))
				_, ok := resp["error"]
				assert.True(t, ok)
			},
		},
		{
			name:       "token.Validがfalse->401",
			authHeader: "Bearer noclaim",
			validateFunc: func(ts string) (*jwt.Token, error) {
				return &jwt.Token{Valid: false, Claims: jwt.MapClaims{"user.id": float64(1)}}, nil
			},
			setupMock:    nil,
			expectStatus: http.StatusUnauthorized,
			checkBody: func(t *testing.T, body []byte) {
				var resp map[string]interface{}
				assert.NoError(t, json.Unmarshal(body, &resp))
				_, ok := resp["error"]
				assert.True(t, ok)
			},
		},
		{
			name:       "ユーザー未検出->401",
			authHeader: "Bearer missing",
			validateFunc: func(ts string) (*jwt.Token, error) {
				return &jwt.Token{Valid: true, Claims: jwt.MapClaims{}}, nil
			},
			setupMock:    nil,
			expectStatus: http.StatusUnauthorized,
			checkBody: func(t *testing.T, body []byte) {
				var resp map[string]interface{}
				assert.NoError(t, json.Unmarshal(body, &resp))
				_, ok := resp["error"]
				assert.True(t, ok)
			},
		},
		{
			name:       "トークン型アサーション失敗->401",
			authHeader: "Bearer missing",
			validateFunc: func(ts string) (*jwt.Token, error) {
				return &jwt.Token{Valid: true, Claims: jwt.RegisteredClaims{}}, nil
			},
			setupMock:    nil,
			expectStatus: http.StatusUnauthorized,
			checkBody: func(t *testing.T, body []byte) {
				var resp map[string]interface{}
				assert.NoError(t, json.Unmarshal(body, &resp))
				_, ok := resp["error"]
				assert.True(t, ok)
			},
		},
		{
			name:       "DBエラー->500",
			authHeader: "Bearer dberr",
			validateFunc: func(ts string) (*jwt.Token, error) {
				return &jwt.Token{Valid: true, Claims: jwt.MapClaims{"user.id": float64(2)}}, nil
			},
			setupMock: func(m *testutil.MockDB) {
				m.On("GetUserForUpdate", mock.Anything, int64(2)).Return(db.User{}, fmt.Errorf("db error"))
			},
			expectStatus: http.StatusInternalServerError,
			checkBody: func(t *testing.T, body []byte) {
				var resp map[string]interface{}
				assert.NoError(t, json.Unmarshal(body, &resp))
				_, ok := resp["error"]
				assert.True(t, ok)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.validateFunc != nil {
				auth.Validate = tt.validateFunc
			} else {
				auth.Validate = origValidate
			}
			router := gin.New()
			mockDB := new(testutil.MockDB)
			if tt.setupMock != nil {
				tt.setupMock(mockDB)
			}

			router.GET("/api/me", handler.MeHandler(mockDB))

			req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectStatus, w.Code)
			if tt.checkBody != nil {
				tt.checkBody(t, w.Body.Bytes())
			}

			mockDB.AssertExpectations(t)
		})
	}
}
