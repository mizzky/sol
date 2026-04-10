package handler_test

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"sol_coffeesys/backend/db"
	"sol_coffeesys/backend/handler"
	"sol_coffeesys/backend/handler/testutil"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRefreshTokenHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.Default()
	mockDB := new(testutil.MockDB)
	mockTG := new(MockTokenGenerator)

	oldRefresh := strings.Repeat("a", 64)
	sum := sha256.Sum256([]byte(oldRefresh))
	oldHash := hex.EncodeToString(sum[:])

	mockDB.On("GetRefreshTokenByHash", mock.Anything, oldHash).Return(
		db.RefreshToken{
			ID:        1,
			UserID:    1,
			TokenHash: oldHash,
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}, nil)

	mockDB.On("GetUserByID", mock.Anything, int64(1)).Return(
		db.User{
			ID:    1,
			Email: "test@example.com",
			Name:  "Test",
			Role:  "member",
		}, nil)

	mockTG.On("GenerateToken", int64(1)).Return("new_access_token", nil)

	var captured db.CreateRefreshTokenParams
	mockDB.On("CreateRefreshToken", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			captured = args.Get(1).(db.CreateRefreshTokenParams)
		}).
		Return(db.RefreshToken{ID: 2, UserID: 1}, nil)

	mockDB.On("RevokeRefreshTokenByHash", mock.Anything, oldHash).Return(nil)

	router.POST("/api/refresh", handler.RefreshTokenHandler(mockDB, mockTG))

	req := httptest.NewRequest(http.MethodPost, "/api/refresh", nil)
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
	assert.True(t, accessCookie.HttpOnly)
	assert.Equal(t, "/", accessCookie.Path)
	assert.Equal(t, 15*60, accessCookie.MaxAge)
	assert.Equal(t, http.SameSiteLaxMode, accessCookie.SameSite)

	assert.True(t, refreshCookie.HttpOnly)
	assert.Equal(t, "/api/refresh", refreshCookie.Path)
	assert.Equal(t, 14*24*60*60, refreshCookie.MaxAge)
	assert.Equal(t, http.SameSiteStrictMode, refreshCookie.SameSite)
	assert.Equal(t, 64, len(refreshCookie.Value))

	sum2 := sha256.Sum256([]byte(refreshCookie.Value))
	gotHash := hex.EncodeToString(sum2[:])
	assert.Equal(t, gotHash, captured.TokenHash)
	assert.WithinDuration(t, time.Now().Add(14*24*time.Hour), captured.ExpiresAt, 5*time.Second)

	mockDB.AssertExpectations(t)
	mockTG.AssertExpectations(t)

}

