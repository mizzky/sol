package handler_test

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

func (m *MockDB) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.User), args.Error(1)
}
