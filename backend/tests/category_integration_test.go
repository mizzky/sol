package tests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sol_coffeesys/backend/auth"
	"sol_coffeesys/backend/db"
	"sol_coffeesys/backend/handler"
	"sol_coffeesys/backend/handler/testutil"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type simpleFake struct {
	users map[int64]db.User
}

func (f *simpleFake) GetUserForUpdate(ctx context.Context, id int64) (db.User, error) {
	u, ok := f.users[id]
	if !ok {
		return db.User{}, sql.ErrNoRows
	}
	return u, nil
}

func TestCreateCategory_AdminAndNonAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	orig := auth.Validate
	defer func() { auth.Validate = orig }()

	tests := []struct {
		name           string
		method         string
		target         string
		body           interface{}
		setupMock      func(m *testutil.MockDB)
		validate       func(string) (*jwt.Token, error)
		expectedStatus int
	}{
		{
			name:   "POST 管理者権限 OK",
			method: http.MethodPost,
			body:   map[string]string{"name": "A"},
			target: "/api/categories",
			setupMock: func(m *testutil.MockDB) {
				m.On("GetUserForUpdate", mock.Anything, int64(1)).
					Return(db.User{ID: 1, Role: "admin"}, nil)
				m.On("CreateCategory", mock.Anything, mock.Anything).
					Return(db.Category{ID: 1, Name: "A"}, nil)
			},
			validate: func(string) (*jwt.Token, error) {
				return &jwt.Token{Valid: true, Claims: jwt.MapClaims{"user.id": float64(1)}}, nil
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:   "POST 非管理者権限 forbidden",
			method: http.MethodPost,
			target: "/api/categories",
			body:   map[string]string{"name": "A"},
			setupMock: func(m *testutil.MockDB) {
				m.On("GetUserForUpdate", mock.Anything, int64(2)).
					Return(db.User{ID: 2, Role: "member"}, nil)
			},
			validate: func(string) (*jwt.Token, error) {
				return &jwt.Token{Valid: true, Claims: jwt.MapClaims{"user.id": float64(2)}}, nil
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:   "PUT 管理者権限 OK",
			method: http.MethodPut,
			target: "/api/categories/1",
			body:   map[string]string{"name": "Updated"},
			setupMock: func(m *testutil.MockDB) {
				m.On("GetUserForUpdate", mock.Anything, int64(1)).
					Return(db.User{ID: 1, Role: "admin"}, nil)
				m.On("UpdateCategory", mock.Anything, mock.Anything).
					Return(db.Category{ID: 1, Name: "Updated"}, nil)
			},
			validate: func(string) (*jwt.Token, error) {
				return &jwt.Token{Valid: true, Claims: jwt.MapClaims{"user.id": float64(1)}}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "DELETE 管理者権限 OK",
			method: http.MethodDelete,
			target: "/api/categories/1",
			body:   nil,
			setupMock: func(m *testutil.MockDB) {
				m.On("GetUserForUpdate", mock.Anything, int64(1)).
					Return(db.User{ID: 1, Role: "admin"}, nil)
				m.On("DeleteCategory", mock.Anything, int64(1)).Return(nil)
			},
			validate: func(string) (*jwt.Token, error) {
				return &jwt.Token{Valid: true, Claims: jwt.MapClaims{"user.id": float64(1)}}, nil
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "トークン無し Unauthorized",
			method:         http.MethodPost,
			target:         "/api/categories",
			body:           map[string]string{"name": "A"},
			setupMock:      nil,
			validate:       nil,
			expectedStatus: http.StatusUnauthorized,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(testutil.MockDB)
			if tt.validate != nil {
				auth.Validate = tt.validate
			} else {
				auth.Validate = orig
			}

			if tt.setupMock != nil {
				tt.setupMock(mockDB)
			}

			router := gin.New()
			router.POST("/api/categories", auth.AdminOnly(mockDB), handler.CreateCategoryHandler(mockDB))
			router.PUT("/api/categories/:id", auth.AdminOnly(mockDB), handler.UpdateCategoryHandler(mockDB))
			router.DELETE("/api/categories/:id", auth.AdminOnly(mockDB), handler.DeleteCategoryHandler(mockDB))

			var reqBody []byte
			if tt.body != nil {
				reqBody, _ = json.Marshal(tt.body)
			}
			req := httptest.NewRequest(tt.method, tt.target, bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			if tt.validate != nil {
				req.Header.Set("Authorization", "Bearer dummy")
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
			mockDB.AssertExpectations(t)
		})
	}

}
