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

func (m *MockDB) CreateCategoryHandler(ctx context.Context, arg db.CreateCategoryParams) (db.Category, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Category), args.Error(1)
}

func (m *MockDB) CreateUserHandler(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.User), args.Error(1)
}
