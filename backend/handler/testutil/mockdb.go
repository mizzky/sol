package testutil

import (
	"context"
	"sol_coffeesys/backend/db"

	"github.com/stretchr/testify/mock"
)

type MockDB struct {
	db.Querier
	mock.Mock
}

func (m *MockDB) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockDB) GetCategory(ctx context.Context, id int64) (db.Category, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.Category), args.Error(1)
}
func (m *MockDB) ListCategories(ctx context.Context) ([]db.Category, error) {
	args := m.Called(ctx)
	return args.Get(0).([]db.Category), args.Error(1)
}

func (m *MockDB) CreateCategory(ctx context.Context, arg db.CreateCategoryParams) (db.Category, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Category), args.Error(1)
}

func (m *MockDB) UpdateCategory(ctx context.Context, arg db.UpdateCategoryParams) (db.Category, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Category), args.Error(1)
}

func (m *MockDB) DeleteCategory(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDB) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockDB) GetUserForUpdate(ctx context.Context, id int64) (db.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockDB) CreateProduct(ctx context.Context, arg db.CreateProductParams) (db.Product, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Product), args.Error(1)
}

func (m *MockDB) GetProduct(ctx context.Context, id int64) (db.Product, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.Product), args.Error(1)
}

func (m *MockDB) ListProducts(ctx context.Context) ([]db.Product, error) {
	args := m.Called(ctx)
	return args.Get(0).([]db.Product), args.Error(1)
}

func (m *MockDB) UpdateProduct(ctx context.Context, arg db.UpdateProductParams) (db.Product, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Product), args.Error(1)
}

func (m *MockDB) DeleteProduct(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDB) GetUserByID(ctx context.Context, id int64) (db.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockDB) UpdateUserRole(ctx context.Context, arg db.UpdateUserRoleParams) (db.User, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockDB) SetResetToken(ctx context.Context, arg db.SetResetTokenParams) (db.User, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockDB) ListCartItemsByUser(ctx context.Context, userID int64) ([]db.ListCartItemsByUserRow, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.ListCartItemsByUserRow), args.Error(1)
}

func (m *MockDB) GetOrCreateCartForUser(ctx context.Context, userID int64) (db.Cart, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(db.Cart), args.Error(1)
}

func (m *MockDB) AddCartItem(ctx context.Context, arg db.AddCartItemParams) (db.CartItem, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.CartItem), args.Error(1)
}

func (m *MockDB) UpdateCartItemQtyByUser(ctx context.Context, arg db.UpdateCartItemQtyByUserParams) (db.CartItem, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.CartItem), args.Error(1)
}

func (m *MockDB) GetCartItemByID(ctx context.Context, id int64) (db.CartItem, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.CartItem), args.Error(1)
}

func (m *MockDB) GetCartByUser(ctx context.Context, userID int64) (db.Cart, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(db.Cart), args.Error(1)
}

func (m *MockDB) RemoveCartItemByUser(ctx context.Context, arg db.RemoveCartItemByUserParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockDB) ClearCart(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockDB) ClearCartByUser(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockDB) GetProductForUpdate(ctx context.Context, id int64) (db.Product, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.Product), args.Error(1)
}

func (m *MockDB) CreateOrder(ctx context.Context, arg db.CreateOrderParams) (db.CreateOrderRow, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.CreateOrderRow), args.Error(1)
}

func (m *MockDB) UpdateProductStock(ctx context.Context, arg db.UpdateProductStockParams) (db.UpdateProductStockRow, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.UpdateProductStockRow), args.Error(1)
}

func (m *MockDB) CreateOrderItem(ctx context.Context, arg db.CreateOrderItemParams) (db.OrderItem, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.OrderItem), args.Error(1)
}

func (m *MockDB) GetOrderByIDForUpdate(ctx context.Context, id int64) (db.GetOrderByIDForUpdateRow, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.GetOrderByIDForUpdateRow), args.Error(1)
}

func (m *MockDB) ListOrderItemsByOrderID(ctx context.Context, orderID int64) ([]db.OrderItem, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.OrderItem), args.Error(1)
}

func (m *MockDB) UpdateOrderStatus(ctx context.Context, arg db.UpdateOrderStatusParams) (db.UpdateOrderStatusRow, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.UpdateOrderStatusRow), args.Error(1)
}

func (m *MockDB) ListOrdersByUser(ctx context.Context, userID int64) ([]db.ListOrdersByUserRow, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.ListOrdersByUserRow), args.Error(1)
}

func (m *MockDB) CreateRefreshToken(ctx context.Context, arg db.CreateRefreshTokenParams) (db.RefreshToken, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return db.RefreshToken{}, args.Error(1)
	}
	return args.Get(0).(db.RefreshToken), args.Error(1)
}

func (m *MockDB) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (db.RefreshToken, error) {
	args := m.Called(ctx, tokenHash)
	if args.Get(0) == nil {
		return db.RefreshToken{}, args.Error(1)
	}
	return args.Get(0).(db.RefreshToken), args.Error(1)
}

func (m *MockDB) RevokeRefreshTokenByHash(ctx context.Context, tokenHash string) error {
	args := m.Called(ctx, tokenHash)
	return args.Error(0)
}
