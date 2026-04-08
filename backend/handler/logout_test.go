package handler_test

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"sol_coffeesys/backend/handler"
	"sol_coffeesys/backend/handler/testutil"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLogoutHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	mockDB := new(testutil.MockDB)

	oldRefresh := strings.Repeat("a", 64)
	sum := sha256.Sum256([]byte(oldRefresh))
	oldHash := hex.EncodeToString(sum[:])

	mockDB.On("RevokeRefreshTokenByHash", mock.Anything, oldHash).Return(nil)

	router.POST("/api/logout", handler.LogoutHandler(mockDB))

	req := httptest.NewRequest(http.MethodPost, "/api/logout", nil)
	req.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: oldRefresh,
		Path:  "/api/refresh",
	})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	resp := w.Result()
	cookies := resp.Cookies()
	var accessCookie, refreshCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "access_token" {
			accessCookie = c
		}
		if c.Name == "refresh_token" {
			refreshCookie = c
		}
	}

	assert.NotNil(t, accessCookie)
	assert.NotNil(t, refreshCookie)

	assert.Equal(t, "", accessCookie.Value)
	assert.Equal(t, "", refreshCookie.Value)
	assert.Equal(t, -1, accessCookie.MaxAge)
	assert.Equal(t, -1, refreshCookie.MaxAge)
	assert.True(t, accessCookie.HttpOnly)
	assert.True(t, refreshCookie.HttpOnly)
	assert.Equal(t, "/", accessCookie.Path)
	assert.Equal(t, "/api/refresh", refreshCookie.Path)
	assert.Equal(t, http.SameSiteLaxMode, accessCookie.SameSite)
	assert.Equal(t, http.SameSiteStrictMode, refreshCookie.SameSite)

	mockDB.AssertExpectations(t)
}

func TestLogoutHandler_Errors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	oldRefresh := strings.Repeat("a", 64)
	sum := sha256.Sum256([]byte(oldRefresh))
	oldHash := hex.EncodeToString(sum[:])

	tests := []struct {
		name           string
		expectedStatus int
		setupMock      func(*testutil.MockDB)
		cookie         *http.Cookie
	}{
		{
			name:           "Cookie欠如",
			expectedStatus: http.StatusOK,
			setupMock:      nil,
			cookie:         nil,
		},
		{
			name:           "DBエラー RevokeRefreshTokenByHash",
			expectedStatus: http.StatusInternalServerError,
			setupMock: func(m *testutil.MockDB) {
				m.On("RevokeRefreshTokenByHash", mock.Anything, oldHash).Return(errors.New("db error"))
			},
			cookie: &http.Cookie{
				Name:  "refresh_token",
				Value: oldRefresh,
				Path:  "/api/refresh",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.Default()
			mockDB := new(testutil.MockDB)

			if tt.setupMock != nil {
				tt.setupMock(mockDB)
			}
			router.POST("/api/logout", handler.LogoutHandler(mockDB))
			req := httptest.NewRequest(http.MethodPost, "/api/logout", nil)
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {

				resp := w.Result()
				cookies := resp.Cookies()
				var accessCookie, refreshCookie *http.Cookie
				for _, c := range cookies {
					if c.Name == "access_token" {
						accessCookie = c
					}
					if c.Name == "refresh_token" {
						refreshCookie = c
					}
				}

				assert.NotNil(t, accessCookie)
				assert.NotNil(t, refreshCookie)

				assert.Equal(t, "", accessCookie.Value)
				assert.Equal(t, "", refreshCookie.Value)
				assert.Equal(t, -1, accessCookie.MaxAge)
				assert.Equal(t, -1, refreshCookie.MaxAge)
				assert.True(t, accessCookie.HttpOnly)
				assert.True(t, refreshCookie.HttpOnly)
				assert.Equal(t, "/", accessCookie.Path)
				assert.Equal(t, "/api/refresh", refreshCookie.Path)
				assert.Equal(t, http.SameSiteLaxMode, accessCookie.SameSite)
				assert.Equal(t, http.SameSiteStrictMode, refreshCookie.SameSite)

			} else {
				resp := w.Result()
				for _, c := range resp.Cookies() {
					assert.NotEqual(t, "access_token", c.Name)
					assert.NotEqual(t, "refresh_token", c.Name)
				}
			}
			mockDB.AssertExpectations(t)
		})
	}
}
