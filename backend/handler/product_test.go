package handler_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sol_coffeesys/backend/db"
	"sol_coffeesys/backend/handler"
	testutil "sol_coffeesys/backend/handler/testutil"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateProductHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	mockDB := new(testutil.MockDB)

	now := time.Now()
	mockDB.On("GetCategory", mock.Anything, int64(1)).Return(db.Category{
		ID:          1,
		Name:        "テストカテゴリ",
		Description: sql.NullString{String: "", Valid: false},
	}, nil)

	mockDB.On("CreateProduct", mock.Anything, db.CreateProductParams{
		Name:          "Coffee",
		Price:         500,
		IsAvailable:   true,
		CategoryID:    1,
		Sku:           "COF-001",
		Description:   sql.NullString{String: "Nice coffee", Valid: true},
		ImageUrl:      sql.NullString{String: "https://example.com/img.png", Valid: true},
		StockQuantity: 10,
	}).Return(db.Product{
		ID:            1,
		Name:          "Coffee",
		Price:         500,
		IsAvailable:   true,
		Sku:           "COF-001",
		Description:   sql.NullString{String: "Nice coffee", Valid: true},
		ImageUrl:      sql.NullString{String: "https://example.com/img.png", Valid: true},
		StockQuantity: 10,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil)

	router.POST("/api/products", handler.CreateProductHandler(mockDB))

	body := map[string]interface{}{
		"name":           "Coffee",
		"price":          500,
		"is_available":   true,
		"category_id":    1,
		"sku":            "COF-001",
		"description":    "Nice coffee",
		"image_url":      "https://example.com/img.png",
		"stock_quantity": 10,
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/products", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockDB.AssertExpectations(t)
}

func TestProductCRUD_HappyPath(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := new(testutil.MockDB)
	mockDB.On("GetCategory", mock.Anything, int64(1)).Return(db.Category{
		ID:          1,
		Name:        "テストカテゴリ",
		Description: sql.NullString{String: "", Valid: false},
	}, nil)

	now := time.Now()
	sample := db.Product{
		ID:            1,
		Name:          "Coffee",
		Price:         500,
		IsAvailable:   true,
		CategoryID:    1,
		Sku:           "COF-001",
		Description:   sql.NullString{String: "Nice coffee", Valid: true},
		ImageUrl:      sql.NullString{String: "https://example.com/img.png", Valid: true},
		StockQuantity: 10,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	mockDB.On("CreateProduct", mock.Anything, mock.Anything).Return(sample, nil)
	mockDB.On("GetProduct", mock.Anything, int64(1)).Return(sample, nil)
	mockDB.On("ListProducts", mock.Anything).Return([]db.Product{sample}, nil)
	mockDB.On("UpdateProduct", mock.Anything, mock.Anything).Return(sample, nil)
	mockDB.On("DeleteProduct", mock.Anything, int64(1)).Return(nil)

	// Create
	{
		router := gin.Default()
		router.POST("/api/products", handler.CreateProductHandler(mockDB))
		body := map[string]interface{}{
			"name":           "Coffee",
			"price":          500,
			"is_available":   true,
			"category_id":    1,
			"sku":            "COF-001",
			"description":    "Nice coffee",
			"image_url":      "https://example.com/img.png",
			"stock_quantity": 10,
		}
		b, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/products", bytes.NewBuffer(b))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
	}

	// Get
	{
		router := gin.Default()
		router.GET("/api/products/:id", handler.GetProductHandler(mockDB))
		req := httptest.NewRequest(http.MethodGet, "/api/products/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// List
	{
		router := gin.Default()
		router.GET("/api/products", handler.ListProductsHandler(mockDB))
		req := httptest.NewRequest(http.MethodGet, "/api/products", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// Update(PUT)
	{
		router := gin.Default()
		router.PUT("/api/products/:id", handler.UpdateProductHandler(mockDB))
		updateBody := map[string]interface{}{
			"name":           "Coffee Updated",
			"price":          600,
			"is_available":   true,
			"category_id":    1,
			"sku":            "COF-001",
			"description":    "Updated",
			"image_url":      "https://example.com/img2.png",
			"stock_quantity": 20,
		}
		b, _ := json.Marshal(updateBody)
		req := httptest.NewRequest(http.MethodPut, "/api/products/1", bytes.NewBuffer(b))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// Delete
	{
		router := gin.Default()
		router.DELETE("/api/products/:id", handler.DeleteProductHandler(mockDB))
		req := httptest.NewRequest(http.MethodDelete, "/api/products/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	}
	mockDB.AssertExpectations(t)
}

func TestCreateProduct_ValidationTable(t *testing.T) {
	tests := []struct {
		name           string
		body           map[string]interface{}
		expectedStatus int
	}{
		{"missing name", map[string]interface{}{"price": 100, "sku": "S1"}, http.StatusBadRequest},
		{"price zero", map[string]interface{}{"name": "X", "price": 0, "sku": "S1"}, http.StatusBadRequest},
		{"missing sku", map[string]interface{}{"name": "X", "price": 100}, http.StatusBadRequest},
		{"invalid json", map[string]interface{}{}, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.Default()
			mockDB := new(testutil.MockDB)
			router.POST("/api/products", handler.CreateProductHandler(mockDB))
			var b []byte
			if tt.name == "invalid json" {
				b = []byte(`{broken json`)
			} else {
				b, _ = json.Marshal(tt.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/products", bytes.NewBuffer(b))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestCreateProduct_CategoryNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	mockDB := new(testutil.MockDB)

	mockDB.On("GetCategory", mock.Anything, int64(1)).Return(db.Category{}, sql.ErrNoRows)

	router.POST("/api/products", handler.CreateProductHandler(mockDB))

	body := map[string]interface{}{
		"name":           "Coffee",
		"price":          500,
		"is_available":   true,
		"category_id":    1,
		"sku":            "COF-001",
		"description":    "Nice coffee",
		"image_url":      "https://example.com/img.png",
		"stock_quantity": 10,
	}

	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/products", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Contains(t, resp["error"], "カテゴリが見つかりません")
	mockDB.AssertExpectations(t)
}

func TestUpdateProduct_CategoryNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	mockDB := new(testutil.MockDB)

	mockDB.On("GetCategory", mock.Anything, int64(1)).Return(db.Category{}, sql.ErrNoRows)
	router.PUT("/api/products/:id", handler.UpdateProductHandler(mockDB))

	updateBody := map[string]interface{}{
		"name":           "Coffee",
		"price":          500,
		"is_available":   true,
		"category_id":    1,
		"sku":            "COF-001",
		"description":    "Nice coffee",
		"image_url":      "https://example.com/img.png",
		"stock_quantity": 10,
	}
	b, _ := json.Marshal(updateBody)
	req := httptest.NewRequest(http.MethodPut, "/api/products/1", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Contains(t, resp["error"], "カテゴリが見つかりません")
	mockDB.AssertExpectations(t)
}

func TestUpdateProduct_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	mockDB := new(testutil.MockDB)

	mockDB.On("GetCategory", mock.Anything, int64(1)).Return(db.Category{
		ID:          1,
		Name:        "テストカテゴリ",
		Description: sql.NullString{String: "", Valid: false},
	}, nil)

	mockDB.On("UpdateProduct", mock.Anything, mock.Anything).Return(db.Product{}, sql.ErrNoRows)
	router.PUT("/api/products/:id", handler.UpdateProductHandler(mockDB))

	updateBody := map[string]interface{}{
		"name":           "Coffee",
		"price":          500,
		"is_available":   true,
		"category_id":    1,
		"sku":            "COF-001",
		"description":    "Nice coffee",
		"image_url":      "https://example.com/img.png",
		"stock_quantity": 10,
	}
	b, _ := json.Marshal(updateBody)
	req := httptest.NewRequest(http.MethodPut, "/api/products/999", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Contains(t, resp["error"], "商品が見つかりません")
	mockDB.AssertExpectations(t)
}

func TestGetProduct_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	mockDB := new(testutil.MockDB)

	mockDB.On("GetProduct", mock.Anything, int64(999)).Return(db.Product{}, sql.ErrNoRows)

	router.GET("/api/products/:id", handler.GetProductHandler(mockDB))
	req := httptest.NewRequest(http.MethodGet, "/api/products/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Contains(t, resp["error"], "商品が見つかりません")
	mockDB.AssertExpectations(t)
}

func TestDeleteProduct_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	mockDB := new(testutil.MockDB)

	mockDB.On("DeleteProduct", mock.Anything, int64(999)).Return(sql.ErrNoRows)

	router.DELETE("/api/products/:id", handler.DeleteProductHandler(mockDB))
	req := httptest.NewRequest(http.MethodDelete, "/api/products/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Contains(t, resp["error"], "商品が見つかりません")
	mockDB.AssertExpectations(t)

}

func TestUpdateProducts_ValidationTable(t *testing.T) {
	tests := []struct {
		name           string
		body           map[string]interface{}
		expectedStatus int
	}{
		{"missing name", map[string]interface{}{"price": 100, "sku": "S1"}, http.StatusBadRequest},
		{"price zero", map[string]interface{}{"name": "X", "price": 0, "sku": "S1"}, http.StatusBadRequest},
		{"missing sku", map[string]interface{}{"name": "X", "price": 100}, http.StatusBadRequest},
		{"invalid json", map[string]interface{}{}, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.Default()
			mockDB := new(testutil.MockDB)
			router.PUT("/api/products/:id", handler.UpdateProductHandler(mockDB))
			var b []byte
			if tt.name == "invalid json" {
				b = []byte(`{broken json`)
			} else {
				b, _ = json.Marshal(tt.body)
			}

			req := httptest.NewRequest(http.MethodPut, "/api/products/1", bytes.NewBuffer(b))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestGetProduct_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	mockDB := new(testutil.MockDB)
	router.GET("/api/products/:id", handler.GetProductHandler(mockDB))

	req := httptest.NewRequest(http.MethodGet, "/api/products/abc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Contains(t, resp["error"], "IDが正しくありません")
	mockDB.AssertNotCalled(t, "GetProduct", mock.Anything, mock.Anything)
}

func TestUpdateProduct_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	mockDB := new(testutil.MockDB)

	router.PUT("/api/products/:id", handler.UpdateProductHandler(mockDB))

	body := map[string]interface{}{
		"name":           "X",
		"price":          100,
		"is_available":   true,
		"category_id":    1,
		"sku":            "S1",
		"stock_quantity": 1,
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPut, "/api/products/invalid-id", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Contains(t, resp["error"], "IDが正しくありません")
	mockDB.AssertNotCalled(t, "GetCategory", mock.Anything, mock.Anything)
	mockDB.AssertNotCalled(t, "UpdateProduct", mock.Anything, mock.Anything)

}

func TestDeleteProduct_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	mockDB := new(testutil.MockDB)

	router.DELETE("/api/products/:id", handler.DeleteProductHandler(mockDB))

	req := httptest.NewRequest(http.MethodDelete, "/api/products/abcd", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Contains(t, resp["error"], "IDが正しくありません")

	mockDB.AssertNotCalled(t, "DeleteProduct", mock.Anything, mock.Anything)
}