func TestRefreshTokenHandler_Errros(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupMock      func(m *testutil.MockDB)
		setupTG        func(tg *MockTokenGenerator)
		cookie         string
		expectedStatus int
	}{
		{
			name:           "Cookie欠如",
			setupMock:      nil,
			cookie:         "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "token未登録",
			setupMock: func(m *testutil.MockDB) {
				oldRefresh := strings.Repeat("a", 64)
				sum := sha256.Sum256([]byte(oldRefresh))
				oldHash := hex.EncodeToString(sum[:])
				m.On("GetRefreshTokenByHash", mock.Anything, oldHash).Return(
					db.RefreshToken{}, sql.ErrNoRows)
			},
			cookie:         strings.Repeat("a", 64),
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "token期限切れ",
			expectedStatus: http.StatusUnauthorized,
			setupMock: func(m *testutil.MockDB) {
				oldRefresh := strings.Repeat("a", 64)
				sum := sha256.Sum256([]byte(oldRefresh))
				oldHash := hex.EncodeToString(sum[:])
				m.On("GetRefreshTokenByHash", mock.Anything, oldHash).Return(
					db.RefreshToken{
						ID:        1,
						UserID:    1,
						TokenHash: oldHash,
						ExpiresAt: time.Now().Add(-1 * time.Hour),
					}, nil)
			},
			cookie: strings.Repeat("a", 64),
		},
		{
			name:           "token失効済み",
			expectedStatus: http.StatusUnauthorized,
			setupMock: func(m *testutil.MockDB) {
				oldRefresh := strings.Repeat("a", 64)
				sum := sha256.Sum256([]byte(oldRefresh))
				oldHash := hex.EncodeToString(sum[:])
				m.On("GetRefreshTokenByHash", mock.Anything, oldHash).Return(db.RefreshToken{
					ID:        1,
					UserID:    1,
					TokenHash: oldHash,
					ExpiresAt: time.Now().Add(1 * time.Hour),
					RevokedAt: sql.NullTime{Valid: true, Time: time.Now().Add(-1 * time.Hour)},
				}, nil)
			},
			cookie: strings.Repeat("a", 64),
		},
		{
			name:           "DBエラー Get",
			expectedStatus: http.StatusInternalServerError,
			setupMock: func(m *testutil.MockDB) {
				oldRefresh := strings.Repeat("a", 64)
				sum := sha256.Sum256([]byte(oldRefresh))
				oldHash := hex.EncodeToString(sum[:])
				m.On("GetRefreshTokenByHash", mock.Anything, oldHash).Return(db.RefreshToken{}, errors.New("db error"))
			},
			cookie: strings.Repeat("a", 64),
		},
		{
			name: "DBエラー Create",
			setupMock: func(m *testutil.MockDB) {
				oldRefersh := strings.Repeat("a", 64)
				sum := sha256.Sum256([]byte(oldRefersh))
				oldHash := hex.EncodeToString(sum[:])
				m.On("GetRefreshTokenByHash", mock.Anything, oldHash).Return(
					db.RefreshToken{ID: 1, UserID: 1, TokenHash: oldHash, ExpiresAt: time.Now().Add(1 * time.Hour)}, nil)
				m.On("GetUserByID", mock.Anything, int64(1)).Return(
					db.User{ID: 1, Email: "test@eample.com", Name: "test", Role: "member"}, nil)
				m.On("CreateRefreshToken", mock.Anything, mock.Anything).Return(
					db.RefreshToken{}, errors.New("db create error"))
			},
			cookie:         strings.Repeat("a", 64),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "DBエラー Revoke",
			setupMock: func(m *testutil.MockDB) {
				oldRefersh := strings.Repeat("a", 64)
				sum := sha256.Sum256([]byte(oldRefersh))
				oldHash := hex.EncodeToString(sum[:])
				m.On("GetRefreshTokenByHash", mock.Anything, oldHash).Return(
					db.RefreshToken{ID: 1, UserID: 1, TokenHash: oldHash, ExpiresAt: time.Now().Add(1 * time.Hour)}, nil)
				m.On("GetUserByID", mock.Anything, int64(1)).Return(
					db.User{ID: 1, Email: "test@eample.com", Name: "test", Role: "member"}, nil)
				m.On("CreateRefreshToken", mock.Anything, mock.Anything).Return(
					db.RefreshToken{ID: 2, UserID: 1}, nil)
				m.On("RevokeRefreshTokenByHash", mock.Anything, oldHash).Return(
					errors.New("db revoke error"))
			},
			cookie:         strings.Repeat("a", 64),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "token生成失敗",
			expectedStatus: http.StatusInternalServerError,
			setupMock: func(m *testutil.MockDB) {
				oldRefresh := strings.Repeat("a", 64)
				sum := sha256.Sum256([]byte(oldRefresh))
				oldHash := hex.EncodeToString(sum[:])
				m.On("GetRefreshTokenByHash", mock.Anything, oldHash).Return(
					db.RefreshToken{
						ID:        1,
						UserID:    1,
						TokenHash: oldHash,
						ExpiresAt: time.Now().Add(1 * time.Hour),
					}, nil)
				m.On("GetUserByID", mock.Anything, int64(1)).Return(db.User{ID: 1, Email: "x", Name: "x", Role: "member"}, nil)
				m.On("CreateRefreshToken", mock.Anything, mock.Anything).Return(
					db.RefreshToken{ID: 2, UserID: 1}, nil)
				m.On("RevokeRefreshTokenByHash", mock.Anything, oldHash).Return(nil)
			},

			setupTG: func(tg *MockTokenGenerator) {
				tg.On("GenerateToken", int64(1)).Return("", errors.New("token gen failed"))
			},
			cookie: strings.Repeat("a", 64),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.Default()
			mockDB := new(testutil.MockDB)
			mockTG := new(MockTokenGenerator)

			if tt.setupMock != nil {
				tt.setupMock(mockDB)
			}
			if tt.setupTG != nil {
				tt.setupTG(mockTG)
			}

			router.POST("/api/refresh", handler.RefreshTokenHandler(mockDB, mockTG))

			req := httptest.NewRequest(http.MethodPost, "/api/refresh", nil)
			if tt.cookie != "" {
				req.AddCookie(&http.Cookie{
					Name:  "refresh_token",
					Value: tt.cookie,
					Path:  "/api/refresh",
				})
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockDB.AssertExpectations(t)
			mockTG.AssertExpectations(t)
		})
	}
}

func TestRevokeRefreshHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	mockDB := new(testutil.MockDB)

	oldRefresh := strings.Repeat("a", 64)
	sum := sha256.Sum256([]byte(oldRefresh))
	oldHash := hex.EncodeToString(sum[:])

	mockDB.On("RevokeRefreshTokenByHash", mock.Anything, oldHash).Return(nil)

	router.POST("/api/refresh/revoke", handler.RevokeRefreshHandler(mockDB))

	req := httptest.NewRequest(http.MethodPost, "/api/refresh/revoke", nil)
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
	assert.Equal(t, "/", accessCookie.Path)
	assert.Equal(t, "/api/refresh", refreshCookie.Path)

	mockDB.AssertExpectations(t)
}

func TestRevokeRefreshHandler_DBError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	mockDB := new(testutil.MockDB)

	oldRefresh := strings.Repeat("a", 64)
	sum := sha256.Sum256([]byte(oldRefresh))
	oldHash := hex.EncodeToString(sum[:])

	mockDB.On("RevokeRefreshTokenByHash", mock.Anything, oldHash).Return(errors.New("db error"))

	router.POST("/api/refresh/revoke", handler.RevokeRefreshHandler(mockDB))

	req := httptest.NewRequest(http.MethodPost, "/api/refresh/revoke", nil)
	req.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: oldRefresh,
		Path:  "/api/refresh",
	})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockDB.AssertExpectations(t)
}
