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
