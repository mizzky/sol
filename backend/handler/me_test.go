package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sol_coffeesys/backend/db"
	"sol_coffeesys/backend/handler/testutil"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMeHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         any
		setupMock      func(*testutil.MockDB)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "正常系：userIDからプロフィールを取得して返す",
			userID: int64(1),
			setupMock: func(m *testutil.MockDB) {
				m.On("GetUserForUpdate", mock.Anything, int64(1)).
					Return(db.User{
						ID:    1,
						Name:  "Taro",
						Email: "taro@example.com",
						Role:  "member",
					}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]any
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				user, ok := response["user"].(map[string]any)
				assert.True(t, ok)
				assert.Equal(t, float64(1), user["id"])
				assert.Equal(t, "Taro", user["name"])
				assert.Equal(t, "taro@example.com", user["email"])
				assert.Equal(t, "member", user["role"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			mockDB := new(testutil.MockDB)
			if tt.setupMock != nil {
				tt.setupMock(mockDB)
			}
			router.GET("/api/me", func(c *gin.Context) {
				c.Set("userID", tt.userID)
				MeHandler(mockDB)(c)
			})

			req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
			mockDB.AssertExpectations(t)
		})
	}
}
