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
			"image-url":      "https://example.com/img2.png",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.Default()
			mockDB := new(testutil.MockDB)
			router.POST("/api/products", handler.CreateProductHandler(mockDB))

			b, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/products", bytes.NewBuffer(b))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
