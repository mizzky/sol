package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sol_coffeesys/backend/db"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockDB struct {
	db.Querier // これで全メソッドを「持っている」ことになる
	mock.Mock
}

func (m *MockDB) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(db.User), args.Error(1)
}

func TestLoginHandler_UserNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := new(MockDB)
	mockDB.On("GetUserByEmail", mock.Anything, "notfound@example.com").Return(db.User{}, errors.New("user not found"))

	r := gin.Default()

	r.POST("/api/login", LoginHandler(mockDB))

	body := `{"email": "notfound@example.com", "password": "password"}`
	req, _ := http.NewRequest("POST", "/api/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
}
