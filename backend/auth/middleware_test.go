package auth_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sol_coffeesys/backend/auth"
	"sol_coffeesys/backend/db"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

type FakeQuerier struct {
	users map[int64]db.User
}

func (f *FakeQuerier) CreateCategory(ctx context.Context, arg db.CreateCategoryParams) (db.Category, error) {
	return db.Category{}, nil
}

func (f *FakeQuerier) CreateProduct(ctx context.Context, arg db.CreateProductParams) (db.Product, error) {
	return db.Product{}, nil
}
func (f *FakeQuerier) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
	return db.User{}, nil
}
func (f *FakeQuerier) DeleteCategory(ctx context.Context, id int64) error {
	return nil
}
func (f *FakeQuerier) GetCategory(ctx context.Context, id int64) (db.Category, error) {
	return db.Category{}, nil
}
func (f *FakeQuerier) GetProduct(ctx context.Context, id int64) (db.Product, error) {
	return db.Product{}, nil
}
func (f *FakeQuerier) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	return db.User{}, nil
}
func (f *FakeQuerier) GetUserForUpdate(ctx context.Context, id int64) (db.User, error) {
	u, ok := f.users[id]
	if !ok {
		return db.User{}, sql.ErrNoRows
	}
	return u, nil
}
func (f *FakeQuerier) ListCategories(ctx context.Context) ([]db.Category, error) {
	return []db.Category{}, nil
}
func (f *FakeQuerier) ListProducts(ctx context.Context) ([]db.Product, error) {
	return []db.Product{}, nil
}
func (f *FakeQuerier) UpdateCategory(ctx context.Context, arg db.UpdateCategoryParams) (db.Category, error) {
	return db.Category{}, nil
}

func (f *FakeQuerier) DeleteProduct(ctx context.Context, id int64) error {
	return nil
}

func (f *FakeQuerier) UpdateProduct(ctx context.Context, arg db.UpdateProductParams) (db.Product, error) {
	return db.Product{}, nil
}

func (f *FakeQuerier) GetUserByID(ctx context.Context, id int64) (db.User, error) {
	u, ok := f.users[id]
	if !ok {
		return db.User{}, sql.ErrNoRows
	}
	return u, nil
}

func (f *FakeQuerier) UpdateUserRole(ctx context.Context, arg db.UpdateUserRoleParams) (db.User, error) {
	u, ok := f.users[arg.ID]
	if !ok {
		return db.User{}, sql.ErrNoRows
	}
	u.Role = arg.Role
	f.users[arg.ID] = u
	return u, nil
}

func (f *FakeQuerier) SetResetToken(ctx context.Context, arg db.SetResetTokenParams) (db.User, error) {
	u, ok := f.users[arg.ID]
	if !ok {
		return db.User{}, sql.ErrNoRows
	}
	u.ResetToken = arg.ResetToken
	f.users[arg.ID] = u
	return u, nil
}

func (f *FakeQuerier) CreateCart(ctx context.Context, userID int64) (db.Cart, error) {
	return db.Cart{}, nil
}

func (f *FakeQuerier) AddCartItem(ctx context.Context, arg db.AddCartItemParams) (db.CartItem, error) {
	return db.CartItem{}, nil
}

func (f *FakeQuerier) GetCartByUser(ctx context.Context, userID int64) (db.Cart, error) {
	return db.Cart{}, nil
}

func (f *FakeQuerier) GetOrCreateCartForUser(ctx context.Context, userID int64) (db.Cart, error) {
	return db.Cart{}, nil
}

func (f *FakeQuerier) GetCartItemByID(ctx context.Context, id int64) (db.CartItem, error) {
	return db.CartItem{}, nil
}

func (f *FakeQuerier) UpdateCartItemQty(ctx context.Context, arg db.UpdateCartItemQtyParams) (db.CartItem, error) {
	return db.CartItem{}, nil
}

func (f *FakeQuerier) UpdateCartItemQtyByUser(ctx context.Context, arg db.UpdateCartItemQtyByUserParams) (db.CartItem, error) {
	return db.CartItem{}, nil
}

func (f *FakeQuerier) RemoveCartItem(ctx context.Context, id int64) error {
	return nil
}

func (f *FakeQuerier) RemoveCartItemByUser(ctx context.Context, arg db.RemoveCartItemByUserParams) error {
	return nil
}

func (f *FakeQuerier) ClearCart(ctx context.Context, cartID int64) error {
	return nil
}

func (f *FakeQuerier) ClearCartByUser(ctx context.Context, userID int64) error {
	return nil
}

func (f *FakeQuerier) ListCartItems(ctx context.Context, cartID int64) ([]db.ListCartItemsRow, error) {
	return []db.ListCartItemsRow{}, nil
}

func (f *FakeQuerier) ListCartItemsByUser(ctx context.Context, cartID int64) ([]db.ListCartItemsByUserRow, error) {
	return []db.ListCartItemsByUserRow{}, nil
}

// DB接続エラー用のQuerier
type BadQuerier struct{ *FakeQuerier }

func (b *BadQuerier) GetUserForUpdate(ctx context.Context, id int64) (db.User, error) {
	return db.User{}, fmt.Errorf("db error")
}

func (b *BadQuerier) GetUserByID(ctx context.Context, id int64) (db.User, error) {
	return db.User{}, sql.ErrConnDone
}

func (b *BadQuerier) UpdateUserRole(ctx context.Context, arg db.UpdateUserRoleParams) (db.User, error) {
	return db.User{}, sql.ErrConnDone
}
func (b *BadQuerier) SetResetToken(ctx context.Context, arg db.SetResetTokenParams) (db.User, error) {
	return db.User{}, sql.ErrConnDone
}

