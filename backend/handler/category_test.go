package handler_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sol_coffeesys/backend/db"
	"sol_coffeesys/backend/handler"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type CategoryResponse struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

func TestCreateCategoryHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		setupMock      func(*MockDB)
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "正常系：カテゴリ作成成功",
			requestBody: map[string]interface{}{
				"name":        "コーヒー豆",
				"description": "各種コーヒー豆を取り扱います",
			},
			expectedStatus: http.StatusCreated,
			setupMock: func(m *MockDB) {
				m.On("CreateCategory", mock.Anything, db.CreateCategoryParams{
					Name:        "コーヒー豆",
					Description: sql.NullString{String: "各種コーヒー豆を取り扱います", Valid: true},
				}).Return(db.Category{
					ID:          1,
					Name:        "コーヒー豆",
					Description: sql.NullString{String: "各種コーヒー豆を取り扱います", Valid: true},
				}, nil)
			},
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotNil(t, response["id"])
				assert.Equal(t, "コーヒー豆", response["name"])
				assert.Equal(t, "各種コーヒー豆を取り扱います", response["description"])
			},
		},
		{
			name: "正常系：descriptionなしでカテゴリ作成",
			requestBody: map[string]interface{}{
				"name": "紅茶",
			},
			expectedStatus: http.StatusCreated,
			setupMock: func(m *MockDB) {
				m.On("CreateCategory", mock.Anything, db.CreateCategoryParams{
					Name:        "紅茶",
					Description: sql.NullString{String: "", Valid: false},
				}).Return(db.Category{
					ID:          1,
					Name:        "紅茶",
					Description: sql.NullString{String: "", Valid: false},
				}, nil)
			},
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
			setupMock:      nil,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "カテゴリ名は必須です")
			},
		},
		{
			name:           "異常系：nameフィールドなし",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
			setupMock:      nil,
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
			name: "DBエラー",
			requestBody: map[string]interface{}{
				"name": "コーヒー豆",
			},
			expectedStatus: http.StatusInternalServerError,
			setupMock: func(m *MockDB) {
				m.On("CreateCategory", mock.Anything, mock.MatchedBy(func(arg db.CreateCategoryParams) bool {
					return arg.Name == "コーヒー豆"
				})).Return(db.Category{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "予期せぬエラーが発生しました")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.Default()
			mockDB := new(MockDB)
			if tt.setupMock != nil {
				tt.setupMock(mockDB)
			}
			router.POST("/api/categories", handler.CreateCategoryHandler(mockDB))

			var body []byte
			if tt.name == "異常系：JSON形式エラー" {
				body = []byte(`{broken json`)
			} else {
				body, _ = json.Marshal(tt.requestBody)
			}

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

func TestUpdateCategory(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		categoryID     int64
		requestBody    map[string]interface{}
		setupMock      func(m *MockDB)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:       "正常系：カテゴリ更新成功",
			categoryID: 1,
			requestBody: map[string]interface{}{
				"name":        "プレミアムコーヒー豆",
				"description": "高級コーヒー豆の取り扱い",
			},
			expectedStatus: http.StatusOK,
			setupMock: func(m *MockDB) {
				m.On("UpdateCategory", mock.Anything, db.UpdateCategoryParams{
					ID:          1,
					Name:        "プレミアムコーヒー豆",
					Description: sql.NullString{String: "高級コーヒー豆の取り扱い", Valid: true},
				}).Return(db.Category{
					ID:          1,
					Name:        "プレミアムコーヒー豆",
					Description: sql.NullString{String: "高級コーヒー豆の取り扱い", Valid: true},
				}, nil)
			},
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var category CategoryResponse
				err := json.Unmarshal(w.Body.Bytes(), &category)
				assert.NoError(t, err)
				assert.Equal(t, "プレミアムコーヒー豆", category.Name)
				assert.NotNil(t, category.Description)
				assert.Equal(t, "高級コーヒー豆の取り扱い", *category.Description)
			},
		},
		{
			name:       "異常系：カテゴリが存在しない",
			categoryID: 999,
			requestBody: map[string]interface{}{
				"name":        "プレミアムコーヒー豆",
				"description": "高級コーヒー豆の取り扱い",
			},
			expectedStatus: http.StatusNotFound,
			setupMock: func(m *MockDB) {
				m.On("UpdateCategory", mock.Anything, mock.Anything).
					Return(db.Category{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "カテゴリが見つかりません")
			},
		},
		{
			name: "異常系：IDが数値以外",
			requestBody: map[string]interface{}{
				"name":        "",
				"description": "高級コーヒー豆の取り扱い",
			},
			expectedStatus: http.StatusBadRequest,
			setupMock:      nil,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "IDが正しくありません")
			},
		},
		{
			name:       "異常系：必須フィールド(name)が空",
			categoryID: 1,
			requestBody: map[string]interface{}{
				"name":        "",
				"description": "高級コーヒー豆の取り扱い",
			},
			expectedStatus: http.StatusBadRequest,
			setupMock:      nil,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "カテゴリ名は必須です")
			},
		},
		{
			name:       "異常系：必須フィールド(name)がnull",
			categoryID: 1,
			requestBody: map[string]interface{}{
				"name":        nil,
				"description": "高級コーヒー豆の取り扱い",
			},
			expectedStatus: http.StatusBadRequest,
			setupMock:      nil,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "カテゴリ名は必須です")
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
			name: "DBエラー",
			requestBody: map[string]interface{}{
				"name":        "コーヒー豆",
				"description": "高級コーヒー豆の取り扱い",
			},
			expectedStatus: http.StatusInternalServerError,
			setupMock: func(m *MockDB) {
				m.On("UpdateCategory", mock.Anything, mock.Anything).
					Return(db.Category{}, fmt.Errorf("DB接続エラー"))
			},
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "予期せぬエラーが発生しました")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.Default()
			mockDB := new(MockDB)

			if tt.setupMock != nil {
				tt.setupMock(mockDB)
			}

			router.PUT("/api/categories/:id", handler.UpdateCategoryHandler(mockDB))

			var body []byte
			if tt.name == "異常系：JSON形式エラー" {
				body = []byte(`{broken json`)
			} else {
				body, _ = json.Marshal(tt.requestBody)
			}

			var req *http.Request
			if tt.name == "異常系：IDが数値以外" {
				req = httptest.NewRequest(http.MethodPut, "/api/categories/abc", bytes.NewBuffer(body))
			} else {
				req = httptest.NewRequest(http.MethodPut, "/api/categories/"+fmt.Sprint(tt.categoryID), bytes.NewBuffer(body))
			}
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

