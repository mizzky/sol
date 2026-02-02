package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sol_coffeesys/backend/db"
	"sol_coffeesys/backend/handler"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockDB struct {
	db.Querier
	mock.Mock
}

func (m *MockDB) CreateCategory(ctx context.Context, arg db.CreateCategoryParams) (db.Category, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Category), args.Error(1)
}

func TestCreateCategory(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "正常系：カテゴリ作成成功",
			requestBody: map[string]interface{}{
				"name":        "コーヒー豆",
				"description": "各種コーヒー豆を取扱います",
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotNil(t, response["id"])
				assert.Equal(t, "コーヒー豆", response["name"])
				assert.Equal(t, "各種コーヒー豆を取扱います", response["description"])
			},
		},
		{
			name: "正常系：descriptionなしでカテゴリ作成",
			requestBody: map[string]interface{}{
				"name": "紅茶",
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "紅茶", response["name"])
			},
		},
		{
			name: "異常系：nameが空",
			requestBody: map[string]interface{}{
				"name":        "",
				"description": "説明",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "name")
			},
		},
		{
			name:           "異常系：nameフィールドなし",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.Default()
			mockDB := new(MockDB)
			mockDB.On("CreateCategory", mock.Anything, mock.AnythingOfType("db.CreateCategoryParams")).
				Return(db.Category{}, nil)
			router.POST("/api/categories", handler.CreateCategory(mockDB))

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/categories", bytes.NewBuffer(body))
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
