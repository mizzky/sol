package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sol_coffeesys/backend/db"
	"sol_coffeesys/backend/handler"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTokenGenerator struct {
	mock.Mock
}

func (m *MockTokenGenerator) GenerateToken(userID int64) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

func TestLoginHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		setupMock      func(m *MockDB)
		setupTokenMock func(tg *MockTokenGenerator)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "正常系：ログイン成功",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"password": "password123",
			},
			expectedStatus: http.StatusOK,
			setupMock: func(m *MockDB) {
				passwordHash, err := handler.HashPassword("password123")
				if err != nil {
					t.Fatalf("パスワードのハッシュ化に失敗しました: %v", err)
				}
				m.On("GetUserByEmail", mock.Anything, "test@example.com").
					Return(db.User{
						ID:           1,
						Email:        "test@example.com",
						PasswordHash: passwordHash,
					}, nil)
			},
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				user := response["user"].(map[string]interface{})
				assert.Equal(t, "test@example.com", user["email"])
				assert.NotNil(t, response["token"])
			},
		},
		{
			name: "異常系：ユーザーが存在しない",
			requestBody: map[string]interface{}{
				"email":    "notfound@example.com",
				"password": "password",
			},
			expectedStatus: http.StatusUnauthorized,
			setupMock: func(m *MockDB) {
				m.On("GetUserByEmail", mock.Anything, "notfound@example.com").
					Return(db.User{}, errors.New("user not found"))
			},
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "メールアドレスまたはパスワードが正しくありません")
			},
		},
		{
			name: "異常系：パスワード相違",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"password": "wrongpassword",
			},
			expectedStatus: http.StatusUnauthorized,
			setupMock: func(m *MockDB) {
				passwordHash, err := handler.HashPassword("password123")
				if err != nil {
					t.Fatalf("パスワードのハッシュ化に失敗しました: %v", err)
				}
				m.On("GetUserByEmail", mock.Anything, "test@example.com").
					Return(db.User{
						ID:           1,
						Email:        "test@example.com",
						PasswordHash: passwordHash,
					}, nil)
			},
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "メールアドレスまたはパスワードが正しくありません")
			},
		},
		{
			name:           "異常系：JSON形式エラー",
			expectedStatus: http.StatusBadRequest,
			setupMock:      nil,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "リクエスト形式が正しくありません")
			},
		},
		{
			name: "異常系：トークン生成エラー",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"password": "password123",
			},
			expectedStatus: http.StatusInternalServerError,
			setupMock: func(m *MockDB) {
				passwordHash, err := handler.HashPassword("password123")
				if err != nil {
					t.Fatalf("パスワードのハッシュ化に失敗しました: %v", err)
				}
				m.On("GetUserByEmail", mock.Anything, "test@example.com").
					Return(db.User{
						ID:           1,
						Email:        "test@example.com",
						PasswordHash: passwordHash,
					}, nil)
			},
			setupTokenMock: func(tg *MockTokenGenerator) {
				tg.On("GenerateToken", int64(1)).
					Return("", errors.New("トークンの生成に失敗しました")) // モック設定を追加
			},
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "トークンの生成に失敗しました")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.Default()
			mockDB := new(MockDB)
			mockTokenGenerator := new(MockTokenGenerator)

			if tt.setupMock != nil {
				tt.setupMock(mockDB)
			}

			if tt.setupTokenMock != nil {
				tt.setupTokenMock(mockTokenGenerator)
			} else {
				// デフォルトの動作を設定
				mockTokenGenerator.On("GenerateToken", mock.Anything).
					Return("default_token", nil)
			}

			router.POST("/api/login", handler.LoginHandler(mockDB, mockTokenGenerator))

			var body []byte
			if tt.name == "異常系：JSON形式エラー" {
				body = []byte(`{broken json`)
			} else {
				body, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}