func (b *BadQuerier) CreateCart(ctx context.Context, userID int64) (db.Cart, error) {
	return db.Cart{}, sql.ErrConnDone
}

func (b *BadQuerier) AddCartItem(ctx context.Context, arg db.AddCartItemParams) (db.CartItem, error) {
	return db.CartItem{}, sql.ErrConnDone
}

func (b *BadQuerier) GetCartItemByID(ctx context.Context, id int64) (db.CartItem, error) {
	return db.CartItem{}, sql.ErrConnDone
}

func (b *BadQuerier) UpdateCartItemQty(ctx context.Context, arg db.UpdateCartItemQtyParams) (db.CartItem, error) {
	return db.CartItem{}, sql.ErrConnDone
}

func (b *BadQuerier) RemoveCartItem(ctx context.Context, id int64) error {
	return sql.ErrConnDone
}

func (b *BadQuerier) ClearCart(ctx context.Context, cartID int64) error {
	return sql.ErrConnDone
}

func (b *BadQuerier) GetOrCreateCartForUser(ctx context.Context, userID int64) (db.Cart, error) {
	return db.Cart{}, sql.ErrConnDone
}

func TestAdminOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)

	users := map[int64]db.User{
		1: {ID: 1, Role: "admin"},
		2: {ID: 2, Role: "member"},
	}
	fq := &FakeQuerier{users: users}

	router := gin.New()
	router.GET("/admin", auth.AdminOnly(fq), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	origValidate := auth.Validate
	t.Cleanup(func() {
		auth.Validate = origValidate
	})

	tests := []struct {
		name           string
		authHeader     string
		validateFunc   func(string) (*jwt.Token, error)
		expectedStatus int
	}{
		{
			name:           "トークン無し->401",
			authHeader:     "",
			validateFunc:   nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "不適切なトークン->401",
			authHeader: "Bearer invalid",
			validateFunc: func(ts string) (*jwt.Token, error) {
				return nil, fmt.Errorf("invalid")
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "権限なし(non-admin)->403",
			authHeader: "Bearer valid-nonadmin",
			validateFunc: func(ts string) (*jwt.Token, error) {
				return &jwt.Token{Valid: true, Claims: jwt.MapClaims{"user.id": float64(2)}}, nil
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:       "権限あり(admin)->204",
			authHeader: "Bearer valid-admin",
			validateFunc: func(ts string) (*jwt.Token, error) {
				return &jwt.Token{Valid: true, Claims: jwt.MapClaims{"user.id": float64(1)}}, nil
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:       "DB該当ユーザー未検出->401",
			authHeader: "Bearer valid-missing-user",
			validateFunc: func(ts string) (*jwt.Token, error) {
				return &jwt.Token{Valid: true, Claims: jwt.MapClaims{"user.id": float64(3)}}, nil
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "クレーム無し->401",
			authHeader: "Bearer non-claim",
			validateFunc: func(ts string) (*jwt.Token, error) {
				return &jwt.Token{Valid: true, Claims: jwt.MapClaims{}}, nil
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "クレーム型一致(文字列数値)->204",
			authHeader: "Bearer valid-admin",
			validateFunc: func(ts string) (*jwt.Token, error) {
				return &jwt.Token{Valid: true, Claims: jwt.MapClaims{"user.id": "1"}}, nil
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:       "クレーム型不一致->401",
			authHeader: "Bearer id-as-string",
			validateFunc: func(ts string) (*jwt.Token, error) {
				return &jwt.Token{Valid: true, Claims: jwt.MapClaims{"user.id": "non-a-number"}}, nil
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "Validateがnullを返す->401",
			authHeader: "Bearer validate-null",
			validateFunc: func(ts string) (*jwt.Token, error) {
				return nil, nil
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "token.Valid == false ->401",
			authHeader: "Bearer valid-false",
			validateFunc: func(ts string) (*jwt.Token, error) {
				return &jwt.Token{Valid: false, Claims: jwt.MapClaims{"user.id": float64(1)}}, nil
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "token.ClaimがJwt.MapClaimでない(&jwt.RegisteredClaims{})->401",
			authHeader: "Bearer invalid-claim",
			validateFunc: func(ts string) (*jwt.Token, error) {
				return &jwt.Token{Valid: true, Claims: &jwt.RegisteredClaims{}}, nil
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "DB接続エラー",
			authHeader: "Bearer DB-connect-err",
			validateFunc: func(ts string) (*jwt.Token, error) {
				return &jwt.Token{Valid: true, Claims: jwt.MapClaims{"user.id": float64(1)}}, nil
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "ヘッダーフォーマット不正->401",
			authHeader:     "invalidheader xyz",
			validateFunc:   nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "ヘッダーフォーマット欠損->401",
			authHeader:     "Bearer",
			validateFunc:   nil,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.validateFunc != nil {
				auth.Validate = tt.validateFunc
			} else {
				auth.Validate = origValidate
			}

			localRouter := gin.New()
			if tt.name == "DB接続エラー" {
				badQ := &BadQuerier{FakeQuerier: &FakeQuerier{users: map[int64]db.User{}}}
				localRouter.GET("/admin", auth.AdminOnly(badQ), func(c *gin.Context) {
					c.Status(http.StatusNoContent)
				})
			} else {
				localRouter.GET("/admin", auth.AdminOnly(fq), func(c *gin.Context) {
					c.Status(http.StatusNoContent)
				})
			}

			req := httptest.NewRequest(http.MethodGet, "/admin", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()
			localRouter.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
			if w.Code == http.StatusNoContent {
				assert.Equal(t, 0, w.Body.Len())
			}
		})
	}
}