func TestGetCategoriesHandler(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedBody   string
		setupMock      func(*MockDB)
	}{
		{
			name:        "正常系：デフォルトのページとリミットで取得",
			queryParams: "",
			setupMock: func(m *MockDB) {
				m.On("GetCategories", mock.Anything, mock.Anything).
					Return([]db.Category{
						{ID: 1, Name: "コーヒー豆", Description: sql.NullString{String: "各種コーヒー豆を取り扱います", Valid: true}},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: `{
			"categories":[
				{
				"id":1,
				"name":"コーヒー豆",
				"description":"各種コーヒー豆を取り扱います"
				}
			],
			"total": 1,
			"page": 1,
			"limit":10
			}`,
		},
		{
			name:        "正常系：ページとリミットを指定して取得",
			queryParams: "?page=2&limit=5",
			setupMock: func(m *MockDB) {
				m.On("GetCategories", mock.Anything, mock.Anything).
					Return([]db.Category{
						{ID: 6, Name: "紅茶", Description: sql.NullString{String: "各種紅茶を取り扱います", Valid: true}},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: `{
			"categories":[
				{
				"id":6,
				"name":"紅茶",
				"description":"各種紅茶を取り扱います"
				}
			],
			"total": 1,
			"page": 2,
			"limit":5
			}`,
		},
		{
			name:           "異常系：不正なページ番号",
			queryParams:    "?page=-1",
			setupMock:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error": "リクエストパラメータが不正です}`,
		},
		{
			name:        "異常系：DB接続エラー",
			queryParams: "",
			setupMock: func(m *MockDB) {
				m.On("GetCategories", mock.Anything, mock.Anything).Return(nil, errors.New("DB接続エラー"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error": :"予期せぬエラーが発生しました"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.Default()
			mockDB := new(MockDB)
			if tt.setupMock != nil {
				tt.setupMock(mockDB)
			}
			req := httptest.NewRequest(http.MethodGet, "/api/categories"+tt.queryParams, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
		})
	}
}
